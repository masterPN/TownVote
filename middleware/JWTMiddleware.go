package middleware

import (
	"LineTownVote/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthorizeJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		// If no authHeader, then send unauthorize
		if len(authHeader) <= len(BEARER_SCHEMA) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		tokenString := authHeader[len(BEARER_SCHEMA):]
		token, err := service.JWTAuthService().ValidateToken(tokenString)
		if token.Valid {
			// claims := token.Claims.(jwt.MapClaims)
			// fmt.Println(claims)
		} else {
			_ = err
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
