package candidates

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "candidates"

type Candidates []struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Dob        string `json:"dob"`
	BioLink    string `json:"bioLink"`
	ImageLink  string `json:"imageLink"`
	Policy     string `json:"policy"`
	VotedCount int32  `json:"votedCount"`
}
type Candidate struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Dob        string `json:"dob"`
	BioLink    string `json:"bioLink"`
	ImageLink  string `json:"imageLink"`
	Policy     string `json:"policy"`
	VotedCount int32  `json:"votedCount"`
}

func GetAllCandidates(db *mongo.Database, ctx context.Context) Candidates {
	cadidatesCollection := db.Collection(collectionName)

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

func GetCandidateDetail(db *mongo.Database, ctx context.Context, candidateId string) Candidate {
	cadidatesCollection := db.Collection(collectionName)

	filter := bson.D{{"id", candidateId}}
	projection := bson.D{
		{"_id", 0},
		{"id", 1},
		{"name", 1},
		{"dob", 1},
		{"bioLink", 1},
		{"imageLink", 1},
		{"policy", 1},
		{"votedCount", 1}}
	opts := options.FindOne().SetProjection(projection)

	var res Candidate
	err := cadidatesCollection.FindOne(ctx, filter, opts).Decode(&res)
	if err != nil {
		panic(err)
	}

	return res
}
