package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Signer struct {
	key string
}

func NewSign(secret string) *Signer {
	return &Signer{key: secret}
}

func (sign *Signer) Hash(message []byte) (hash string, err error) {
	h := hmac.New(sha256.New, []byte(sign.key))

	_, err = h.Write(message)
	if err != nil {
		return "", err
	}

	hash = hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func (sign *Signer) Check(hash string, message []byte) bool {
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
