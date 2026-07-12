package dto

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	Message      string       `json:"message"`
}
