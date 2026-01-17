package tools

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, secret []byte, lifeTime time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(lifeTime).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
