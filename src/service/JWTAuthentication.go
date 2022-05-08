package service

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTService interface {
	GenerateToken(id_no string, id_name string) string
	ValidateToken(token string) (*jwt.Token, error)
}
type jwtServices struct {
	secretKey string
	issure    string
}
type authCustomClaims struct {
	Id_no   string `json:"id_no"`
	Id_name string `json:"id_name"`
	jwt.StandardClaims
}

func JWTAuthService() JWTService {
	return &jwtServices{
		secretKey: "secretlyKey",
		issure:    "KorKorTor",
	}
}

func (service *jwtServices) GenerateToken(id_no string, id_name string) string {
	claims := &authCustomClaims{
		id_no,
		id_name,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
			Issuer:    service.issure,
			IssuedAt:  time.Now().Unix(),
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
