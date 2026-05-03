package cli

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
	"github.com/wfdx165/use-browser/internal/selector"
)

var (
	clickTimeout int
	clickDouble  bool
)

var clickCmd = &cobra.Command{
	Use:   "click <selector>",
	Short: "Click an element",
	Long: `Click an element by CSS selector or ref (@e1, @e2, ...).

Examples:
  use-browser click @e1           # Click by ref from snapshot
  use-browser click "#submit"     # Click by CSS selector
  use-browser click ".btn"        # Click by CSS class`,
	Args: cobra.ExactArgs(1),
	RunE: runClick,
}

func init() {
	clickCmd.Flags().IntVar(&clickTimeout, "timeout", 0, "Timeout in ms")
	clickCmd.Flags().BoolVar(&clickDouble, "double", false, "Double-click")
	rootCmd.AddCommand(clickCmd)
}

func runClick(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	input := args[0]

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if clickTimeout > 0 {
		timeout = time.Duration(clickTimeout) * time.Millisecond
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

	if selector.IsRef(input) {
		return fmt.Errorf("click by ref not yet supported, use CSS selector")
	}

	if clickDouble {
		if err := client.Run(timeout, chromedp.DoubleClick(input, chromedp.BySearch)); err != nil {
			return fmt.Errorf("double-click failed: %w", err)
		}
	} else {
		if err := client.Run(timeout, chromedp.Click(input, chromedp.BySearch)); err != nil {
			return fmt.Errorf("click failed: %w", err)
		}
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"selector": input}})
	} else {
		fmt.Printf("Clicked: %s\n", input)
	}
	return nil
}
