package checkpoint

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStoreSaveLoad(t *testing.T) {
	store := NewStore(t.TempDir())

	cp := &Checkpoint{
		Session:   "sess-1",
		Provider:  "claude",
		Created:   time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
		Objective: "build a REST API",
	}

	if err := store.Save(cp); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if cp.ID == "" {
		t.Fatal("Save() did not assign an ID")
	}

	loaded, err := store.Load(cp.Session, cp.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Objective != cp.Objective {
		t.Errorf("loaded.Objective = %q, want %q", loaded.Objective, cp.Objective)
	}
	if loaded.Provider != cp.Provider {
		t.Errorf("loaded.Provider = %q, want %q", loaded.Provider, cp.Provider)
	}
}

func TestStoreSaveRequiresSession(t *testing.T) {
	store := NewStore(t.TempDir())
	err := store.Save(&Checkpoint{Objective: "no session set"})
	if err == nil {
		t.Fatal("Save() with empty Session should error")
	}
}

func TestStoreListAndLatest(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	base := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		cp := &Checkpoint{
			Session:   "sess-1",
			Provider:  "claude",
			Created:   base.Add(time.Duration(i) * time.Minute),
			Objective: "step",
		}
		if err := store.Save(cp); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	list, err := store.List("sess-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("List() returned %d checkpoints, want 3", len(list))
	}
	for i := 0; i < len(list)-1; i++ {
		if list[i].Created.After(list[i+1].Created) {
			t.Errorf("List() not sorted oldest-first at index %d", i)
		}
	}

	latest, err := store.Latest("sess-1")
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if !latest.Created.Equal(base.Add(2 * time.Minute)) {
		t.Errorf("Latest().Created = %v, want %v", latest.Created, base.Add(2*time.Minute))
	}
}

func TestStoreListMissingSession(t *testing.T) {
	store := NewStore(t.TempDir())
	list, err := store.List("does-not-exist")
	if err != nil {
		t.Fatalf("List() on missing session error = %v", err)
	}
	if list != nil {
		t.Errorf("List() on missing session = %v, want nil", list)
	}
}

func TestStoreLatestAny(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	empty, err := store.LatestAny()
	if err != nil {
		t.Fatalf("LatestAny() on empty store error = %v", err)
	}
	if empty != nil {
		t.Errorf("LatestAny() on empty store = %v, want nil", empty)
	}

	cp := &Checkpoint{Session: "sess-1", Provider: "claude", Created: time.Now().UTC(), Objective: "x"}
	if err := store.Save(cp); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	latest, err := store.LatestAny()
	if err != nil {
		t.Fatalf("LatestAny() error = %v", err)
	}
	if latest == nil || latest.Session != "sess-1" {
		t.Errorf("LatestAny() = %v, want session sess-1", latest)
	}
}

func TestIDFromTimestampIsFilesystemSafe(t *testing.T) {
	cp := &Checkpoint{Created: time.Date(2026, 7, 30, 12, 0, 0, 123456789, time.UTC)}
	id := idFromTimestamp(cp)
	if filepath.Base(id) != id {
		t.Errorf("idFromTimestamp() = %q contains path separators", id)
	}
}
