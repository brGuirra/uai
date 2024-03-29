package token

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")     // ErrInvalidToken is the error for invalid token
	ErrExpiredToken = errors.New("token has expired") // ErrExpiredToken is the error for expired token
)

// TokenScope is the type for token scope
type TokenScope string

const (
	ScopeActivation     = TokenScope("activation")     // ScopeActivation is the scope for activating user
	ScopeAuthentication = TokenScope("authentication") // ScopeAuthentication is the scope for authenticating user
	ScopePasswordReset  = TokenScope("password_reset") // ScopePasswordReset is the scope for resetting user password
)

// Payload is the payload for a token
type Payload struct {
	UserID string
	Scope  TokenScope
}

// getDurationForScope returns the duration for a given scope
func getDurationForScope(scope TokenScope) time.Duration {
	switch scope {
	case ScopeActivation:
		return 72 * time.Hour
	case ScopePasswordReset:
		return 15 * time.Minute
	default:
		return time.Hour
	}
}
