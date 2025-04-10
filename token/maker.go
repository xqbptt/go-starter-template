package token

// Maker is an interface for managing tokens
type Maker interface {
	CreateAccessToken(userID int64) (string, error)

	CreateRefreshToken(userID int64, tokenHash string) (string, error)

	ValidateAccessToken(token string) (*Payload, error)

	ValidateRefreshToken(token string) (*RefreshPayload, error)

	ValidateRefreshHash(hash string, userID int64, originalHash string) error
}
