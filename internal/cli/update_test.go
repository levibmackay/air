package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateCheckDevVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.2.0","html_url":"https://github.com/levibmackay/air/releases/tag/v1.2.0"}`))
	}))
	defer ts.Close()

	var buf bytes.Buffer
	err := checkForUpdates(&buf, "dev", true, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "dev") || !strings.Contains(out, "v1.2.0") {
		t.Errorf("expected dev version output with latest version v1.2.0, got: %s", out)
	}
}

func TestUpdateCheckUpToDate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.2.0","html_url":"https://github.com/levibmackay/air/releases/tag/v1.2.0"}`))
	}))
	defer ts.Close()

	var buf bytes.Buffer
	err := checkForUpdates(&buf, "v1.2.0", false, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "AIR is up to date") {
		t.Errorf("expected up to date output, got: %s", out)
	}
}
