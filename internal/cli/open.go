package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	openTimeout int
	openWaitFor string
)

var openCmd = &cobra.Command{
	Use:     "open [url]",
	Aliases: []string{"goto", "navigate"},
	Short:   "Launch or connect to browser, optionally navigate to URL",
	Long: `Launch a new browser instance or connect to an existing one.
If a URL is provided, navigate to it.

Examples:
  use-browser open                      # Launch browser on about:blank
  use-browser open https://example.com  # Launch and navigate
  use-browser goto https://example.com  # Alias for open`,
	RunE: runOpen,
}

func init() {
	openCmd.Flags().IntVar(&openTimeout, "timeout", 0, "Navigation timeout in ms")
	openCmd.Flags().StringVar(&openWaitFor, "wait-for", "", "Wait state: load, domcontentloaded, networkidle")

	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	ctx := SignalContext(context.Background())

	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

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

	timeout := cfg.DefaultTimeout
	if openTimeout > 0 {
		timeout = openTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	url := ""
	if len(args) > 0 {
		url = args[0]
	}

	if url == "" {
		if cfg.JSON {
			output := map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"url":   "about:blank",
					"title": "",
				},
			}
			printJSON(output)
		} else {
			fmt.Println("Browser launched on about:blank")
		}
		return nil
	}

	var title string
	tasks := []chromedp.Action{
		chromedp.Navigate(url),
	}

	switch openWaitFor {
	case "load":
		tasks = append(tasks, chromedp.WaitReady("body"))
	case "domcontentloaded":
		tasks = append(tasks, chromedp.WaitReady("body", chromedp.ByQuery))
	case "networkidle":
		tasks = append(tasks, chromedp.WaitIdle())
	}

	tasks = append(tasks, chromedp.Title(&title))

	if err := client.Run(ctx, tasks...); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"url":   url,
				"title": title,
			},
		}
		printJSON(output)
	} else {
		fmt.Printf("Navigated to %s\n", url)
		if title != "" {
			fmt.Printf("Title: %s\n", title)
		}
	}

	return nil
}

func printJSON(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}
