package model

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
