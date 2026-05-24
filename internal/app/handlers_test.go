package app

import "testing"

func TestValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "valid", email: "person@example.com", want: true},
		{name: "missing domain dot", email: "person@example", want: false},
		{name: "missing at", email: "person", want: false},
		{name: "display name", email: "Person <person@example.com>", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validEmail(test.email); got != test.want {
				t.Fatalf("validEmail(%q) = %v, want %v", test.email, got, test.want)
			}
		})
	}
}
