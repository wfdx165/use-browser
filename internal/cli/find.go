package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	findName  string
	findExact bool
)

var findCmd = &cobra.Command{
	Use:   "find <strategy> <value> [action] [action-value]",
	Short: "Find elements by semantic locators",
	Long: `Find elements by semantic properties and optionally perform an action.

Strategies: role, text, label, placeholder, alt, title, testid, first, last, nth
Actions: click, fill, type, hover, focus, check, uncheck, text

Examples:
  use-browser find role button click --name "Submit"
  use-browser find text "Sign In" click
  use-browser find label "Email" fill "test@test.com"
  use-browser find testid "login-btn" click
  use-browser find first ".item" click
  use-browser find nth 2 "a" text`,
	Args: cobra.MinimumNArgs(2),
	RunE: runFind,
}

func init() {
	findCmd.Flags().StringVar(&findName, "name", "", "Filter role by accessible name")
	findCmd.Flags().BoolVar(&findExact, "exact", false, "Require exact text match")
	rootCmd.AddCommand(findCmd)
}

func runFind(cmd *cobra.Command, args []string) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	strategy := args[0]

	var value, cssSel, action, actionValue string
	switch strategy {
	case "first", "last":
		if len(args) < 2 {
			return fmt.Errorf("find %s requires a selector", strategy)
		}
		cssSel = args[1]
		if len(args) > 2 {
			action = args[2]
		}
		if len(args) > 3 {
			actionValue = args[3]
		}
	case "nth":
		if len(args) < 3 {
			return fmt.Errorf("find nth requires: find nth <n> <selector>")
		}
		value = args[1] // the n
		cssSel = args[2]
		if len(args) > 3 {
			action = args[3]
		}
		if len(args) > 4 {
			actionValue = args[4]
		}
	default:
		// role, text, label, placeholder, alt, title, testid
		if len(args) < 2 {
			return fmt.Errorf("find %s requires a value", strategy)
		}
		value = args[1]
		if len(args) > 2 {
			action = args[2]
		}
		if len(args) > 3 {
			actionValue = args[3]
		}
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

	var sel string
	if strategy == "first" || strategy == "last" || strategy == "nth" {
		// CSS-based selection
		var nodes []map[string]interface{}
		if err := client.Run(timeout, chromedp.Evaluate(
			fmt.Sprintf(`(function(){const all=document.querySelectorAll(%q);return Array.from(all).map(el=>({tag:el.tagName.toLowerCase(),text:(el.textContent||'').trim().substring(0,80)}));})()`, cssSel),
			&nodes)); err != nil {
			return fmt.Errorf("find %s failed: %w", strategy, err)
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no element found with selector %q", cssSel)
		}

		idx := 0
		if strategy == "last" {
			idx = len(nodes) - 1
		} else if strategy == "nth" {
			n, _ := strconv.Atoi(value)
			if n < 1 || n > len(nodes) {
				return fmt.Errorf("nth %d out of range (found %d)", n, len(nodes))
			}
			idx = n - 1
		}
		sel = cssSel
		if strategy == "nth" {
			sel = fmt.Sprintf("%s:nth-of-type(%s)", cssSel, value)
		}

		// Print match info if no action
		if action == "" {
			if cfg.JSON {
				printJSON(map[string]interface{}{
					"success": true,
					"data": map[string]interface{}{
						"selector": sel,
						"index":    idx + 1,
						"total":    len(nodes),
						"match":    nodes[idx],
					},
				})
			} else {
				fmt.Printf("Found %d match(es), selected [%d]: <%s> %s\n", len(nodes), idx+1, nodes[idx]["tag"], nodes[idx]["text"])
			}
			return nil
		}
	} else {
		// Semantic selection via JS
		js := buildFindJS(strategy, value, findName, findExact)
		var results []map[string]interface{}
		if err := client.Run(timeout, chromedp.Evaluate(js, &results)); err != nil {
			return fmt.Errorf("find failed: %w", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("no element found with %s=%q", strategy, value)
		}

		first := results[0]
		var ok bool
		sel, ok = first["selector"].(string)
		if !ok || sel == "" {
			return fmt.Errorf("could not determine selector for match")
		}

		// Print match info if no action
		if action == "" {
			if cfg.JSON {
				printJSON(map[string]interface{}{"success": true, "data": results})
			} else {
				fmt.Printf("Found %d match(es)\n", len(results))
				for i, r := range results {
					fmt.Printf("  [%d] <%s> %s (selector: %s)\n", i+1, r["tag"], r["text"], r["selector"])
				}
			}
			return nil
		}
	}

	// Perform action
	if err := performFindAction(client, timeout, sel, action, actionValue); err != nil {
		return err
	}

	if cfg.JSON {
		printJSON(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"selector": sel,
				"action":   action,
			},
		})
	} else {
		fmt.Printf("Found and %s: %s\n", action, sel)
	}
	return nil
}

func buildFindJS(strategy, value, nameFilter string, exact bool) string {
	return fmt.Sprintf(`(function() {
		const strategy = %q;
		const value = %q;
		const nameFilter = %q;
		const exact = %v;
		function getName(el) {
			return el.getAttribute('aria-label') || 
				(el.getAttribute('aria-labelledby') && document.getElementById(el.getAttribute('aria-labelledby'))?.textContent) || 
				el.title || el.alt || el.placeholder || '';
		}
		function textMatch(el) {
			const t = (el.textContent || '').trim();
			if (!t) return false;
			return exact ? t === value : t.includes(value);
		}
		const all = document.querySelectorAll('*');
		const results = [];
		for (const el of all) {
			let m = false;
			if (strategy === 'role') {
				const role = el.getAttribute('role') || el.tagName.toLowerCase();
				m = role === value || (value === 'button' && el.tagName === 'BUTTON') || (value === 'link' && el.tagName === 'A') || (value === 'textbox' && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA'));
				if (m && nameFilter && !getName(el).includes(nameFilter)) m = false;
			} else if (strategy === 'text') {
				m = textMatch(el);
			} else if (strategy === 'label') {
				const label = document.querySelector('label[for="' + el.id + '"');
				m = label && (exact ? label.textContent.trim() === value : label.textContent.includes(value));
			} else if (strategy === 'placeholder') {
				m = el.placeholder && (exact ? el.placeholder === value : el.placeholder.includes(value));
			} else if (strategy === 'alt') {
				m = el.alt && (exact ? el.alt === value : el.alt.includes(value));
			} else if (strategy === 'title') {
				m = el.title && (exact ? el.title === value : el.title.includes(value));
			} else if (strategy === 'testid') {
				m = el.getAttribute('data-testid') === value;
			}
			if (m) {
				let sel = '';
				if (el.id) sel = '#' + el.id;
				else if (el.className && typeof el.className === 'string') sel = '.' + el.className.split(' ').slice(0,2).join('.');
				else sel = el.tagName.toLowerCase();
				results.push({tag: el.tagName.toLowerCase(), text: (el.textContent||'').trim().substring(0,80), name: getName(el).substring(0,80), selector: sel});
				if (results.length >= 5) break;
			}
		}
		return results;
	})()`, strategy, value, nameFilter, exact)
}

func performFindAction(client *cdp.Client, timeout time.Duration, sel, action, value string) error {
	var tasks []chromedp.Action
	switch action {
	case "click":
		tasks = append(tasks, chromedp.Click(sel, chromedp.BySearch))
	case "fill":
		tasks = append(tasks, chromedp.Clear(sel, chromedp.BySearch), chromedp.SendKeys(sel, value, chromedp.BySearch))
	case "type":
		tasks = append(tasks, chromedp.SendKeys(sel, value, chromedp.BySearch))
	case "hover":
		tasks = append(tasks, chromedp.ScrollIntoView(sel, chromedp.BySearch))
		tasks = append(tasks, chromedp.Query(sel, chromedp.BySearch))
		// Hover is approximated by scrolling into view and focusing
		// Full implementation would need mouse move to element center
	case "focus":
		tasks = append(tasks, chromedp.Focus(sel, chromedp.BySearch))
	case "check":
		tasks = append(tasks, chromedp.SetAttributeValue(sel, "checked", "true", chromedp.BySearch))
	case "uncheck":
		tasks = append(tasks, chromedp.RemoveAttribute(sel, "checked", chromedp.BySearch))
	case "text":
		var text string
		tasks = append(tasks, chromedp.Text(sel, &text, chromedp.BySearch))
		if err := client.Run(timeout, tasks...); err != nil {
			return fmt.Errorf("find %s failed: %w", action, err)
		}
		fmt.Println(text)
		return nil
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	if err := client.Run(timeout, tasks...); err != nil {
		return fmt.Errorf("find %s failed: %w", action, err)
	}
	return nil
}
