package recension

// Outcome is the server comparison result for one submitted testcase.
type Outcome struct {
	Team     string  `json:"team"`
	Suite    string  `json:"suite"`
	Version  string  `json:"version"`
	Testcase string  `json:"testcase"`
	Verdict  string  `json:"verdict"` // sent | pass | diff
	Score    float64 `json:"score"`
}

type submitResponse struct {
	Outcomes []Outcome `json:"outcomes"`
}
