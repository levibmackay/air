package agent

import (
	"fmt"
	"strings"
)

// ParseProviderKey splits a config entry like "ollama:qwen3-coder:30b" into
// its registry key ("ollama") and parameter ("qwen3-coder:30b"). Entries
// with no colon (e.g. "claude") return an empty param.
func ParseProviderKey(entry string) (key, param string) {
	key, param, found := strings.Cut(entry, ":")
	if !found {
		return entry, ""
	}
	return key, param
}

// BuildProviders resolves a config's ordered provider entries into Agents
// via registry, preserving order. It fails fast if any entry names a
// provider the registry doesn't know about.
func BuildProviders(entries []string, registry *Registry) ([]Agent, error) {
	agents := make([]Agent, 0, len(entries))
	for _, entry := range entries {
		key, param := ParseProviderKey(entry)
		a, err := registry.Build(key, param)
		if err != nil {
			return nil, fmt.Errorf("provider entry %q: %w", entry, err)
		}
		agents = append(agents, a)
	}
	return agents, nil
}
