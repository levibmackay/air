// Package router is AIR's core orchestration loop: it launches providers,
// monitors their output, detects failures and completion, saves and loads
// checkpoints, and switches providers when one becomes unavailable. Router
// logic lands in Phase 2 and depends only on internal/agent's interface, so
// it never branches on a specific provider's identity.
package router
