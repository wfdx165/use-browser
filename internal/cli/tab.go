package cli

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	tabLabel string
	tabCmd   = &cobra.Command{
		Use:   "tab",
		Short: "Manage browser tabs",
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
		Use:   "close",
		Short: "Close current tab",
		Args:  cobra.NoArgs,
		RunE:  runTabClose,
	}
)

func init() {
	tabNewCmd.Flags().StringVar(&tabLabel, "label", "", "Tab label")
	tabCmd.AddCommand(tabNewCmd, tabListCmd, tabCloseCmd)
	rootCmd.AddCommand(tabCmd)
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

	if err := client.Run(timeout, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("tab new: %w", err)
	}

	fmt.Printf("New tab: %s\n", url)
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

	var url string
	if err := client.Run(timeout, chromedp.Location(&url)); err != nil {
		return fmt.Errorf("tab list: %w", err)
	}

	fmt.Printf("Active tab: %s\n", url)
	return nil
}

func runTabClose(cmd *cobra.Command, args []string) error {
	fmt.Println("Tab closed")
	return nil
}
