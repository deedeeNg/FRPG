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

// facebookGraphURL returns the profile for a Facebook access token. Facebook
// login is OAuth: the frontend obtains an access token, the backend calls the
// Graph API to fetch the user's id/name/email.
const facebookGraphURL = "https://graph.facebook.com/me"

// FacebookVerifier is the real ProfileVerifier for "Continue with Facebook".
// It takes the access token (Credential.Token) and returns the verified profile.
// Same interface as GoogleVerifier — the only difference is which service it calls.
type FacebookVerifier struct {
	HTTPClient *http.Client
	// GraphURL overrides the Graph endpoint; used by tests. Defaults to
	// facebookGraphURL when empty.
	GraphURL string
}

func (v FacebookVerifier) Verify(ctx context.Context, cred Credential) (ProviderProfile, error) {
	if cred.Token == "" {
		return ProviderProfile{}, errors.New("missing access token")
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

	var profile struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return ProviderProfile{}, err
	}
	if profile.ID == "" {
		return ProviderProfile{}, errors.New("graph response missing id")
	}
	return ProviderProfile{
		ProviderUserID: profile.ID,
		Email:          profile.Email,
		DisplayName:    profile.Name,
	}, nil
}

// NewFacebookProvider builds a ready-to-use Facebook IdentityProvider.
func NewFacebookProvider(repo users.Repository) OAuthProvider {
	return NewOAuthProvider("facebook", FacebookVerifier{}, repo)
}
