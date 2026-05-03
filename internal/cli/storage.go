package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	storageCmd = &cobra.Command{
		Use:   "storage",
		Short: "Manage localStorage and sessionStorage",
	}

	storageLocalCmd = &cobra.Command{
		Use:   "local [key|set key value|clear]",
		Short: "Manage localStorage",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runStorageLocal,
	}

	storageSessionCmd = &cobra.Command{
		Use:   "session [key|set key value|clear]",
		Short: "Manage sessionStorage",
		Args:  cobra.MinimumNArgs(0),
		RunE:  runStorageSession,
	}
)

func init() {
	storageCmd.AddCommand(storageLocalCmd, storageSessionCmd)
	rootCmd.AddCommand(storageCmd)
}

func runStorageLocal(cmd *cobra.Command, args []string) error {
	return runStorageOp("localStorage", args)
}

func runStorageSession(cmd *cobra.Command, args []string) error {
	return runStorageOp("sessionStorage", args)
}

func runStorageOp(storageType string, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

	manager := browser.NewManager(cfg)
	if err := manager.Start(context.Background()); err != nil {
		return err
	}
	defer manager.Close()

	client, err := cdp.NewClientFromManager(manager)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(args) == 0 {
		var result map[string]string
		if err := client.Run(timeout, chromedp.Evaluate(
			fmt.Sprintf(`(function(){var o={};for(var i=0;i<%s.length;i++){var k=%s.key(i);o[k]=%s.getItem(k);}return o;})()`, storageType, storageType, storageType),
			&result,
		)); err != nil {
			return fmt.Errorf("failed to read %s: %w", storageType, err)
		}

		for k, v := range result {
			fmt.Printf("%s = %s\n", k, v)
		}
		return nil
	}

	if len(args) == 1 {
		var value string
		key := args[0]
		if err := client.Run(timeout, chromedp.Evaluate(
			fmt.Sprintf(`%s.getItem(%q)||''`, storageType, key),
			&value,
		)); err != nil {
			return fmt.Errorf("failed to get key: %w", err)
		}
		fmt.Println(value)
		return nil
	}

	if len(args) >= 2 && args[0] == "set" {
		key := args[1]
		val := ""
		if len(args) > 2 {
			val = args[2]
		}
		if err := client.Run(timeout, chromedp.Evaluate(
			fmt.Sprintf(`%s.setItem(%q,%q)`, storageType, key, val),
			nil,
		)); err != nil {
			return fmt.Errorf("failed to set item: %w", err)
		}
		fmt.Printf("%s[%s] set\n", storageType, key)
		return nil
	}

	if len(args) == 1 && args[0] == "clear" {
		if err := client.Run(timeout, chromedp.Evaluate(
			fmt.Sprintf(`%s.clear()`, storageType),
			nil,
		)); err != nil {
			return fmt.Errorf("failed to clear %s: %w", storageType, err)
		}
		fmt.Printf("%s cleared\n", storageType)
		return nil
	}

	return fmt.Errorf("invalid storage command")
}
