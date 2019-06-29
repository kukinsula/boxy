package usecase

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Tokener struct {
	secret []byte
}

type GenerateTokenParams struct {
	Audience  string
	ExpiresIn time.Duration
	NotBefore time.Duration
	Issuer    string
	Subject   string

	Email string `json:"email"`
	UUID  string `json:"uuid"`
}

func NewTokener(secret string) *Tokener {
	return &Tokener{
		secret: []byte(secret),
	}
}

func (tokener *Tokener) Generate(params GenerateTokenParams) (string, error) {
	now := time.Now()

	claims := struct {
		jwt.StandardClaims

		Email string `json:"email"`
		UUID  string `json:"uuid"`
	}{
		StandardClaims: jwt.StandardClaims{
			Audience:  params.Audience,
			ExpiresAt: now.Add(params.ExpiresIn).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    params.Issuer,
			// NotBefore: params.NotBefore,
			Subject: params.Subject,
		},

		Email: params.Email,
		UUID:  params.UUID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(tokener.secret)
}

func (tokener *Tokener) Verify(str string) (interface{}, error) {
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		return tokener.secret, nil
	})

	if token.Valid {
		return token, nil
	}

	if err, ok := err.(*jwt.ValidationError); ok {
		return nil, err
	} else {
		return nil, fmt.Errorf("Tokener: couldn't handle token: %s", err)
	}
}
