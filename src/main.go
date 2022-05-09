package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LineTownVote/api/candidates"
	"LineTownVote/api/election"
	"LineTownVote/api/vote"
	"LineTownVote/controller"
	"LineTownVote/middleware"
	"LineTownVote/service"
	"LineTownVote/websocket_mod"
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

	// API
	api := router.Group("/api")
	api.Use(middleware.AuthorizeJWT())
	{
		api.GET("/candidates", func(c *gin.Context) {
			candidates.APIGetCandidatesHandler(c, voteDB, ctx)
		})

		api.GET("/candidates/:candidateID", func(c *gin.Context) {
			candidates.APIGetCandidateDetailHandler(c, voteDB, ctx)
		})

		api.POST("/candidates", func(c *gin.Context) {
			candidates.APIPostCreateCandidateHandler(c, voteDB, ctx)
		})

		api.PUT("/candidates/:candidateID", func(c *gin.Context) {
			candidates.APIPutUpdateCandidateHandler(c, voteDB, ctx)
		})

		api.DELETE("/candidates/:candidateID", func(c *gin.Context) {
			candidates.APIDeleteCandidateHandler(c, voteDB, ctx)
		})

		api.POST("/vote/status", func(c *gin.Context) {
			vote.APIPostCheckStatusHandler(c, voteDB, ctx)
		})

		api.POST("/vote", func(c *gin.Context) {
			vote.APIPostVote(c, voteDB, ctx)
		})

		api.POST("/election/toggle", func(c *gin.Context) {
			election.APIPostToggle(c, voteDB, ctx)
		})

		api.GET("/election/result", func(c *gin.Context) {
			election.APIGetResult(c, voteDB, ctx)
		})

		api.GET("/election/export", func(c *gin.Context) {
			election.APIGetExport(c, voteDB, ctx)
		})
	}

	// Websocket
	router.GET("/ws/results/:candidateID", func(c *gin.Context) {
		websocket_mod.Handler(c, voteDB, ctx)
	})

	router.Run()

}
