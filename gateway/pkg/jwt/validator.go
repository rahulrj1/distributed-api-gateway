package jwt

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const algorithmRS256 = "RS256"

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

// Header represents JWT header
type Header struct {
	Alg string `json:"alg"`
}

// Claims represents JWT payload claims
type Claims struct {
	Sub      string `json:"sub"`       // User ID
	ClientID string `json:"client_id"` // For rate limiting
	Exp      int64  `json:"exp"`       // Expiration timestamp
	Iss      string `json:"iss"`       // Issuer
}

// Validator validates JWT tokens using RS256
type Validator struct {
	publicKey *rsa.PublicKey
	issuer    string // Optional: expected issuer
}

// NewValidator creates a validator from PEM-encoded public key file
func NewValidator(publicKeyPath, issuer string) (*Validator, error) {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Validator{publicKey: rsaPub, issuer: issuer}, nil
}

// Validate verifies token and returns claims. LLD §3 steps:
// 1. Extract parts, 2. Check alg=RS256, 3. Verify signature, 4. Check exp, 5. Check iss
func (v *Validator) Validate(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Verify header has RS256 algorithm
	headerBytes, err := decodeSegment(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var header Header
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Alg != algorithmRS256 {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Verify signature: decrypt(signature, public_key) should match SHA256(header.payload)
	// If someone tampered with payload, hashes won't match → rejected
	signedContent := parts[0] + "." + parts[1]
	signature, err := decodeSegment(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := verifyRS256Signature(v.publicKey, signedContent, signature); err != nil {
		return nil, ErrInvalidSignature
	}

	// Check expiration
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, ErrExpiredToken
	}

	// Check issuer if configured
	if v.issuer != "" && claims.Iss != v.issuer {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

func decodeSegment(seg string) ([]byte, error) {
	switch len(seg) % 4 { // Add padding if needed
	case 2:
		seg += "=="
	case 3:
		seg += "="
	}
	return base64.URLEncoding.DecodeString(seg)
}

func verifyRS256Signature(key *rsa.PublicKey, message string, signature []byte) error {
	hash := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature)
}
