package token

import (
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrSigningMethod = errors.New("signing method err")
)

const defaultSecretKey = "go-wms-token"

func Generate(claims jwt.Claims, secret string) (string, error) {
	if secret == "" {
		secret = defaultSecretKey
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ExtractTokenFromRequest(r *http.Request) (string, error) {
	token, err := request.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		return "", errors.New("ExtractTokenFromRequestFailed")
	}

	return token, nil
}

func VerityExtractTokenFromRequest(r *http.Request, claims jwt.Claims, secret string) (*jwt.Token, error) {
	token, err := request.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		return nil, err
	}

	return Verify(token, claims, secret)
}

func Verify(token string, claims jwt.Claims, secret string) (*jwt.Token, error) {
	verifyToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrSigningMethod
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	return verifyToken, nil
}

func TokenMapClaims(token string, secret string) map[string]interface{} {
	parMap := make(map[string]interface{})

	parseAuth, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
		// return []byte("My Secret"), nil
		return []byte(secret), nil
	})
	if err != nil {
		return parMap
	}

	claim, ok := parseAuth.Claims.(jwt.MapClaims)
	if !ok {
		return parMap
	}

	for key, val := range claim {
		parMap[key] = val
	}

	return parMap
}
