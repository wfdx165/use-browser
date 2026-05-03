package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	cdpClient "github.com/wfdx165/use-browser/internal/cdp"
)

var (
	waitTimeout int
	waitText    string
	waitURL     string
	waitLoad    string
)

var waitCmd = &cobra.Command{
	Use:   "wait [selector|ms]",
	Short: "Wait for an element, time, or condition",
	Long: `Wait for various conditions.

Examples:
  use-browser wait "#loading"       # Wait for element visible
  use-browser wait 5000             # Wait 5000ms
  use-browser wait --text "Welcome" # Wait for text
  use-browser wait --load networkidle`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWait,
}

func init() {
	waitCmd.Flags().IntVar(&waitTimeout, "timeout", 0, "Timeout in ms")
	waitCmd.Flags().StringVar(&waitText, "text", "", "Wait for text")
	waitCmd.Flags().StringVar(&waitURL, "url", "", "Wait for URL pattern")
	waitCmd.Flags().StringVar(&waitLoad, "load", "", "Load state: load, domcontentloaded, networkidle")
	rootCmd.AddCommand(waitCmd)
}

func runWait(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if waitTimeout > 0 {
		timeout = time.Duration(waitTimeout) * time.Millisecond
	}

	manager := browser.NewManager(cfg)
	if err := manager.Start(cmd.Context()); err != nil {
		return err
	}
	defer manager.Close()

	client, err := cdpClient.NewClientFromManager(manager)
	if err != nil {
		return err
	}
	defer client.Close()

	input := ""
	if len(args) > 0 {
		input = args[0]
	}

	if waitLoad != "" {
		switch waitLoad {
		case "load":
			return client.Run(timeout, chromedp.WaitReady("body"))
		case "domcontentloaded":
			return client.Run(timeout, chromedp.WaitReady("body"))
		case "networkidle":
			return client.Run(timeout, chromedp.WaitVisible("body"))
		}
	}

	if waitText != "" {
		return waitForTextContent(client, waitText, timeout)
	}

	if input != "" {
		if ms, err := parseMS(input); err == nil {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			if cfg.JSON {
				printJSON(map[string]interface{}{"success": true})
			} else {
				fmt.Printf("Waited %dms\n", ms)
			}
			return nil
		}
		return client.Run(timeout, chromedp.WaitVisible(input, chromedp.BySearch))
	}

	fmt.Println("No wait condition specified")
	return nil
}

func parseMS(s string) (int, error) {
	var ms int
	_, err := fmt.Sscanf(s, "%d", &ms)
	return ms, err
}

func waitForTextContent(client *cdpClient.Client, text string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var bodyText string
		if err := client.Run(5*time.Second, chromedp.Text("body", &bodyText)); err == nil {
			if strings.Contains(bodyText, text) {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for text %q", text)
}
