package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"frpg-backend/internal/users"
)

// googleTokenInfoURL verifies a Google ID token. A production-hardened setup
// verifies the JWT signature locally against Google's JWKS; this endpoint is a
// dependency-free equivalent that is fine to start with. Either way, it sits
// behind ProfileVerifier, so swapping in JWKS verification later is a one-file
// change and does not touch the provider or its tests.
const googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// GoogleVerifier is the real ProfileVerifier for "Sign in with Google". It takes
// the ID token the frontend received from Google (Credential.Token) and returns
// the verified profile. Tests use a fake verifier instead of this.
type GoogleVerifier struct {
	HTTPClient *http.Client
	// Audience, if set, must equal the token's aud (your OAuth client_id).
	Audience string
	// TokenInfoURL overrides Google's endpoint; used by tests. Defaults to
	// googleTokenInfoURL when empty.
	TokenInfoURL string
}

func (v GoogleVerifier) Verify(ctx context.Context, cred Credential) (ProviderProfile, error) {
	if cred.Token == "" {
		return ProviderProfile{}, errors.New("missing id token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	base := v.TokenInfoURL
	if base == "" {
		base = googleTokenInfoURL
	}
	endpoint := base + "?" + url.Values{"id_token": {cred.Token}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ProviderProfile{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return ProviderProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ProviderProfile{}, fmt.Errorf("token rejected (status %d)", resp.StatusCode)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Aud   string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return ProviderProfile{}, err
	}
	if v.Audience != "" && claims.Aud != v.Audience {
		return ProviderProfile{}, errors.New("token audience mismatch")
	}
	return ProviderProfile{
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		DisplayName:    claims.Name,
	}, nil
}

// NewGoogleProvider builds a ready-to-use Google IdentityProvider. At deploy time
// you supply the audience (client_id); nothing else changes.
func NewGoogleProvider(audience string, repo users.Repository) OAuthProvider {
	return NewOAuthProvider("google", GoogleVerifier{Audience: audience}, repo)
}
