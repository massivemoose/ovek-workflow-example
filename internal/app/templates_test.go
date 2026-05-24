package app

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/massivemoose/ovek-workflow-example/internal/workflow"
)

func TestRenderHomeShowsWorkflowEmptyState(t *testing.T) {
	rec := httptest.NewRecorder()

	renderHome(rec, HomeData{})

	body := rec.Body.String()
	if !strings.Contains(body, "No workflow results yet.") {
		t.Fatalf("home body does not include workflow empty state:\n%s", body)
	}
	if strings.Contains(body, "PB_SUPERUSER_PASSWORD") {
		t.Fatal("home body exposed PocketBase credential environment names")
	}
}

func TestRenderHomeShowsRecentWorkflowResults(t *testing.T) {
	rec := httptest.NewRecorder()

	renderHome(rec, HomeData{
		WorkflowResults: []workflow.WorkflowResult{
			{
				Workflow:     "digest",
				RunID:        "run_1234567890abcdef",
				Status:       "succeeded",
				FinishedAt:   time.Date(2026, 5, 24, 10, 5, 0, 0, time.UTC),
				TotalSignups: 3,
				NewSignups:   1,
				Summary:      "Digest found 3 total signups, including 1 new since the previous successful run.",
			},
		},
	})

	body := rec.Body.String()
	for _, want := range []string{
		"Recent workflow results",
		"digest",
		"succeeded",
		"run_1234",
		"3 total",
		"1 new",
		"Digest found 3 total signups",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("home body missing %q:\n%s", want, body)
		}
	}
}
