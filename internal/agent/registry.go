package agent

import "fmt"

// Factory constructs an Agent, optionally parameterized (e.g. an Ollama
// model name parsed out of a config entry like "ollama:qwen3-coder:30b").
type Factory func(param string) (Agent, error)

// Registry maps a provider's config key to the Factory that builds it.
type Registry struct {
	factories map[string]Factory
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds a provider Factory under the given key. Registering the same
// key twice is a programmer error and panics, matching the pattern used by
// database/sql and similar stdlib registries.
func (r *Registry) Register(key string, factory Factory) {
	if _, exists := r.factories[key]; exists {
		panic(fmt.Sprintf("agent: provider %q already registered", key))
	}
	r.factories[key] = factory
}

// Build constructs the Agent registered under key, passing param through to
// its Factory (e.g. a model name for local-model providers).
func (r *Registry) Build(key, param string) (Agent, error) {
	factory, ok := r.factories[key]
	if !ok {
		return nil, fmt.Errorf("agent: no provider registered for %q", key)
	}
	return factory(param)
}

// Keys returns the provider keys currently registered.
func (r *Registry) Keys() []string {
	keys := make([]string, 0, len(r.factories))
	for k := range r.factories {
		keys = append(keys, k)
	}
	return keys
}
