package authutils_test

import (
	"os"
	"testing"
	"time"

	"github.com/gurch101/gowebutils/pkg/authutils"
)

func TestMain(m *testing.M) {
	// Set test encryption key
	err := os.Setenv("ENCRYPTION_KEY", "0123456789ABCDEF0123456789ABCDEF")

	if err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "valid data",
			data: map[string]any{
				"key1": "value1",
				"key2": "42",
			},
			wantErr: false,
		},
		{
			name:    "empty map",
			data:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "nil map",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := authutils.Encrypt(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			decrypted, err := authutils.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)

				return
			}

			// Compare original and decrypted data
			for k, v := range tt.data {
				if decrypted[k] != v {
					t.Errorf("Decrypt() got = %v %T, want %v %T for key %s", decrypted[k], decrypted[k], v, v, k)
				}
			}
		})
	}
}

func TestDecrypt_Invalid(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "invalid-base64!@#$",
			wantErr:    true,
		},
		{
			name:       "empty string",
			ciphertext: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authutils.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInviteToken(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: map[string]any{
				"user_id": "123",
				"email":   "test@example.com",
			},
			wantErr: false,
		},
		{
			name:    "empty payload",
			payload: map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authutils.CreateInviteToken(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateInviteToken() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			// Verify token
			decoded, err := authutils.VerifyInviteToken(token)
			if err != nil {
				t.Errorf("VerifyInviteToken() error = %v", err)

				return
			}

			// Check payload values
			for k, v := range tt.payload {
				if decoded[k] != v {
					t.Errorf("VerifyInviteToken() got = %v, want %v for key %s", decoded[k], v, k)
				}
			}
		})
	}
}

func TestInviteToken_Expiration(t *testing.T) {
	payload := map[string]any{
		"user_id": 123,
	}

	token, err := authutils.CreateInviteToken(payload)
	if err != nil {
		t.Fatalf("CreateInviteToken() error = %v", err)
	}

	// Verify valid token
	_, err = authutils.VerifyInviteToken(token)
	if err != nil {
		t.Errorf("VerifyInviteToken() error = %v", err)
	}

	// Modify expiration time to test expired token
	decoded, err := authutils.Decrypt(token)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	decoded["expires_at"] = time.Now().UTC().Add(-time.Hour)

	expired, err := authutils.Encrypt(decoded)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Verify expired token
	_, err = authutils.VerifyInviteToken(expired)
	if err == nil {
		t.Error("VerifyInviteToken() expected error for expired token")
	}
}
