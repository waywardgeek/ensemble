package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"ensemble/internal/common"
)

type discovery struct {
	// Issuer must exactly match the configured trusted authority.
	Issuer string `json:"issuer"`
	// Authorization is the browser destination, not a token exchange endpoint.
	Authorization string `json:"authorization_endpoint"`
	// Token accepts form-encoded code and refresh grants.
	Token string `json:"token_endpoint"`
	// JWKS publishes verification keys for signed identity tokens.
	JWKS string `json:"jwks_uri"`
	// Revocation ends a renewable session without deleting its registration.
	Revocation string `json:"revocation_endpoint"`
}

func (s *store) discover(ctx context.Context) (out discovery, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(s.options.Issuer, "/")+"/.well-known/openid-configuration", nil)
	if err != nil {
		return out, errors.New("invalid OAuth issuer")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return out, errors.New("OAuth discovery unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&out) != nil {
		return out, errors.New("invalid OAuth discovery")
	}
	if out.Issuer != s.options.Issuer {
		return out, errors.New("OAuth discovery issuer mismatch")
	}
	for _, endpoint := range []string{out.Authorization, out.Token, out.JWKS, out.Revocation} {
		u, err := url.Parse(endpoint)
		if err != nil || u.Host == "" || u.User != nil {
			return out, errors.New("invalid OAuth endpoint")
		}
		if u.Scheme != "https" {
			issuer, _ := url.Parse(s.options.Issuer)
			if u.Scheme != "http" || u.Host != issuer.Host || !net.ParseIP(u.Hostname()).IsLoopback() {
				return out, errors.New("OAuth endpoints require HTTPS")
			}
		}
	}
	return out, nil
}

// Login keeps an unverified attempt separate from the active session. The
// browser callback carries a code, never tokens; token exchange stays server-side.
func (s *store) Login(ctx context.Context, launch func(string) error) (common.AuthStatus, error) {
	var p profile
	var host string
	err := s.locked(ctx, func(value *saved) error {
		if value.HostID == "" {
			var err error
			value.HostID, err = hostID()
			if err != nil {
				return err
			}
		}
		host = value.HostID
		p = value.Profiles[value.Active]
		if p.ClientID == "" && value.Pending != "" {
			p.ClientID = value.Pending
		}
		return s.save(value)
	})
	if err != nil {
		return common.AuthStatus{}, err
	}
	config, err := s.discover(ctx)
	if err != nil {
		return common.AuthStatus{}, err
	}
	entropy, err := random(96)
	if err != nil {
		return common.AuthStatus{}, errors.New("cannot generate OAuth challenge")
	}
	encode := base64.RawURLEncoding.EncodeToString
	verifier, state, nonce := encode(entropy[:32]), encode(entropy[32:64]), encode(entropy[64:])
	challenge := sha256.Sum256([]byte(verifier))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return common.AuthStatus{}, errors.New("cannot start OAuth callback")
	}
	redirect := "http://" + listener.Addr().String() + "/auth/callback"
	callback := make(chan url.Values, 1)
	var once sync.Once
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Expected redirect", 405)
			return
		}
		values := r.URL.Query()
		once.Do(func() { callback <- values })
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, "Return to Ensemble to see the sign-in result.\n")
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	defer server.Close()
	go server.Serve(listener)
	clientID := p.ClientID
	if clientID == "" {
		clientID = "dynamic_agent_client"
	}
	params := url.Values{"client_id": {clientID}, "ext_agent_host_id": {host}, "response_type": {"code"}, "redirect_uri": {redirect}, "scope": {"openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"}, "resource": {"https://api.openai.com/v1"}, "state": {state}, "nonce": {nonce}, "code_challenge_method": {"S256"}, "code_challenge": {encode(challenge[:])}}
	if clientID == "dynamic_agent_client" {
		params.Set("agent_name_hint", "Ensemble")
	}
	// Omit optional ID-token hints: --print-url must never print a retained token.
	if launch == nil {
		return common.AuthStatus{}, errors.New("OAuth requires a browser launcher")
	}
	if err := launch(config.Authorization + "?" + params.Encode()); err != nil {
		return common.AuthStatus{}, errors.New("cannot open sign-in browser")
	}
	var values url.Values
	select {
	case values = <-callback:
	case <-ctx.Done():
		return common.AuthStatus{}, ctx.Err()
	}
	server.Close()
	if values.Get("state") != state {
		return common.AuthStatus{}, errors.New("OAuth callback state mismatch")
	}
	if values.Get("error") != "" {
		return common.AuthStatus{}, errors.New("OAuth authorization declined or failed")
	}
	issued := values.Get("client_id")
	if clientID == "dynamic_agent_client" {
		if issued == "" || issued == clientID {
			return common.AuthStatus{}, errors.New("OAuth registration did not issue a client ID")
		}
		clientID = issued
	} else if issued != "" && issued != clientID {
		return common.AuthStatus{}, errors.New("OAuth callback client ID mismatch")
	}
	if values.Get("code") == "" {
		return common.AuthStatus{}, errors.New("OAuth callback missing code")
	}
	// Retain registration before exchange: a failed code needs a new attempt,
	// not a second registration. Do not promote it to the active account yet.
	if err := s.locked(ctx, func(value *saved) error { value.Pending = clientID; return s.save(value) }); err != nil {
		return common.AuthStatus{}, err
	}
	token, err := s.exchange(ctx, config.Token, url.Values{"grant_type": {"authorization_code"}, "client_id": {clientID}, "code": {values.Get("code")}, "code_verifier": {verifier}, "redirect_uri": {redirect}, "resource": {"https://api.openai.com/v1"}})
	if err != nil {
		return common.AuthStatus{}, err
	}
	claims, err := s.identity(ctx, config, token.IDToken, clientID, nonce)
	if err != nil {
		return common.AuthStatus{}, err
	}
	if p.Subject != "" && claims.Subject != p.Subject {
		return common.AuthStatus{}, errors.New("OAuth returning account identity mismatch")
	}
	p.ClientID, p.Subject, p.Email = clientID, claims.Subject, claims.Email
	replace(&p, token)
	var out common.AuthStatus
	err = s.locked(ctx, func(value *saved) error {
		value.Profiles[clientID] = p
		value.Active = clientID
		value.Pending = ""
		value.Billing = common.CredentialChatGPT
		if err := s.save(value); err != nil {
			return err
		}
		out = status(value)
		return nil
	})
	return out, err
}

type tokenResponse struct {
	// AccessToken is the transport bearer, not the OIDC identity token.
	AccessToken string `json:"access_token"`
	// RefreshToken replaces the prior refresh token in the same atomic save.
	RefreshToken string `json:"refresh_token"`
	// IDToken must be verified before trusting any identity claim.
	IDToken string `json:"id_token"`
	// ExpiresIn supplies the actual lifetime in seconds.
	ExpiresIn int `json:"expires_in"`
	// Scope may be absent on renewal, which retains the existing grant.
	Scope *string `json:"scope"`
	// TokenType must identify bearer authentication when present.
	TokenType string `json:"token_type"`
}
type tokenError string

func (e tokenError) Error() string { return "OAuth token exchange rejected: " + string(e) }
func terminal(code string) bool {
	switch code {
	case "invalid_grant", "invalid_refresh_token", "token_expired", "refresh_token_expired", "refresh_token_invalidated", "refresh_token_reused":
		return true
	}
	return false
}

// Bounded exponential retries preserve credentials on transport and 5xx failure.
// Errors never echo endpoint bodies, which may contain codes or tokens.
func (s *store) exchange(ctx context.Context, endpoint string, values url.Values) (tokenResponse, error) {
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return tokenResponse{}, ctx.Err()
			case <-time.After(time.Duration(1<<attempt) * 100 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
		if err != nil {
			return tokenResponse{}, errors.New("invalid OAuth token endpoint")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil || resp.StatusCode >= 500 {
			continue
		}
		if resp.StatusCode != 200 {
			var failure struct {
				// Error is an OAuth error identifier, never arbitrary display prose.
				Error string `json:"error"`
			}
			json.Unmarshal(data, &failure)
			if terminal(failure.Error) {
				return tokenResponse{}, tokenError(failure.Error)
			}
			return tokenResponse{}, errors.New("OAuth token exchange rejected")
		}
		var out tokenResponse
		if json.Unmarshal(data, &out) != nil || out.AccessToken == "" || out.RefreshToken == "" || out.ExpiresIn <= 0 || (out.TokenType != "" && !strings.EqualFold(out.TokenType, "Bearer")) {
			return out, errors.New("invalid OAuth token response")
		}
		return out, nil
	}
	return tokenResponse{}, errors.New("OAuth token service unavailable; credentials retained")
}
func replace(p *profile, t tokenResponse) {
	p.AccessToken = t.AccessToken
	p.RefreshToken = t.RefreshToken
	p.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
	if t.IDToken != "" {
		p.IDToken = t.IDToken
	}
	if t.Scope != nil {
		p.Scopes = strings.Fields(*t.Scope)
	}
}
func clear(p *profile) {
	p.AccessToken = ""
	p.RefreshToken = ""
	p.IDToken = ""
	p.ExpiresAt = time.Time{}
	p.Scopes = nil
}

// GetBearerToken refreshes under the cross-process profile lock before exposure.
func (s *store) GetBearerToken(ctx context.Context) (bearer string, err error) {
	err = s.locked(ctx, func(value *saved) error {
		p := value.Profiles[value.Active]
		if p.AccessToken == "" {
			return errors.New("ChatGPT sign-in required")
		}
		if !status(value).PlanEnabled {
			return errors.New("ChatGPT plan permission was not granted")
		}
		if !time.Now().Before(p.ExpiresAt.Add(-5 * time.Minute)) {
			if err := s.renew(ctx, value, &p); err != nil {
				return err
			}
		}
		if !status(value).PlanEnabled {
			return errors.New("ChatGPT plan permission was not granted")
		}
		bearer = p.AccessToken
		return nil
	})
	return
}
func (s *store) renew(ctx context.Context, value *saved, p *profile) error {
	config, err := s.discover(ctx)
	if err != nil {
		return err
	}
	token, err := s.exchange(ctx, config.Token, url.Values{"grant_type": {"refresh_token"}, "client_id": {p.ClientID}, "refresh_token": {p.RefreshToken}, "resource": {"https://api.openai.com/v1"}})
	if err != nil {
		var rejection tokenError
		if errors.As(err, &rejection) && terminal(string(rejection)) {
			clear(p)
			value.Profiles[value.Active] = *p
			if saveErr := s.save(value); saveErr != nil {
				return saveErr
			}
		}
		return err
	}
	if token.IDToken != "" {
		claims, err := s.identity(ctx, config, token.IDToken, p.ClientID, "")
		if err != nil {
			return err
		}
		if claims.Subject != p.Subject {
			return errors.New("OAuth refresh identity mismatch")
		}
	}
	replace(p, token)
	value.Profiles[value.Active] = *p
	return s.save(value)
}

// Refresh avoids rotating twice when another request already replaced rejected.
func (s *store) Refresh(ctx context.Context, rejected string) error {
	return s.locked(ctx, func(value *saved) error {
		p := value.Profiles[value.Active]
		if p.AccessToken == "" {
			return errors.New("ChatGPT sign-in required")
		}
		if p.AccessToken != rejected {
			return nil
		}
		return s.renew(ctx, value, &p)
	})
}

// Logout always removes local secrets, even if remote revocation cannot finish.
func (s *store) Logout(ctx context.Context) error {
	return s.locked(ctx, func(value *saved) error {
		p := value.Profiles[value.Active]
		var remote error
		if p.RefreshToken != "" {
			config, err := s.discover(ctx)
			remote = err
			if err == nil {
				values := url.Values{"token": {p.RefreshToken}, "token_type_hint": {"refresh_token"}, "client_id": {p.ClientID}}
				remote = errors.New("remote revocation was not confirmed; disconnect Ensemble in ChatGPT settings")
				for i := 0; i < 3; i++ {
					if i > 0 {
						select {
						case <-ctx.Done():
							i = 3
						case <-time.After(time.Duration(1<<i) * 100 * time.Millisecond):
						}
						if i == 3 {
							break
						}
					}
					req, _ := http.NewRequestWithContext(ctx, http.MethodPost, config.Revocation, strings.NewReader(values.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					resp, err := s.client.Do(req)
					if err != nil {
						continue
					}
					resp.Body.Close()
					if resp.StatusCode == 200 {
						remote = nil
						break
					}
					if resp.StatusCode < 500 {
						break
					}
				}
			}
		}
		clear(&p)
		if p.ClientID != "" {
			value.Profiles[value.Active] = p
		}
		return errors.Join(s.save(value), remote)
	})
}

// Models uses exactly the active inference credential and preserves catalog order.
func (s *store) Models(ctx context.Context, base string) ([]common.ModelOption, error) {
	bearer, err := s.GetBearerToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/v1/models", nil)
	if err != nil {
		return nil, errors.New("invalid model endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.New("model catalog unavailable")
	}
	defer resp.Body.Close()
	var catalog struct {
		// Models is the account-specific subscription catalog, not the API-key data array.
		Models []struct {
			// Visibility distinguishes picker entries from hidden routing aliases.
			Visibility string `json:"visibility"`
			// Slug is the exact model identifier accepted by inference.
			Slug string `json:"slug"`
			// DisplayName is the human-facing model label.
			DisplayName string `json:"display_name"`
		} `json:"models"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&catalog) != nil {
		return nil, errors.New("model catalog request failed")
	}
	var out []common.ModelOption
	for _, m := range catalog.Models {
		if m.Visibility == "list" && m.Slug != "" {
			out = append(out, common.ModelOption{Slug: m.Slug, DisplayName: m.DisplayName})
		}
	}
	return out, nil
}
