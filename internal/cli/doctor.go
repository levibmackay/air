package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/levibmackay/air/internal/config"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose AIR's config and check which providers are installed and available",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Printf("config: FAIL (%v)\n", err)
				return err
			}
			fmt.Println("config: OK")
			fmt.Printf("  providers: %v\n", cfg.Providers)
			fmt.Printf("  checkpoint_interval: %s\n", cfg.CheckpointInterval)

			home, err := config.AirHome()
			if err != nil {
				fmt.Printf("air home: FAIL (%v)\n", err)
				return err
			}
			fmt.Printf("air home: %s\n", home)

			return printProviderStatus(cfg.Providers)
		},
	}
}
