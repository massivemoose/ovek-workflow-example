package pocketbase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/massivemoose/ovek-workflow-example/internal/config"
	"github.com/massivemoose/ovek-workflow-example/internal/workflow"
)

const (
	signupsCollection         = "signups"
	workflowResultsCollection = "workflow_results"
)

var (
	ErrEmailAlreadySignedUp = errors.New("email already signed up")
	ErrInvalidEmail         = errors.New("invalid email")
)

type Client struct {
	baseURL string
	email   string
	pass    string
	token   string
	client  *http.Client
}

type listResponse[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

type collection struct {
	Name   string  `json:"name"`
	Fields []field `json:"fields"`
}

type field struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"values,omitempty"`
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		baseURL: cfg.PocketBaseURL,
		email:   cfg.SuperuserEmail,
		pass:    cfg.SuperuserPass,
		token:   cfg.SuperuserToken,
		client:  &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (pb *Client) EnsureRequiredCollections(ctx context.Context) error {
	if err := pb.EnsureSignupsCollection(ctx); err != nil {
		return err
	}
	if err := pb.EnsureWorkflowResultsCollection(ctx); err != nil {
		return err
	}
	return nil
}

func (pb *Client) EnsureSignupsCollection(ctx context.Context) error {
	fields := []field{
		{Name: "email", Type: "email", Required: true},
		{Name: "source", Type: "text"},
	}
	body := map[string]any{
		"name":       signupsCollection,
		"type":       "base",
		"listRule":   nil,
		"viewRule":   nil,
		"createRule": nil,
		"updateRule": nil,
		"deleteRule": nil,
		"fields":     fields,
		"indexes": []string{
			"CREATE UNIQUE INDEX idx_signups_email ON signups (email)",
		},
	}
	return pb.ensureCollection(ctx, signupsCollection, fields, body)
}

func (pb *Client) EnsureWorkflowResultsCollection(ctx context.Context) error {
	fields := []field{
		{Name: "workflow", Type: "text", Required: true},
		{Name: "run_id", Type: "text"},
		{Name: "status", Type: "text", Required: true},
		{Name: "started_at", Type: "date"},
		{Name: "finished_at", Type: "date"},
		{Name: "total_signups", Type: "number"},
		{Name: "new_signups", Type: "number"},
		{Name: "latest_signup_created_at", Type: "text"},
		{Name: "summary", Type: "text"},
	}
	body := map[string]any{
		"name":       workflowResultsCollection,
		"type":       "base",
		"listRule":   nil,
		"viewRule":   nil,
		"createRule": nil,
		"updateRule": nil,
		"deleteRule": nil,
		"fields":     fields,
	}
	return pb.ensureCollection(ctx, workflowResultsCollection, fields, body)
}

func (pb *Client) ensureCollection(ctx context.Context, name string, requiredFields []field, createBody map[string]any) error {
	token, err := pb.authToken(ctx)
	if err != nil {
		return err
	}

	exists, err := pb.getCollection(ctx, token, name)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return pb.createCollection(ctx, token, name, createBody)
		}
		return err
	}

	return verifyFields(name, exists.Fields, requiredFields)
}

var errNotFound = errors.New("not found")

func (pb *Client) getCollection(ctx context.Context, token string, name string) (collection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pb.baseURL+"/api/collections/"+name, nil)
	if err != nil {
		return collection{}, err
	}
	req.Header.Set("Authorization", token)

	resp, err := pb.client.Do(req)
	if err != nil {
		return collection{}, fmt.Errorf("check collection %s: %w", name, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var c collection
		if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
			return collection{}, err
		}
		return c, nil
	case http.StatusNotFound:
		return collection{}, errNotFound
	default:
		return collection{}, fmt.Errorf("check collection %s returned HTTP %d", name, resp.StatusCode)
	}
}

func verifyFields(collectionName string, existing []field, required []field) error {
	byName := make(map[string]field, len(existing))
	for _, field := range existing {
		byName[field.Name] = field
	}

	var missing []string
	for _, field := range required {
		if _, ok := byName[field.Name]; !ok {
			missing = append(missing, field.Name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("collection %s is missing required fields: %s", collectionName, strings.Join(missing, ", "))
	}
	return nil
}

func (pb *Client) createCollection(ctx context.Context, token string, name string, body map[string]any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("create collection %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		if _, err := pb.getCollection(ctx, token, name); err == nil {
			return nil
		}
	}

	return fmt.Errorf("create collection %s returned HTTP %d", name, resp.StatusCode)
}

func (pb *Client) CreateSignup(ctx context.Context, email string) error {
	token, err := pb.authToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]string{
		"email":  email,
		"source": "ovek-example",
	}

	resp, err := pb.post(ctx, token, signupsCollection, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return SignupBadRequestError(body)
	default:
		return fmt.Errorf("create signup returned HTTP %d", resp.StatusCode)
	}
}

func SignupBadRequestError(body []byte) error {
	bodyText := strings.ToLower(string(body))
	switch {
	case strings.Contains(bodyText, "validation_invalid_email"),
		strings.Contains(bodyText, "invalid email"):
		return ErrInvalidEmail
	case strings.Contains(bodyText, "validation_not_unique"),
		strings.Contains(bodyText, "already"),
		strings.Contains(bodyText, "unique"),
		strings.Contains(bodyText, "idx_signups_email"):
		return ErrEmailAlreadySignedUp
	default:
		return fmt.Errorf("create signup returned HTTP %d", http.StatusBadRequest)
	}
}

func (pb *Client) ListSignups(ctx context.Context) ([]workflow.Signup, error) {
	token, err := pb.authToken(ctx)
	if err != nil {
		return nil, err
	}

	var all []workflow.Signup
	page := 1
	for {
		values := url.Values{}
		values.Set("page", fmt.Sprint(page))
		values.Set("perPage", "200")

		var list listResponse[workflow.Signup]
		if err := pb.getList(ctx, token, signupsCollection, values, &list); err != nil {
			return nil, err
		}
		all = append(all, list.Items...)
		if list.TotalPages == 0 || page >= list.TotalPages {
			break
		}
		page++
	}

	return all, nil
}

func (pb *Client) LatestSuccessfulWorkflowResult(ctx context.Context, workflowName string) (*workflow.WorkflowResult, error) {
	token, err := pb.authToken(ctx)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("page", "1")
	values.Set("perPage", "1")
	values.Set("sort", "-finished_at")
	values.Set("filter", fmt.Sprintf(`workflow = "%s" && status = "succeeded"`, workflowName))

	var list listResponse[workflow.WorkflowResult]
	if err := pb.getList(ctx, token, workflowResultsCollection, values, &list); err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, nil
	}
	return &list.Items[0], nil
}

func (pb *Client) ListWorkflowResults(ctx context.Context, limit int) ([]workflow.WorkflowResult, error) {
	token, err := pb.authToken(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}

	values := url.Values{}
	values.Set("page", "1")
	values.Set("perPage", fmt.Sprint(limit))
	values.Set("sort", "-finished_at")

	var list listResponse[workflow.WorkflowResult]
	if err := pb.getList(ctx, token, workflowResultsCollection, values, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (pb *Client) CreateWorkflowResult(ctx context.Context, result workflow.WorkflowResult) (string, error) {
	token, err := pb.authToken(ctx)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"workflow":                 result.Workflow,
		"run_id":                   result.RunID,
		"status":                   result.Status,
		"started_at":               result.StartedAt.UTC().Format(timeFormat),
		"finished_at":              result.FinishedAt.UTC().Format(timeFormat),
		"total_signups":            result.TotalSignups,
		"new_signups":              result.NewSignups,
		"latest_signup_created_at": result.LatestSignupCreatedAt,
		"summary":                  result.Summary,
	}

	resp, err := pb.post(ctx, token, workflowResultsCollection, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create workflow result returned HTTP %d", resp.StatusCode)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}
	return created.ID, nil
}

const timeFormat = "2006-01-02 15:04:05.000Z"

func (pb *Client) getList(ctx context.Context, token string, collectionName string, values url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pb.baseURL+"/api/collections/"+collectionName+"/records?"+values.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("list %s: %w", collectionName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return fmt.Errorf("list %s returned HTTP %d: %s", collectionName, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (pb *Client) post(ctx context.Context, token string, collectionName string, body any) (*http.Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections/"+collectionName+"/records", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create %s record: %w", collectionName, err)
	}
	return resp, nil
}

func (pb *Client) authToken(ctx context.Context) (string, error) {
	if strings.TrimSpace(pb.token) != "" {
		return pb.token, nil
	}
	if strings.TrimSpace(pb.email) == "" || strings.TrimSpace(pb.pass) == "" {
		return "", errors.New("set PB_SUPERUSER_TOKEN or PB_SUPERUSER_EMAIL and PB_SUPERUSER_PASSWORD")
	}

	body := map[string]string{
		"identity": pb.email,
		"password": pb.pass,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections/_superusers/auth-with-password", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("authenticate PocketBase superuser: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authenticate PocketBase superuser returned HTTP %d", resp.StatusCode)
	}

	var auth struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return "", err
	}
	if auth.Token == "" {
		return "", errors.New("PocketBase auth response did not include a token")
	}

	return auth.Token, nil
}
