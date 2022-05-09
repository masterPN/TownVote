package election

import (
	"LineTownVote/api/candidates"
	"context"
	"encoding/csv"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionStatusName = "status"
const collectionResultName = "result"

type Toggle struct {
	Status string
	Enable bool
}
type Results []struct {
	CandidateId int32  `json:"candidateId"`
	NationalId  string `json:"nationalId"`
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
	results, err := candidates.GetAllCandidates(db, ctx)
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
	for i := 0; i < len(results); i++ {
		results[i].Percentage = strconv.Itoa(int((float32(results[i].VotedCount)/float32(total))*100)) + "%"
	}

	return results, nil
}

func GetCSVExport(db *mongo.Database, ctx context.Context) {
	resultCollection := db.Collection(collectionResultName)
	// fields := []string{"candidateId", "nationalId"}

	cursor, err := resultCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	var records Results
	if err = cursor.All(ctx, &records); err != nil {
		panic(err)
	}

	csvFile, err := os.Create("results.csv")
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(csvFile)
	defer w.Flush()

	var data [][]string
	data = append(data, []string{"Candidate id", "National id"})
	for _, record := range records {
		row := []string{strconv.Itoa(int(record.CandidateId)), record.NationalId}
		data = append(data, row)
	}
	err = w.WriteAll(data)
	if err != nil {
		panic(err)
	}
}

func APIPostToggle(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	var bodyInput Toggle
	c.BindJSON(&bodyInput)
	err := ToggleElection(voteDB, ctx, bodyInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "enable": bodyInput.Enable})
	}
}

func APIGetResult(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	res, err := GetResult(voteDB, ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		panic(err)
	}
	c.JSON(http.StatusOK, res)
}

func APIGetExport(c *gin.Context, voteDB *mongo.Database, ctx context.Context) {
	GetCSVExport(voteDB, ctx)
	fileName := "results.csv"
	targetPath := filepath.Join("./", fileName)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "text/csv")
	c.FileAttachment(targetPath, fileName)
}
