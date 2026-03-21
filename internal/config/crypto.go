package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	cryptoAlgorithm      = "aes-256-gcm"
	keyLength            = 32
	ivLength             = 12
	legacyIVLength       = 16
	authTagLength        = 16
	separator            = ":"
	versionPrefix        = "v2"
	masterKeySalt        = "708u-slack-cli-master-key-salt-v2"
	masterKeyIterations  = 100000
	legacySalt           = "708u-slack-cli-salt-v1"
	legacyPassword       = "708u-slack-cli-key"
	legacyKeyIterations  = 100000
	envMasterKeyName     = "SLACK_CLI_MASTER_KEY"
	defaultSecretsDir    = ".slack-cli-secrets"
	defaultConfigDirName = ".slack-cli"
	masterKeyFileName    = "master.key"
)

var hexPattern = regexp.MustCompile(`^[0-9a-fA-F]+$`)

// CryptoOption configures a TokenCryptoService.
type CryptoOption func(*TokenCryptoService)

// WithMasterKey injects a master key directly.
func WithMasterKey(key string) CryptoOption {
	return func(s *TokenCryptoService) {
		s.injectedMasterKey = key
	}
}

// WithKeyFilePath overrides the default key file path.
func WithKeyFilePath(path string) CryptoOption {
	return func(s *TokenCryptoService) {
		s.keyFilePath = path
	}
}

// WithLegacyKeyFilePath overrides the default legacy key file path.
func WithLegacyKeyFilePath(path string) CryptoOption {
	return func(s *TokenCryptoService) {
		s.legacyKeyFilePath = path
	}
}

// TokenCryptoService handles encryption and decryption of tokens using
// AES-256-GCM (v2) with legacy AES-256-CBC decryption support.
type TokenCryptoService struct {
	keyFilePath       string
	legacyKeyFilePath string
	injectedMasterKey string
	cachedMasterKey   []byte
	mu                sync.Mutex
}

// NewTokenCryptoService creates a new crypto service with the given options.
func NewTokenCryptoService(opts ...CryptoOption) *TokenCryptoService {
	s := &TokenCryptoService{}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *TokenCryptoService) getKeyFilePath() string {
	if s.keyFilePath != "" {
		return s.keyFilePath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, defaultSecretsDir, masterKeyFileName)
}

func (s *TokenCryptoService) getLegacyKeyFilePath() string {
	if s.legacyKeyFilePath != "" {
		return s.legacyKeyFilePath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, defaultConfigDirName, masterKeyFileName)
}

func deriveLegacyKey() ([]byte, error) {
	return pbkdf2.Key(sha256.New, legacyPassword, []byte(legacySalt), legacyKeyIterations, keyLength)
}

func deriveMasterKey(secret string) ([]byte, error) {
	return pbkdf2.Key(sha256.New, secret, []byte(masterKeySalt), masterKeyIterations, keyLength)
}

func parseFileKey(contents string) ([]byte, error) {
	keyHex := strings.TrimSpace(contents)
	if matched, _ := regexp.MatchString(`^[0-9a-fA-F]{64}$`, keyHex); !matched {
		return nil, &ConfigurationError{Msg: "Invalid token encryption key format"}
	}
	return hex.DecodeString(keyHex)
}

func readKeyFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return parseFileKey(string(data))
}

func writeKeyFile(keyFilePath, keyHex string) ([]byte, error) {
	keyDir := filepath.Dir(keyFilePath)
	if err := os.MkdirAll(keyDir, dirPermission); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, err
	}

	// Use O_CREATE|O_EXCL to fail if file already exists (atomic creation).
	f, err := os.OpenFile(keyFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePermission)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.WriteString(keyHex + "\n"); err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func (s *TokenCryptoService) createKeyFile() ([]byte, error) {
	keyFilePath := s.getKeyFilePath()
	randBytes := make([]byte, keyLength)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, &ConfigurationError{Msg: "Failed to generate random key"}
	}
	keyHex := hex.EncodeToString(randBytes)
	key, err := writeKeyFile(keyFilePath, keyHex)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return readKeyFromFile(keyFilePath)
		}
		return nil, &ConfigurationError{Msg: "Failed to initialize token encryption key"}
	}
	return key, nil
}

func (s *TokenCryptoService) migrateLegacyKeyFile() ([]byte, error) {
	legacyPath := s.getLegacyKeyFilePath()
	legacyKey, err := readKeyFromFile(legacyPath)
	if err != nil {
		return nil, err
	}
	legacyKeyHex := hex.EncodeToString(legacyKey)
	keyFilePath := s.getKeyFilePath()
	if _, err := writeKeyFile(keyFilePath, legacyKeyHex); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return legacyKey, nil
		}
		if _, statErr := os.Stat(keyFilePath); statErr != nil {
			return legacyKey, nil
		}
	}
	return readKeyFromFile(keyFilePath)
}

func (s *TokenCryptoService) getMasterKey() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cachedMasterKey != nil {
		return s.cachedMasterKey, nil
	}

	// 1. Injected key
	if s.injectedMasterKey != "" {
		key, err := deriveMasterKey(s.injectedMasterKey)
		if err != nil {
			return nil, &ConfigurationError{Msg: "Failed to derive master key"}
		}
		s.cachedMasterKey = key
		return key, nil
	}

	// 2. Environment variable
	if envKey := strings.TrimSpace(os.Getenv(envMasterKeyName)); envKey != "" {
		key, err := deriveMasterKey(envKey)
		if err != nil {
			return nil, &ConfigurationError{Msg: "Failed to derive master key from environment"}
		}
		s.cachedMasterKey = key
		return key, nil
	}

	// 3. Key file
	key, err := readKeyFromFile(s.getKeyFilePath())
	if err == nil {
		s.cachedMasterKey = key
		return key, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		var cfgErr *ConfigurationError
		if errors.As(err, &cfgErr) {
			return nil, cfgErr
		}
		return nil, &ConfigurationError{Msg: "Failed to load token encryption key"}
	}

	// 4. Try legacy key file migration
	key, err = s.migrateLegacyKeyFile()
	if err == nil {
		s.cachedMasterKey = key
		return key, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		var cfgErr *ConfigurationError
		if errors.As(err, &cfgErr) {
			return nil, cfgErr
		}
		return nil, &ConfigurationError{Msg: "Failed to migrate token encryption key"}
	}

	// 5. Auto-generate
	key, err = s.createKeyFile()
	if err != nil {
		return nil, err
	}
	s.cachedMasterKey = key
	return key, nil
}

// Encrypt encrypts a token using AES-256-GCM and returns the v2 format string.
// Format: "v2:iv_hex:cipher_hex:authtag_hex"
func (s *TokenCryptoService) Encrypt(token string) (string, error) {
	key, err := s.getMasterKey()
	if err != nil {
		return "", err
	}

	iv := make([]byte, ivLength)
	if _, err := rand.Read(iv); err != nil {
		return "", &ConfigurationError{Msg: "Failed to encrypt token"}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to encrypt token"}
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, ivLength)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to encrypt token"}
	}

	// GCM Seal appends the auth tag to the ciphertext.
	sealed := aesGCM.Seal(nil, iv, []byte(token), nil)

	// Split sealed into ciphertext and auth tag.
	ciphertext := sealed[:len(sealed)-authTagLength]
	authTag := sealed[len(sealed)-authTagLength:]

	return strings.Join([]string{
		versionPrefix,
		hex.EncodeToString(iv),
		hex.EncodeToString(ciphertext),
		hex.EncodeToString(authTag),
	}, separator), nil
}

// Decrypt decrypts a v2 or legacy encrypted token string.
func (s *TokenCryptoService) Decrypt(data string) (string, error) {
	if data == "" {
		return "", &ValidationError{Msg: "Invalid encrypted data format"}
	}

	if s.IsCurrentFormat(data) {
		return s.decryptCurrentFormat(data)
	}

	if s.isLegacyEncrypted(data) {
		return s.decryptLegacyFormat(data)
	}

	return "", &ValidationError{Msg: "Invalid encrypted data format"}
}

// IsEncrypted checks if a value looks like an encrypted token (v2 or legacy).
func (s *TokenCryptoService) IsEncrypted(value string) bool {
	return s.IsCurrentFormat(value) || s.isLegacyEncrypted(value)
}

// IsCurrentFormat checks if the value matches the v2 encryption format.
func (s *TokenCryptoService) IsCurrentFormat(value string) bool {
	if value == "" {
		return false
	}
	parts := strings.Split(value, separator)
	if len(parts) != 4 || parts[0] != versionPrefix {
		return false
	}
	ivHex := parts[1]
	cipherHex := parts[2]
	authTagHex := parts[3]

	if !hexPattern.MatchString(ivHex) || len(ivHex) != ivLength*2 {
		return false
	}
	if cipherHex != "" && !hexPattern.MatchString(cipherHex) {
		return false
	}
	if len(cipherHex)%2 != 0 {
		return false
	}
	if !hexPattern.MatchString(authTagHex) || len(authTagHex) != authTagLength*2 {
		return false
	}
	return true
}

func (s *TokenCryptoService) isLegacyEncrypted(value string) bool {
	if value == "" {
		return false
	}
	parts := strings.Split(value, separator)
	if len(parts) != 2 {
		return false
	}
	ivHex := parts[0]
	cipherHex := parts[1]

	return hexPattern.MatchString(ivHex) &&
		len(ivHex) == legacyIVLength*2 &&
		hexPattern.MatchString(cipherHex) &&
		len(cipherHex) > 0 &&
		len(cipherHex)%2 == 0
}

func (s *TokenCryptoService) decryptCurrentFormat(data string) (string, error) {
	if !s.IsCurrentFormat(data) {
		return "", &ValidationError{Msg: "Invalid encrypted data format"}
	}
	parts := strings.Split(data, separator)
	iv, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	ciphertext, err := hex.DecodeString(parts[2])
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	authTag, err := hex.DecodeString(parts[3])
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}

	key, err := s.getMasterKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	aesGCM, err := cipher.NewGCMWithNonceSize(block, ivLength)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}

	// Go's GCM expects ciphertext + authTag concatenated.
	sealed := append(ciphertext, authTag...)
	plaintext, err := aesGCM.Open(nil, iv, sealed, nil)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	return string(plaintext), nil
}

func (s *TokenCryptoService) decryptLegacyFormat(data string) (string, error) {
	if !s.isLegacyEncrypted(data) {
		return "", &ValidationError{Msg: "Invalid encrypted data format"}
	}
	parts := strings.Split(data, separator)
	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}

	key, err := deriveLegacyKey()
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to derive legacy key"}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding.
	if len(plaintext) == 0 {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	padLen := int(plaintext[len(plaintext)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(plaintext) {
		return "", &ConfigurationError{Msg: "Failed to decrypt token"}
	}
	for i := range padLen {
		if plaintext[len(plaintext)-1-i] != byte(padLen) {
			return "", &ConfigurationError{Msg: "Failed to decrypt token"}
		}
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	return string(plaintext), nil
}
