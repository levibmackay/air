package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/levibmackay/air/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View AIR's effective configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			out, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("config: marshal: %w", err)
			}
			fmt.Print(string(out))
			return nil
		},
	}

	cmd.AddCommand(newConfigInitCmd())
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write AIR's default configuration to ~/.air/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := config.AirHome()
			if err != nil {
				return err
			}
			path := filepath.Join(home, "config.yaml")

			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config init: %s already exists", path)
			}

			out, err := yaml.Marshal(config.Default())
			if err != nil {
				return fmt.Errorf("config init: marshal: %w", err)
			}
			if err := os.WriteFile(path, out, 0o644); err != nil {
				return fmt.Errorf("config init: write %s: %w", path, err)
			}
			fmt.Printf("wrote %s\n", path)
			return nil
		},
	}
}
