package websocket_mod

import (
	"LineTownVote/api/candidates"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type message struct {
	Id         string `json:"id"`
	VotedCount int32  `json:"votedCount"`
}

func Handler(c *gin.Context, db *mongo.Database, ctx context.Context) {
	// declare
	candidateID := c.Param("candidateID")

	// init websocket
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil, 0, 0)
	if err != nil {
		panic(err)
	}
	defer ws.Close()

	for {
		// get candidate result
		res, err := candidates.GetCandidateDetail(db, ctx, candidateID)
		if err != nil {
			panic(err)
		}

		// format response
		msg := message{
			Id:         candidateID,
			VotedCount: res.VotedCount,
		}

		err = ws.WriteJSON(msg)
		if err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}
}
