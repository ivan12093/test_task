package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"frontend/internal/domain"
	"frontend/internal/pkg/httpclient"
)

const (
	defaultTimeout     = 10 * time.Second
	loginURI           = "/api/auth/login"
	signupURI          = "/api/auth/signup"
	logoutURI          = "/api/auth/logout"
	googleAuthURI      = "/api/auth/google/url"
	jsonContentType    = "application/json"
	checkAuthStatusURI = "/api/auth/status"
)

var googleAuthPurposeMap = map[domain.GoogleAuthPurpose]string{
	domain.GoogleAuthPurposeLogin:  "login",
	domain.GoogleAuthPurposeSignUp: "signup",
}

type Gateway struct {
	client     *http.Client
	apiBaseURL string
}

func NewGateway(apiBaseURL string) *Gateway {
	return &Gateway{
		client:     httpclient.New(&http.Client{Timeout: defaultTimeout}),
		apiBaseURL: apiBaseURL,
	}
}

func (g *Gateway) makeRequestWithBody(ctx context.Context, method, url string, data interface{}) (*http.Response, error) {
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", jsonContentType)

	return g.client.Do(req)
}

func (g *Gateway) makeRequestWithoutBody(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return g.client.Do(req)
}

func (g *Gateway) Login(ctx context.Context, email, password string) (*domain.LoginResult, error) {
	resp, err := g.makeRequestWithBody(ctx, http.MethodPost, g.apiBaseURL+loginURI, loginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	var respDTO loginResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	var status domain.ResponseStatus
	if resp.StatusCode == http.StatusOK {
		status = domain.ResponseStatusSuccess
	} else {
		status = domain.ResponseStatusError
	}
	return &domain.LoginResult{
		Status:     status,
		Message:    respDTO.Message,
		Error:      respDTO.Error,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}, nil
}

func (g *Gateway) GetGoogleAuthURL(ctx context.Context, purpose domain.GoogleAuthPurpose) (*domain.GoogleAuthResult, error) {
	resp, err := g.makeRequestWithoutBody(ctx, http.MethodGet, g.apiBaseURL+googleAuthURI+"?purpose="+googleAuthPurposeMap[purpose])
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respDTO googleAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&respDTO); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var status domain.ResponseStatus
	if resp.StatusCode == http.StatusOK {
		status = domain.ResponseStatusSuccess
	} else {
		status = domain.ResponseStatusError
	}

	return &domain.GoogleAuthResult{
		Status:     status,
		URL:        respDTO.URL,
		Error:      respDTO.Error,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}, nil
}

func (g *Gateway) SignUp(ctx context.Context, email, password string) (*domain.SignUpResult, error) {
	resp, err := g.makeRequestWithBody(ctx, http.MethodPost, g.apiBaseURL+signupURI, signUpRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	var respDTO signUpResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	var status domain.ResponseStatus
	if resp.StatusCode == http.StatusCreated {
		status = domain.ResponseStatusSuccess
	} else {
		status = domain.ResponseStatusError
	}
	return &domain.SignUpResult{
		Status:     status,
		Message:    respDTO.Message,
		Error:      respDTO.Error,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}, nil
}

func (g *Gateway) Logout(ctx context.Context) (*domain.LogoutResult, error) {
	resp, err := g.makeRequestWithoutBody(ctx, http.MethodPost, g.apiBaseURL+logoutURI)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respDTO logoutResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to logout: %w", err)
	}

	var status domain.ResponseStatus
	if resp.StatusCode == http.StatusOK {
		status = domain.ResponseStatusSuccess
	} else {
		status = domain.ResponseStatusError
	}
	return &domain.LogoutResult{
		Status:     status,
		Message:    respDTO.Message,
		Error:      respDTO.Error,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}, nil
}

func (g *Gateway) CheckAuthStatus(ctx context.Context) (*domain.AuthStatusResult, error) {
	resp, err := g.makeRequestWithoutBody(ctx, http.MethodGet, g.apiBaseURL+checkAuthStatusURI)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respDTO authStatusResponse
	err = json.NewDecoder(resp.Body).Decode(&respDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to check auth status: %w", err)
	}

	var status domain.ResponseStatus
	if resp.StatusCode == http.StatusOK {
		status = domain.ResponseStatusSuccess
	} else {
		status = domain.ResponseStatusError
	}
	return &domain.AuthStatusResult{
		Status:          status,
		IsAuthenticated: respDTO.IsAuthenticated,
		StatusCode:      resp.StatusCode,
		Cookies:         resp.Cookies(),
	}, nil
}
