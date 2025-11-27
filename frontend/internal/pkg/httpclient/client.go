package httpclient

import (
	"net/http"

	"frontend/internal/pkg/cookies"
)

const (
	csrfTokenCookieName = "csrf_token"
	csrfTokenHeaderName = "X-CSRF-Token"
)

type CookieTransport struct {
	Base http.RoundTripper
}

func (t *CookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, cookie := range cookies.FromContext(req.Context()) {
		req.AddCookie(cookie)
		if cookie.Name == csrfTokenCookieName && cookie.Value != "" {
			req.Header.Set(csrfTokenHeaderName, cookie.Value)
		}
	}

	if req.Header.Get(csrfTokenHeaderName) == "" {
		for _, cookie := range req.Cookies() {
			if cookie.Name == csrfTokenCookieName && cookie.Value != "" {
				req.Header.Set(csrfTokenHeaderName, cookie.Value)
				break
			}
		}
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

func New(client *http.Client) *http.Client {
	if client.Transport == nil {
		client.Transport = &CookieTransport{}
	} else {
		client.Transport = &CookieTransport{Base: client.Transport}
	}
	return client
}
