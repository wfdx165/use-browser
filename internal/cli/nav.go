package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	navTimeout int
)

var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Go back in history",
	RunE:  runBack,
}

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Go forward in history",
	RunE:  runForward,
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current page",
	RunE:  runReload,
}

func init() {
	backCmd.Flags().IntVar(&navTimeout, "timeout", 0, "Timeout in ms")
	forwardCmd.Flags().IntVar(&navTimeout, "timeout", 0, "Timeout in ms")
	reloadCmd.Flags().IntVar(&navTimeout, "timeout", 0, "Timeout in ms")

	rootCmd.AddCommand(backCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(reloadCmd)
}

func runBack(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if navTimeout > 0 {
		timeout = time.Duration(navTimeout) * time.Millisecond
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

	if err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, err := runtime.Evaluate("history.back()").Do(ctx)
		return err
	})); err != nil {
		return fmt.Errorf("back failed: %w", err)
	}

	fmt.Println("Navigated back")
	return nil
}

func runForward(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if navTimeout > 0 {
		timeout = time.Duration(navTimeout) * time.Millisecond
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

	if err := client.Run(timeout, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, err := runtime.Evaluate("history.forward()").Do(ctx)
		return err
	})); err != nil {
		return fmt.Errorf("forward failed: %w", err)
	}

	fmt.Println("Navigated forward")
	return nil
}

func runReload(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if navTimeout > 0 {
		timeout = time.Duration(navTimeout) * time.Millisecond
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

	if err := client.Run(timeout, chromedp.Reload()); err != nil {
		return fmt.Errorf("reload failed: %w", err)
	}

	fmt.Println("Page reloaded")
	return nil
}
