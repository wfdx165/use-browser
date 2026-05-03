package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/config"
)

var closeAll bool

var closeCmd = &cobra.Command{
	Use:     "close",
	Aliases: []string{"quit", "exit"},
	Short:   "Close browser",
	Long: `Close the browser instance.
If --session flag is set, closes only that session's browser.
If --all flag is set, closes all session browsers.`,
	RunE: runClose,
}

func init() {
	closeCmd.Flags().BoolVar(&closeAll, "all", false, "Close all active sessions")
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	configDir := config.DefaultConfigDir()
	pidDir := filepath.Join(configDir, "pids")

	if closeAll {
		entries, err := os.ReadDir(pidDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No active sessions")
				return nil
			}
			return fmt.Errorf("failed to read pid directory: %w", err)
		}

		for _, entry := range entries {
			if err := killSession(filepath.Join(pidDir, entry.Name())); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close session %s: %v\n", entry.Name(), err)
			}
		}
		fmt.Println("All sessions closed")
		return nil
	}

	sessionName := ""
	if cfg != nil && cfg.Session != "" {
		sessionName = cfg.Session
	}

	if sessionName == "" {
		sessionName = "default"
	}

	pidFile := filepath.Join(pidDir, sessionName+".pid")
	if err := killSession(pidFile); err != nil {
		return err
	}

	fmt.Println("Browser closed")
	return nil
}

func killSession(pidFile string) error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read pid file %s: %w", pidFile, err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
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
		os.Remove(pidFile)
		return nil
	case <-time.After(5 * time.Second):
		process.Kill()
		os.Remove(pidFile)
		return nil
	}
}
