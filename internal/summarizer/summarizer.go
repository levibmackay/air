// Package summarizer compresses a checkpoint into a concise resume prompt:
// completed work, remaining tasks, files changed, current git diff, known
// issues, and latest output — never the full conversation history. Provider
// plugins call Compress inside their Resume implementation to build the
// prompt they hand to their own CLI.
package summarizer

import (
	"fmt"
	"strings"

	"github.com/levibmackay/air/internal/checkpoint"
)

// maxOutputTail bounds how much of a checkpoint's raw terminal output is
// echoed back into the resume prompt, so a long-running session doesn't
// balloon the next provider's context.
const maxOutputTail = 2000

// Compress builds a resume prompt from a checkpoint. The result is meant to
// stand in for the full conversation history when handing a task to the
// next provider.
func Compress(cp *checkpoint.Checkpoint) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are resuming work on the following objective:\n%s\n", cp.Objective)

	if len(cp.CompletedWork) > 0 {
		b.WriteString("\nCompleted so far:\n")
		for _, item := range cp.CompletedWork {
			fmt.Fprintf(&b, "- %s\n", item)
		}
	}

	if len(cp.RemainingTasks) > 0 {
		b.WriteString("\nRemaining:\n")
		for _, item := range cp.RemainingTasks {
			fmt.Fprintf(&b, "- %s\n", item)
		}
	}

	if len(cp.ModifiedFiles) > 0 {
		b.WriteString("\nFiles modified so far:\n")
		for _, f := range cp.ModifiedFiles {
			fmt.Fprintf(&b, "- %s\n", f)
		}
	}

	if cp.GitDiff != "" {
		fmt.Fprintf(&b, "\nCurrent git diff:\n```diff\n%s\n```\n", cp.GitDiff)
	}

	if len(cp.Errors) > 0 {
		b.WriteString("\nKnown issues from the previous provider:\n")
		for _, e := range cp.Errors {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}

	if cp.ConversationSum != "" {
		fmt.Fprintf(&b, "\nSummary of prior progress:\n%s\n", cp.ConversationSum)
	}

	if tail := tailOf(cp.TerminalOutput, maxOutputTail); tail != "" {
		fmt.Fprintf(&b, "\nLatest output from the previous provider:\n%s\n", tail)
	}

	b.WriteString("\nPick up from here. Do not repeat completed work.\n")

	return b.String()
}

// tailOf returns the last n runes of s, so multi-byte output isn't cut
// mid-character.
func tailOf(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}
