//go:build ignore
// +build ignore

// Usage: go run scripts/generate_jwt.go -sub user123 -client_id app1 -exp 1h
// Generates a JWT signed with keys/private.pem

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	sub := flag.String("sub", "user123", "Subject (user ID)")
	clientID := flag.String("client_id", "default", "Client ID for rate limiting")
	expDuration := flag.Duration("exp", time.Hour, "Token expiration duration")
	keyPath := flag.String("key", "keys/private.pem", "Path to private key")
	flag.Parse()

	privateKey, err := loadPrivateKey(*keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading private key: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nGenerate keys with:\n")
		fmt.Fprintf(os.Stderr, "  mkdir -p keys\n")
		fmt.Fprintf(os.Stderr, "  openssl genrsa -out keys/private.pem 2048\n")
		fmt.Fprintf(os.Stderr, "  openssl rsa -in keys/private.pem -pubout -out keys/public.pem\n")
		os.Exit(1)
	}

	token, err := generateToken(privateKey, *sub, *clientID, *expDuration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	// Try PKCS1 first, then PKCS8
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return key, nil
}

func generateToken(key *rsa.PrivateKey, sub, clientID string, exp time.Duration) (string, error) {
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"sub":       sub,
		"client_id": clientID,
		"exp":       time.Now().Add(exp).Unix(),
		"iat":       time.Now().Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	message := headerB64 + "." + payloadB64

	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return message + "." + signatureB64, nil
}
