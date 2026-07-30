// Package cost tracks elapsed time, and — where a provider reports it —
// token usage and estimated cost, per provider run within a session. None
// of AIR's current providers (internal/providers/claude, gemini) parse
// token/cost figures out of their CLI's output yet, so those fields stay
// zero until a provider plugin populates them; Entry and Ledger exist now so
// `air status` has a stable shape to report against as that lands.
package cost

import (
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
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

	current := Entry{
		Provider:  checkpoints[0].Provider,
		StartedAt: checkpoints[0].Created,
		EndedAt:   checkpoints[0].Created,
	}
	for _, cp := range checkpoints[1:] {
		if cp.Provider != current.Provider {
			ledger.Entries = append(ledger.Entries, current)
			current = Entry{Provider: cp.Provider, StartedAt: cp.Created}
		}
		current.EndedAt = cp.Created
	}
	ledger.Entries = append(ledger.Entries, current)

	return ledger
}
