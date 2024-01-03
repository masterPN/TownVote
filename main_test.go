package main

import (
	"LineTownVote/api/candidates"
	"LineTownVote/api/election"
	"LineTownVote/api/vote"
	"LineTownVote/controller"
	"LineTownVote/middleware"
	"LineTownVote/service"
	"LineTownVote/websocket_mod"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Bearer = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZF9ubyI6IjEyMzQ1Njc4OTAxMjMiLCJpZF9sYXNlckNvZGUiOiJKVDEyMyIsImV4cCI6MTcwNTE2MTM1MywiaWF0IjoxNzA0Mjk3MzUzLCJpc3MiOiJLb3JLb3JUb3IifQ.1yivRGTxy7e0Jb3NyLhbA5LEEnId2ceu13o4lS46jDA"

func setupRouter() *gin.Engine {
	// Set MongoDB router
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var ctx context.Context
	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}

	// If collection status is empty, then init status
	voteDB := client.Database("LineTownVoteDB")
	voteDB.Collection("status").Drop(ctx)
	_, _ = voteDB.Collection("status").InsertOne(ctx, bson.D{
		{Key: "candidateContinuouslyCount", Value: 0},
		{Key: "voteToggle", Value: false},
	})

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

	return router
}

func Test_PostCreateCandidate_Success(t *testing.T) {
	Test_DeleteCandidate_Success(t)

	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPost, "/api/candidates", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_GetAllCandidates_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_GetAllCandidates_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_GetCandidateDetail_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_GetCandidateDetail_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates/1", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_GetCandidateDetail_Error_IdNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates/0", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func Test_GetCandidateDetail_Error_BadParam(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/candidates/z", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func Test_PostCreateCandidate_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPost, "/api/candidates", bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_PostCreateCandidate_Error(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPost, "/api/candidates", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func Test_PutUpdateCandidate_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"id": "-1"
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPut, "/api/candidates/-1", bytes.NewBuffer(reqBody))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_PutUpdateCandidate_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"id": "1",
		"name": "Orange",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPut, "/api/candidates/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_PutUpdateCandidate_Error_IdsDontMatched(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"id": "-1",
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPut, "/api/candidates/1", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func Test_PutUpdateCandidate_Error_IdNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{
		"id": "-1",
		"name": "Brown",
		"dob": "August 8, 2011",
		"bioLink": "https://line.fandom.com/wiki/Brown",
		"imageLink": "https://static.wikia.nocookie.net/line/images/b/bb/2015-brown.png/revision/latest/scale-to-width-down/700?cb=20150808131630",
		"policy": "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown"
	  }`)
	req, _ := http.NewRequest(http.MethodPut, "/api/candidates/-1", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func Test_DeleteCandidate_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/candidates/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_DeleteCandidate_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/candidates/1", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_DeleteCandidate_Error_IdNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/candidates/1", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func Test_PostVoteStatus_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/vote/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_PostVoteStatus_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{"nationalId": "1234567890123"}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/vote/status", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_PostVote_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/vote", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_PostVote_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{"nationalId": "1234567890123", "candidateId": 1}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/vote", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_PostToggleElection_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/election/toggle", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_PostToggleElection_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{"enable": true}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/election/toggle", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_PostToggleElection_Error(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	var reqBody = []byte(`{"enable": tru}`)
	req, _ := http.NewRequest(http.MethodPost, "/api/election/toggle", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", Bearer)
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func Test_GetElectionResult_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/election/result", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_GetElectionResult_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/election/result", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func Test_GetElectionResultExport_Unauthorized(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/election/export", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_GetElectionResultExport_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/election/export", nil)
	req.Header.Set("Authorization", Bearer)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func Test_SocketCandidate(t *testing.T) {
	Test_PostCreateCandidate_Success(t)

	router := setupRouter()
	_ = router

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws://127.0.0.1:8080/ws/results/1"

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// read response and check to see if it's what we expect.
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	pString := strings.TrimSuffix(string(p), "\n")
	assert.Equal(t, string(`{"id":"1","votedCount":0}`), pString)
}
