package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// withHome points HOME at a temp dir for the duration of the test so Load
// never touches the real ~/.air, and returns to the working directory the
// test started in.
func withHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // windows
	return home
}

func withCwd(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestLoadDefaultsWhenNoConfigFiles(t *testing.T) {
	home := withHome(t)
	withCwd(t, home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0] != "claude" {
		t.Errorf("cfg.Providers = %v, want [claude]", cfg.Providers)
	}
	if cfg.CheckpointInterval != 2*time.Minute {
		t.Errorf("cfg.CheckpointInterval = %v, want 2m", cfg.CheckpointInterval)
	}
}

func TestLoadGlobalConfigOverridesDefaults(t *testing.T) {
	home := withHome(t)
	withCwd(t, home)

	airDir := filepath.Join(home, ".air")
	if err := os.MkdirAll(airDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := "providers:\n  - codex\n  - gemini\nretry_failed: false\ncheckpoint_interval: 5m\n"
	if err := os.WriteFile(filepath.Join(airDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Providers) != 2 || cfg.Providers[0] != "codex" || cfg.Providers[1] != "gemini" {
		t.Errorf("cfg.Providers = %v, want [codex gemini]", cfg.Providers)
	}
	if cfg.RetryFailed {
		t.Error("cfg.RetryFailed = true, want false from global config")
	}
	if cfg.CheckpointInterval != 5*time.Minute {
		t.Errorf("cfg.CheckpointInterval = %v, want 5m", cfg.CheckpointInterval)
	}
}

func TestLoadProjectConfigOverridesGlobal(t *testing.T) {
	home := withHome(t)

	airDir := filepath.Join(home, ".air")
	if err := os.MkdirAll(airDir, 0o755); err != nil {
		t.Fatal(err)
	}
	globalYAML := "providers:\n  - codex\ncheckpoint_interval: 5m\n"
	if err := os.WriteFile(filepath.Join(airDir, "config.yaml"), []byte(globalYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	project := t.TempDir()
	projectYAML := "providers:\n  - gemini\n"
	if err := os.WriteFile(filepath.Join(project, "air.yaml"), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	withCwd(t, project)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0] != "gemini" {
		t.Errorf("cfg.Providers = %v, want [gemini] (project override)", cfg.Providers)
	}
	if cfg.CheckpointInterval != 5*time.Minute {
		t.Errorf("cfg.CheckpointInterval = %v, want 5m (inherited from global)", cfg.CheckpointInterval)
	}
}

func TestValidateRejectsEmptyProviders(t *testing.T) {
	cfg := &Config{Providers: nil, CheckpointInterval: time.Minute}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() with empty Providers should error")
	}
}

func TestValidateRejectsNonPositiveInterval(t *testing.T) {
	cfg := &Config{Providers: []string{"claude"}, CheckpointInterval: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() with zero CheckpointInterval should error")
	}
}
