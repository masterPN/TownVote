package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LineTownVote/api/candidates"
)

// Connection URI
const uri = "mongodb://localhost:27017"

func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "The is a message from voter!"})
}

func main() {
	// Set MongoDB router
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	voteDB := client.Database("LineTownVoteDB")

	// Set root api router
	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/", GetTest)

	router.GET("/api/candidates", func(c *gin.Context) {
		res := candidates.GetAllCandidates(voteDB, ctx)
		c.JSON(http.StatusOK, res)
	})

	router.Run()

}
