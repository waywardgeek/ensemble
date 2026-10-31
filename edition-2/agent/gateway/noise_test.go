package gateway

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestNoiseKK exercises the public cryptographic implementation in both
// directions, including wrong static keys and altered transport ciphertext.
func TestNoiseKK(t *testing.T) {
	a, _ := GenerateKeypair()
	b, _ := GenerateKeypair()
	wrong, _ := GenerateKeypair()
	for _, bad := range []bool{false, true} {
		peer := b.Public
		if bad {
			peer = wrong.Public
		}
		initiator, err := NewNoiseKK("initiator", a, peer, []byte("test"))
		if err != nil {
			t.Fatal(err)
		}
		responder, _ := NewNoiseKK("responder", b, a.Public, []byte("test"))
		first, err := initiator.WriteHandshakeMessage([]byte("hello"))
		if err != nil {
			t.Fatal(err)
		}
		got, err := responder.ReadHandshakeMessage(first)
		if bad {
			if err == nil {
				t.Fatal("wrong static key authenticated")
			}
			continue
		}
		if err != nil || string(got) != "hello" {
			t.Fatalf("first payload %q: %v", got, err)
		}
		second, err := responder.WriteHandshakeMessage([]byte("welcome"))
		if err != nil {
			t.Fatal(err)
		}
		got, err = initiator.ReadHandshakeMessage(second)
		if err != nil || string(got) != "welcome" {
			t.Fatalf("second payload %q: %v", got, err)
		}
		if !initiator.IsHandshakeComplete() || !responder.IsHandshakeComplete() {
			t.Fatal("handshake incomplete")
		}
		for _, pair := range [][2]*NoiseKK{{initiator, responder}, {responder, initiator}} {
			cipher, err := pair[0].Encrypt([]byte("private transport"), []byte("binding"))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(cipher, []byte("private transport")) {
				t.Fatal("transport plaintext exposed")
			}
			plain, err := pair[1].Decrypt(cipher, []byte("binding"))
			if err != nil || string(plain) != "private transport" {
				t.Fatalf("transport: %q %v", plain, err)
			}
			cipher, _ = pair[0].Encrypt([]byte("tampered"), nil)
			cipher[len(cipher)-1] ^= 1
			if _, err = pair[1].Decrypt(cipher, nil); err == nil {
				t.Fatal("tampered ciphertext accepted")
			}
		}
	}
}

// TestIdentityFiles verifies stable private identity, owner-only creation, and
// refusal to silently replace an existing malformed file.
func TestIdentityFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "identity.key")
	first, err := LoadOrCreateKeypair(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateKeypair(path)
	if err != nil || !bytes.Equal(first.Public, second.Public) {
		t.Fatal("identity changed", err)
	}
	parsed, err := ParsePublicKey(EncodePublicKey(first.Public))
	if err != nil || !bytes.Equal(parsed, first.Public) {
		t.Fatal("public encoding", err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatal("private key permissions", info.Mode())
	}
	os.WriteFile(path, []byte("broken identity"), 0600)
	if _, err = LoadOrCreateKeypair(path); err == nil {
		t.Fatal("malformed identity replaced")
	}
}
