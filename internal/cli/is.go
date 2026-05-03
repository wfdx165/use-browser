package cli

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	cdpClient "github.com/wfdx165/use-browser/internal/cdp"
)

var isTimeout int

var isVisibleCmd = &cobra.Command{
	Use:   "visible <selector>",
	Short: "Check if element is visible",
	Args:  cobra.ExactArgs(1),
	RunE:  runIsVisible,
}

var isEnabledCmd = &cobra.Command{
	Use:   "enabled <selector>",
	Short: "Check if element is enabled",
	Args:  cobra.ExactArgs(1),
	RunE:  runIsEnabled,
}

var isCheckedCmd = &cobra.Command{
	Use:   "checked <selector>",
	Short: "Check if checkbox is checked",
	Args:  cobra.ExactArgs(1),
	RunE:  runIsChecked,
}

var isCmd = &cobra.Command{
	Use:   "is",
	Short: "Check element state",
}

func init() {
	isCmd.PersistentFlags().IntVar(&isTimeout, "timeout", 0, "Timeout in ms")
	isCmd.AddCommand(isVisibleCmd, isEnabledCmd, isCheckedCmd)
	rootCmd.AddCommand(isCmd)
}

func isTimeoutDur() time.Duration {
	timeout := cfg.DefaultTimeout
	if isTimeout > 0 {
		timeout = isTimeout
	}
	return time.Duration(timeout) * time.Millisecond
}

func runIsVisible(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
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

	var visible bool
	js := fmt.Sprintf(`(function(){
		const el = document.querySelector(%q);
		if (!el) return false;
		const style = window.getComputedStyle(el);
		return style.display !== 'none' && style.visibility !== 'hidden' && style.opacity !== '0' && el.offsetParent !== null;
	})()`, args[0])

	if err := client.Run(isTimeoutDur(), chromedp.Evaluate(js, &visible)); err != nil {
		return fmt.Errorf("is visible failed: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"visible": visible}})
	} else {
		fmt.Println(visible)
	}
	return nil
}

func runIsEnabled(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
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

	var enabled bool
	js := fmt.Sprintf(`(function(){
		const el = document.querySelector(%q);
		return el ? !el.disabled : false;
	})()`, args[0])

	if err := client.Run(isTimeoutDur(), chromedp.Evaluate(js, &enabled)); err != nil {
		return fmt.Errorf("is enabled failed: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"enabled": enabled}})
	} else {
		fmt.Println(enabled)
	}
	return nil
}

func runIsChecked(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
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

	var checked bool
	js := fmt.Sprintf(`(function(){
		const el = document.querySelector(%q);
		return el ? el.checked : false;
	})()`, args[0])

	if err := client.Run(isTimeoutDur(), chromedp.Evaluate(js, &checked)); err != nil {
		return fmt.Errorf("is checked failed: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"checked": checked}})
	} else {
		fmt.Println(checked)
	}
	return nil
}
