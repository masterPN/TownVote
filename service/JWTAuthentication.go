package service

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(id_no string, id_laserCode string) string
	ValidateToken(token string) (*jwt.Token, error)
}
type jwtServices struct {
	secretKey string
	issure    string
}
type authCustomClaims struct {
	Id_no        string `json:"id_no"`
	Id_laserCode string `json:"id_laserCode"`
	jwt.RegisteredClaims
}

func JWTAuthService() JWTService {
	return &jwtServices{
		secretKey: "secretlyKey",
		issure:    "KorKorTor",
	}
}

func (service *jwtServices) GenerateToken(id_no string, id_laserCode string) string {
	claims := &authCustomClaims{
		id_no,
		id_laserCode,
		jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(10 * 24 * time.Hour)},
			Issuer:    service.issure,
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		panic(err)
	}

	return t
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("Invalid token", token.Header["alg"])
		}
		return []byte(service.secretKey), nil
	})
}
