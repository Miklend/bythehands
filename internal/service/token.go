package service

import (
	"crypto/rand"
	"encoding/base64"
)

func newInviteToken() (string, error) {
	// 18 random bytes => 24 chars base64url (no padding).
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
