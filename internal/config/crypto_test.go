package config

import (
	"path/filepath"
	"testing"
)

func newTestCrypto(t *testing.T) *TokenCryptoService {
	t.Helper()
	dir := t.TempDir()
	return NewTokenCryptoService(
		WithKeyFilePath(filepath.Join(dir, "master.key")),
	)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	cs := newTestCrypto(t)

	plaintext := "xoxb-very-secret-token"
	encrypted, err := cs.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted should differ from plaintext")
	}

	if !cs.IsCurrentFormat(encrypted) {
		t.Error("encrypted should be in current format")
	}

	decrypted, err := cs.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	cs := newTestCrypto(t)

	encrypted, err := cs.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}

	decrypted, err := cs.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	cs := newTestCrypto(t)
	plaintext := "test-token"

	enc1, _ := cs.Encrypt(plaintext)
	enc2, _ := cs.Encrypt(plaintext)

	if enc1 == enc2 {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random IV)")
	}
}

func TestIsEncrypted(t *testing.T) {
	cs := newTestCrypto(t)

	encrypted, _ := cs.Encrypt("test")
	if !cs.IsEncrypted(encrypted) {
		t.Error("expected IsEncrypted to return true for encrypted value")
	}

	if cs.IsEncrypted("plaintext-token") {
		t.Error("expected IsEncrypted to return false for plaintext")
	}

	if cs.IsEncrypted("") {
		t.Error("expected IsEncrypted to return false for empty string")
	}
}

func TestIsCurrentFormat(t *testing.T) {
	cs := newTestCrypto(t)

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty", "", false},
		{"plaintext", "xoxb-token", false},
		{"invalid parts", "v2:ab:cd", false},
		{"valid v2", "v2:" + "aabbccddaabbccddaabbccdd" + ":" + "aabb" + ":" + "aabbccddaabbccddaabbccddaabbccdd", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cs.IsCurrentFormat(tt.value); got != tt.expected {
				t.Errorf("IsCurrentFormat(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestDecrypt_InvalidFormat(t *testing.T) {
	cs := newTestCrypto(t)

	_, err := cs.Decrypt("not-encrypted")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestDecrypt_EmptyString(t *testing.T) {
	cs := newTestCrypto(t)

	_, err := cs.Decrypt("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestWithMasterKey(t *testing.T) {
	cs := NewTokenCryptoService(WithMasterKey("my-secret-key"))

	encrypted, err := cs.Encrypt("test-token")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	decrypted, err := cs.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted != "test-token" {
		t.Errorf("expected test-token, got %s", decrypted)
	}
}

func TestDifferentMasterKeys_CannotDecrypt(t *testing.T) {
	cs1 := NewTokenCryptoService(WithMasterKey("key-one"))
	cs2 := NewTokenCryptoService(WithMasterKey("key-two"))

	encrypted, _ := cs1.Encrypt("secret")
	_, err := cs2.Decrypt(encrypted)
	if err == nil {
		t.Error("expected decrypt to fail with different master key")
	}
}
