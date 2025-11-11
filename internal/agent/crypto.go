package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"

	"github.com/goccy/go-json"
)

type Encrypter struct {
	pub *rsa.PublicKey
}

type encryptedContainer struct {
	Alg string `json:"alg"`
	K   string `json:"k"`
	N   string `json:"n"`
	D   string `json:"d"`
	V   int    `json:"v"`
}

func NewEncrypterFromFile(path string) (*Encrypter, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pub, err := parsePublicKey(b)
	if err != nil {
		return nil, err
	}
	return &Encrypter{pub: pub}, nil
}

func (e *Encrypter) Encrypt(data []byte) ([]byte, map[string]string, error) {
	k := make([]byte, 32)
	_, err := rand.Read(k)
	if err != nil {
		return nil, nil, err
	}
	n := make([]byte, 12)
	_, err = rand.Read(n)
	if err != nil {
		return nil, nil, err
	}

	a, err := aes.NewCipher(k)
	if err != nil {
		return nil, nil, err
	}
	g, err := cipher.NewGCM(a)
	if err != nil {
		return nil, nil, err
	}
	c := g.Seal(nil, n, data, nil)

	ek, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.pub, k, nil)
	if err != nil {
		return nil, nil, err
	}

	cont := encryptedContainer{
		Alg: "aes256gcm+rsa-oaep",
		K:   base64.StdEncoding.EncodeToString(ek),
		N:   base64.StdEncoding.EncodeToString(n),
		D:   base64.StdEncoding.EncodeToString(c),
		V:   1,
	}
	body, err := json.Marshal(cont)
	if err != nil {
		return nil, nil, err
	}

	headers := map[string]string{
		"X-Encrypted": "aes256gcm+rsa-oaep;v=1",
	}
	return body, headers, nil
}

func parsePublicKey(b []byte) (*rsa.PublicKey, error) {
	var block *pem.Block
	rest := b
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case "CERTIFICATE":
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			switch pk := cert.PublicKey.(type) {
			case *rsa.PublicKey:
				return pk, nil
			default:
				return nil, errors.New("unsupported certificate public key type")
			}
		case "PUBLIC KEY":
			key, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			switch pk := key.(type) {
			case *rsa.PublicKey:
				return pk, nil
			default:
				return nil, errors.New("unsupported public key type")
			}
		case "RSA PUBLIC KEY":
			key, err := x509.ParsePKCS1PublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return key, nil
		}
	}
	return nil, errors.New("no valid public key found")
}
