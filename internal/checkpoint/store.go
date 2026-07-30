package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store persists checkpoints as one JSON file per checkpoint, under
// <baseDir>/<session>/<id>.json.
type Store struct {
	baseDir string
}

// NewStore returns a Store rooted at baseDir (typically ~/.air/checkpoints).
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// idFromTimestamp turns a checkpoint's Created time into a filesystem-safe,
// lexicographically-sortable ID.
func idFromTimestamp(cp *Checkpoint) string {
	return strings.NewReplacer(":", "-", ".", "-").Replace(cp.Created.UTC().Format("20060102T150405.000000000Z"))
}

// Save writes cp to disk, assigning it an ID derived from its Created time
// if it doesn't already have one.
func (s *Store) Save(cp *Checkpoint) error {
	if cp.Session == "" {
		return fmt.Errorf("checkpoint: Session is required")
	}
	if cp.ID == "" {
		cp.ID = idFromTimestamp(cp)
	}

	dir := filepath.Join(s.baseDir, cp.Session)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("checkpoint: create session dir: %w", err)
	}

	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}

	path := filepath.Join(dir, cp.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("checkpoint: write: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("checkpoint: finalize: %w", err)
	}
	return nil
}

// Load reads a single checkpoint by session and ID.
func (s *Store) Load(session, id string) (*Checkpoint, error) {
	path := filepath.Join(s.baseDir, session, id+".json")
	return loadFile(path)
}

func loadFile(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("checkpoint: read: %w", err)
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint: unmarshal %s: %w", path, err)
	}
	return &cp, nil
}

// List returns all checkpoints for a session, oldest first.
func (s *Store) List(session string) ([]*Checkpoint, error) {
	dir := filepath.Join(s.baseDir, session)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checkpoint: list %s: %w", session, err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	checkpoints := make([]*Checkpoint, 0, len(names))
	for _, name := range names {
		cp, err := loadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}
	return checkpoints, nil
}

// Latest returns the most recent checkpoint for a session, or nil if none
// exist.
func (s *Store) Latest(session string) (*Checkpoint, error) {
	checkpoints, err := s.List(session)
	if err != nil {
		return nil, err
	}
	if len(checkpoints) == 0 {
		return nil, nil
	}
	return checkpoints[len(checkpoints)-1], nil
}

// Sessions returns the IDs of every session with at least one checkpoint,
// most-recently-modified first.
func (s *Store) Sessions() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("checkpoint: list sessions: %w", err)
	}

	type sessionInfo struct {
		name    string
		modTime int64
	}
	var infos []sessionInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, sessionInfo{name: e.Name(), modTime: info.ModTime().UnixNano()})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].modTime > infos[j].modTime })

	sessions := make([]string, len(infos))
	for i, info := range infos {
		sessions[i] = info.name
	}
	return sessions, nil
}

// LatestAny returns the most recent checkpoint across all sessions, or nil
// if the store is empty. Used by `air resume` when no session is specified.
func (s *Store) LatestAny() (*Checkpoint, error) {
	sessions, err := s.Sessions()
	if err != nil {
		return nil, err
	}
	for _, session := range sessions {
		cp, err := s.Latest(session)
		if err != nil {
			return nil, err
		}
		if cp != nil {
			return cp, nil
		}
	}
	return nil, nil
}
