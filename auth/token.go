package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func genIdToken(user any) (string, error) {
	// RS256 for asymmetric signature, private key to sign in server, public key to verify in client
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"kid": "keyid",
		"jku": "http://localhost:8080/auth/jwks.json",
		// public should be in jwks mapped with kid
		"pk": "publickey",
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
	return token.SignedString(hs256Secret)
}
