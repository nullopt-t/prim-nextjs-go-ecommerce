package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// Hash computes the SHA-256 cryptographic digest of an input string and returns it as a lowercase hex-encoded string.
func Hash(s string) (string, error) {
	// Initialize a new SHA-256 hash.Hash instance
	hash := sha256.New()

	// Write the bytes of the input string into the digest state
	_, err := hash.Write([]byte(s))
	if err != nil {
		return "", err
	}

	// Calculate the checksum and encode the binary digest into a hex string
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Equal computes the SHA-256 hash of the plain string s and compares it with the existing hex-encoded hashed string
// using constant-time comparison to prevent timing attacks.
func Equal(hashed, s string) (bool, error) {
	// Re-hash the raw input string
	sh, err := Hash(s)
	if err != nil {
		return false, err
	}

	// Constant-time comparison prevents side-channel timing attacks
	if subtle.ConstantTimeCompare([]byte(sh), []byte(hashed)) != 1 {
		return false, nil
	}

	return true, nil
}

