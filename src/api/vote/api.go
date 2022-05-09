package vote

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionStatusName = "status"
const collectionResultName = "result"

type VoteInput struct {
	NationalId  string
	CandidateId int
}
type Status struct {
	VoteToggle bool
}

func CheckVoteStatus(db *mongo.Database, ctx context.Context, bodyInput VoteInput) (bool, string, error) {
	resultCollection := db.Collection(collectionResultName)
	statusCollection := db.Collection(collectionStatusName)

	// If the election stop count, then the voter can't vote.
	filter := bson.D{}
	var resStatus Status
	err := statusCollection.FindOne(ctx, filter).Decode(&resStatus)
	if err != nil {
		return false, "", err
	}
	if !resStatus.VoteToggle {
		return false, "Election is closed", nil
	}

	// If the voter had already voted, then the voter can't vote.
	filter = bson.D{{"nationalId", bodyInput.NationalId}}
	var resFindVoter VoteInput
	err = resultCollection.FindOne(ctx, filter).Decode(&resFindVoter)
	if err == nil {
		return false, "Already voted", nil
	}

	// The voter can vote.
	return true, "", nil
}

func Vote(db *mongo.Database, ctx context.Context, bodyInput VoteInput) (bool, string, error) {
	// Check if the voter can vote.
	canVote, errorMsg, err := CheckVoteStatus(db, ctx, bodyInput)
	if err != nil {
		return false, "", err
	}

	// If the voter can't vote, then return false
	if !canVote {
		return canVote, errorMsg, err
	}

	// Save vote
	resultCollection := db.Collection(collectionResultName)
	insertRes, err := resultCollection.InsertOne(ctx, bson.D{
		{Key: "candidateId", Value: bodyInput.CandidateId},
		{Key: "nationalId", Value: bodyInput.NationalId},
	})
	if err != nil {
		return false, "", err
	}
	_ = insertRes

	// Save success
	return true, "", nil
}

func APIPostCheckStatusHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput VoteInput
	c.BindJSON(&bodyInput)
	res, _, err := CheckVoteStatus(voteDB, ctx, bodyInput)
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
}

func APIPostVote(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput VoteInput
	c.BindJSON(&bodyInput)
	res, errMsg, err := Vote(voteDB, ctx, bodyInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	if !res {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": errMsg})
	}
	if res {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
