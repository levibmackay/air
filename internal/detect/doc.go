// Package detect holds shared heuristics for recognizing rate limits,
// quota errors, context-window overflow, and crashes in provider output.
// Providers compose these helpers inside their Agent.DetectRateLimit
// implementations; the router itself never inspects provider output
// directly.
package detect
