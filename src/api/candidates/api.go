package candidates

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Candidates []struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Dob        string `json:"dob"`
	BioLink    string `json:"bioLink"`
	ImageLink  string `json:"imageLink"`
	Policy     string `json:"policy"`
	VotedCount int32  `json:"votedCount"`
}

func GetAllCandidates(db *mongo.Database, ctx context.Context) Candidates {
	cadidatesCollection := db.Collection("candidates")

	cursor, err := cadidatesCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	var res Candidates
	if err = cursor.All(ctx, &res); err != nil {
		panic(err)
	}

	return res
}
