package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connection URI
const uri = "mongodb://localhost:27017"

func GetTest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "The is a message from voter!"})
}

func main() {
	// Set root api router
	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1"})

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
	cadidatesCollection := voteDB.Collection("candidates")

	cursor, err := cadidatesCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	var candidates []bson.M
	if err = cursor.All(ctx, &candidates); err != nil {
		panic(err)
	}
	fmt.Println(candidates)

	router.GET("/", GetTest)

	router.Run()
}
