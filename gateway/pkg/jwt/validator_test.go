package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"testing"
	"time"
)

func TestValidator(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKeyPath := createTempPublicKey(t, privateKey)
	defer os.Remove(publicKeyPath)

	validator, err := NewValidator(publicKeyPath, "")
	if err != nil {
		t.Fatalf("NewValidator failed: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		token := createToken(t, privateKey, "user123", "client1", "", time.Hour)
		claims, err := validator.Validate(token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if claims.Sub != "user123" {
			t.Errorf("Expected sub=user123, got %s", claims.Sub)
		}
		if claims.ClientID != "client1" {
			t.Errorf("Expected client_id=client1, got %s", claims.ClientID)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := createToken(t, privateKey, "user123", "client1", "", -time.Hour)
		_, err := validator.Validate(token)
		if err != ErrExpiredToken {
			t.Errorf("Expected ErrExpiredToken, got %v", err)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		token := createToken(t, otherKey, "user123", "client1", "", time.Hour)
		_, err := validator.Validate(token)
		if err != ErrInvalidSignature {
			t.Errorf("Expected ErrInvalidSignature, got %v", err)
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := validator.Validate("not.a.valid.token.format")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := validator.Validate("")
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestValidatorWithIssuer(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKeyPath := createTempPublicKey(t, privateKey)
	defer os.Remove(publicKeyPath)

	validator, _ := NewValidator(publicKeyPath, "expected-issuer")

	t.Run("correct issuer", func(t *testing.T) {
		token := createToken(t, privateKey, "user123", "client1", "expected-issuer", time.Hour)
		_, err := validator.Validate(token)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		token := createToken(t, privateKey, "user123", "client1", "wrong-issuer", time.Hour)
		_, err := validator.Validate(token)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for wrong issuer, got %v", err)
		}
	})
}

// --- Helpers ---

func createTempPublicKey(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	pubASN1, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})
	f, _ := os.CreateTemp("", "public-*.pem")
	f.Write(pubPEM)
	f.Close()
	return f.Name()
}

func createToken(t *testing.T, key *rsa.PrivateKey, sub, clientID, issuer string, exp time.Duration) string {
	t.Helper()

	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"sub":       sub,
		"client_id": clientID,
		"exp":       time.Now().Add(exp).Unix(),
	}
	if issuer != "" {
		payload["iss"] = issuer
	}

	hJSON, _ := json.Marshal(header)
	pJSON, _ := json.Marshal(payload)

	hB64 := base64.RawURLEncoding.EncodeToString(hJSON)
	pB64 := base64.RawURLEncoding.EncodeToString(pJSON)

	msg := hB64 + "." + pB64
	hash := sha256.Sum256([]byte(msg))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return msg + "." + sigB64
}
