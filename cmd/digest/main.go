package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/massivemoose/ovek-workflow-example/internal/config"
	"github.com/massivemoose/ovek-workflow-example/internal/pocketbase"
	"github.com/massivemoose/ovek-workflow-example/internal/workflow"
)

func main() {
	cfg := config.LoadWorkflow(os.Getenv)
	runID := cfg.WorkflowRunID
	if runID == "" {
		runID = "local-" + time.Now().UTC().Format("20060102T150405Z")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	started := time.Now().UTC()
	pb := pocketbase.NewClient(cfg)

	log.Printf("digest run_id=%s starting", runID)

	if err := pb.EnsureRequiredCollections(ctx); err != nil {
		log.Fatalf("ensure PocketBase collections: %v", err)
	}

	resultID, result, err := runDigest(ctx, pb, runID, started)
	if err != nil {
		finished := time.Now().UTC()
		failed := workflow.WorkflowResult{
			Workflow:   workflow.DigestWorkflowName,
			RunID:      runID,
			Status:     "failed",
			StartedAt:  started,
			FinishedAt: finished,
			Summary:    err.Error(),
		}
		if id, writeErr := pb.CreateWorkflowResult(ctx, failed); writeErr != nil {
			log.Printf("write failed workflow result: %v", writeErr)
		} else {
			log.Printf("failed result_record_id=%s", id)
		}
		log.Fatalf("digest failed: %v", err)
	}

	log.Printf("digest run_id=%s status=%s total_signups=%d new_signups=%d latest_signup_created_at=%q result_record_id=%s",
		result.RunID,
		result.Status,
		result.TotalSignups,
		result.NewSignups,
		result.LatestSignupCreatedAt,
		resultID,
	)
}

type digestStore interface {
	ListSignups(context.Context) ([]workflow.Signup, error)
	LatestSuccessfulWorkflowResult(context.Context, string) (*workflow.WorkflowResult, error)
	CreateWorkflowResult(context.Context, workflow.WorkflowResult) (string, error)
}

func runDigest(ctx context.Context, store digestStore, runID string, started time.Time) (string, workflow.WorkflowResult, error) {
	signups, err := store.ListSignups(ctx)
	if err != nil {
		return "", workflow.WorkflowResult{}, err
	}

	previous, err := store.LatestSuccessfulWorkflowResult(ctx, workflow.DigestWorkflowName)
	if err != nil {
		return "", workflow.WorkflowResult{}, err
	}

	result := workflow.CalculateDigest(workflow.DigestInput{
		RunID:              runID,
		Signups:            signups,
		PreviousSuccessful: previous,
		Started:            started,
		Finished:           time.Now().UTC(),
	})

	id, err := store.CreateWorkflowResult(ctx, result)
	if err != nil {
		return "", workflow.WorkflowResult{}, err
	}

	return id, result, nil
}
