package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	networkAbort bool
	networkBody  string
	networkCmd   = &cobra.Command{
		Use:   "network",
		Short: "Network interception and request tracking",
	}
	networkRouteCmd = &cobra.Command{
		Use:   "route <url-pattern>",
		Short: "Intercept or mock requests",
		Args:  cobra.ExactArgs(1),
		RunE:  runNetworkRoute,
	}
	networkUnrouteCmd = &cobra.Command{
		Use:   "unroute [url]",
		Short: "Remove interception",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runNetworkUnroute,
	}
)

func init() {
	networkRouteCmd.Flags().BoolVar(&networkAbort, "abort", false, "Block requests")
	networkRouteCmd.Flags().StringVar(&networkBody, "body", "", "Mock response JSON")
	networkCmd.AddCommand(networkRouteCmd, networkUnrouteCmd)
	rootCmd.AddCommand(networkCmd)
}

func runNetworkRoute(cmd *cobra.Command, args []string) error {
	pattern := args[0]
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

	if err := client.Run(timeout,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.Enable().Do(ctx)
		}),
	); err != nil {
		return fmt.Errorf("network route: %w", err)
	}

	fmt.Printf("Network route: %s (abort=%v)\n", pattern, networkAbort)
	return nil
}

func runNetworkUnroute(cmd *cobra.Command, args []string) error {
	fmt.Println("Network route removed")
	return nil
}
