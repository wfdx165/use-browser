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
	consoleCmd = &cobra.Command{
		Use:   "console",
		Short: "View browser console messages",
		Args:  cobra.NoArgs,
		RunE:  runConsole,
	}
	errorsCmd = &cobra.Command{
		Use:   "errors",
		Short: "View page errors",
		Args:  cobra.NoArgs,
		RunE:  runErrors,
	}
	highlightCmd = &cobra.Command{
		Use:   "highlight <selector>",
		Short: "Highlight an element on the page",
		Args:  cobra.ExactArgs(1),
		RunE:  runHighlight,
	}
)

func init() {
	rootCmd.AddCommand(consoleCmd, errorsCmd, highlightCmd)
}

func runConsole(cmd *cobra.Command, args []string) error {
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

	if err := client.Run(timeout,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return runtime.Enable().Do(ctx)
		}),
	); err != nil {
		return fmt.Errorf("console: %w", err)
	}

	fmt.Println("Console messages enabled")
	return nil
}

func runErrors(cmd *cobra.Command, args []string) error {
	fmt.Println("No errors tracked")
	return nil
}

func runHighlight(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	sel := args[0]
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

	js := fmt.Sprintf(`
(function(){
	var el=document.querySelector(%q);
	if(!el)return'not found';
	el.style.outline='2px solid red';
	el.style.outlineOffset='2px';
	el.scrollIntoView({behavior:'auto',block:'center'});
	return'highlighted';
})()`, sel)

	var result string
	if err := client.Run(timeout, chromedp.Evaluate(js, &result)); err != nil {
		return fmt.Errorf("highlight failed: %w", err)
	}

	fmt.Println(result)
	return nil
}
