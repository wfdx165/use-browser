package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	cdpClient "github.com/wfdx165/use-browser/internal/cdp"
)

var (
	getTimeout int
)

var getTextCmd = &cobra.Command{
	Use:   "text <selector>",
	Short: "Get text content of an element",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetText,
}

var getHTMLCmd = &cobra.Command{
	Use:   "html <selector>",
	Short: "Get innerHTML of an element",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetHTML,
}

var getValueCmd = &cobra.Command{
	Use:   "value <selector>",
	Short: "Get value of an input element",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetValue,
}

var getTitleCmd = &cobra.Command{
	Use:   "title",
	Short: "Get page title",
	Args:  cobra.NoArgs,
	RunE:  runGetTitle,
}

var getURLCmd = &cobra.Command{
	Use:   "url",
	Short: "Get current page URL",
	Args:  cobra.NoArgs,
	RunE:  runGetURL,
}

var getCountCmd = &cobra.Command{
	Use:   "count <selector>",
	Short: "Count matching elements",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetCount,
}

var getAttrCmd = &cobra.Command{
	Use:   "attr <selector> <attribute>",
	Short: "Get attribute value of an element",
	Args:  cobra.ExactArgs(2),
	RunE:  runGetAttr,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get page or element information",
}

func init() {
	getCmd.PersistentFlags().IntVar(&getTimeout, "timeout", 0, "Timeout in ms")
	getCmd.AddCommand(getTextCmd, getHTMLCmd, getValueCmd, getTitleCmd, getURLCmd, getCountCmd, getAttrCmd)
	rootCmd.AddCommand(getCmd)
}

func setupBrowser() (*cdpClient.Client, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config not initialized")
	}

	ctx := context.Background()
	manager := browser.NewManager(cfg)
	if err := manager.Start(ctx); err != nil {
		return nil, nil, err
	}

	client, err := cdpClient.NewClientFromManager(manager)
	if err != nil {
		manager.Close()
		return nil, nil, err
	}

	return client, func() {
		client.Close()
		manager.Close()
	}, nil
}

func getTimeoutDur() time.Duration {
	timeout := cfg.DefaultTimeout
	if getTimeout > 0 {
		timeout = getTimeout
	}
	return time.Duration(timeout) * time.Millisecond
}

func runGetText(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var text string
	if err := client.Run(getTimeoutDur(), chromedp.Text(args[0], &text, chromedp.BySearch)); err != nil {
		return fmt.Errorf("get text failed: %w", err)
	}

	fmt.Println(strings.TrimSpace(text))
	return nil
}

func runGetHTML(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var text string
	if err := client.Run(getTimeoutDur(), chromedp.OuterHTML(args[0], &text, chromedp.BySearch)); err != nil {
		return fmt.Errorf("get html failed: %w", err)
	}

	fmt.Println(text)
	return nil
}

func runGetValue(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var value string
	if err := client.Run(getTimeoutDur(), chromedp.Value(args[0], &value, chromedp.BySearch)); err != nil {
		return fmt.Errorf("get value failed: %w", err)
	}

	fmt.Println(value)
	return nil
}

func runGetTitle(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var title string
	if err := client.Run(getTimeoutDur(), chromedp.Title(&title)); err != nil {
		return fmt.Errorf("get title failed: %w", err)
	}

	fmt.Println(title)
	return nil
}

func runGetURL(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var url string
	if err := client.Run(getTimeoutDur(), chromedp.Location(&url)); err != nil {
		return fmt.Errorf("get url failed: %w", err)
	}

	fmt.Println(url)
	return nil
}

func runGetCount(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var nodes []*cdp.Node
	if err := client.Run(getTimeoutDur(), chromedp.Nodes(args[0], &nodes, chromedp.ByQueryAll)); err != nil {
		return fmt.Errorf("get count failed: %w", err)
	}

	fmt.Println(len(nodes))
	return nil
}

func runGetAttr(cmd *cobra.Command, args []string) error {
	client, cleanup, err := setupBrowser()
	if err != nil {
		return err
	}
	defer cleanup()

	var value string
	if err := client.Run(getTimeoutDur(), chromedp.AttributeValue(args[0], args[1], &value, nil, chromedp.BySearch)); err != nil {
		return fmt.Errorf("get attr failed: %w", err)
	}

	fmt.Println(value)
	return nil
}
