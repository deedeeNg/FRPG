// Package facebook implements domain.ProfileVerifier for "Continue with Facebook".
package facebook

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
	// graphURL returns the profile for a Facebook access token.
	graphURL = "https://graph.facebook.com/me"
	// debugTokenURL reports which app a token belongs to — Facebook's equivalent
	// of Google's audience check. It needs an app access token to call.
	debugTokenURL = "https://graph.facebook.com/debug_token"
)

// Verifier is the real domain.ProfileVerifier for Facebook. Same interface as the
// Google verifier — only the service it calls differs.
type Verifier struct {
	HTTPClient *http.Client
	// AppID + AppToken enable the app-audience check: when both are set, the token
	// must belong to AppID (verified via debug_token, authenticated by AppToken,
	// which is "{app-id}|{app-secret}"). Empty disables the check (dev/tests only).
	AppID    string
	AppToken string
	// GraphURL / DebugTokenURL override the Graph endpoints; used by tests.
	GraphURL      string
	DebugTokenURL string
}

func (v Verifier) Verify(ctx context.Context, cred domain.Credential) (domain.ProviderProfile, error) {
	if cred.Token == "" {
		return domain.ProviderProfile{}, errors.New("missing access token")
	}
	client := v.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	if v.AppID != "" && v.AppToken != "" {
		if err := v.checkApp(ctx, client, cred.Token); err != nil {
			return domain.ProviderProfile{}, err
		}
	}

	base := v.GraphURL
	if base == "" {
		base = graphURL
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

// checkApp confirms the access token was issued for our app, closing the
// token-substitution hole (a token minted for another app is rejected).
func (v Verifier) checkApp(ctx context.Context, client *http.Client, token string) error {
	base := v.DebugTokenURL
	if base == "" {
		base = debugTokenURL
	}
	endpoint := base + "?" + url.Values{
		"input_token":  {token},
		"access_token": {v.AppToken},
	}.Encode()
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
	var out struct {
		Data struct {
			AppID   string `json:"app_id"`
			IsValid bool   `json:"is_valid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if !out.Data.IsValid {
		return errors.New("facebook token is not valid")
	}
	if out.Data.AppID != v.AppID {
		return errors.New("token app mismatch")
	}
	return nil
}
