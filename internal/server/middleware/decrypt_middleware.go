package middleware

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"

	"encoding/pem"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type Decrypter struct {
	priv *rsa.PrivateKey
}

type encryptedContainer struct {
	Alg string `json:"alg"`
	K   string `json:"k"`
	N   string `json:"n"`
	D   string `json:"d"`
	V   int    `json:"v"`
}

func NewDecrypterFromFile(path string) (*Decrypter, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	priv, err := parsePrivateKey(b)
	if err != nil {
		return nil, err
	}
	return &Decrypter{priv: priv}, nil
}

func parsePrivateKey(b []byte) (*rsa.PrivateKey, error) {
	var block *pem.Block
	rest := b
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case "RSA PRIVATE KEY":
			key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			return key, nil
		case "PRIVATE KEY":
			keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
			switch k := keyAny.(type) {
			case *rsa.PrivateKey:
				return k, nil
			default:
				return nil, errors.New("unsupported private key type")
			}
		}
	}
	return nil, errors.New("no valid private key found")
}

type DecryptMiddleware struct {
	decrypter *Decrypter
	logger    *zap.Logger
}

func NewDecryptMiddleware(decrypter *Decrypter, logger *zap.Logger) *DecryptMiddleware {
	return &DecryptMiddleware{decrypter: decrypter, logger: logger}
}

func (m *DecryptMiddleware) DecryptBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := r.Header.Get("X-Encrypted")
		if enc == "" {
			next.ServeHTTP(w, r)
			return
		}
		if m.decrypter == nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if enc != "aes256gcm+rsa-oaep;v=1" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var cont encryptedContainer
		if err := json.Unmarshal(body, &cont); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if cont.Alg != "aes256gcm+rsa-oaep" || cont.V != 1 || cont.K == "" || cont.N == "" || cont.D == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		ek, err := base64.StdEncoding.DecodeString(cont.K)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		nonce, err := base64.StdEncoding.DecodeString(cont.N)
		if err != nil || len(nonce) != 12 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		ct, err := base64.StdEncoding.DecodeString(cont.D)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		k, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.decrypter.priv, ek, nil)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		a, err := aes.NewCipher(k)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		g, err := cipher.NewGCM(a)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		pt, err := g.Open(nil, nonce, ct, nil)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(pt))
		next.ServeHTTP(w, r)
	})
}
