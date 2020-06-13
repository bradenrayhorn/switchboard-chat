package utils

import (
	"fmt"
	"github.com/bradenrayhorn/switchboard-chat/config"
	"github.com/dgrijalva/jwt-go"
)

func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.RsaPublic, nil
	})
}
