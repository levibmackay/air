package detect

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   Kind
	}{
		{"429", "request failed with status 429", RateLimit},
		{"rate limit phrase", "You have hit the rate limit for this model", RateLimit},
		{"quota exceeded", "Error: quota exceeded for this billing period", RateLimit},
		{"daily limit", "You've reached your daily limit", RateLimit},
		{"context exceeded", "Error: context length exceeded", ContextOverflow},
		{"context overflow phrase", "context overflow, please start a new session", ContextOverflow},
		{"connection lost", "connection lost to server", ConnectionLost},
		{"econnreset", "read tcp: ECONNRESET", ConnectionLost},
		{"503", "upstream returned 503", Unavailable},
		{"service unavailable", "Service Unavailable", Unavailable},
		{"timeout", "the request timed out after 30s", Timeout},
		{"deadline exceeded", "context deadline exceeded", Timeout},
		{"unknown", "here is your generated code", Unknown},
		{"empty", "", Unknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Classify(tc.output); got != tc.want {
				t.Errorf("Classify(%q) = %v, want %v", tc.output, got, tc.want)
			}
		})
	}
}

func TestKindString(t *testing.T) {
	if Unknown.String() != "unknown" {
		t.Errorf("Unknown.String() = %q, want %q", Unknown.String(), "unknown")
	}
	if RateLimit.String() != "rate_limit" {
		t.Errorf("RateLimit.String() = %q, want %q", RateLimit.String(), "rate_limit")
	}
}
