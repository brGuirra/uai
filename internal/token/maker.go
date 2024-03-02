package token

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific userId and scope
	CreateToken(userID string, scope TokenScope) string

	// VerifyToken checks if a token is valid or not
	VerifyToken(token string) (*Payload, error)
}
