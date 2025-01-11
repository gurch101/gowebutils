package authutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gurch101/gowebutils/pkg/parser"
)

var ErrEncryption = errors.New("failed to encrypt data")

var ErrDecryption = errors.New("failed to decrypt data")

var ErrInviteExpiresAt = errors.New("invite token has expired")

var ErrInvalidPayload = errors.New("invalid payload")

const inviteTokenExpiresAt = time.Hour * 24 * 7

// Encrypt a map[string]any.
func Encrypt(data map[string]any) (string, error) {
	if data == nil {
		return "", ErrInvalidPayload
	}

	key := []byte(parser.ParseEnvStringPanic("ENCRYPTION_KEY"))

	// Serialize the map to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("%w: json marshal failed: %w", ErrEncryption, err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("%w: aes new cipher failed: %w", ErrEncryption, err)
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: gcm new failed: %w", ErrEncryption, err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: rand reader failed: %w", ErrEncryption, err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt a string that was encrypted with Encrypt.
func Decrypt(ciphertext string) (map[string]any, error) {
	key := []byte(parser.ParseEnvStringPanic("ENCRYPTION_KEY"))

	ciphertextBytes, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %w", ErrDecryption, err)
	}
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: aes new cipher failed: %w", ErrDecryption, err)
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: gcm new failed: %w", ErrDecryption, err)
	}

	// Extract the nonce from the ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return nil, fmt.Errorf("%w: ciphertext too short: %w", ErrDecryption, err)
	}

	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: gcm open failed: %w", ErrDecryption, err)
	}

	// Deserialize the JSON back into a map
	var data map[string]any
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("%w: json unmarshal failed: %w", ErrDecryption, err)
	}

	return data, nil
}

func CreateInviteToken(payload map[string]any) (string, error) {
	inviteTokenPayload := make(map[string]any)

	for key, value := range payload {
		inviteTokenPayload[key] = value
	}

	inviteTokenPayload["expires_at"] = time.Now().UTC().Add(inviteTokenExpiresAt)

	encrypted, err := Encrypt(inviteTokenPayload)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt invite token: %w", err)
	}

	return encrypted, nil
}

func VerifyInviteToken(token string) (map[string]any, error) {
	decrypted, err := Decrypt(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt invite token: %w", err)
	}

	expiresAt, ok := decrypted["expires_at"].(string)
	if !ok {
		return nil, fmt.Errorf("%w, invalid expires_at in invite token", ErrInviteExpiresAt)
	}

	parsedTime, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at in invite token: %w", err)
	}

	if parsedTime.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: invite token expired", ErrInviteExpiresAt)
	}

	return decrypted, nil
}
