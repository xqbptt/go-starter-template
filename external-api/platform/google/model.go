package google

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

type UserInfoResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
