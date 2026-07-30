// Package cost tracks elapsed time, and — where a provider reports it —
// token usage and estimated cost, per provider run within a session.
package cost

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
)

var (
	reInputTokens  = regexp.MustCompile(`(?i)(?:input|prompt)\s*tokens?:?\s*([0-9,]+)`)
	reOutputTokens = regexp.MustCompile(`(?i)(?:output|completion)\s*tokens?:?\s*([0-9,]+)`)
	reCostUSD      = regexp.MustCompile(`(?i)(?:cost|estimated cost):?\s*\$([0-9]+\.[0-9]+)`)
)

// Entry records usage for one continuous run of a single provider within a
// session.
type Entry struct {
	Provider         string
	StartedAt        time.Time
	EndedAt          time.Time
	InputTokens      int
	OutputTokens     int
	EstimatedCostUSD float64
}

// Duration is how long this provider run lasted.
func (e Entry) Duration() time.Duration { return e.EndedAt.Sub(e.StartedAt) }

// Ledger aggregates Entries for a session.
type Ledger struct {
	Entries []Entry
}

// TotalDuration sums every Entry's Duration.
func (l *Ledger) TotalDuration() time.Duration {
	var total time.Duration
	for _, e := range l.Entries {
		total += e.Duration()
	}
	return total
}

// TotalTokens sums input and output tokens across every Entry.
func (l *Ledger) TotalTokens() (input, output int) {
	for _, e := range l.Entries {
		input += e.InputTokens
		output += e.OutputTokens
	}
	return input, output
}

// TotalCostUSD sums estimated cost across every Entry.
func (l *Ledger) TotalCostUSD() float64 {
	var total float64
	for _, e := range l.Entries {
		total += e.EstimatedCostUSD
	}
	return total
}

// ParseTokensAndCost extracts token counts and cost from provider output.
func ParseTokensAndCost(output string) (input, outputTokens int, costUSD float64) {
	if matches := reInputTokens.FindStringSubmatch(output); len(matches) > 1 {
		input, _ = strconv.Atoi(strings.ReplaceAll(matches[1], ",", ""))
	}
	if matches := reOutputTokens.FindStringSubmatch(output); len(matches) > 1 {
		outputTokens, _ = strconv.Atoi(strings.ReplaceAll(matches[1], ",", ""))
	}
	if matches := reCostUSD.FindStringSubmatch(output); len(matches) > 1 {
		costUSD, _ = strconv.ParseFloat(matches[1], 64)
	}
	return input, outputTokens, costUSD
}

// FromCheckpoints builds a Ledger from a session's checkpoint history
// (oldest first, as returned by checkpoint.Store.List). Consecutive
// checkpoints from the same provider collapse into a single Entry spanning
// their first and last timestamps, so a provider that ran for several
// checkpoint intervals before switching still shows as one run.
func FromCheckpoints(checkpoints []*checkpoint.Checkpoint) *Ledger {
	ledger := &Ledger{}
	if len(checkpoints) == 0 {
		return ledger
	}

	updateEntryMetrics := func(entry *Entry, cp *checkpoint.Checkpoint) {
		in, out, costVal := ParseTokensAndCost(cp.TerminalOutput)
		if in > entry.InputTokens {
			entry.InputTokens = in
		}
		if out > entry.OutputTokens {
			entry.OutputTokens = out
		}
		if costVal > entry.EstimatedCostUSD {
			entry.EstimatedCostUSD = costVal
		}
	}

	current := Entry{
		Provider:  checkpoints[0].Provider,
		StartedAt: checkpoints[0].Created,
		EndedAt:   checkpoints[0].Created,
	}
	updateEntryMetrics(&current, checkpoints[0])

	for _, cp := range checkpoints[1:] {
		if cp.Provider != current.Provider {
			ledger.Entries = append(ledger.Entries, current)
			current = Entry{Provider: cp.Provider, StartedAt: cp.Created}
		}
		current.EndedAt = cp.Created
		updateEntryMetrics(&current, cp)
	}
	ledger.Entries = append(ledger.Entries, current)

	return ledger
}
