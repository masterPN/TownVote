package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "The is a message from voter!"})
}

func main() {
	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/", GetTest)

	router.Run()
}
