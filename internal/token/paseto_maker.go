package token

import (
	"time"

	"aidanwoods.dev/go-paseto"
)

// PasetoMaker is a PASETO token maker
type PasetoMaker struct {
	parser *paseto.Parser
	key    paseto.V4SymmetricKey
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(hexKey string) (Maker, error) {
	key, err := paseto.V4SymmetricKeyFromHex(hexKey)
	if err != nil {
		return nil, err
	}

	parser := paseto.NewParser()

	maker := &PasetoMaker{
		parser: &parser,
		key:    key,
	}

	return maker, nil
}

// CreateToken creates a new encrypted token with user id and scope
func (maker *PasetoMaker) CreateToken(userID string, scope TokenScope) string {
	token := paseto.NewToken()

	token.SetString("user_id", userID)
	token.SetString("scope", string(scope))
	token.SetIssuedAt(time.Now())
	token.SetExpiration(time.Now().Add(getDurationForScope(scope)))

	return token.V4Encrypt(maker.key, nil)
}

// VerifyToken checks if a token is valid and if the scope is correct
func (maker *PasetoMaker) VerifyToken(encrypted string, requiredScope TokenScope) (*Payload, error) {
	token, err := maker.parser.ParseV4Local(maker.key, encrypted, nil)
	if err != nil {
		switch err.Error() {
		case "this token has expired":
			return nil, ErrExpiredToken
		default:
			return nil, ErrInvalidToken
		}
	}

	userID, err := token.GetString("user_id")
	if err != nil {
		return nil, ErrInvalidToken
	}

	scope, err := token.GetString("scope")
	if err != nil {
		return nil, ErrInvalidToken
	}

	if TokenScope(scope) != requiredScope {
		return nil, ErrInvalidToken
	}

	return &Payload{
		UserID: userID,
		Scope:  TokenScope(scope),
	}, nil
}
