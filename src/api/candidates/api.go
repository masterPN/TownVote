package candidates

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAllCandidates(db *mongo.Database, ctx context.Context) []primitive.M {
	cadidatesCollection := db.Collection("candidates")

	cursor, err := cadidatesCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	var candidates []bson.M
	if err = cursor.All(ctx, &candidates); err != nil {
		panic(err)
	}

	return candidates
}
