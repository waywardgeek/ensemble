package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ensemble/internal/common"
)

const bootstrapClient = "dynamic_agent_client"
const inferenceScope = "chatgpt.tokens.use.direct"
const resource = "https://api.openai.com/v1"
const loginScopes = "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"

func randomURL(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func hostID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
func allowedEndpoint(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.User == nil && u.Fragment == "" && u.Host != "" && (u.Scheme == "https" || (u.Scheme == "http" && u.Hostname() == "127.0.0.1"))
}
func (s *Service) endpoints(ctx context.Context) (Options, error) {
	o := s.options
	if !allowedEndpoint(o.Issuer) {
		return o, fmt.Errorf("OAuth issuer must use HTTPS or literal loopback")
	}
	if o.AuthorizationURL == "" || o.TokenURL == "" || o.JWKSURL == "" {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, o.Issuer+"/.well-known/openid-configuration", nil)
		var v struct {
			Issuer        string `json:"issuer"`
			Authorization string `json:"authorization_endpoint"`
			Token         string `json:"token_endpoint"`
			JWKS          string `json:"jwks_uri"`
			Revocation    string `json:"revocation_endpoint"`
		}
		if err := s.doJSON(req, &v); err != nil {
			return o, fmt.Errorf("OAuth discovery unavailable")
		}
		if v.Issuer != o.Issuer {
			return o, fmt.Errorf("OAuth discovery issuer mismatch")
		}
		if o.AuthorizationURL == "" {
			o.AuthorizationURL = v.Authorization
		}
		if o.TokenURL == "" {
			o.TokenURL = v.Token
		}
		if o.JWKSURL == "" {
			o.JWKSURL = v.JWKS
		}
		if o.RevocationURL == "" {
			o.RevocationURL = v.Revocation
		}
	}
	for _, u := range []string{o.AuthorizationURL, o.TokenURL, o.JWKSURL} {
		if !allowedEndpoint(u) {
			return o, fmt.Errorf("unsafe OAuth endpoint")
		}
	}
	if o.RevocationURL != "" && !allowedEndpoint(o.RevocationURL) {
		return o, fmt.Errorf("unsafe OAuth revocation endpoint")
	}
	return o, nil
}
func (s *Service) Login(ctx context.Context, open func(string) error) (common.AuthStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	var out common.AuthStatus
	err := s.locked(ctx, func(st *store) error {
		o, err := s.endpoints(ctx)
		if err != nil {
			return err
		}
		if st.Client == nil {
			st.Client = &clientIdentity{}
		}
		if st.Client.Issuer != "" && st.Client.Issuer != o.Issuer {
			return fmt.Errorf("profile issuer differs from configured issuer")
		}
		if st.Client.HostID == "" {
			id, err := hostID()
			if err != nil {
				return err
			}
			st.Client.HostID = id
			st.Client.Issuer = o.Issuer
			if err = s.save(st); err != nil {
				return err
			}
		}
		redirect := o.RedirectURI
		if redirect == "" {
			redirect = "http://127.0.0.1:0/auth/callback"
		}
		u, err := url.Parse(redirect)
		if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Path != "/auth/callback" || u.RawPath != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
			return fmt.Errorf("OAuth callback requires literal 127.0.0.1 and /auth/callback")
		}
		listener, err := net.Listen("tcp4", u.Host)
		if err != nil {
			return fmt.Errorf("OAuth loopback listener unavailable")
		}
		defer listener.Close()
		u.Host = listener.Addr().String()
		redirect = u.String()
		verifier, err := randomURL(32)
		if err != nil {
			return err
		}
		state, err := randomURL(32)
		if err != nil {
			return err
		}
		nonce, err := randomURL(32)
		if err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(verifier))
		type callback struct {
			code, client string
			err          error
		}
		done := make(chan callback, 1)
		server := &http.Server{ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, MaxHeaderBytes: 8192}
		server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			q, queryErr := url.ParseQuery(r.URL.RawQuery)
			valid := queryErr == nil && r.Method == http.MethodGet && r.Host == u.Host && r.URL.Path == "/auth/callback" && r.URL.RawPath == "" && len(q["state"]) == 1 && subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(state)) == 1
			for _, key := range []string{"code", "error", "iss", "client_id"} {
				if len(q[key]) > 1 {
					valid = false
				}
			}
			if q.Get("iss") != "" && q.Get("iss") != o.Issuer {
				valid = false
			}
			id := q.Get("client_id")
			if st.Client.ClientID == "" {
				if id == "" || id == bootstrapClient {
					valid = false
				}
			} else if id != "" && id != st.Client.ClientID {
				valid = false
			}
			if !valid {
				http.Error(w, "Invalid authorization callback", 400)
				return
			}
			answer := callback{code: q.Get("code"), client: id}
			if q.Get("error") != "" || answer.code == "" {
				answer.err = fmt.Errorf("authorization denied or code missing")
			}
			select {
			case done <- answer:
				fmt.Fprintln(w, "Authorization received. Return to Ensemble.")
			default:
				http.Error(w, "Authorization already received", 409)
			}
		})
		go func() { _ = server.Serve(listener) }()
		defer server.Close()
		id := st.Client.ClientID
		if id == "" {
			id = bootstrapClient
		}
		authURL, err := url.Parse(o.AuthorizationURL)
		if err != nil {
			return fmt.Errorf("invalid OAuth authorization endpoint")
		}
		q := authURL.Query()
		for k, v := range map[string]string{"response_type": "code", "client_id": id, "redirect_uri": redirect, "scope": loginScopes, "resource": resource, "state": state, "nonce": nonce, "code_challenge": base64.RawURLEncoding.EncodeToString(sum[:]), "code_challenge_method": "S256", "ext_agent_host_id": st.Client.HostID} {
			q.Set(k, v)
		}
		if id == bootstrapClient {
			q.Set("agent_name_hint", "Ensemble")
		}
		authURL.RawQuery = q.Encode()
		if open == nil {
			return fmt.Errorf("OAuth login needs a browser presenter")
		}
		if err := open(authURL.String()); err != nil {
			return fmt.Errorf("could not open OAuth authorization page")
		}
		var cb callback
		select {
		case cb = <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
		_ = server.Close()
		if cb.err != nil {
			return cb.err
		}
		if st.Client.ClientID == "" {
			st.Client.ClientID = cb.client
			if err := s.save(st); err != nil {
				return err
			}
		}
		token, err := s.tokens(ctx, url.Values{"grant_type": {"authorization_code"}, "client_id": {st.Client.ClientID}, "code": {cb.code}, "redirect_uri": {redirect}, "code_verifier": {verifier}, "resource": {resource}})
		if err != nil {
			return fmt.Errorf("OAuth code exchange failed")
		}
		identity, err := s.verify(ctx, token.ID, st.Client.ClientID, nonce)
		if err != nil {
			return err
		}
		if token.Refresh == "" {
			return fmt.Errorf("OAuth exchange omitted refresh token")
		}
		key := profileKey(st.Client, identity.Subject)
		g := grant{Subject: identity.Subject, Account: identity.Subject, Scopes: strings.Fields(token.Scope), Access: token.Access, Refresh: token.Refresh, Issuer: o.Issuer, ClientID: st.Client.ClientID, Eligible: hasInference(strings.Fields(token.Scope)), Expires: time.Now().Add(time.Duration(token.Expires) * time.Second), Generation: st.Accounts[key].Generation + 1}
		st.Accounts[key] = g
		st.Selected = key
		st.Mode = common.AuthChatGPT
		out = status(st, "")
		return nil
	})
	return out, err
}
