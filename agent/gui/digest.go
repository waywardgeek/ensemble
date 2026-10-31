package gui

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const digestNonceLifetime = 5 * time.Minute
const digestReplayLimit = 4096
const digestFailureLimit = 1024

type digestFailure struct {
	count   int
	until   time.Time
	updated time.Time
}

type digestAuthentication struct {
	parent     *Server
	credential digestCredentials
	key        [32]byte
	mu         sync.Mutex
	replays    map[string]time.Time
	failures   map[string]digestFailure
}

// ConfigureDigestAuth loads private GUI credentials before serving any requests.
// It protects assets and WebSocket upgrades without changing the core Agent.
func (s *Server) ConfigureDigestAuth(path string) error {
	credential, err := loadDigestCredentials(path)
	if err != nil {
		return err
	}
	guard := &digestAuthentication{parent: s, credential: credential,
		replays: make(map[string]time.Time), failures: make(map[string]digestFailure)}
	if _, err := rand.Read(guard.key[:]); err != nil {
		return fmt.Errorf("initialize GUI authentication: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serving || s.closed {
		return fmt.Errorf("configure GUI authentication before serving requests")
	}
	s.digest = guard
	return nil
}

func (a *digestAuthentication) nonce(now time.Time) (string, error) {
	var data [56]byte
	binary.BigEndian.PutUint64(data[:8], uint64(now.Unix()))
	if _, err := rand.Read(data[8:24]); err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, a.key[:])
	mac.Write(data[:24])
	copy(data[24:], mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}

func (a *digestAuthentication) nonceTime(value string, now time.Time) (time.Time, bool) {
	data, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(data) != 56 {
		return time.Time{}, false
	}
	mac := hmac.New(sha256.New, a.key[:])
	mac.Write(data[:24])
	if !hmac.Equal(data[24:], mac.Sum(nil)) {
		return time.Time{}, false
	}
	issued := time.Unix(int64(binary.BigEndian.Uint64(data[:8])), 0)
	if issued.After(now.Add(time.Second)) {
		return time.Time{}, false
	}
	return issued.Add(digestNonceLifetime), true
}

func (a *digestAuthentication) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		source, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			source = r.RemoteAddr
		}
		a.mu.Lock()
		failure := a.failures[source]
		a.mu.Unlock()
		if now.Before(failure.until) {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(time.Until(failure.until).Seconds())+1)))
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Too many authentication attempts", http.StatusTooManyRequests)
			return
		}
		headers := r.Header.Values("Authorization")
		if len(headers) == 0 {
			a.challenge(w, now, false)
			return
		}
		if len(headers) != 1 {
			a.failed(source, now)
			a.challenge(w, now, false)
			return
		}
		fields, valid := digestFields(headers[0])
		if !valid || !a.validResponse(r, fields) {
			a.failed(source, now)
			a.challenge(w, now, false)
			return
		}
		expires, valid := a.nonceTime(fields["nonce"], now)
		if !valid {
			a.failed(source, now)
			a.challenge(w, now, false)
			return
		}
		if !now.Before(expires) {
			a.challenge(w, now, true)
			return
		}
		count, _ := strconv.ParseUint(fields["nc"], 16, 32)
		key := fields["nonce"] + ":" + fields["cnonce"] + ":" + strconv.FormatUint(count, 16)
		a.mu.Lock()
		for previous, expiry := range a.replays {
			if !now.Before(expiry) {
				delete(a.replays, previous)
			}
		}
		_, replay := a.replays[key]
		full := len(a.replays) >= digestReplayLimit
		if !replay && !full {
			a.replays[key] = expires
			delete(a.failures, source)
		}
		a.mu.Unlock()
		if replay {
			a.failed(source, now)
			a.challenge(w, now, false)
			return
		}
		if full {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Authentication temporarily busy", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *digestAuthentication) validResponse(r *http.Request, fields map[string]string) bool {
	count, err := strconv.ParseUint(fields["nc"], 16, 32)
	_, extendedUsername := fields["username*"]
	if extendedUsername || err != nil || count == 0 || len(fields["nc"]) != 8 || fields["qop"] != "auth" ||
		!strings.EqualFold(fields["algorithm"], "SHA-256") || fields["realm"] != a.credential.Realm ||
		fields["uri"] != r.URL.RequestURI() || fields["username"] != a.credential.Username ||
		fields["cnonce"] == "" || len(fields["cnonce"]) > 256 || fields["nonce"] == "" ||
		(fields["userhash"] != "" && fields["userhash"] != "false") {
		return false
	}
	response, err := hex.DecodeString(fields["response"])
	if err != nil || len(response) != sha256.Size {
		return false
	}
	requestHash := sha256.Sum256([]byte(r.Method + ":" + fields["uri"]))
	expected := sha256.Sum256([]byte(a.credential.HA1 + ":" + fields["nonce"] + ":" +
		fields["nc"] + ":" + fields["cnonce"] + ":auth:" + hex.EncodeToString(requestHash[:])))
	return subtle.ConstantTimeCompare(expected[:], response) == 1
}

func (a *digestAuthentication) challenge(w http.ResponseWriter, now time.Time, stale bool) {
	nonce, err := a.nonce(now)
	if err != nil {
		a.parent.root.Logf("GUI authentication could not create a challenge")
		http.Error(w, "Authentication unavailable", http.StatusServiceUnavailable)
		return
	}
	value := "Digest realm=" + strconv.Quote(a.credential.Realm) + ", algorithm=SHA-256, qop=\"auth\", charset=UTF-8, nonce=" + strconv.Quote(nonce)
	if stale {
		value += ", stale=true"
	}
	w.Header().Set("WWW-Authenticate", value)
	w.Header().Set("Cache-Control", "no-store")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (a *digestAuthentication) failed(source string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for address, failure := range a.failures {
		if now.Sub(failure.updated) >= 10*time.Minute {
			delete(a.failures, address)
		}
	}
	failure := a.failures[source]
	if failure.count == 0 && len(a.failures) >= digestFailureLimit {
		oldest := source
		var earliest time.Time
		for address, previous := range a.failures {
			if earliest.IsZero() || previous.updated.Before(earliest) {
				oldest, earliest = address, previous.updated
			}
		}
		delete(a.failures, oldest)
	}
	failure.count = min(failure.count+1, 8)
	failure.updated = now
	if failure.count >= 3 {
		failure.until = now.Add(time.Second * time.Duration(1<<min(failure.count-3, 5)))
	}
	a.failures[source] = failure
}

func digestFields(header string) (map[string]string, bool) {
	if len(header) > 8192 {
		return nil, false
	}
	scheme, rest, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, "Digest") {
		return nil, false
	}
	fields := make(map[string]string)
	for rest = strings.TrimSpace(rest); rest != ""; rest = strings.TrimSpace(rest) {
		name, tail := digestToken(rest)
		name = strings.ToLower(name)
		if name == "" {
			return nil, false
		}
		rest = strings.TrimSpace(tail)
		if !strings.HasPrefix(rest, "=") {
			return nil, false
		}
		rest = strings.TrimSpace(rest[1:])
		var value string
		if strings.HasPrefix(rest, "\"") {
			rest = rest[1:]
			var decoded strings.Builder
			closed := false
			for len(rest) > 0 {
				character := rest[0]
				rest = rest[1:]
				if character == '"' {
					closed = true
					break
				}
				if character == '\\' {
					if rest == "" {
						return nil, false
					}
					character, rest = rest[0], rest[1:]
				}
				if character < 32 || character == 127 {
					return nil, false
				}
				decoded.WriteByte(character)
			}
			if !closed {
				return nil, false
			}
			value = decoded.String()
		} else {
			value, rest = digestToken(rest)
			if value == "" {
				return nil, false
			}
		}
		if _, duplicate := fields[name]; duplicate {
			return nil, false
		}
		fields[name] = value
		rest = strings.TrimSpace(rest)
		if rest != "" {
			if rest[0] != ',' {
				return nil, false
			}
			rest = rest[1:]
			if strings.TrimSpace(rest) == "" {
				return nil, false
			}
		}
	}
	return fields, len(fields) > 0
}

func digestToken(text string) (string, string) {
	index := 0
	for index < len(text) {
		character := text[index]
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(character))) {
			break
		}
		index++
	}
	return text[:index], text[index:]
}
