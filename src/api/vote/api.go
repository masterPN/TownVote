package vote

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionVoterName = "voter"
const collectionStatusName = "status"

type VoteInput struct {
	NationalId  string
	CandidateId int
}
type Status struct {
	VoteToggle bool
}

func CheckVoteStatus(db *mongo.Database, ctx context.Context, bodyInput VoteInput) (bool, error) {
	voterCollection := db.Collection(collectionVoterName)
	statusCollection := db.Collection(collectionStatusName)

	// If the election stop count, then the voter can't vote.
	filter := bson.D{}
	var resStatus Status
	err := statusCollection.FindOne(ctx, filter).Decode(&resStatus)
	if err != nil {
		return false, err
	}
	if !resStatus.VoteToggle {
		return false, nil
	}

	// If the voter had already voted, then the voter can't vote.
	filter = bson.D{{"nationalId", bodyInput.NationalId}}
	var resFindVoter VoteInput
	err = voterCollection.FindOne(ctx, filter).Decode(&resFindVoter)
	if err == nil {
		return false, nil
	}

	// The voter can vote.
	return true, nil
}
