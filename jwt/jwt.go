package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrSigningMethod = errors.New("signing method err")
)

type JWT struct {
	secretKey string
	duration  time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

func NewJWT(secretKey string, duration time.Duration) *JWT {
	return &JWT{secretKey: secretKey, duration: duration}
}

func (j *JWT) Generate(username string) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.duration).Unix(),
		},
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.secretKey))
}

func (j *JWT) Verify(token string) (*UserClaims, error) {
	verifyToken, err := jwt.ParseWithClaims(
		token,
		&UserClaims{},
		func(t *jwt.Token) (interface{}, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrSigningMethod
			}

			return []byte(j.secretKey), nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := verifyToken.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
