// Package google implements domain.ProfileVerifier for "Sign in with Google".
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"frpg-backend/internal/domain"
)

// tokenInfoURL verifies a Google ID token. A production-hardened setup verifies
// the JWT signature locally against Google's JWKS; this endpoint is a
// dependency-free equivalent that is fine to start with.
const tokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// Verifier is the real domain.ProfileVerifier for Google.
type Verifier struct {
	HTTPClient *http.Client
	// Audience, if set, must equal the token's aud (your OAuth client_id).
	Audience string
	// TokenInfoURL overrides Google's endpoint; used by tests.
	TokenInfoURL string
}

func (v Verifier) Verify(ctx context.Context, cred domain.Credential) (domain.ProviderProfile, error) {
	if cred.Token == "" {
		return domain.ProviderProfile{}, errors.New("missing id token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	base := v.TokenInfoURL
	if base == "" {
		base = tokenInfoURL
	}
	endpoint := base + "?" + url.Values{"id_token": {cred.Token}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.ProviderProfile{}, fmt.Errorf("token rejected (status %d)", resp.StatusCode)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Aud   string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return domain.ProviderProfile{}, err
	}
	if v.Audience != "" && claims.Aud != v.Audience {
		return domain.ProviderProfile{}, errors.New("token audience mismatch")
	}
	return domain.ProviderProfile{
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		DisplayName:    claims.Name,
	}, nil
}
