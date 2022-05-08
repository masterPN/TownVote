package vote

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionVoterName = "voter"
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
	voterCollection := db.Collection(collectionVoterName)
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
	err = voterCollection.FindOne(ctx, filter).Decode(&resFindVoter)
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
