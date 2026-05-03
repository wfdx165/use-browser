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
	dialogAcceptCmd = &cobra.Command{
		Use:   "accept [text]",
		Short: "Accept dialog",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDialogAccept,
	}
	dialogCmd = &cobra.Command{
		Use:   "dialog",
		Short: "Handle JavaScript dialogs",
	}
)

func init() {
	dialogCmd.AddCommand(dialogAcceptCmd)
	rootCmd.AddCommand(dialogCmd)
}

func runDialogAccept(cmd *cobra.Command, args []string) error {
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
		chromedp.Evaluate("window.alert=function(){};window.confirm=function(){return true;};window.prompt=function(){return'';}", nil),
	); err != nil {
		return fmt.Errorf("dialog accept failed: %w", err)
	}

	fmt.Println("Dialogs set to auto-accept")
	return nil
}
