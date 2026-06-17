package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(s string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(s))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
