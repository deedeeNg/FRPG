// Package google implements domain.ProfileVerifier for "Sign in with Google".
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"frpg-backend/internal/domain"
)

// userInfoURL returns the profile for a Google OAuth access token. The browser
// obtains that access token via the Google Identity Services token client (a
// popup), which lets the frontend keep a custom-styled button. In exchange we
// identify the user from the userinfo endpoint instead of validating an ID
// token's aud — see ARCHITECTURE.md "Next goals" for that tradeoff.
const userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

// Verifier is the real domain.ProfileVerifier for Google.
type Verifier struct {
	HTTPClient *http.Client
	// UserInfoURL overrides Google's endpoint; used by tests.
	UserInfoURL string
}

func (v Verifier) Verify(ctx context.Context, cred domain.Credential) (domain.ProviderProfile, error) {
	if cred.Token == "" {
		return domain.ProviderProfile{}, errors.New("missing access token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	base := v.UserInfoURL
	if base == "" {
		base = userInfoURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	if err != nil {
		return domain.ProviderProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cred.Token)

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
	}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return domain.ProviderProfile{}, err
	}
	if claims.Sub == "" {
		return domain.ProviderProfile{}, errors.New("userinfo response missing sub")
	}
	return domain.ProviderProfile{
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		DisplayName:    claims.Name,
	}, nil
}
