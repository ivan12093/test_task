package profile

type profileResponse struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type profileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}
