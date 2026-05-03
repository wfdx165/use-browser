package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	evalBase64  bool
	evalStdin   bool
)

var evalCmd = &cobra.Command{
	Use:   "eval <javascript>",
	Short: "Evaluate JavaScript in the page",
	Long: `Execute JavaScript code in the current page context.

Examples:
  use-browser eval "document.title"
  use-browser eval "window.innerWidth"
  use-browser eval -b "ZG9jdW1lbnQudGl0bGU="
  use-browser eval --stdin < script.js`,
	Args: cobra.MaximumNArgs(1),
	RunE: runEval,
}

func init() {
	evalCmd.Flags().BoolVarP(&evalBase64, "base64", "b", false, "Script is base64 encoded")
	evalCmd.Flags().BoolVar(&evalStdin, "stdin", false, "Read script from stdin")

	rootCmd.AddCommand(evalCmd)
}

func runEval(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	script := ""
	if evalStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		script = string(data)
	} else if len(args) > 0 {
		script = args[0]
	} else {
		return fmt.Errorf("javascript expression required")
	}

	if evalBase64 {
		decoded, err := base64.StdEncoding.DecodeString(script)
		if err != nil {
			return fmt.Errorf("invalid base64: %w", err)
		}
		script = string(decoded)
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

	var result string
	tasks := []chromedp.Action{
		chromedp.Evaluate(script, &result),
	}

	if err := client.Run(0, tasks...); err != nil {
		return fmt.Errorf("eval failed: %w", err)
	}

	if cfg.JSON {
		output := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"result": result,
			},
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(output)
	} else {
		fmt.Println(result)
	}

	return nil
}
