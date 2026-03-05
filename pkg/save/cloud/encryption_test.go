package cloud

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		password string
	}{
		{
			name:     "empty data",
			data:     []byte{},
			password: "test-password",
		},
		{
			name:     "small data",
			data:     []byte("hello world"),
			password: "my-secret-password-123",
		},
		{
			name:     "large data",
			data:     make([]byte, 1024*1024), // 1MB
			password: "strong-password",
		},
		{
			name:     "special characters in password",
			data:     []byte("test data"),
			password: "p@ssw0rd!#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.data) > 1000 {
				rand.Read(tt.data)
			}

			encrypted, err := Encrypt(tt.data, tt.password)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if len(encrypted) < SaltSize+NonceSize+16 {
				t.Fatalf("encrypted data too small: got %d bytes", len(encrypted))
			}

			decrypted, err := Decrypt(encrypted, tt.password)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("decrypted data doesn't match original")
			}
		})
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	data := []byte("secret message")
	password := "correct-password"

	encrypted, err := Encrypt(data, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, "wrong-password")
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got: %v", err)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    error
	}{
		{
			name:       "too short",
			ciphertext: make([]byte, 10),
			wantErr:    ErrInvalidCiphertext,
		},
		{
			name:       "empty",
			ciphertext: []byte{},
			wantErr:    ErrInvalidCiphertext,
		},
		{
			name:       "minimum size but invalid",
			ciphertext: make([]byte, SaltSize+NonceSize+16),
			wantErr:    ErrDecryptionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.ciphertext, "password")
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEncryptionUniqueness(t *testing.T) {
	data := []byte("same data")
	password := "same password"

	encrypted1, err := Encrypt(data, password)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}

	encrypted2, err := Encrypt(data, password)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("identical inputs produced identical ciphertext (nonce reuse)")
	}

	decrypted1, _ := Decrypt(encrypted1, password)
	decrypted2, _ := Decrypt(encrypted2, password)

	if !bytes.Equal(decrypted1, data) || !bytes.Equal(decrypted2, data) {
		t.Error("decryption failed for unique ciphertexts")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "test-password"
	salt1 := make([]byte, SaltSize)
	salt2 := make([]byte, SaltSize)
	rand.Read(salt1)
	rand.Read(salt2)

	key1 := deriveKey(password, salt1)
	key2 := deriveKey(password, salt1)
	key3 := deriveKey(password, salt2)

	if len(key1) != KeySize {
		t.Errorf("key size mismatch: got %d, want %d", len(key1), KeySize)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("same password and salt produced different keys")
	}

	if bytes.Equal(key1, key3) {
		t.Error("different salts produced identical keys")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	data := make([]byte, 1024*100) // 100KB
	rand.Read(data)
	password := "benchmark-password"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(data, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	data := make([]byte, 1024*100) // 100KB
	rand.Read(data)
	password := "benchmark-password"

	encrypted, err := Encrypt(data, password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Decrypt(encrypted, password)
		if err != nil {
			b.Fatal(err)
		}
	}
}
