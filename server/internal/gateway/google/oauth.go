package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/internal/domain"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
const googleProviderName = "google"

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

	client := g.config.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body: %s", resp.StatusCode, string(body))
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
