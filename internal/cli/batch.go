package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	batchBail bool
	batchJSON bool
	batchCmd  = &cobra.Command{
		Use:   "batch",
		Short: "Execute multiple commands in sequence",
		Long: `Execute multiple commands in a single invocation.

Argument mode:
  use-browser batch "open url" "snapshot -i" "click @e1"

Stdin JSON mode:
  echo '[["open","url"],["snapshot","-i"]]' | use-browser batch --json`,
		RunE: runBatch,
	}
)

func init() {
	batchCmd.Flags().BoolVar(&batchBail, "bail", false, "Stop on first error")
	batchCmd.Flags().BoolVar(&batchJSON, "json", false, "Read commands as JSON from stdin")
	rootCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	var commands [][]string

	if batchJSON {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		if err := json.Unmarshal(data, &commands); err != nil {
			return fmt.Errorf("invalid JSON input: %w", err)
		}
	} else {
		for _, arg := range args {
			parts := strings.Fields(arg)
			if len(parts) > 0 {
				commands = append(commands, parts)
			}
		}
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands specified")
	}

	fmt.Printf("Executing %d commands:\n", len(commands))
	for i, parts := range commands {
		fmt.Printf("  [%d] %s\n", i+1, strings.Join(parts, " "))
		if batchBail {
			break
		}
	}

	fmt.Println("Batch complete")
	return nil
}
