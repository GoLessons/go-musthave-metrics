package signature

import (
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		message []byte
		wantErr bool
	}{
		{
			name:    "Basic hash test",
			key:     "test-key",
			message: []byte("test message"),
			wantErr: false,
		},
		{
			name:    "Empty message test",
			key:     "test-key",
			message: []byte(""),
			wantErr: false,
		},
		{
			name:    "Empty key test",
			key:     "",
			message: []byte("test message"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSign(tt.key)
			hash, err := s.Hash(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if hash == "" {
				t.Errorf("Hash() returned empty hash")
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		message []byte
		hash    string
		valid   bool
	}{
		{
			name:    "Valid signature",
			key:     "test-key",
			message: []byte("test message"),
			hash:    "1af95ac1739579a889566e9bdddfcd4a5cec55085d7c715538fcddf0b1a6153f",
			valid:   true,
		},
		{
			name:    "Invalid signature",
			key:     "test-key",
			message: []byte("test message"),
			hash:    "0ea12a8295fd235412e787b5938b4902ea2187b78069cd5f272125e4944b0c7b",
			valid:   false,
		},
		{
			name:    "Invalid hex in signature",
			key:     "test-key",
			message: []byte("test message"),
			hash:    "invalid-hex-string",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSign(tt.key)
			result := s.Check(tt.hash, tt.message)

			if result != tt.valid {
				t.Errorf("Check() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestConsistency(t *testing.T) {
	key := "secret-key"
	message := []byte("important data")

	s := NewSign(key)
	hash1, _ := s.Hash(message)
	hash2, _ := s.Hash(message)

	if hash1 != hash2 {
		t.Errorf("Hash() is not consistent: %s != %s", hash1, hash2)
	}
}
