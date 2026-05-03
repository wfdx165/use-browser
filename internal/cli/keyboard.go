package cli

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var keyboardTimeout int

var keyboardTypeCmd = &cobra.Command{
	Use:   "type <text>",
	Short: "Type text with real keystrokes (no selector, current focus)",
	Args:  cobra.ExactArgs(1),
	RunE:  runKeyboardType,
}

var keyboardInsertCmd = &cobra.Command{
	Use:   "inserttext <text>",
	Short: "Insert text without key events (no selector)",
	Args:  cobra.ExactArgs(1),
	RunE:  runKeyboardInsert,
}

var keyboardCmd = &cobra.Command{
	Use:   "keyboard",
	Short: "Keyboard input without element selector",
}

func init() {
	keyboardCmd.PersistentFlags().IntVar(&keyboardTimeout, "timeout", 0, "Timeout in ms")
	keyboardCmd.AddCommand(keyboardTypeCmd, keyboardInsertCmd)
	rootCmd.AddCommand(keyboardCmd)
}

func keyboardTimeoutDur() time.Duration {
	timeout := cfg.DefaultTimeout
	if keyboardTimeout > 0 {
		timeout = keyboardTimeout
	}
	return time.Duration(timeout) * time.Millisecond
}

func runKeyboardType(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	text := args[0]

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

	if err := client.Run(keyboardTimeoutDur(), chromedp.KeyEvent(text)); err != nil {
		return fmt.Errorf("keyboard type failed: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"text": text}})
	} else {
		fmt.Printf("Typed: %s\n", text)
	}
	return nil
}

func runKeyboardInsert(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	text := args[0]

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

	// Set value on currently focused element
	js := fmt.Sprintf(`(function(){
		const el = document.activeElement;
		if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) {
			if (el.isContentEditable) {
				document.execCommand('insertText', false, %q);
			} else {
				el.value = %q;
			}
			return true;
		}
		return false;
	})()`, text, text)

	var ok bool
	if err := client.Run(keyboardTimeoutDur(), chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("keyboard insert failed: %w", err)
	}

	if !ok {
		return fmt.Errorf("no input element is currently focused")
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"text": text}})
	} else {
		fmt.Printf("Inserted: %s\n", text)
	}
	return nil
}
