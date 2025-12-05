package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/internal/domain"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleUserInfoURL    = "https://www.googleapis.com/oauth2/v3/userinfo"
	googleProviderName   = "google"
	httpClientTimeout    = 30 * time.Second
	maxErrorResponseSize = 1024
)

type OAuthGateway struct {
	config oauth2.Config
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewOAuthGateway(config GoogleOAuthConfig) *OAuthGateway {
	return &OAuthGateway{
		config: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *OAuthGateway) GetOAuthUserInfo(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
	conf := g.config
	conf.RedirectURL = fmt.Sprintf("%s/%s", conf.RedirectURL, purpose)
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	httpClient := g.config.Client(ctx, token)
	httpClient.Timeout = httpClientTimeout

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		limitedReader := io.LimitReader(resp.Body, maxErrorResponseSize)
		io.Copy(io.Discard, limitedReader)
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	googleUser := googleUserDTO{}
	err = json.NewDecoder(resp.Body).Decode(&googleUser)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &domain.OAuthUserInfo{
		Sub:          googleUser.Sub,
		Email:        googleUser.Email,
		FullName:     googleUser.Name,
		ProviderName: googleProviderName,
	}, nil
}

func (g *OAuthGateway) GetGoogleAuthURL(ctx context.Context, purpose, state string) string {
	conf := g.config
	conf.RedirectURL = fmt.Sprintf("%s/%s", conf.RedirectURL, purpose)
	return conf.AuthCodeURL(state)
}
