package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	tabLabel string
	tabCmd   = &cobra.Command{
		Use:   "tab [id]",
		Short: "Manage browser tabs",
		Long: `List, switch, open, or close tabs.

Examples:
  use-browser tab                    # List all tabs
  use-browser tab 2                  # Switch to tab 2
  use-browser tab new [url]          # Open new tab
  use-browser tab close [id]         # Close tab (current or by id)`,
		Args: cobra.ArbitraryArgs,
		RunE: runTab,
	}
	tabNewCmd = &cobra.Command{
		Use:   "new [url]",
		Short: "Open a new tab",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTabNew,
	}
	tabListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tabs",
		Args:  cobra.NoArgs,
		RunE:  runTabList,
	}
	tabCloseCmd = &cobra.Command{
		Use:   "close [id]",
		Short: "Close tab (current or by index)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTabClose,
	}
)

func init() {
	tabNewCmd.Flags().StringVar(&tabLabel, "label", "", "Tab label")
	tabCmd.AddCommand(tabNewCmd, tabListCmd, tabCloseCmd)
	rootCmd.AddCommand(tabCmd)
}

func runTab(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runTabList(cmd, args)
	}

	// Check if first arg is a number (tab index to switch)
	if n, err := strconv.Atoi(args[0]); err == nil {
		return runTabSwitch(cmd.Context(), n)
	}

	return fmt.Errorf("unknown tab command: %s", args[0])
}

func runTabSwitch(ctx context.Context, n int) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	if n < 1 {
		return fmt.Errorf("tab index must be >= 1")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

	manager := browser.NewManager(cfg)
	if err := manager.Start(ctx); err != nil {
		return err
	}
	defer manager.Close()

	client, err := cdp.NewClientFromManager(manager)
	if err != nil {
		return err
	}
	defer client.Close()

	// Get all targets
	targets, err := getTargets(ctx, client, timeout)
	if err != nil {
		return err
	}

	// Filter page targets
	var pageTargets []*target.Info
	for _, t := range targets {
		if t.Type == "page" {
			pageTargets = append(pageTargets, t)
		}
	}

	if n > len(pageTargets) {
		return fmt.Errorf("tab %d not found (only %d tabs)", n, len(pageTargets))
	}

	targetID := target.ID(pageTargets[n-1].TargetID)

	// Activate the target
	if err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		return target.ActivateTarget(targetID).Do(ctx)
	})); err != nil {
		return fmt.Errorf("tab switch: %w", err)
	}

	fmt.Printf("Switched to tab %d: %s\n", n, pageTargets[n-1].URL)
	return nil
}

func runTabNew(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

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

	url := "about:blank"
	if len(args) > 0 {
		url = args[0]
	}

	var targetID target.ID
	if err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		tid, err := target.CreateTarget(url).Do(ctx)
		if err != nil {
			return err
		}
		targetID = tid
		return nil
	})); err != nil {
		return fmt.Errorf("tab new: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{
			"success":  true,
			"data":     map[string]interface{}{"url": url, "targetId": targetID},
		})
	} else {
		fmt.Printf("New tab: %s (id: %s)\n", url, targetID)
	}
	return nil
}

func runTabList(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

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

	targets, err := getTargets(cmd.Context(), client, timeout)
	if err != nil {
		return err
	}

	// Filter and display page targets
	idx := 1
	for _, t := range targets {
		if t.Type == "page" {
			active := ""
			if t.Attached {
				active = " (active)"
			}
			fmt.Printf("  [%d] %s - %s%s\n", idx, t.Title, t.URL, active)
			idx++
		}
	}

	return nil
}

func runTabClose(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

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

	var targetID target.ID

	if len(args) > 0 {
		// Close specific tab by index
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("tab index must be a number: %s", args[0])
		}

		targets, err := getTargets(cmd.Context(), client, timeout)
		if err != nil {
			return err
		}

		var pageTargets []*target.Info
		for _, t := range targets {
			if t.Type == "page" {
				pageTargets = append(pageTargets, t)
			}
		}

		if n < 1 || n > len(pageTargets) {
			return fmt.Errorf("tab %d not found", n)
		}

		targetID = target.ID(pageTargets[n-1].TargetID)
	} else {
		// Close current/active tab
		targets, err := getTargets(cmd.Context(), client, timeout)
		if err != nil {
			return err
		}

		for _, t := range targets {
			if t.Type == "page" && t.Attached {
				targetID = target.ID(t.TargetID)
				break
			}
		}

		if targetID == "" {
			return fmt.Errorf("no active tab found")
		}
	}

	if err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		return target.CloseTarget(targetID).Do(ctx)
	})); err != nil {
		return fmt.Errorf("tab close: %w", err)
	}

	fmt.Println("Tab closed")
	return nil
}

func getTargets(ctx context.Context, client *cdp.Client, timeout time.Duration) ([]*target.Info, error) {
	var targets []*target.Info
	err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		res, err := target.GetTargets().Do(ctx)
		if err != nil {
			return err
		}
		targets = res
		return nil
	}))
	return targets, err
}
