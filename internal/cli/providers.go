package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/agent"
	"github.com/levibmackay/air/internal/config"
	"github.com/levibmackay/air/internal/providers"
)

func newProvidersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "providers",
		Short: "List configured providers and their availability",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			return printProviderStatus(cfg.Providers)
		},
	}
}

// printProviderStatus resolves each config entry through the registry and
// prints its installed/available status, one line per provider.
func printProviderStatus(entries []string) error {
	reg := providers.Registry()
	fmt.Println("providers:")
	for _, entry := range entries {
		key, param := agent.ParseProviderKey(entry)
		a, err := reg.Build(key, param)
		if err != nil {
			fmt.Printf("  %-20s unknown (%v)\n", entry, err)
			continue
		}

		installed := a.DetectInstalled()
		available, availErr := a.IsAvailable()

		status := "not installed"
		switch {
		case availErr != nil:
			status = fmt.Sprintf("error: %v", availErr)
		case available:
			status = "available"
		case installed:
			status = "installed, not available"
		}
		fmt.Printf("  %-20s %s\n", a.Name(), status)
	}
	return nil
}
