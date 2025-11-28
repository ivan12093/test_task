package profile

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
	defaultTimeout   = 10 * time.Second
	getProfileURI    = "/api/profile"
	updateProfileURI = "/api/profile"
	jsonContentType  = "application/json"
)

type gateway struct {
	client     *http.Client
	apiBaseURL string
}

func NewGateway(apiBaseURL string) Gateway {
	return &gateway{
		client:     httpclient.New(&http.Client{Timeout: defaultTimeout}),
		apiBaseURL: apiBaseURL,
	}
}

func (g *gateway) makeRequestWithBody(ctx context.Context, method, url string, data interface{}) (*http.Response, error) {
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

func (g *gateway) makeRequestWithoutBody(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return g.client.Do(req)
}

func (g *gateway) GetProfile(ctx context.Context) (*domain.ProfileResult, error) {
	resp, err := g.makeRequestWithoutBody(ctx, http.MethodGet, g.apiBaseURL+getProfileURI)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	result := &domain.ProfileResult{
		Status:     domain.ResponseStatusSuccess,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode != http.StatusOK {
		result.Status = domain.ResponseStatusError
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			result.Error = errorResp.Error
		} else {
			result.Error = fmt.Sprintf("failed to get profile: status %d", resp.StatusCode)
		}
		return result, nil
	}

	var profileResp profileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result.Profile = &domain.Profile{
		FullName: profileResp.FullName,
		Phone:    profileResp.Phone,
		Email:    profileResp.Email,
	}

	return result, nil
}

func (g *gateway) UpdateProfile(ctx context.Context, profile *domain.Profile) (*domain.ProfileResult, error) {
	req := profileRequest{
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Email:    profile.Email,
	}

	resp, err := g.makeRequestWithBody(ctx, http.MethodPut, g.apiBaseURL+updateProfileURI, req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	result := &domain.ProfileResult{
		Status:     domain.ResponseStatusSuccess,
		Cookies:    resp.Cookies(),
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode != http.StatusOK {
		result.Status = domain.ResponseStatusError
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			result.Error = errorResp.Error
		} else {
			result.Error = fmt.Sprintf("failed to update profile: status %d", resp.StatusCode)
		}
		return result, nil
	}

	result.Profile = profile

	return result, nil
}
