package auth

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type googleAuthResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type logoutResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type authStatusResponse struct {
	IsAuthenticated bool   `json:"authenticated"`
	Error           string `json:"error"`
}
