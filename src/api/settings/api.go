package settings

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionStatusName = "status"

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
