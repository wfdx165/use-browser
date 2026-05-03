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
	frameCmd = &cobra.Command{
		Use:   "frame [selector|main]",
		Short: "Switch to iframe (or 'main')",
		Args:  cobra.ExactArgs(1),
		RunE:  runFrame,
	}
)

func init() {
	rootCmd.AddCommand(frameCmd)
}

func runFrame(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	input := args[0]
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

	if input == "main" {
		if err := client.Run(timeout,
			chromedp.Evaluate("window.frameElement=null", nil),
		); err != nil {
			return fmt.Errorf("frame main failed: %w", err)
		}
		fmt.Println("Switched to main frame")
		return nil
	}

	if err := client.Run(timeout,
		chromedp.Evaluate(
			fmt.Sprintf(`
(function(){
	var f = document.querySelector(%q);
	if(!f) return 'not found';
	if(f.tagName.toLowerCase()!=='iframe') return 'not an iframe';
	return 'frame selected';
})()`, input),
			nil,
		),
	); err != nil {
		return fmt.Errorf("frame switch failed: %w", err)
	}

	fmt.Printf("Frame selected: %s\n", input)
	return nil
}
