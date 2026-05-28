package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var aesKey []byte

const keyFilePath = "data/.encryption_key"

// InitKey loads or generates the encryption passphrase, derives a 32-byte AES key,
// and stores it in the package-level aesKey variable. Must be called once at startup.
func InitKey() error {
	// 1. ENCRYPTION_KEY env var
	if key := os.Getenv("ENCRYPTION_KEY"); key != "" {
		h := sha256.Sum256([]byte(key))
		aesKey = h[:]
		return nil
	}
	// 2. data/.encryption_key file
	if data, err := os.ReadFile(keyFilePath); err == nil {
		h := sha256.Sum256(data)
		aesKey = h[:]
		return nil
	}
	// 3. Auto-generate
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}
	passphrase := hex.EncodeToString(randomBytes)
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(keyFilePath, []byte(passphrase), 0600); err != nil {
		return fmt.Errorf("write .encryption_key: %w", err)
	}
	h := sha256.Sum256([]byte(passphrase))
	aesKey = h[:]
	return nil
}

// Encrypt encrypts plaintext with AES-256-GCM and returns base64(iv + ciphertext + tag).
func Encrypt(plaintext string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("crypto not initialized: call InitKey first")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)
	result := append(iv, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts a base64(iv + ciphertext + tag) string with AES-256-GCM.
func Decrypt(encoded string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("crypto not initialized: call InitKey first")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:nonceSize]
	ct := data[nonceSize:]
	plaintext, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}
