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

var cookiesCmd = &cobra.Command{
	Use:   "cookies [set key val|clear]",
	Short: "Manage browser cookies",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runCookies,
}

func init() {
	rootCmd.AddCommand(cookiesCmd)
}

func runCookies(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
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

	if len(args) >= 2 && args[0] == "set" {
		name := args[1]
		val := ""
		if len(args) > 2 {
			val = args[2]
		}

		js := fmt.Sprintf(`document.cookie=%q+'='+%q+';path=/;max-age=86400'`, name, val)
		if err := client.Run(timeout, chromedp.Evaluate(js, nil)); err != nil {
			return fmt.Errorf("failed to set cookie: %w", err)
		}
		fmt.Printf("Cookie %s set\n", name)
		return nil
	}

	if len(args) >= 1 && args[0] == "clear" {
		if err := client.Run(timeout, chromedp.Evaluate(`(function(){var cs=document.cookie.split(';');for(var i=0;i<cs.length;i++){var e=cs[i].indexOf('=');var n=e>-1?cs[i].substr(0,e):cs[i];document.cookie=n+'=;expires=Thu,01 Jan 1970 00:00:00 GMT;path=/';}})()`, nil)); err != nil {
			return fmt.Errorf("failed to clear cookies: %w", err)
		}
		fmt.Println("Cookies cleared")
		return nil
	}

	var cookieStr string
	if err := client.Run(timeout, chromedp.Evaluate(`document.cookie`, &cookieStr)); err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}

	if cookieStr == "" {
		if cfg.JSON {
			printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"cookies": []interface{}{}}})
		} else {
			fmt.Println("No cookies")
		}
		return nil
	}

	if cfg.JSON {
		var list []map[string]string
		pairs := strings.Split(cookieStr, ";")
		for _, p := range pairs {
			parts := strings.SplitN(strings.TrimSpace(p), "=", 2)
			if len(parts) >= 1 {
				entry := map[string]string{"name": parts[0]}
				if len(parts) == 2 {
					entry["value"] = parts[1]
				}
				list = append(list, entry)
			}
		}
		printJSON(map[string]interface{}{"success": true, "data": map[string]interface{}{"cookies": list}})
	} else {
		fmt.Println(cookieStr)
	}

	return nil
}
