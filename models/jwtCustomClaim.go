package models

import (
	"os"

	"github.com/golang-jwt/jwt"
)

type CustomClaim struct {
	*jwt.StandardClaims
	Username string
}

func (c *CustomClaim) GetClaim(tokenString string) (CustomClaim, error) {
	claims := c

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if !token.Valid || err != nil {
		return *claims, err
	}

	return *claims, nil
}
