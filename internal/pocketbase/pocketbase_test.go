package pocketbase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/massivemoose/ovek-workflow-example/internal/config"
)

func TestSignupBadRequestError(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{name: "invalid email", body: `{"data":{"email":{"code":"validation_invalid_email"}}}`, want: ErrInvalidEmail},
		{name: "duplicate email", body: `{"data":{"email":{"code":"validation_not_unique"}}}`, want: ErrEmailAlreadySignedUp},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := SignupBadRequestError([]byte(test.body)); !errors.Is(got, test.want) {
				t.Fatalf("SignupBadRequestError(%q) = %v, want %v", test.body, got, test.want)
			}
		})
	}
}

func TestGetListErrorIncludesPocketBaseBody(t *testing.T) {
	pb := NewClient(config.Config{
		PocketBaseURL:  "http://pocketbase.test",
		SuperuserToken: "test-token",
	})
	pb.client = &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"message":"invalid sort field"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err := pb.getList(context.Background(), "test-token", "signups", url.Values{}, &listResponse[struct{}]{})
	if err == nil {
		t.Fatal("getList returned nil, want HTTP error")
	}

	got := err.Error()
	for _, want := range []string{"list signups returned HTTP 400", "invalid sort field"} {
		if !strings.Contains(got, want) {
			t.Fatalf("error = %q, want it to contain %q", got, want)
		}
	}
}

func TestListSignupsDoesNotDependOnPocketBaseSort(t *testing.T) {
	var gotQuery url.Values
	pb := NewClient(config.Config{
		PocketBaseURL:  "http://pocketbase.test",
		SuperuserToken: "test-token",
	})
	pb.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotQuery = r.URL.Query()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"page":1,
					"perPage":200,
					"totalItems":0,
					"totalPages":0,
					"items":[]
				}`)),
				Header: make(http.Header),
			}, nil
		}),
	}

	if _, err := pb.ListSignups(context.Background()); err != nil {
		t.Fatalf("ListSignups returned error: %v", err)
	}

	if got := gotQuery.Get("sort"); got != "" {
		t.Fatalf("sort query = %q, want empty because digest sorts signups in Go", got)
	}
}

func TestLatestSuccessfulWorkflowResultDoesNotSortByCreated(t *testing.T) {
	var gotQuery url.Values
	pb := NewClient(config.Config{
		PocketBaseURL:  "http://pocketbase.test",
		SuperuserToken: "test-token",
	})
	pb.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotQuery = r.URL.Query()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"page":1,
					"perPage":1,
					"totalItems":0,
					"totalPages":0,
					"items":[]
				}`)),
				Header: make(http.Header),
			}, nil
		}),
	}

	if _, err := pb.LatestSuccessfulWorkflowResult(context.Background(), "digest"); err != nil {
		t.Fatalf("LatestSuccessfulWorkflowResult returned error: %v", err)
	}

	if got := gotQuery.Get("sort"); got != "-finished_at" {
		t.Fatalf("sort query = %q, want -finished_at without unsupported -created fallback", got)
	}
}

func TestListWorkflowResultsDoesNotSortByCreated(t *testing.T) {
	var gotQuery url.Values
	pb := NewClient(config.Config{
		PocketBaseURL:  "http://pocketbase.test",
		SuperuserToken: "test-token",
	})
	pb.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotQuery = r.URL.Query()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"page":1,
					"perPage":5,
					"totalItems":0,
					"totalPages":0,
					"items":[]
				}`)),
				Header: make(http.Header),
			}, nil
		}),
	}

	if _, err := pb.ListWorkflowResults(context.Background(), 5); err != nil {
		t.Fatalf("ListWorkflowResults returned error: %v", err)
	}

	if got := gotQuery.Get("sort"); got != "-finished_at" {
		t.Fatalf("sort query = %q, want -finished_at without unsupported -created fallback", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
