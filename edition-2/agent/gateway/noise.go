// Package gateway provides the optional authenticated message client. It has no
// dependency on Ensemble's actors, tools, GUI, or internal runtime packages.
package gateway

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/flynn/noise"
)

// Keypair is an X25519 identity. Never log or send its private half.
// This key agrees on secrets; it is not an Ed25519 signing key despite some
// historical configurations using that label. Only public halves belong in
// the gateway's agent registry or another peer's configuration.
type Keypair struct {
	// Private is the 32-byte secret retained by its owning application.
	Private []byte
	// Public is the corresponding 32-byte peer configuration value.
	Public []byte
}

// GenerateKeypair obtains a fresh identity from the operating system RNG.
func GenerateKeypair() (Keypair, error) {
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return Keypair{}, err
	}
	return Keypair{Private: k.Bytes(), Public: k.PublicKey().Bytes()}, nil
}

// EncodePublicKey labels an X25519 public key, avoiding the misleading ed25519 label.
func EncodePublicKey(key []byte) string { return "x25519:" + base64.StdEncoding.EncodeToString(key) }

// ParsePublicKey accepts historical labels but requires exactly 32 decoded bytes.
func ParsePublicKey(text string) ([]byte, error) {
	if _, suffix, ok := strings.Cut(strings.TrimSpace(text), ":"); ok {
		text = suffix
	}
	key, err := base64.StdEncoding.DecodeString(text)
	if err != nil || len(key) != 32 {
		return nil, errors.New("public key must encode 32 X25519 bytes")
	}
	return key, nil
}

// LoadOrCreateKeypair stores only the private half with owner-only permissions.
// Loading derives the public half anew; malformed existing files are never replaced.
func LoadOrCreateKeypair(path string) (Keypair, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key, err := GenerateKeypair()
		if err != nil {
			return Keypair{}, err
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return Keypair{}, err
		}
		// Exclusive creation resolves two first-run processes without replacing
		// an identity another process just published. Losing that race means
		// loading the winner, not continuing with our now-unregistered key.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			return LoadOrCreateKeypair(path)
		}
		if err != nil {
			return Keypair{}, err
		}
		_, writeErr := f.WriteString(base64.StdEncoding.EncodeToString(key.Private) + "\n")
		return key, errors.Join(writeErr, f.Close())
	}
	if err != nil {
		return Keypair{}, err
	}
	secret, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return Keypair{}, errors.New("invalid private key encoding")
	}
	k, err := ecdh.X25519().NewPrivateKey(secret)
	if err != nil {
		return Keypair{}, err
	}
	return Keypair{Private: k.Bytes(), Public: k.PublicKey().Bytes()}, nil
}

// NoiseKK wraps the public Noise implementation, not a custom cryptographic scheme.
// Handshake calls are sequential. After completion one reader and one writer may
// operate concurrently; each direction has its own cipher state and nonce.
type NoiseKK struct {
	handshake     *noise.HandshakeState
	initiator     bool
	send, receive *noise.CipherState
}

// NewNoiseKK binds the known peer key and application prologue to a KK handshake.
// KK assumes both static public keys were distributed beforehand. The initial
// ephemeral public value is visible; secrecy comes from the library's derived
// keys, not from hiding every handshake byte. The prologue binds the intended
// application identity to that authenticated transcript.
func NewNoiseKK(role string, key Keypair, peer, prologue []byte) (*NoiseKK, error) {
	if role != "initiator" && role != "responder" {
		return nil, errors.New("role must be initiator or responder")
	}
	if len(key.Private) != 32 || len(key.Public) != 32 || len(peer) != 32 {
		return nil, errors.New("Noise KK requires 32-byte keys")
	}
	h, err := noise.NewHandshakeState(noise.Config{CipherSuite: noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256), Pattern: noise.HandshakeKK, Initiator: role == "initiator", StaticKeypair: noise.DHKey{Private: key.Private, Public: key.Public}, PeerStatic: peer, Prologue: prologue})
	return &NoiseKK{handshake: h, initiator: role == "initiator"}, err
}

// Split returns initiator-to-responder first, irrespective of which endpoint
// called the library. Reversing these for the responder is what makes each
// endpoint's outgoing nonce match only its peer's incoming nonce.
func (n *NoiseKK) finish(a, b *noise.CipherState) {
	if a == nil {
		return
	}
	if n.initiator {
		n.send, n.receive = a, b
	} else {
		n.send, n.receive = b, a
	}
}

// WriteHandshakeMessage advances the handshake and returns a wire message.
func (n *NoiseKK) WriteHandshakeMessage(payload []byte) ([]byte, error) {
	data, a, b, err := n.handshake.WriteMessage(nil, payload)
	if err == nil {
		n.finish(a, b)
	}
	return data, err
}

// ReadHandshakeMessage authenticates a peer handshake message before returning payload.
func (n *NoiseKK) ReadHandshakeMessage(message []byte) ([]byte, error) {
	data, a, b, err := n.handshake.ReadMessage(nil, message)
	if err == nil {
		n.finish(a, b)
	}
	return data, err
}

// IsHandshakeComplete reports whether both transport ciphers are established.
func (n *NoiseKK) IsHandshakeComplete() bool { return n.send != nil && n.receive != nil }

// Encrypt advances the outgoing nonce; callers serialize outgoing frames.
func (n *NoiseKK) Encrypt(data, associated []byte) ([]byte, error) {
	if n.send == nil {
		return nil, errors.New("handshake incomplete")
	}
	return n.send.Encrypt(nil, associated, data)
}

// Decrypt authenticates the next incoming frame; failures terminate the connection.
// Continuing after an authentication failure would make frame order ambiguous.
// The caller closes the socket instead of guessing whether to advance a nonce.
func (n *NoiseKK) Decrypt(data, associated []byte) ([]byte, error) {
	if n.receive == nil {
		return nil, errors.New("handshake incomplete")
	}
	return n.receive.Decrypt(nil, associated, data)
}
