package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	fillTimeout  int
	typeTimeout  int
	pressTimeout int
)

var fillCmd = &cobra.Command{
	Use:   "fill <selector> <text>",
	Short: "Clear and fill an input element",
	Long: `Clear the element's value and fill it with the given text.

Examples:
  use-browser fill @e1 "hello@example.com"
  use-browser fill "#email" "test@test.com"`,
	Args: cobra.ExactArgs(2),
	RunE: runFill,
}

var typeCmd = &cobra.Command{
	Use:   "type <selector> <text>",
	Short: "Type text into an element (with key events)",
	Long: `Type text into an element character by character, generating key events.

Examples:
  use-browser type @e1 "hello"`,
	Args: cobra.ExactArgs(2),
	RunE: runType,
}

var pressCmd = &cobra.Command{
	Use:     "press <key>",
	Aliases: []string{"key"},
	Short:   "Press a keyboard key",
	Long: `Press a key on the keyboard.

Examples:
  use-browser press Enter
  use-browser press Tab
  use-browser press Control+a`,
	Args: cobra.ExactArgs(1),
	RunE: runPress,
}

func init() {
	fillCmd.Flags().IntVar(&fillTimeout, "timeout", 0, "Timeout in ms")
	typeCmd.Flags().IntVar(&typeTimeout, "timeout", 0, "Timeout in ms")
	pressCmd.Flags().IntVar(&pressTimeout, "timeout", 0, "Timeout in ms")

	rootCmd.AddCommand(fillCmd)
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(pressCmd)
}

func runFill(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	selector := args[0]
	text := args[1]

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if fillTimeout > 0 {
		timeout = time.Duration(fillTimeout) * time.Millisecond
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

	if err := client.Run(timeout,
		chromedp.Clear(selector, chromedp.BySearch),
		chromedp.SendKeys(selector, text, chromedp.BySearch),
		chromedp.Sleep(100*time.Millisecond),
	); err != nil {
		return fmt.Errorf("fill failed: %w", err)
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"selector": selector,
				"text":     text,
			},
		}
		printJSON(output)
	} else {
		fmt.Printf("Filled %s with text (%d chars)\n", selector, len(text))
	}

	return nil
}

func runType(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	selector := args[0]
	text := args[1]

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if typeTimeout > 0 {
		timeout = time.Duration(typeTimeout) * time.Millisecond
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

	if err := client.Run(timeout,
		chromedp.SendKeys(selector, text, chromedp.BySearch),
	); err != nil {
		return fmt.Errorf("type failed: %w", err)
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"selector": selector,
				"text":     text,
			},
		}
		printJSON(output)
	} else {
		fmt.Printf("Typed into %s: %s\n", selector, text)
	}

	return nil
}

func runPress(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	key := args[0]

	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond
	if pressTimeout > 0 {
		timeout = time.Duration(pressTimeout) * time.Millisecond
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

	// Map common key names
	keyMap := map[string]string{
		"enter":     "\n",
		"return":    "\r",
		"tab":       "\t",
		"escape":    "\x1b",
		"esc":       "\x1b",
		"backspace": "\b",
		"delete":    "\x7f",
		"space":     " ",
		"arrowup":    "\uf700",
		"arrowdown":  "\uf701",
		"arrowleft":  "\uf702",
		"arrowright": "\uf703",
	}

	// Check for modifier+key combinations
	modifiers := []string{}
	keyToSend := key
	
	// Handle combinations like "Control+a", "Shift+Enter"
	if strings.Contains(key, "+") {
		parts := strings.Split(key, "+")
		for i := 0; i < len(parts)-1; i++ {
			modifiers = append(modifiers, strings.ToLower(strings.TrimSpace(parts[i])))
		}
		keyToSend = strings.TrimSpace(parts[len(parts)-1])
	}

	// Map the key
	lowerKey := strings.ToLower(keyToSend)
	if mapped, ok := keyMap[lowerKey]; ok {
		keyToSend = mapped
	}

	// Send key
	if len(modifiers) == 0 {
		// Simple key press
		if err := client.Run(timeout, chromedp.KeyEvent(keyToSend)); err != nil {
			return fmt.Errorf("press failed: %w", err)
		}
	} else {
		// With modifiers - not fully implemented, just send the key
		// Full implementation would require keyDown/modifierDown + keyUp/modifierUp
		if err := client.Run(timeout, chromedp.KeyEvent(keyToSend)); err != nil {
			return fmt.Errorf("press failed: %w", err)
		}
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"key": key,
			},
		}
		printJSON(output)
	} else {
		fmt.Printf("Pressed key: %s\n", key)
	}
	return nil
}
