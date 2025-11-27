package auth

import "net/http"

type pageData struct {
	Email                string
	Error                string
	Message              string
	Title                string
	HeaderTitle          string
	HeaderSubtitle       string
	FormAction           string
	SubmitButtonText     string
	GoogleAuthURL        string
	GoogleButtonText     string
	PasswordAutocomplete string
	FooterText           string
	FooterLink           string
	FooterLinkText       string
}

type pageDataOptions struct {
	Email   string
	Error   string
	Message string
}

func newLoginPageData(opts pageDataOptions) pageData {
	data := pageData{
		Title:                "Login",
		HeaderTitle:          "Welcome",
		HeaderSubtitle:       "Sign in to your account",
		FormAction:           "/login",
		SubmitButtonText:     "Sign In",
		GoogleAuthURL:        "/login/google",
		GoogleButtonText:     "Sign in with Google",
		PasswordAutocomplete: "current-password",
		FooterText:           "Don't have an account?",
		FooterLink:           "/signup",
		FooterLinkText:       "Sign Up",
	}

	if opts.Email != "" {
		data.Email = opts.Email
	}
	if opts.Error != "" {
		data.Error = opts.Error
	}
	if opts.Message != "" {
		data.Message = opts.Message
	}

	return data
}

func newSignUpPageData(opts pageDataOptions) pageData {
	data := pageData{
		Title:                "Sign Up",
		HeaderTitle:          "Create Account",
		HeaderSubtitle:       "Sign up for a new account",
		FormAction:           "/signup",
		SubmitButtonText:     "Sign Up",
		GoogleAuthURL:        "/signup/google",
		GoogleButtonText:     "Sign up with Google",
		PasswordAutocomplete: "new-password",
		FooterText:           "Already have an account?",
		FooterLink:           "/login",
		FooterLinkText:       "Sign In",
	}

	if opts.Email != "" {
		data.Email = opts.Email
	}
	if opts.Error != "" {
		data.Error = opts.Error
	}
	if opts.Message != "" {
		data.Message = opts.Message
	}

	return data
}

func setCookies(w http.ResponseWriter, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
}
