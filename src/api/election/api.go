package election

import (
	"LineTownVote/api/candidates"
	"context"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionStatusName = "status"
const collectionResultName = "result"

type Toggle struct {
	Status string
	Enable bool
}

func ToggleElection(db *mongo.Database, ctx context.Context, bodyInput Toggle) error {
	statusCollection := db.Collection(collectionStatusName)
	filter := bson.D{}
	update := bson.M{
		"$set": bson.M{
			"voteToggle": bodyInput.Enable,
		},
	}
	var updateRes bson.D
	err := statusCollection.FindOneAndUpdate(ctx, filter, update).Decode(&updateRes)
	if err != nil {
		return err
	}

	return nil
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
