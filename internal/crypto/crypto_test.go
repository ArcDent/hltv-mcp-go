package crypto

import (
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase-for-unit-tests")
	if err := InitKey(); err != nil {
		t.Fatal(err)
	}

	plaintext := "sk-test-api-key-1234567890abcdef"
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "" || encrypted == plaintext {
		t.Error("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	if _, err := Decrypt("!!not-valid-base64!!"); err == nil {
		t.Error("expected error for invalid base64")
	}
	if _, err := Decrypt(""); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestDecryptTooShort(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	if _, err := Decrypt("YWJj"); err == nil {
		t.Error("expected error for too-short ciphertext (abc in base64)")
	}
}

func TestEncryptWithoutInit(t *testing.T) {
	oldKey := aesKey
	aesKey = nil
	defer func() { aesKey = oldKey }()
	if _, err := Encrypt("test"); err == nil {
		t.Error("expected error when not initialized")
	}
}

func TestDecryptWithoutInit(t *testing.T) {
	oldKey := aesKey
	aesKey = nil
	defer func() { aesKey = oldKey }()
	if _, err := Decrypt("dGVzdA=="); err == nil {
		t.Error("expected error when not initialized")
	}
}

func TestDeterministicWithSameKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "test-passphrase")
	InitKey()

	plaintext := "my-secret-key"
	enc1, _ := Encrypt(plaintext)
	enc2, _ := Encrypt(plaintext)

	// Different IV each time means different ciphertext
	if enc1 == enc2 {
		t.Error("encryptions should produce different ciphertext due to random IV")
	}

	// But both should decrypt to the same plaintext
	dec1, _ := Decrypt(enc1)
	dec2, _ := Decrypt(enc2)
	if dec1 != plaintext || dec2 != plaintext {
		t.Error("both should decrypt to original plaintext")
	}
}

func TestInitKeyFilePersistence(t *testing.T) {
	// Don't write to real data/.encryption_key; the test should trigger
	// auto-generation but we can't easily test file write in unit tests
	// without filesystem isolation. Test that ENCRYPTION_KEY env works.
	t.Setenv("ENCRYPTION_KEY", "from-env")
	if err := InitKey(); err != nil {
		t.Fatal(err)
	}
	if len(aesKey) != 32 {
		t.Errorf("expected 32-byte AES key, got %d", len(aesKey))
	}
}
