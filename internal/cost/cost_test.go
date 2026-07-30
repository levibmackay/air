package cost

import (
	"testing"
	"time"

	"github.com/levibmackay/air/internal/checkpoint"
)

func at(seconds int) time.Time {
	return time.Date(2026, 7, 30, 12, 0, seconds, 0, time.UTC)
}

func TestFromCheckpointsEmpty(t *testing.T) {
	ledger := FromCheckpoints(nil)
	if len(ledger.Entries) != 0 {
		t.Errorf("Entries = %v, want empty", ledger.Entries)
	}
	if ledger.TotalDuration() != 0 {
		t.Errorf("TotalDuration() = %v, want 0", ledger.TotalDuration())
	}
}

func TestFromCheckpointsGroupsByProvider(t *testing.T) {
	checkpoints := []*checkpoint.Checkpoint{
		{Provider: "claude", Created: at(0), TerminalOutput: "Input tokens: 1,000\nOutput tokens: 500\nCost: $0.05"},
		{Provider: "claude", Created: at(60), TerminalOutput: "Input tokens: 2,000\nOutput tokens: 800\nCost: $0.10"},
		{Provider: "claude", Created: at(120)},
		{Provider: "gemini", Created: at(180)},
		{Provider: "gemini", Created: at(240)},
	}

	ledger := FromCheckpoints(checkpoints)
	if len(ledger.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(ledger.Entries))
	}

	claudeEntry := ledger.Entries[0]
	if claudeEntry.Provider != "claude" {
		t.Errorf("Entries[0].Provider = %q, want claude", claudeEntry.Provider)
	}
	if claudeEntry.Duration() != 120*time.Second {
		t.Errorf("claude Duration() = %v, want 120s", claudeEntry.Duration())
	}
	if claudeEntry.InputTokens != 2000 || claudeEntry.OutputTokens != 800 {
		t.Errorf("claude tokens = (%d, %d), want (2000, 800)", claudeEntry.InputTokens, claudeEntry.OutputTokens)
	}
	if claudeEntry.EstimatedCostUSD != 0.10 {
		t.Errorf("claude cost = %v, want 0.10", claudeEntry.EstimatedCostUSD)
	}

	geminiEntry := ledger.Entries[1]
	if geminiEntry.Provider != "gemini" {
		t.Errorf("Entries[1].Provider = %q, want gemini", geminiEntry.Provider)
	}
	if geminiEntry.Duration() != 60*time.Second {
		t.Errorf("gemini Duration() = %v, want 60s", geminiEntry.Duration())
	}

	if got := ledger.TotalDuration(); got != 180*time.Second {
		t.Errorf("TotalDuration() = %v, want 180s", got)
	}
}

func TestLedgerTotals(t *testing.T) {
	ledger := &Ledger{Entries: []Entry{
		{InputTokens: 100, OutputTokens: 50, EstimatedCostUSD: 0.10},
		{InputTokens: 200, OutputTokens: 75, EstimatedCostUSD: 0.20},
	}}

	inTok, outTok := ledger.TotalTokens()
	if inTok != 300 || outTok != 125 {
		t.Errorf("TotalTokens() = (%d, %d), want (300, 125)", inTok, outTok)
	}
	if got := ledger.TotalCostUSD(); got < 0.2999 || got > 0.3001 {
		t.Errorf("TotalCostUSD() = %v, want ~0.30", got)
	}
}

func TestParseTokensAndCost(t *testing.T) {
	sample := `
Processing task...
Prompt tokens: 1,500
Completion tokens: 350
Estimated cost: $0.045
`
	in, out, cost := ParseTokensAndCost(sample)
	if in != 1500 {
		t.Errorf("input tokens = %d, want 1500", in)
	}
	if out != 350 {
		t.Errorf("output tokens = %d, want 350", out)
	}
	if cost != 0.045 {
		t.Errorf("cost = %v, want 0.045", cost)
	}
}
