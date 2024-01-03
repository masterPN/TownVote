package candidates

import (
	"LineTownVote/model"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionStatusName = "status"
const collectionName = "candidates"
const collectionResultName = "result"

var NilCandidates model.Candidates
var nilCandidate model.Candidate

func GetAllCandidates(db *mongo.Database, ctx context.Context) (model.Candidates, error) {
	candidatesCollection := db.Collection(collectionName)
	resultCollection := db.Collection(collectionResultName)

	cursor, err := candidatesCollection.Find(ctx, bson.M{})
	if err != nil {
		return NilCandidates, err
	}
	var res model.Candidates
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

func GetCandidateDetail(db *mongo.Database, ctx context.Context, candidateId string) (model.Candidate, error) {
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

	var res model.Candidate
	err := candidatesCollection.FindOne(ctx, filter, opts).Decode(&res)
	if err != nil {
		return nilCandidate, err
	}

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

func CreateCandidate(db *mongo.Database, ctx context.Context, inputForm model.Candidate) (model.Candidate, error) {
	candidatesCollection := db.Collection(collectionName)
	statusCollection := db.Collection(collectionStatusName)

	// Check if found the same bioLink then abort
	filter := bson.D{{"bioLink", inputForm.BioLink}}
	var resCheck model.Candidate
	err := candidatesCollection.FindOne(ctx, filter).Decode(&resCheck)
	if err == nil {
		return nilCandidate, errors.New("duplicate bio link")
	}

	// Prepare form
	// Get candidate continuously count
	filter = bson.D{}
	var resStatus model.Status
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

func UpdateCandidate(db *mongo.Database, ctx context.Context, bodyInput model.Candidate, candidateID string) (model.Candidate, error) {
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

// API Handler

// @BasePath /api

// GetCandidates godoc
// @Summary Get Candidates
// @Description Get all candidates
// @Tags candidates
// @Produce json
// @Response 200 {object} model.Candidates "OK"
// @Router /candidates [get]
func APIGetCandidatesHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	res, err := GetAllCandidates(voteDB, ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

// GetCandidate godoc
// @Summary Get Candidate
// @Description Get selected candidate
// @Tags candidates
// @Produce json
// @Response 200 {object} model.Candidate "OK"
// @Router /candidates/{id} [get]
func APIGetCandidateDetailHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	candidateID := c.Param("candidateID")
	num, _ := strconv.Atoi(candidateID)
	compareParam := strings.Compare(strconv.Itoa(num), candidateID)
	if compareParam != 0 {
		c.JSON(http.StatusBadRequest, "")
		fmt.Println("das")
		return
	}

	res, err := GetCandidateDetail(voteDB, ctx, candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

// CreatCandidate godoc
// @Summary Create Candidate
// @Description Create candidate
// @Tags candidates
// @Produce json
// @Param candidate body model.Candidate true "candidate detail"
// @Response 200 {object} model.Candidate "OK"
// @Router /candidates [post]
func APIPostCreateCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput model.Candidate
	c.BindJSON(&bodyInput)
	res, err := CreateCandidate(voteDB, ctx, bodyInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

// UpdateCandidate godoc
// @Summary Update Candidate
// @Description Update candidate
// @Tags candidates
// @Produce json
// @Param candidate body model.Candidate true "candidate detail"
// @Response 200 {object} model.Candidate "OK"
// @Router /candidates/{id} [put]
func APIPutUpdateCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput model.Candidate
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

// DeleteCandidate godoc
// @Summary Delete Candidate
// @Description Delete candidate
// @Tags candidates
// @Produce json
// @Response 200 {object} model.ApiDeleteCandidateHandlerResponse "OK"
// @Router /candidates/{id} [delete]
func APIDeleteCandidateHandler(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	candidateID := c.Param("candidateID")
	err := DeleteCandidate(voteDB, ctx, candidateID)

	var res model.ApiDeleteCandidateHandlerResponse
	if err != nil {
		res.Status = "error"
		res.Message = "Candidate not found"
		c.JSON(http.StatusInternalServerError, res)
		panic(err)
	}
	res.Status = "ok"
	c.JSON(http.StatusOK, res)
}
