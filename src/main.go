package main

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LineTownVote/api/candidates"
	"LineTownVote/api/election"
	"LineTownVote/api/vote"
	"LineTownVote/controller"
	"LineTownVote/middleware"
	"LineTownVote/service"
)

// Connection URI
const uri = "mongodb://localhost:27017"

func main() {
	// Set MongoDB router
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var ctx context.Context
	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	voteDB := client.Database("LineTownVoteDB")

	// Set root api router
	router := gin.Default()
	router.SetTrustedProxies([]string{"127.0.0.1"})

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
			res, _, err := vote.CheckVoteStatus(voteDB, ctx, bodyInput)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			if !res {
				c.JSON(http.StatusOK, gin.H{"status": false})
			}
			if res {
				c.JSON(http.StatusOK, gin.H{"status": true})
			}
		})

		api.POST("/vote", func(c *gin.Context) {
			var bodyInput vote.VoteInput
			c.BindJSON(&bodyInput)
			res, errMsg, err := vote.Vote(voteDB, ctx, bodyInput)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
			}
			if !res {
				c.JSON(http.StatusOK, gin.H{"status": "error", "message": errMsg})
			}
			if res {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			}
		})

		api.POST("/election/toggle", func(c *gin.Context) {
			var bodyInput election.Toggle
			c.BindJSON(&bodyInput)
			err := election.ToggleElection(voteDB, ctx, bodyInput)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
			} else {
				c.JSON(http.StatusOK, gin.H{"status": "ok", "enable": bodyInput.Enable})
			}
		})

		api.GET("/election/result", func(c *gin.Context) {
			res, err := election.GetResult(voteDB, ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

		api.GET("/election/export", func(c *gin.Context) {
			election.GetCSVExport(voteDB, ctx)
			// c.FileAttachment("./results.csv", "results.csv")
			fileName := "results.csv"
			targetPath := filepath.Join("./", fileName)
			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", "attachment; filename="+fileName)
			c.Header("Content-Type", "text/csv")
			c.FileAttachment(targetPath, fileName)
		})
	}

	router.Run()

}
