package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// Hash computes an HMAC-SHA-256 digest of the input string keyed by a server-side pepper,
// and returns it as a lowercase hex-encoded string. The pepper prevents precomputation
// attacks (e.g. rainbow tables) over the small 6-digit OTP space if Redis contents leak.
func Hash(s, pepper string) (string, error) {
	// Initialize a new HMAC-SHA256 hash.Hash instance keyed by the server-side pepper
	mac := hmac.New(sha256.New, []byte(pepper))

	// Write the bytes of the input string into the digest state
	_, err := mac.Write([]byte(s))
	if err != nil {
		return "", err
	}

	// Calculate the checksum and encode the binary digest into a hex string
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Equal computes the HMAC-SHA-256 of the plain string s using the same server-side pepper
// and compares it with the existing hex-encoded hashed string using constant-time comparison
// to prevent timing attacks.
func Equal(hashed, s, pepper string) (bool, error) {
	// Re-compute the expected MAC of the raw input string
	sh, err := Hash(s, pepper)
	if err != nil {
		return false, fmt.Errorf("compute expected hash: %w", err)
	}

	// Constant-time comparison prevents side-channel timing attacks
	if subtle.ConstantTimeCompare([]byte(sh), []byte(hashed)) != 1 {
		return false, nil
	}

	return true, nil
}
