package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type sign struct {
	key string
}

func NewSign(secret string) *sign {
	return &sign{key: secret}
}

func (sign *sign) Hash(message []byte) (hash string, err error) {
	h := hmac.New(sha256.New, []byte(sign.key))

	_, err = h.Write(message)
	if err != nil {
		return "", err
	}

	hash = hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func (sign *sign) Check(hash string, message []byte) bool {
	h := hmac.New(sha256.New, []byte(sign.key))
	_, err := h.Write(message)
	if err != nil {
		return false
	}

	expectedMAC, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	return hmac.Equal(expectedMAC, h.Sum(nil))
}
