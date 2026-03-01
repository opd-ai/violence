package chat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// KeyExchange performs ECDH key exchange over a network connection
// and derives an AES-256-GCM key for encrypted chat communication.
//
// Protocol:
// 1. Generate ephemeral ECDH P-256 keypair
// 2. Send public key to peer (65 bytes)
// 3. Receive peer's public key
// 4. Compute shared secret via ECDH
// 5. Derive AES-256 key using HKDF-SHA3-256
//
// Returns 32-byte AES key on success.
func PerformKeyExchange(conn net.Conn) ([]byte, error) {
	if conn == nil {
		return nil, errors.New("connection is nil")
	}

	// Generate ephemeral ECDH keypair using P-256 curve
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Get our public key bytes
	ourPublicKey := privateKey.PublicKey().Bytes()

	// Send our public key to peer
	if err := sendPublicKey(conn, ourPublicKey); err != nil {
		return nil, err
	}

	// Receive peer's public key
	peerPublicKeyBytes, err := receivePublicKey(conn)
	if err != nil {
		return nil, err
	}

	// Parse peer's public key
	peerPublicKey, err := curve.NewPublicKey(peerPublicKeyBytes)
	if err != nil {
		return nil, err
	}

	// Perform ECDH to get shared secret
	sharedSecret, err := privateKey.ECDH(peerPublicKey)
	if err != nil {
		return nil, err
	}

	// Derive AES-256 key from shared secret using HKDF
	aesKey := deriveKey(sharedSecret, 32)

	return aesKey, nil
}

// sendPublicKey sends a 65-byte P-256 public key over the connection
func sendPublicKey(conn net.Conn, publicKey []byte) error {
	// Send length prefix (2 bytes, big-endian)
	lengthBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBuf, uint16(len(publicKey)))
	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}

	// Send public key bytes
	if _, err := conn.Write(publicKey); err != nil {
		return err
	}

	return nil
}

// receivePublicKey receives a public key from the connection
func receivePublicKey(conn net.Conn) ([]byte, error) {
	// Read length prefix
	lengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint16(lengthBuf)
	if length == 0 || length > 1024 { // Sanity check
		return nil, errors.New("invalid public key length")
	}

	// Read public key bytes
	publicKey := make([]byte, length)
	if _, err := io.ReadFull(conn, publicKey); err != nil {
		return nil, err
	}

	return publicKey, nil
}

// deriveKey derives an encryption key from shared secret using HKDF-SHA3-256
func deriveKey(sharedSecret []byte, keyLen int) []byte {
	// HKDF with SHA3-256
	// Salt: nil (not required for ECDH as shared secret is already random)
	// Info: context string to bind key to application
	info := []byte("violence-chat-encryption-v1")
	hkdf := hkdf.New(sha3.New256, sharedSecret, nil, info)

	key := make([]byte, keyLen)
	if _, err := io.ReadFull(hkdf, key); err != nil {
		panic(err) // Should never fail with valid inputs
	}

	return key
}

// EncryptMessage encrypts a plaintext message using AES-256-GCM
func EncryptMessage(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptMessage decrypts a ciphertext message using AES-256-GCM
func DecryptMessage(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:nonceSize]
	ciphertextData := ciphertext[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
