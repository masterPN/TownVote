package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LineTownVote/api/candidates"
	"LineTownVote/api/vote"
	"LineTownVote/controller"
	"LineTownVote/middleware"
	"LineTownVote/service"
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

	// JWT
	var loginService service.LoginService = service.StaticLoginService()
	var jwtService service.JWTService = service.JWTAuthService()
	var loginController controller.LoginController = controller.LoginHandler(loginService, jwtService)
	router.POST("/get_token", func(c *gin.Context) {
		token := loginController.Login(c)
		if token != "" {
			c.JSON(http.StatusOK, gin.H{
				"token": token,
			})
		} else {
			c.JSON(http.StatusUnauthorized, nil)
		}
	})

	api := router.Group("/api")
	api.Use(middleware.AuthorizeJWT())
	{
		api.GET("/candidates", func(c *gin.Context) {
			res, err := candidates.GetAllCandidates(voteDB, ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

		api.GET("/candidates/:candidateID", func(c *gin.Context) {
			candidateID := c.Param("candidateID")
			res, err := candidates.GetCandidateDetail(voteDB, ctx, candidateID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

		api.POST("/candidates", func(c *gin.Context) {
			var bodyInput candidates.Candidate
			c.BindJSON(&bodyInput)
			res, err := candidates.CreateCandidate(voteDB, ctx, bodyInput)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

		api.PUT("/candidates/:candidateID", func(c *gin.Context) {
			var bodyInput candidates.Candidate
			c.BindJSON(&bodyInput)
			candidateID := c.Param("candidateID")
			if candidateID != bodyInput.Id {
				c.JSON(http.StatusBadRequest, "ID on Param and Body don't match.")
				panic("ID on Param and Body don't match.")
			}
			res, err := candidates.UpdateCandidate(voteDB, ctx, bodyInput, candidateID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

		api.DELETE("/candidates/:candidateID", func(c *gin.Context) {
			candidateID := c.Param("candidateID")
			err := candidates.DeleteCandidate(voteDB, ctx, candidateID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Candidate not found"})
				panic(err)
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.POST("/vote/status", func(c *gin.Context) {
			var bodyInput vote.VoteInput
			c.BindJSON(&bodyInput)
			res, err := vote.CheckVoteStatus(voteDB, ctx, bodyInput)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": false})
				panic(err)
			}
			if res == false {
				c.JSON(http.StatusOK, gin.H{"status": false})
			}
			if res == true {
				c.JSON(http.StatusOK, gin.H{"status": true})
			}
		})
	}

	router.Run()

}
