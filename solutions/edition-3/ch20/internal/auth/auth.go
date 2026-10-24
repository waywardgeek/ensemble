package auth

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Options is application wiring, never saved conversation configuration.
type Options struct {
	Directory, Issuer, AuthorizationURL, TokenURL, JWKSURL, RevocationURL, ModelsURL, RedirectURI string
	HTTPClient                                                                                    *http.Client
}
type Service struct {
	options Options
	client  *http.Client
}

func New(o Options) *Service {
	if o.Issuer == "" {
		o.Issuer = "https://auth.openai.com"
	}
	if o.ModelsURL == "" {
		o.ModelsURL = "https://api.openai.com/v1/models"
	}
	c := &http.Client{Timeout: 30 * time.Second}
	if o.HTTPClient != nil {
		*c = *o.HTTPClient
	}
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Service{options: o, client: c}
}

var errRejected = errors.New("ChatGPT authorization expired or revoked; log in again")
var errTransient = errors.New("auth service temporarily unavailable")

func (s *Service) doJSON(req *http.Request, out any) error {
	if !allowedEndpoint(req.URL.String()) {
		return errors.New("unsafe auth endpoint")
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return errTransient
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(b) > 1<<20 {
		return errTransient
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(b, &failure)
		switch failure.Error {
		case "invalid_grant", "invalid_refresh_token", "token_expired", "refresh_token_expired", "refresh_token_invalidated", "refresh_token_reused":
			return errRejected
		}
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			return errTransient
		}
		return fmt.Errorf("auth service returned HTTP %d", resp.StatusCode)
	}
	if out != nil && json.Unmarshal(b, out) != nil {
		return errors.New("invalid auth service response")
	}
	return nil
}
func (s *Service) request(ctx context.Context, method, endpoint, ctype string, body io.Reader, out any) (int, error) {
	r, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return 0, errors.New("invalid auth request")
	}
	r.Header.Set("Content-Type", ctype)
	return 0, s.doJSON(r, out)
}

type tokenReply struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	ID      string `json:"id_token"`
	Type    string `json:"token_type"`
	Expires int64  `json:"expires_in"`
	Scope   string `json:"scope"`
}

func (s *Service) tokens(ctx context.Context, form url.Values) (t tokenReply, err error) {
	o, err := s.endpoints(ctx)
	if err != nil {
		return t, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		_, err = s.request(ctx, http.MethodPost, o.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()), &t)
		if !errors.Is(err, errTransient) {
			break
		}
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return t, ctx.Err()
			case <-time.After(time.Duration(100*(1<<attempt)) * time.Millisecond):
			}
		}
	}
	if err != nil {
		return t, err
	}
	if t.Access == "" || !strings.EqualFold(t.Type, "bearer") || t.Expires <= 0 || t.Expires > 365*24*3600 {
		return t, errors.New("invalid auth token response")
	}
	return t, nil
}
func hasInference(scopes []string) bool {
	for _, s := range scopes {
		if s == inferenceScope {
			return true
		}
	}
	return false
}
func billing(st *store) common.AuthMode {
	if st.Mode == "" {
		return common.AuthAPI
	}
	return st.Mode
}
func status(st *store, profile string) common.AuthStatus {
	if profile == "" {
		profile = st.Selected
	}
	g, ok := st.Accounts[profile]
	if !ok {
		return common.AuthStatus{Mode: billing(st), Reason: "no ChatGPT login"}
	}
	reason := g.Reason
	if reason == "" && !g.Eligible {
		reason = "identity verified; plan inference scope not granted"
	}
	return common.AuthStatus{Mode: billing(st), Account: g.Account, Eligible: g.Eligible && g.Refresh != "" && g.Reason == "", ExpiresAt: g.Expires, Reason: reason}
}
func (s *Service) Status(ctx context.Context) (out common.AuthStatus, err error) {
	err = s.locked(ctx, func(st *store) error { out = status(st, ""); return nil })
	return
}
func (s *Service) Mode(ctx context.Context) (out common.AuthMode, err error) {
	err = s.locked(ctx, func(st *store) error {
		out = billing(st)
		if out != common.AuthAPI && out != common.AuthChatGPT {
			return errors.New("invalid stored billing selection")
		}
		return nil
	})
	return
}
func (s *Service) Resolve(ctx context.Context, c common.Config) (out common.Credential, err error) {
	mode := c.AuthMode
	if mode == "" {
		mode, err = s.Mode(ctx)
		if err != nil {
			return out, err
		}
	}
	if mode == common.AuthAPI {
		if c.APIKey == "" {
			return out, errors.New("API billing selected but OPENAI_API_KEY is missing")
		}
		return common.Credential{Mode: mode, Token: c.APIKey}, nil
	}
	if mode != common.AuthChatGPT {
		return out, errors.New("automatic billing fallback is not implemented")
	}
	err = s.locked(ctx, func(st *store) error {
		profile := st.Selected
		if c.Account != "" {
			profile = ""
			for id, g := range st.Accounts {
				if id == c.Account || g.Account == c.Account {
					if profile != "" {
						return errors.New("ambiguous ChatGPT account selector")
					}
					profile = id
				}
			}
		}
		g, ok := st.Accounts[profile]
		if !ok {
			return errors.New("ChatGPT login required; API billing was not selected")
		}
		if c.Surface != common.Responses {
			return errors.New("ChatGPT inference requires Responses")
		}
		if st.Client == nil || st.Client.ClientID != g.ClientID || st.Client.Issuer != s.options.Issuer || g.Issuer != s.options.Issuer || g.Refresh == "" || g.Reason != "" {
			return errRejected
		}
		if !g.Eligible || !hasInference(g.Scopes) {
			return errors.New("ChatGPT identity has no plan inference grant; API fallback refused")
		}
		if time.Until(g.Expires) <= 5*time.Minute {
			t, refreshErr := s.tokens(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {g.Refresh}, "client_id": {g.ClientID}, "resource": {resource}})
			if refreshErr != nil {
				if errors.Is(refreshErr, errRejected) {
					invalidate(&g, "ChatGPT authorization revoked; log in again")
					st.Accounts[profile] = g
				}
				return refreshErr
			}
			if t.Refresh == "" {
				return errors.New("refresh rotation omitted replacement token")
			}
			if t.ID != "" {
				id, verifyErr := s.verify(ctx, t.ID, g.ClientID, "")
				if verifyErr != nil || id.Subject != g.Subject {
					invalidate(&g, "renewed identity could not be verified; log in again")
					st.Accounts[profile] = g
					return errRejected
				}
			}
			g.Access = t.Access
			g.Refresh = t.Refresh
			g.Expires = time.Now().Add(time.Duration(t.Expires) * time.Second)
			g.Generation++
			if t.Scope != "" {
				g.Scopes = strings.Fields(t.Scope)
				g.Eligible = hasInference(g.Scopes)
			}
			st.Accounts[profile] = g
			if !g.Eligible {
				return errors.New("renewed grant does not permit plan inference")
			}
		}
		out = common.Credential{Mode: mode, Token: g.Access, Account: g.Account, Profile: profile, Generation: g.Generation}
		return nil
	})
	return
}
func invalidate(g *grant, reason string) {
	g.Access = ""
	g.Refresh = ""
	g.Eligible = false
	g.Reason = reason
	g.Generation++
}

// An inference 401 does not revoke a refresh grant. Expire only the generation
// actually rejected, then let Resolve renew once through the same source.
func (s *Service) Reject(ctx context.Context, c common.Credential) error {
	if c.Mode != common.AuthChatGPT {
		return nil
	}
	return s.locked(ctx, func(st *store) error {
		g, ok := st.Accounts[c.Profile]
		if ok && g.Generation == c.Generation {
			g.Expires = time.Time{}
			st.Accounts[c.Profile] = g
		}
		return nil
	})
}
func (s *Service) Logout(ctx context.Context, account string, all bool) error {
	return s.locked(ctx, func(st *store) error {
		var revoke []string
		for id, g := range st.Accounts {
			if all || (account == "" && id == st.Selected) || account == id || account == g.Account {
				revoke = append(revoke, g.Refresh)
				delete(st.Accounts, id)
				if st.Selected == id {
					st.Selected = ""
				}
			}
		}
		if len(revoke) == 0 {
			return nil
		}
		// Commit local removal before network revocation; even interruption must
		// not revive a locally logged-out grant on the next process start.
		if err := s.save(st); err != nil {
			return err
		}
		o, err := s.endpoints(ctx)
		if err != nil || o.RevocationURL == "" {
			return errors.New("local tokens removed; remote revocation could not be confirmed")
		}
		for _, token := range revoke {
			if token != "" {
				_, err = s.request(ctx, http.MethodPost, o.RevocationURL, "application/x-www-form-urlencoded", strings.NewReader(url.Values{"token": {token}, "client_id": {st.Client.ClientID}}.Encode()), nil)
				if err != nil {
					return errors.New("local tokens removed; remote revocation could not be confirmed")
				}
			}
		}
		return nil
	})
}

// Models is an explicit user command, never implicit discovery during inference.
func (s *Service) Models(ctx context.Context, c common.Config) ([]common.AuthModel, error) {
	c.Vendor = common.OpenAI
	c.Surface = common.Responses
	cred, err := s.Resolve(ctx, c)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.options.ModelsURL, nil)
	if err != nil {
		return nil, errors.New("invalid model catalog endpoint")
	}
	req.Header.Set("Authorization", "Bearer "+cred.Token)
	var result struct {
		Models []struct {
			Slug       string `json:"slug"`
			Name       string `json:"display_name"`
			Visibility string `json:"visibility"`
		} `json:"models"`
	}
	if err = s.doJSON(req, &result); err != nil {
		return nil, err
	}
	out := []common.AuthModel{}
	for _, m := range result.Models {
		if m.Slug != "" && m.Visibility == "list" {
			out = append(out, common.AuthModel{Slug: m.Slug, Name: m.Name})
		}
	}
	return out, nil
}
