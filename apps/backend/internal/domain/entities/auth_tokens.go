package entities

// AuthTokens holds the access and refresh token pair returned to clients.
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
