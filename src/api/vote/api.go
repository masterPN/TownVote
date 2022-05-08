package vote

import (
	"LineTownVote/api/candidates"
	"context"
	"strconv"

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

func GetResult(db *mongo.Database, ctx context.Context) (candidates.Candidates, error) {
	// Get all candidates' detail
	Results, err := candidates.GetAllCandidates(db, ctx)
	if err != nil {
		return candidates.NilCandidates, err
	}

	// Calculate percentage and append to the data
	resultCollection := db.Collection(collectionResultName)
	// Count all record
	filter := bson.D{}
	total, err := resultCollection.CountDocuments(ctx, filter)
	if err != nil {
		return candidates.NilCandidates, err
	}
	// Calculate each candidate
	for i := 0; i < len(Results); i++ {
		Results[i].Percentage = strconv.Itoa(int((float32(Results[i].VotedCount)/float32(total))*100)) + "%"
	}

	return Results, nil

}
