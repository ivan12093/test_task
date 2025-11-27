package domain

import "net/http"

type Profile struct {
	FullName string
	Phone    string
	Email    string
}

type ProfileResult struct {
	Status     ResponseStatus
	Profile    *Profile
	Error      string
	Cookies    []*http.Cookie
	StatusCode int
}
