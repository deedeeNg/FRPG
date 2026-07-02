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

const (
	// userInfoURL returns the profile (sub/email/name) for an access token.
	userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	// tokenInfoURL reports the token's audience so we can confirm it was minted
	// for THIS app (client_id), not some other app with email/profile scope.
	tokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"
)

// Verifier is the real domain.ProfileVerifier for Google. The browser sends an
// OAuth access token (from the GIS token client, so a custom button works); we
// first check its audience, then read the profile from userinfo.
type Verifier struct {
	HTTPClient *http.Client
	// Audience, when set, must equal the token's aud/azp (your OAuth client_id).
	// Empty disables the check (dev/tests only) — set it in production.
	Audience string
	// UserInfoURL / TokenInfoURL override Google's endpoints; used by tests.
	UserInfoURL  string
	TokenInfoURL string
}

func (v Verifier) Verify(ctx context.Context, cred domain.Credential) (domain.ProviderProfile, error) {
	if cred.Token == "" {
		return domain.ProviderProfile{}, errors.New("missing access token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	if v.Audience != "" {
		if err := v.checkAudience(ctx, client, cred.Token); err != nil {
			return domain.ProviderProfile{}, err
		}
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

// checkAudience confirms the access token was issued to our client_id, closing
// the token-substitution hole (a token minted for another app is rejected).
func (v Verifier) checkAudience(ctx context.Context, client *http.Client, token string) error {
	base := v.TokenInfoURL
	if base == "" {
		base = tokenInfoURL
	}
	endpoint := base + "?" + url.Values{"access_token": {token}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token rejected (status %d)", resp.StatusCode)
	}
	var info struct {
		Aud string `json:"aud"`
		Azp string `json:"azp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return err
	}
	if info.Aud != v.Audience && info.Azp != v.Audience {
		return errors.New("token audience mismatch")
	}
	return nil
}
