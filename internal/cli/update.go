package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update AIR to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkForUpdates(cmd.OutOrStdout(), version, checkOnly, "https://api.github.com/repos/levibmackay/air/releases/latest")
		},
	}

	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check for available updates without installing")

	return cmd
}

func checkForUpdates(w io.Writer, currentVer string, checkOnly bool, releaseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create update check request: %w", err)
	}
	req.Header.Set("User-Agent", "AIR-CLI/"+currentVer)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Current version: %s\nUnable to check for updates: %v\n", currentVer, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "Current version: %s\nNo releases found on remote repository.\n", currentVer)
		return nil
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVer, "v")

	if currentVer == "dev" {
		fmt.Fprintf(w, "AIR version: dev (latest released version: %s)\n", release.TagName)
		return nil
	}

	if current == latest {
		fmt.Fprintf(w, "AIR is up to date (%s).\n", currentVer)
		return nil
	}

	fmt.Fprintf(w, "A new version of AIR is available: %s (current: %s)\n", release.TagName, currentVer)
	if checkOnly {
		fmt.Fprintln(w, "Run 'air update' to upgrade.")
		return nil
	}

	fmt.Fprintf(w, "To update AIR, run:\n  go install github.com/levibmackay/air/cmd/air@latest\nOr view release at: %s\n", release.HTMLURL)
	return nil
}
