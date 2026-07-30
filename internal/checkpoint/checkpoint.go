// Package checkpoint defines AIR's checkpoint format and JSON-file store.
// A checkpoint captures enough state for any provider to resume a task
// without replaying the full conversation history.
package checkpoint

import "time"

// Checkpoint is a snapshot of an in-progress AIR session, saved to disk so
// the session can be resumed by the same or a different provider.
type Checkpoint struct {
	ID       string    `json:"id"`
	Session  string    `json:"session"`
	Provider string    `json:"provider"`
	Created  time.Time `json:"created"`

	Objective       string   `json:"objective"`
	CompletedWork   []string `json:"completed_work"`
	RemainingTasks  []string `json:"remaining_tasks"`
	GitDiff         string   `json:"git_diff"`
	ModifiedFiles   []string `json:"modified_files"`
	TerminalOutput  string   `json:"terminal_output"`
	ConversationSum string   `json:"conversation_summary"`
	Errors          []string `json:"errors"`
}
