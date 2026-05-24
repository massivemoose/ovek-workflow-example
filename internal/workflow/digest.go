package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

const DigestWorkflowName = "digest"

type Signup struct {
	ID      string `json:"id,omitempty"`
	Email   string `json:"email,omitempty"`
	Created string `json:"created,omitempty"`
}

type WorkflowResult struct {
	ID                    string    `json:"id,omitempty"`
	Workflow              string    `json:"workflow"`
	RunID                 string    `json:"run_id"`
	Status                string    `json:"status"`
	StartedAt             time.Time `json:"started_at"`
	FinishedAt            time.Time `json:"finished_at"`
	TotalSignups          int       `json:"total_signups"`
	NewSignups            int       `json:"new_signups"`
	LatestSignupCreatedAt string    `json:"latest_signup_created_at"`
	Summary               string    `json:"summary"`
}

func (r *WorkflowResult) UnmarshalJSON(data []byte) error {
	type rawResult struct {
		ID                    string `json:"id,omitempty"`
		Workflow              string `json:"workflow"`
		RunID                 string `json:"run_id"`
		Status                string `json:"status"`
		StartedAt             string `json:"started_at"`
		FinishedAt            string `json:"finished_at"`
		TotalSignups          int    `json:"total_signups"`
		NewSignups            int    `json:"new_signups"`
		LatestSignupCreatedAt string `json:"latest_signup_created_at"`
		Summary               string `json:"summary"`
	}

	var raw rawResult
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	started, err := parsePocketBaseTime(raw.StartedAt)
	if err != nil {
		return err
	}
	finished, err := parsePocketBaseTime(raw.FinishedAt)
	if err != nil {
		return err
	}

	*r = WorkflowResult{
		ID:                    raw.ID,
		Workflow:              raw.Workflow,
		RunID:                 raw.RunID,
		Status:                raw.Status,
		StartedAt:             started,
		FinishedAt:            finished,
		TotalSignups:          raw.TotalSignups,
		NewSignups:            raw.NewSignups,
		LatestSignupCreatedAt: raw.LatestSignupCreatedAt,
		Summary:               raw.Summary,
	}
	return nil
}

func parsePocketBaseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	formats := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.000Z",
		"2006-01-02 15:04:05Z",
	}
	var lastErr error
	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

type DigestInput struct {
	RunID              string
	Signups            []Signup
	PreviousSuccessful *WorkflowResult
	Started            time.Time
	Finished           time.Time
}

func CalculateDigest(input DigestInput) WorkflowResult {
	signups := append([]Signup(nil), input.Signups...)
	sort.SliceStable(signups, func(i, j int) bool {
		return signups[i].Created < signups[j].Created
	})

	latest := ""
	if len(signups) > 0 {
		latest = signups[len(signups)-1].Created
	}

	previousHighWater := ""
	if input.PreviousSuccessful != nil {
		previousHighWater = input.PreviousSuccessful.LatestSignupCreatedAt
	}

	newSignups := 0
	for _, signup := range signups {
		if previousHighWater == "" || signup.Created > previousHighWater {
			newSignups++
		}
	}

	total := len(signups)
	summary := "No signups yet."
	switch {
	case total == 0:
	case newSignups == 0:
		summary = fmt.Sprintf("Digest found %d total signups, with no new signups since the previous successful run.", total)
	default:
		summary = fmt.Sprintf("Digest found %d total signups, including %d new since the previous successful run.", total, newSignups)
	}

	return WorkflowResult{
		Workflow:              DigestWorkflowName,
		RunID:                 input.RunID,
		Status:                "succeeded",
		StartedAt:             input.Started,
		FinishedAt:            input.Finished,
		TotalSignups:          total,
		NewSignups:            newSignups,
		LatestSignupCreatedAt: latest,
		Summary:               summary,
	}
}

func ShortRunID(runID string) string {
	if len(runID) <= 8 {
		return runID
	}
	return runID[:8]
}
