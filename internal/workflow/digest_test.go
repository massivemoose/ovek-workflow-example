package workflow

import (
	"encoding/json"
	"testing"
	"time"
)

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}

func TestCalculateDigestNoSignups(t *testing.T) {
	result := CalculateDigest(DigestInput{
		RunID:    "run_empty",
		Started:  mustTime(t, "2026-05-24T10:00:00Z"),
		Finished: mustTime(t, "2026-05-24T10:00:03Z"),
	})

	if result.TotalSignups != 0 || result.NewSignups != 0 {
		t.Fatalf("counts = total %d new %d, want 0/0", result.TotalSignups, result.NewSignups)
	}
	if result.LatestSignupCreatedAt != "" {
		t.Fatalf("LatestSignupCreatedAt = %q, want empty", result.LatestSignupCreatedAt)
	}
	if result.Summary != "No signups yet." {
		t.Fatalf("Summary = %q, want empty-state summary", result.Summary)
	}
}

func TestCalculateDigestFirstRunCountsAllSignupsAsNew(t *testing.T) {
	result := CalculateDigest(DigestInput{
		RunID: "run_first",
		Signups: []Signup{
			{Created: "2026-05-24T10:00:00Z"},
			{Created: "2026-05-24T10:03:00Z"},
		},
		Started:  mustTime(t, "2026-05-24T10:04:00Z"),
		Finished: mustTime(t, "2026-05-24T10:04:02Z"),
	})

	if result.TotalSignups != 2 || result.NewSignups != 2 {
		t.Fatalf("counts = total %d new %d, want 2/2", result.TotalSignups, result.NewSignups)
	}
	if result.LatestSignupCreatedAt != "2026-05-24T10:03:00Z" {
		t.Fatalf("LatestSignupCreatedAt = %q, want newest signup timestamp", result.LatestSignupCreatedAt)
	}
	if result.Summary != "Digest found 2 total signups, including 2 new since the previous successful run." {
		t.Fatalf("Summary = %q, want first-run summary", result.Summary)
	}
}

func TestCalculateDigestLaterRunWithNoNewSignups(t *testing.T) {
	previous := WorkflowResult{LatestSignupCreatedAt: "2026-05-24T10:03:00Z"}

	result := CalculateDigest(DigestInput{
		RunID: "run_later",
		Signups: []Signup{
			{Created: "2026-05-24T10:00:00Z"},
			{Created: "2026-05-24T10:03:00Z"},
		},
		PreviousSuccessful: &previous,
		Started:            mustTime(t, "2026-05-24T11:00:00Z"),
		Finished:           mustTime(t, "2026-05-24T11:00:02Z"),
	})

	if result.TotalSignups != 2 || result.NewSignups != 0 {
		t.Fatalf("counts = total %d new %d, want 2/0", result.TotalSignups, result.NewSignups)
	}
	if result.Summary != "Digest found 2 total signups, with no new signups since the previous successful run." {
		t.Fatalf("Summary = %q, want no-new summary", result.Summary)
	}
}

func TestCalculateDigestLaterRunWithNewSignups(t *testing.T) {
	previous := WorkflowResult{LatestSignupCreatedAt: "2026-05-24T10:03:00Z"}

	result := CalculateDigest(DigestInput{
		RunID: "run_later",
		Signups: []Signup{
			{Created: "2026-05-24T10:00:00Z"},
			{Created: "2026-05-24T10:03:00Z"},
			{Created: "2026-05-24T10:08:00Z"},
			{Created: "2026-05-24T10:12:00Z"},
		},
		PreviousSuccessful: &previous,
		Started:            mustTime(t, "2026-05-24T11:00:00Z"),
		Finished:           mustTime(t, "2026-05-24T11:00:02Z"),
	})

	if result.TotalSignups != 4 || result.NewSignups != 2 {
		t.Fatalf("counts = total %d new %d, want 4/2", result.TotalSignups, result.NewSignups)
	}
	if result.LatestSignupCreatedAt != "2026-05-24T10:12:00Z" {
		t.Fatalf("LatestSignupCreatedAt = %q, want newest signup timestamp", result.LatestSignupCreatedAt)
	}
}

func TestWorkflowResultUnmarshalPocketBaseDates(t *testing.T) {
	body := []byte(`{
		"id":"abc123",
		"workflow":"digest",
		"run_id":"run_123",
		"status":"succeeded",
		"started_at":"2026-05-24 10:00:00.000Z",
		"finished_at":"2026-05-24 10:00:03.000Z",
		"total_signups":2,
		"new_signups":1,
		"latest_signup_created_at":"2026-05-24 09:59:00.000Z",
		"summary":"Digest found 2 total signups."
	}`)

	var result WorkflowResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal WorkflowResult: %v", err)
	}

	if result.FinishedAt.UTC().Format(time.RFC3339) != "2026-05-24T10:00:03Z" {
		t.Fatalf("FinishedAt = %s, want parsed PocketBase timestamp", result.FinishedAt.UTC().Format(time.RFC3339))
	}
}
