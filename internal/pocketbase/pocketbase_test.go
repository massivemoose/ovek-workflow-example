package pocketbase

import (
	"errors"
	"testing"
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
