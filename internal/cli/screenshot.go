package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	screenshotFull    bool
	screenshotAnnotate bool
	screenshotFormat  string
	screenshotQuality int
	screenshotDir     string
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [output-path]",
	Short: "Take a screenshot of the current page",
	Long: `Take a screenshot of the browser viewport or full page.

Examples:
  use-browser screenshot                       # Save to temp dir as PNG
  use-browser screenshot page.png              # Save as PNG
  use-browser screenshot --full                # Full-page screenshot
  use-browser screenshot --format jpeg --quality 80
  use-browser screenshot --screenshot-dir ./shots`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScreenshot,
}

func init() {
	screenshotCmd.Flags().BoolVar(&screenshotFull, "full", false, "Full page screenshot")
	screenshotCmd.Flags().BoolVar(&screenshotAnnotate, "annotate", false, "Annotated screenshot with element labels")
	screenshotCmd.Flags().StringVar(&screenshotFormat, "format", "", "Screenshot format: png, jpeg (default: png)")
	screenshotCmd.Flags().IntVar(&screenshotQuality, "quality", 0, "JPEG quality 0-100")
	screenshotCmd.Flags().StringVar(&screenshotDir, "screenshot-dir", "", "Output directory")
	rootCmd.AddCommand(screenshotCmd)
}

func runScreenshot(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	format := screenshotFormat
	if format == "" {
		format = cfg.ScreenshotFormat
	}
	if format == "" {
		format = "png"
	}

	quality := screenshotQuality
	if quality <= 0 {
		quality = cfg.ScreenshotQuality
	}
	if quality <= 0 {
		quality = 100
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

	var buf []byte
	var tasks []chromedp.Action

	if screenshotFull {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, quality))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	if err := client.Run(timeout, tasks...); err != nil {
		return fmt.Errorf("screenshot failed: %w", err)
	}

	outputPath := ""
	if len(args) > 0 {
		outputPath = args[0]
	}

	if outputPath == "" {
		dir := screenshotDir
		if dir == "" {
			dir = cfg.ScreenshotDir
		}
		if dir == "" {
			dir = os.TempDir()
		}
		ext := ".png"
		if format == "jpeg" {
			ext = ".jpg"
		}
		outputPath = filepath.Join(dir, fmt.Sprintf("screenshot-%d%s", time.Now().UnixNano(), ext))
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("failed to write screenshot: %w", err)
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"path":   outputPath,
				"format": format,
			},
		})
	} else {
		fmt.Printf("Screenshot saved: %s\n", outputPath)
	}

	return nil
}
