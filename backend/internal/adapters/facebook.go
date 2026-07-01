package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"frpg-backend/internal/domain"
)

// facebookGraphURL returns the profile for a Facebook access token.
const facebookGraphURL = "https://graph.facebook.com/me"

// FacebookVerifier is the real domain.ProfileVerifier for "Continue with
// Facebook". Same interface as GoogleVerifier — only the service it calls differs.
type FacebookVerifier struct {
	HTTPClient *http.Client
	// GraphURL overrides the Graph endpoint; used by tests.
	GraphURL string
}

func (v FacebookVerifier) Verify(ctx context.Context, cred domain.Credential) (domain.ProviderProfile, error) {
	if cred.Token == "" {
		return domain.ProviderProfile{}, errors.New("missing access token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	base := v.GraphURL
	if base == "" {
		base = facebookGraphURL
	}
	endpoint := base + "?" + url.Values{
		"fields":       {"id,name,email"},
		"access_token": {cred.Token},
	}.Encode()

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

	var profile struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return domain.ProviderProfile{}, err
	}
	if profile.ID == "" {
		return domain.ProviderProfile{}, errors.New("graph response missing id")
	}
	return domain.ProviderProfile{
		ProviderUserID: profile.ID,
		Email:          profile.Email,
		DisplayName:    profile.Name,
	}, nil
}
