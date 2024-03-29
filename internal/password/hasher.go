package password

// Hasher is a password hashing interface
type Hasher interface {
	// Hash returns a hashed version of the plaintext password or an error if fails
	Hash(plaintextPassword string) (string, error)

	// Matches returns true if the plaintext password matches the hashed password or an error if fails
	Matches(plaintextPassword, hashedPassword string) (bool, error)
}
