package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type identity struct {
	Subject, Account string
}

func (s *Service) verify(ctx context.Context, token, clientID, nonce string) (identity, error) {
	bad := errors.New("ChatGPT identity validation failed")
	var out identity
	pieces := strings.Split(token, ".")
	if len(pieces) != 3 {
		return out, bad
	}
	decode := base64.RawURLEncoding.DecodeString
	header, err := decode(pieces[0])
	if err != nil {
		return out, bad
	}
	var h struct {
		Alg  string   `json:"alg"`
		Kid  string   `json:"kid"`
		Crit []string `json:"crit"`
	}
	if json.Unmarshal(header, &h) != nil || h.Alg != "RS256" || h.Kid == "" || len(h.Crit) > 0 {
		return out, bad
	}
	var keys struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	o, err := s.endpoints(ctx)
	if err != nil {
		return out, err
	}
	if _, err = s.request(ctx, http.MethodGet, o.JWKSURL, "", nil, &keys); err != nil {
		return out, err
	}
	var key *rsa.PublicKey
	for _, k := range keys.Keys {
		if k.Kid != h.Kid || k.Kty != "RSA" || (k.Alg != "" && k.Alg != "RS256") || (k.Use != "" && k.Use != "sig") {
			continue
		}
		if key != nil {
			return out, bad
		}
		n, e1 := decode(k.N)
		eb, e2 := decode(k.E)
		if e1 != nil || e2 != nil || len(eb) == 0 || len(eb) > 4 {
			return out, bad
		}
		exponent := 0
		for _, b := range eb {
			exponent = exponent*256 + int(b)
		}
		key = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
		if key.N.BitLen() < 2048 || exponent < 3 || exponent%2 == 0 {
			return out, bad
		}
	}
	if key == nil {
		return out, bad
	}
	signature, err := decode(pieces[2])
	if err != nil {
		return out, bad
	}
	hash := sha256.Sum256([]byte(pieces[0] + "." + pieces[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature) != nil {
		return out, bad
	}
	payload, err := decode(pieces[1])
	if err != nil {
		return out, bad
	}
	var claims struct {
		Issuer          string          `json:"iss"`
		Subject         string          `json:"sub"`
		Audience        json.RawMessage `json:"aud"`
		AuthorizedParty string          `json:"azp"`
		Nonce           string          `json:"nonce"`
		Expires         int64           `json:"exp"`
		Issued          int64           `json:"iat"`
		NotBefore       int64           `json:"nbf"`
	}
	if json.Unmarshal(payload, &claims) != nil {
		return out, bad
	}
	var audiences []string
	var audience string
	if json.Unmarshal(claims.Audience, &audience) == nil {
		audiences = []string{audience}
	} else if json.Unmarshal(claims.Audience, &audiences) != nil {
		return out, bad
	}
	found := false
	for _, a := range audiences {
		if a == clientID {
			found = true
		}
	}
	now := time.Now().Unix()
	if claims.Issuer != s.options.Issuer || !found || (len(audiences) > 1 && claims.AuthorizedParty != clientID) || (claims.AuthorizedParty != "" && claims.AuthorizedParty != clientID) || claims.Subject == "" || claims.Expires <= now || claims.Issued <= 0 || claims.Issued > now+30 || claims.Expires <= claims.Issued || claims.NotBefore > now+30 || (nonce != "" && subtle.ConstantTimeCompare([]byte(nonce), []byte(claims.Nonce)) != 1) {
		return out, bad
	}
	out.Subject = claims.Subject
	out.Account = claims.Subject
	return out, nil
}
