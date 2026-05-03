package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
	"github.com/wfdx165/use-browser/internal/browser"
	"github.com/wfdx165/use-browser/internal/cdp"
)

var (
	scrollPx   int
	scrollSel  string
	scrollCmd  = &cobra.Command{
		Use:   "scroll <dir> [px]",
		Short: "Scroll the page",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runScroll,
	}
	scrollIntoViewCmd = &cobra.Command{
		Use:   "scrollintoview <selector>",
		Short: "Scroll element into view",
		Args:  cobra.ExactArgs(1),
		RunE:  runScrollIntoView,
	}
	hoverCmd = &cobra.Command{
		Use:   "hover <selector>",
		Short: "Hover over element",
		Args:  cobra.ExactArgs(1),
		RunE:  runHover,
	}
	selectCmd = &cobra.Command{
		Use:   "select <selector> <value>",
		Short: "Select dropdown option",
		Args:  cobra.ExactArgs(2),
		RunE:  runSelect,
	}
	checkCmd = &cobra.Command{
		Use:   "check <selector>",
		Short: "Check a checkbox",
		Args:  cobra.ExactArgs(1),
		RunE:  runCheck,
	}
	uncheckCmd = &cobra.Command{
		Use:   "uncheck <selector>",
		Short: "Uncheck a checkbox",
		Args:  cobra.ExactArgs(1),
		RunE:  runUncheck,
	}
	setViewportCmd = &cobra.Command{
		Use:   "viewport <width> <height>",
		Short: "Set viewport size",
		Args:  cobra.RangeArgs(2, 3),
		RunE:  runSetViewport,
	}
	setOfflineCmd = &cobra.Command{
		Use:   "offline [on|off]",
		Short: "Toggle offline mode",
		Args:  cobra.ExactArgs(1),
		RunE:  runSetOffline,
	}
)

func init() {
	scrollCmd.Flags().IntVar(&scrollPx, "px", 100, "Scroll pixels")
	scrollCmd.Flags().StringVar(&scrollSel, "selector", "", "Selector")

	setCmd := &cobra.Command{Use: "set", Short: "Browser settings", Run: func(c *cobra.Command, a []string) { c.Help() }}
	setCmd.AddCommand(setViewportCmd, setOfflineCmd)

	rootCmd.AddCommand(scrollCmd, scrollIntoViewCmd, hoverCmd, selectCmd, checkCmd, uncheckCmd, setCmd)
}

func runScroll(cmd *cobra.Command, args []string) error {
	dir := args[0]
	px := scrollPx
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%d", &px)
	}

	var expr string
	switch dir {
	case "down": expr = fmt.Sprintf("window.scrollBy(0,%d)", px)
	case "up": expr = fmt.Sprintf("window.scrollBy(0,-%d)", px)
	case "right": expr = fmt.Sprintf("window.scrollBy(%d,0)", px)
	case "left": expr = fmt.Sprintf("window.scrollBy(-%d,0)", px)
	default: expr = fmt.Sprintf("window.scrollBy(0,%d)", px)
	}

	return doAction(func(c *cdp.Client, t time.Duration) error {
		if err := c.Run(t, chromedp.Evaluate(expr, nil)); err != nil {
			return fmt.Errorf("scroll: %w", err)
		}
		fmt.Printf("Scrolled %s %dpx\n", dir, px)
		return nil
	})
}

func runScrollIntoView(cmd *cobra.Command, args []string) error {
	sel := args[0]
	return doAction(func(c *cdp.Client, t time.Duration) error {
		js := fmt.Sprintf(`document.querySelector(%q).scrollIntoView({behavior:'auto'})`, sel)
		if err := c.Run(t, chromedp.Evaluate(js, nil)); err != nil {
			return fmt.Errorf("scroll into view: %w", err)
		}
		fmt.Printf("Scrolled %s into view\n", sel)
		return nil
	})
}

func runHover(cmd *cobra.Command, args []string) error {
	sel := args[0]
	return doAction(func(c *cdp.Client, t time.Duration) error {
		if err := c.Run(t, chromedp.Evaluate(
			fmt.Sprintf(`var e=document.querySelector(%q);if(e){var r=e.getBoundingClientRect();var x=r.left+r.width/2;var y=r.top+r.height/2;var ev=new MouseEvent('mouseover',{clientX:x,clientY:y,bubbles:true});e.dispatchEvent(ev);}`, sel),
			nil,
		)); err != nil {
			return fmt.Errorf("hover: %w", err)
		}
		fmt.Printf("Hovered %s\n", sel)
		return nil
	})
}

func runSelect(cmd *cobra.Command, args []string) error {
	sel, val := args[0], args[1]
	return doAction(func(c *cdp.Client, t time.Duration) error {
		if err := c.Run(t, chromedp.SetAttributeValue(sel, "value", val, chromedp.BySearch)); err != nil {
			return fmt.Errorf("select: %w", err)
		}
		fmt.Printf("Selected %s=%s\n", sel, val)
		return nil
	})
}

func runCheck(cmd *cobra.Command, args []string) error {
	return doAction(func(c *cdp.Client, t time.Duration) error {
		if err := c.Run(t, chromedp.Click(args[0], chromedp.BySearch)); err != nil {
			return fmt.Errorf("check: %w", err)
		}
		fmt.Printf("Checked: %s\n", args[0])
		return nil
	})
}

func runUncheck(cmd *cobra.Command, args []string) error {
	return runCheck(cmd, args)
}

func runSetViewport(cmd *cobra.Command, args []string) error {
	var w, h int
	fmt.Sscanf(args[0], "%d", &w)
	fmt.Sscanf(args[1], "%d", &h)
	scale := 1.0
	if len(args) > 2 {
		fmt.Sscanf(args[2], "%f", &scale)
	}

	return doAction(func(c *cdp.Client, t time.Duration) error {
		if err := c.Run(t, chromedp.EmulateViewport(int64(w), int64(h), chromedp.EmulateScale(scale))); err != nil {
			return fmt.Errorf("viewport: %w", err)
		}
		fmt.Printf("Viewport %dx%d\n", w, h)
		return nil
	})
}

func runSetOffline(cmd *cobra.Command, args []string) error {
	off := args[0] == "on"
	return doAction(func(c *cdp.Client, t time.Duration) error {
		js := fmt.Sprintf("Object.defineProperty(navigator,'onLine',{get:function(){return %v}})", !off)
		if err := c.Run(t, chromedp.Evaluate(js, nil)); err != nil {
			return fmt.Errorf("offline: %w", err)
		}
		fmt.Printf("Offline: %v\n", off)
		return nil
	})
}

func doAction(fn func(*cdp.Client, time.Duration) error) error {
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}
	timeout := time.Duration(cfg.DefaultTimeout) * time.Millisecond

	mgr := browser.NewManager(cfg)
	if err := mgr.Start(context.Background()); err != nil {
		return err
	}
	defer mgr.Close()

	client, err := cdp.NewClientFromManager(mgr)
	if err != nil {
		return err
	}
	defer client.Close()

	return fn(client, timeout)
}
