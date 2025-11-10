package agent

import (
	"bytes"
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/goccy/go-json"
)

func TestEncrypterFromCert_Success(t *testing.T) {
	path := filepath.Join("..", "..", "var", "keys", "cert.pem")
	_, err := NewEncrypterFromFile(path)
	if err != nil {
		t.Fatalf("NewEncrypterFromFile error: %v", err)
	}
}

func TestEncrypterFromFile_NotFound_Error(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.pem")
	_, err := NewEncrypterFromFile(missing)
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestEncrypter_Encrypt_ReturnsHeadersAndBody(t *testing.T) {
	path := filepath.Join("..", "..", "var", "keys", "cert.pem")
	e, err := NewEncrypterFromFile(path)
	if err != nil {
		t.Fatalf("NewEncrypterFromFile error: %v", err)
	}

	data := []byte("hello")
	body, headers, err := e.Encrypt(data)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("empty encrypted body")
	}

	if headers["X-Encrypted"] != "aes256gcm+rsa-oaep;v=1" {
		t.Fatalf("missing or invalid X-Encrypted header")
	}

	var cont encryptedContainer
	if err := json.Unmarshal(body, &cont); err != nil {
		t.Fatalf("container unmarshal error: %v", err)
	}
	if cont.Alg != "aes256gcm+rsa-oaep" {
		t.Fatalf("unexpected alg: %s", cont.Alg)
	}
	if cont.V != 1 {
		t.Fatalf("unexpected version: %d", cont.V)
	}
	if cont.K == "" || cont.N == "" || cont.D == "" {
		t.Fatalf("container fields must not be empty")
	}

	nonce, err := base64.StdEncoding.DecodeString(cont.N)
	if err != nil {
		t.Fatalf("nonce base64 decode error: %v", err)
	}
	if len(nonce) != 12 {
		t.Fatalf("unexpected nonce length: %d", len(nonce))
	}

	ek, err := base64.StdEncoding.DecodeString(cont.K)
	if err != nil {
		t.Fatalf("encrypted key base64 decode error: %v", err)
	}
	if len(ek) == 0 {
		t.Fatalf("empty encrypted key")
	}

	ct, err := base64.StdEncoding.DecodeString(cont.D)
	if err != nil {
		t.Fatalf("ciphertext base64 decode error: %v", err)
	}
	if len(ct) == 0 {
		t.Fatalf("empty ciphertext")
	}
	if bytes.Equal(ct, data) {
		t.Fatalf("ciphertext must differ from plaintext")
	}
}
