package auth

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/puoklam/chat-app-backend/env"
)

var (
	ed25519Secret any
)

func init() {
	data, _ := os.ReadFile(env.ED25519_PRIV_KEY_PATH)
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		os.Exit(1)
	}
	ed25519Secret = key
}

func genIdToken(user any) (string, error) {
	// RS256 for asymmetric signature, private key to sign in server, public key to verify in client
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"kid": "keyid",
		"jku": "http://localhost:8080/auth/jwks.json",
		// public should be in jwks mapped with kid
		// time for re-authenticate (signin)
		"exp":   time.Now().Add(24 * 165 * 10 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "https://chat.test.com",
		"nonce": "test",
		"user":  user,
	})
	return token.SignedString(ed25519Secret)
}

func genAccessToken(aud, sub string) (string, error) {
	// HS256 for symmetric signature, sign and verify in server
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "https://chat.test.com",
		Audience:  aud,
		Subject:   sub,
	})
	return token.SignedString(env.HS256_SECRET)
}
