package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
	"github.com/wfdx165/use-browser/internal/snapshot"
)

var (
	snapshotInteractive bool
	snapshotURLs        bool
	snapshotCompact     bool
	snapshotDepth       int
	snapshotSelector    string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Get accessibility tree with refs",
	Long: `Get the accessibility tree of the current page with element refs (@e1, @e2, ...).

The refs can be used with other commands like click, fill, etc.

Examples:
  use-browser snapshot              # Full accessibility tree
  use-browser snapshot -i           # Interactive elements only
  use-browser snapshot -c           # Compact (no empty structural elements)
  use-browser snapshot -d 3         # Limit depth to 3 levels
  use-browser snapshot -s "#main"   # Scope to CSS selector
  use-browser snapshot --json       # JSON output`,
	RunE: runSnapshot,
}

func init() {
	snapshotCmd.Flags().BoolVarP(&snapshotInteractive, "interactive", "i", false, "Interactive elements only")
	snapshotCmd.Flags().BoolVarP(&snapshotURLs, "urls", "u", false, "Include link URLs")
	snapshotCmd.Flags().BoolVarP(&snapshotCompact, "compact", "c", false, "Remove empty structural elements")
	snapshotCmd.Flags().IntVarP(&snapshotDepth, "depth", "d", 0, "Limit tree depth")
	snapshotCmd.Flags().StringVarP(&snapshotSelector, "selector", "s", "", "Scope to CSS selector")

	rootCmd.AddCommand(snapshotCmd)
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	manager := browser.NewManager(cfg)
	if err := manager.Start(cmd.Context()); err != nil {
		return err
	}
	defer manager.Close()

	client, err := cdp.NewClientFromManager(manager)
	if err != nil {
		return err
	}
	defer client.Close()

	var snap *snapshot.Snapshot
	if err := client.Run(0, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		snap, err = snapshot.BuildTree(ctx, snapshotInteractive, snapshotCompact, snapshotDepth)
		return err
	})); err != nil {
		return fmt.Errorf("failed to build snapshot: %w", err)
	}

	if snap == nil {
		return fmt.Errorf("snapshot is nil")
	}

	if cfg.JSON {
		data, err := snapshot.FormatJSON(snap, snapshot.FormatOptions{
			InteractiveOnly: snapshotInteractive,
			URLs:            snapshotURLs,
		})
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		os.Stdout.Write(data)
		os.Stdout.Write([]byte("\n"))
	} else {
		output := snapshot.FormatText(snap, snapshot.FormatOptions{
			InteractiveOnly: snapshotInteractive,
			URLs:            snapshotURLs,
			Compact:         snapshotCompact,
			Depth:           snapshotDepth,
		})
		fmt.Print(output)
	}

	return nil
}
