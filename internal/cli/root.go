package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/config"
	"github.com/wfdx165/use-browser/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "use-browser",
	Short: "Browser automation CLI for AI agents",
	Long: `use-browser is a browser automation CLI for AI agents.

It connects to an existing Chrome browser via Chrome DevTools Protocol (CDP)
or launches a new Chrome instance automatically.

Examples:
  use-browser open https://example.com
  use-browser snapshot
  use-browser click @e1
  use-browser fill @e2 "hello"
  use-browser screenshot
  use-browser close`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRunE: applyFlagOverrides,
}

var cfg *config.Config

func applyFlagOverrides(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		var err error
		cfg, err = config.Load(cfgConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	if cfgCDP != "" {
		cfg.CDP = cfgCDP
	}
	if cfgAutoConnect {
		cfg.AutoConnect = cfgAutoConnect
	}
	if cfgHeadless {
		cfg.Headed = false
	}
	if cfgHeaded {
		cfg.Headed = cfgHeaded
	}
	if cfgExecutablePath != "" {
		cfg.ExecutablePath = cfgExecutablePath
	}
	if cfgState != "" {
		cfg.State = cfgState
	}
	if cfgProxy != "" {
		cfg.Proxy = cfgProxy
	}
	if cfgIgnoreHTTPS {
		cfg.IgnoreHTTPS = cfgIgnoreHTTPS
	}
	if cfgAllowFileAccess {
		cfg.AllowFileAccess = cfgAllowFileAccess
	}
	if cfgUserAgent != "" {
		cfg.UserAgent = cfgUserAgent
	}
	if cfgColorScheme != "" {
		cfg.ColorScheme = cfgColorScheme
	}
	if cfgDownloadPath != "" {
		cfg.DownloadPath = cfgDownloadPath
	}
	if cfgScreenshotDir != "" {
		cfg.ScreenshotDir = cfgScreenshotDir
	}
	if cfgScreenshotFormat != "" {
		cfg.ScreenshotFormat = cfgScreenshotFormat
	}
	if cfgScreenshotQuality != 0 {
		cfg.ScreenshotQuality = cfgScreenshotQuality
	}
	if cfgAnnotate {
		cfg.Annotate = cfgAnnotate
	}
	if cfgJSON {
		cfg.JSON = cfgJSON
	}
	if cfgVerbose {
		cfg.Verbose = cfgVerbose
	}
	if cfgDebug {
		cfg.Debug = cfgDebug
	}

	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgCDP, "cdp", "", "CDP port or WebSocket URL")
	rootCmd.PersistentFlags().BoolVar(&cfgAutoConnect, "auto-connect", false, "Auto-discover running Chrome")
	rootCmd.PersistentFlags().BoolVar(&cfgHeaded, "headed", false, "Show browser window (default: true)")
	rootCmd.PersistentFlags().BoolVar(&cfgHeadless, "headless", false, "Run browser in headless mode")
	rootCmd.PersistentFlags().StringVar(&cfgExecutablePath, "executable-path", "", "Custom Chrome executable path")
	rootCmd.PersistentFlags().StringVar(&cfgState, "state", "", "Load storage state from JSON file")
	rootCmd.PersistentFlags().StringVar(&cfgProxy, "proxy", "", "Proxy server URL")
	rootCmd.PersistentFlags().BoolVar(&cfgIgnoreHTTPS, "ignore-https-errors", false, "Ignore HTTPS certificate errors")
	rootCmd.PersistentFlags().BoolVar(&cfgAllowFileAccess, "allow-file-access", false, "Allow file:// URL access")
	rootCmd.PersistentFlags().StringVar(&cfgUserAgent, "user-agent", "", "Custom User-Agent string")
	rootCmd.PersistentFlags().StringVar(&cfgColorScheme, "color-scheme", "", "Color scheme: dark, light, no-preference")
	rootCmd.PersistentFlags().StringVar(&cfgDownloadPath, "download-path", "", "Default download directory")
	rootCmd.PersistentFlags().StringVar(&cfgScreenshotDir, "screenshot-dir", "", "Default screenshot directory")
	rootCmd.PersistentFlags().StringVar(&cfgScreenshotFormat, "screenshot-format", "", "Screenshot format: png, jpeg")
	rootCmd.PersistentFlags().IntVar(&cfgScreenshotQuality, "screenshot-quality", 0, "JPEG quality 0-100")
	rootCmd.PersistentFlags().BoolVar(&cfgAnnotate, "annotate", false, "Annotated screenshot with numbered labels")
	rootCmd.PersistentFlags().BoolVar(&cfgJSON, "json", false, "JSON output")
	rootCmd.PersistentFlags().BoolVarP(&cfgVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&cfgConfigPath, "config", "", "Custom config file path")
	rootCmd.PersistentFlags().BoolVar(&cfgDebug, "debug", false, "Debug output")

	rootCmd.AddCommand(versionCmd)
}

var (
	cfgCDP              string
	cfgAutoConnect      bool
	cfgHeaded           bool
	cfgHeadless         bool
	cfgExecutablePath   string
	cfgState            string
	cfgProxy            string
	cfgIgnoreHTTPS      bool
	cfgAllowFileAccess  bool
	cfgUserAgent        string
	cfgColorScheme      string
	cfgDownloadPath     string
	cfgScreenshotDir    string
	cfgScreenshotFormat string
	cfgScreenshotQuality int
	cfgAnnotate         bool
	cfgJSON             bool
	cfgVerbose          bool
	cfgDebug            bool
	cfgConfigPath       string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s (commit: %s, built: %s)\n", version.Name, version.Version, version.Commit, version.Date)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func SignalContext(parent context.Context) context.Context {
	ctx, cancel := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}
