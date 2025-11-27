package domain

import "net/http"

type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
)

type GoogleAuthPurpose string

const (
	GoogleAuthPurposeLogin  GoogleAuthPurpose = "login"
	GoogleAuthPurposeSignUp GoogleAuthPurpose = "signup"
)

type LoginResult struct {
	Status     ResponseStatus
	Message    string
	Error      string
	Cookies    []*http.Cookie
	StatusCode int
}

type GoogleAuthResult struct {
	Status     ResponseStatus
	URL        string
	Error      string
	Cookies    []*http.Cookie
	StatusCode int
}

type SignUpResult struct {
	Status     ResponseStatus
	Message    string
	Error      string
	Cookies    []*http.Cookie
	StatusCode int
}

type LogoutResult struct {
	Status     ResponseStatus
	Message    string
	Error      string
	Cookies    []*http.Cookie
	StatusCode int
}

type AuthStatusResult struct {
	Status          ResponseStatus
	IsAuthenticated bool
	StatusCode      int
	Cookies         []*http.Cookie
}
