// Package detect holds shared, provider-agnostic heuristics for recognizing
// rate limits, quota errors, context-window overflow, and connectivity
// failures in raw CLI output. Providers compose these helpers inside their
// Agent.DetectRateLimit implementations; the router itself never inspects
// provider output directly — it only calls the Agent interface.
package detect

import "regexp"

// Kind classifies a recognized failure signal in provider output.
type Kind int

const (
	// Unknown means no known failure pattern matched.
	Unknown Kind = iota
	// RateLimit covers 429s, "rate limit", "quota exceeded", and "daily limit".
	RateLimit
	// ContextOverflow covers "context exceeded" and similar context-window errors.
	ContextOverflow
	// ConnectionLost covers dropped connections and network errors.
	ConnectionLost
	// Unavailable covers upstream API/service unavailability.
	Unavailable
	// Timeout covers request or operation timeouts.
	Timeout
)

func (k Kind) String() string {
	switch k {
	case RateLimit:
		return "rate_limit"
	case ContextOverflow:
		return "context_overflow"
	case ConnectionLost:
		return "connection_lost"
	case Unavailable:
		return "unavailable"
	case Timeout:
		return "timeout"
	default:
		return "unknown"
	}
}

// patterns is ordered: the first Kind whose pattern matches wins, so more
// specific signals (context overflow) are checked before more general ones
// where wording could otherwise overlap.
var patterns = []struct {
	kind Kind
	re   *regexp.Regexp
}{
	{ContextOverflow, regexp.MustCompile(`(?i)context\s*(window\s*)?(length\s*)?exceeded|context\s+overflow|maximum\s+context\s+length`)},
	{RateLimit, regexp.MustCompile(`(?i)\b429\b|rate[\s-]?limit|quota\s+exceeded|daily\s+limit`)},
	{ConnectionLost, regexp.MustCompile(`(?i)connection\s+(lost|reset|refused)|econnreset|network\s+error`)},
	{Unavailable, regexp.MustCompile(`(?i)\b503\b|service\s+unavailable|api\s+unavailable`)},
	{Timeout, regexp.MustCompile(`(?i)\btimed?[\s-]?out\b|deadline\s+exceeded`)},
}

// Classify inspects a chunk of output and returns the first recognized
// failure Kind, or Unknown if none of the known patterns match.
func Classify(output string) Kind {
	for _, p := range patterns {
		if p.re.MatchString(output) {
			return p.kind
		}
	}
	return Unknown
}
