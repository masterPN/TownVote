package candidates

import (
	"context"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionStatName = "stat"
const collectionName = "candidates"

type Stat struct {
	CandidateContinuouslyCount int32 `json:"candidateContinuouslyCount"`
}
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
	candidatesCollection := db.Collection(collectionName)

	cursor, err := candidatesCollection.Find(ctx, bson.M{})
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
	candidatesCollection := db.Collection(collectionName)

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
	err := candidatesCollection.FindOne(ctx, filter, opts).Decode(&res)
	if err != nil {
		panic(err)
	}

	return res
}

func CreateCandidate(db *mongo.Database, ctx context.Context, inputForm Candidate) Candidate {
	candidatesCollection := db.Collection(collectionName)
	statCollection := db.Collection(collectionStatName)

	// Check if found the same bioLink then abort
	filter := bson.D{{"bioLink", inputForm.BioLink}}
	var resCheck Candidate
	err := candidatesCollection.FindOne(ctx, filter).Decode(&resCheck)
	if err == nil {
		panic("duplicate bio link")
	}

	// Prepare form
	// Get candidate continuously count
	filter = bson.D{}
	var resStat Stat
	err = statCollection.FindOne(ctx, filter).Decode(&resStat)
	if err != nil {
		panic(err)
	}
	// Insert form to collection
	currentCount := resStat.CandidateContinuouslyCount + 1
	insertResult, err := candidatesCollection.InsertOne(ctx, bson.D{
		{Key: "id", Value: strconv.Itoa(int(currentCount))},
		{Key: "name", Value: inputForm.Name},
		{Key: "dob", Value: inputForm.Dob},
		{Key: "bioLink", Value: inputForm.BioLink},
		{Key: "imageLink", Value: inputForm.ImageLink},
		{Key: "policy", Value: inputForm.Policy},
		{Key: "votedCount", Value: 0},
	})
	if err != nil {
		panic(err)
	}
	_ = insertResult

	// If insert candidate success that update stat
	filter = bson.D{{"candidateContinuouslyCount", resStat.CandidateContinuouslyCount}}
	update := bson.M{
		"$set": bson.M{
			"candidateContinuouslyCount": currentCount,
		},
	}
	var updateStat bson.D
	err = statCollection.FindOneAndUpdate(ctx, filter, update).Decode(&updateStat)
	if err != nil {
		panic(err)
	}

	return GetCandidateDetail(db, ctx, strconv.Itoa(int(currentCount)))

	// fmt.Println(insertResult)
}
