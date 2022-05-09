package candidates

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionStatusName = "status"
const collectionName = "candidates"
const collectionResultName = "result"

type Status struct {
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
	Percentage string `json:"percentage,omitempty"`
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

var NilCandidates Candidates
var nilCandidate Candidate

func GetAllCandidates(db *mongo.Database, ctx context.Context) (Candidates, error) {
	candidatesCollection := db.Collection(collectionName)
	resultCollection := db.Collection(collectionResultName)

	cursor, err := candidatesCollection.Find(ctx, bson.M{})
	if err != nil {
		return NilCandidates, err
	}
	var res Candidates
	if err = cursor.All(ctx, &res); err != nil {
		return NilCandidates, err
	}

	// Count votes
	for i := 0; i < len(res); i++ {
		currentId, err := strconv.Atoi(res[i].Id)
		filter := bson.D{{"candidateId", currentId}}
		currentCount, err := resultCollection.CountDocuments(ctx, filter)
		if err != nil {
			currentCount = 0
		}
		res[i].VotedCount = int32(currentCount)
	}

	return res, err
}

func GetCandidateDetail(db *mongo.Database, ctx context.Context, candidateId string) (Candidate, error) {
	candidatesCollection := db.Collection(collectionName)
	resultCollection := db.Collection(collectionResultName)

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

	// Count votes
	currentId, _ := strconv.Atoi(res.Id)
	filter = bson.D{{"candidateId", currentId}}
	currentCount, err := resultCollection.CountDocuments(ctx, filter)
	if err != nil {
		currentCount = 0
	}
	res.VotedCount = int32(currentCount)

	return res, err
}

func CreateCandidate(db *mongo.Database, ctx context.Context, inputForm Candidate) (Candidate, error) {
	candidatesCollection := db.Collection(collectionName)
	statusCollection := db.Collection(collectionStatusName)

	// Check if found the same bioLink then abort
	filter := bson.D{{"bioLink", inputForm.BioLink}}
	var resCheck Candidate
	err := candidatesCollection.FindOne(ctx, filter).Decode(&resCheck)
	if err == nil {
		return nilCandidate, errors.New("duplicate bio link")
	}

	// Prepare form
	// Get candidate continuously count
	filter = bson.D{}
	var resStatus Status
	err = statusCollection.FindOne(ctx, filter).Decode(&resStatus)
	if err != nil {
		return nilCandidate, err
	}
	// Insert form to collection
	currentCount := resStatus.CandidateContinuouslyCount + 1
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
		return nilCandidate, err
	}
	_ = insertResult

	// If insert candidate success that update stat
	filter = bson.D{{"candidateContinuouslyCount", resStatus.CandidateContinuouslyCount}}
	update := bson.M{
		"$set": bson.M{
			"candidateContinuouslyCount": currentCount,
		},
	}
	var updateStatus bson.D
	err = statusCollection.FindOneAndUpdate(ctx, filter, update).Decode(&updateStatus)
	if err != nil {
		return nilCandidate, err
	}

	return GetCandidateDetail(db, ctx, strconv.Itoa(int(currentCount)))
}

func UpdateCandidate(db *mongo.Database, ctx context.Context, bodyInput Candidate, candidateID string) (Candidate, error) {
	// Check if the candidate is store in DB
	_, err := GetCandidateDetail(db, ctx, candidateID)
	if err != nil {
		return nilCandidate, err
	}

	// Update the candidate
	candidatesCollection := db.Collection(collectionName)
	filter := bson.D{{"id", candidateID}}
	update := bson.M{
		"$set": bson.M{
			"name":      bodyInput.Name,
			"dob":       bodyInput.Dob,
			"bioLink":   bodyInput.BioLink,
			"imageLink": bodyInput.ImageLink,
			"policy":    bodyInput.Policy,
		},
	}
	var updateCandidateRes bson.D
	err = candidatesCollection.FindOneAndUpdate(ctx, filter, update).Decode(&updateCandidateRes)
	if err != nil {
		return nilCandidate, err
	}

	return GetCandidateDetail(db, ctx, candidateID)
}

func DeleteCandidate(db *mongo.Database, ctx context.Context, candidateID string) error {
	filter := bson.D{{"id", candidateID}}
	var deleteCandidateRes bson.D
	err := db.Collection(collectionName).FindOneAndDelete(ctx, filter).Decode(&deleteCandidateRes)

	return err
}

func APIGetCandidatesHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	res, err := GetAllCandidates(voteDB, ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

func APIGetCandidateDetailHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	candidateID := c.Param("candidateID")
	res, err := GetCandidateDetail(voteDB, ctx, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

func APIPostCreateCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput Candidate
	c.BindJSON(&bodyInput)
	res, err := CreateCandidate(voteDB, ctx, bodyInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

func APIPutUpdateCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput Candidate
	c.BindJSON(&bodyInput)
	candidateID := c.Param("candidateID")
	if candidateID != bodyInput.Id {
		c.JSON(http.StatusBadRequest, "ID on Param and Body don't match.")
		panic("ID on Param and Body don't match.")
	}
	res, err := UpdateCandidate(voteDB, ctx, bodyInput, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

func APIDeleteCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	candidateID := c.Param("candidateID")
	err := DeleteCandidate(voteDB, ctx, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Candidate not found"})
		panic(err)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
