package app

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/massivemoose/ovek-workflow-example/internal/pocketbase"
	"github.com/massivemoose/ovek-workflow-example/internal/workflow"
)

type Store interface {
	CreateSignup(context.Context, string) error
	ListWorkflowResults(context.Context, int) ([]workflow.WorkflowResult, error)
}

func Routes(store Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleHome(store))
	mux.HandleFunc("POST /signup", handleSignup(store))
	mux.HandleFunc("GET /success", handleSuccess)
	mux.HandleFunc("GET /failure", handleFailure)
	mux.HandleFunc("GET /healthz", handleHealthz)
	return mux
}

func handleHome(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		results, err := store.ListWorkflowResults(ctx, 5)
		data := HomeData{WorkflowResults: results}
		if err != nil {
			data.ResultsError = "Workflow results are temporarily unavailable."
		}

		renderHome(w, data)
	}
}

func handleSignup(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectFailure(w, r, "form")
			return
		}

		email := strings.TrimSpace(r.FormValue("email"))
		if !validEmail(email) {
			redirectFailure(w, r, "invalid")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := store.CreateSignup(ctx, email); err != nil {
			switch {
			case errors.Is(err, pocketbase.ErrEmailAlreadySignedUp):
				redirectFailure(w, r, "duplicate")
			case errors.Is(err, pocketbase.ErrInvalidEmail):
				redirectFailure(w, r, "invalid")
			default:
				redirectFailure(w, r, "save")
			}
			return
		}

		http.Redirect(w, r, "/success", http.StatusSeeOther)
	}
}

func handleSuccess(w http.ResponseWriter, r *http.Request) {
	renderResult(w, resultData{
		Kind:    "success",
		Title:   "You're on the list.",
		Message: "Thanks for signing up.",
	})
}

func handleFailure(w http.ResponseWriter, r *http.Request) {
	data := resultData{
		Kind:    "error",
		Title:   "Whoops...",
		Message: "Try again in a moment.",
	}

	switch r.URL.Query().Get("reason") {
	case "form":
		data.Message = "The form could not be read. Try submitting it again."
	case "invalid":
		data.Message = "Enter a valid email address."
	case "duplicate":
		data.Message = "That email is already signed up."
	case "save":
		data.Message = "Something went wrong while saving the signup. Try again soon."
	}

	renderResult(w, data)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func redirectFailure(w http.ResponseWriter, r *http.Request, reason string) {
	http.Redirect(w, r, "/failure?reason="+reason, http.StatusSeeOther)
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	if addr.Address != email || strings.Contains(addr.Name, "@") {
		return false
	}

	local, domain, ok := strings.Cut(email, "@")
	return ok && local != "" && strings.Contains(domain, ".")
}
