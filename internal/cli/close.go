package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/config"
)

var closeCmd = &cobra.Command{
	Use:     "close",
	Aliases: []string{"quit", "exit"},
	Short:   "Close browser",
	Long: `Close the browser instance.`,
	RunE: runClose,
}

func init() {
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	if err := killSession(); err != nil {
		return err
	}

	fmt.Println("Browser closed")
	return nil
}

func killSession() error {
	configDir := config.DefaultConfigDir()
	pidDir := filepath.Join(configDir, "pids")

	pidFile := filepath.Join(pidDir, "default.pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read pid file %s: %w", pidFile, err)
	}

	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid); err != nil {
		return fmt.Errorf("failed to parse pid from %s: %w", pidFile, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	process.Signal(syscall.SIGINT)

	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		process.Kill()
	}

	os.Remove(pidFile)
	os.Remove(filepath.Join(pidDir, "default.cdp"))

	return nil
}
