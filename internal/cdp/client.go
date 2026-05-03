package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/target"
	"github.com/wfdx165/use-browser/internal/browser"
)

type Client struct {
	allocCtx context.Context
	allocCancel context.CancelFunc
	browserCtx context.Context
	browserCancel context.CancelFunc
}

func NewClient(cdpURL string) (*Client, error) {
	opts := []chromedp.ContextOption{
		chromedp.WithBrowserOption(
			chromedp.WithBrowserErrorf(func(string, ...interface{}) {}),
		),
	}

	targetID, err := findExistingPage(cdpURL)
	if err == nil && targetID != "" {
		opts = append(opts, chromedp.WithTargetID(target.ID(targetID)))
	}

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), cdpURL)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx, opts...)

	return &Client{
		allocCtx:     allocCtx,
		allocCancel:  allocCancel,
		browserCtx:   browserCtx,
		browserCancel: browserCancel,
	}, nil
}

func findExistingPage(cdpURL string) (string, error) {
	parsed, err := url.Parse(cdpURL)
	if err != nil || parsed.Host == "" {
		return "", err
	}

	host := parsed.Host
	if !strings.Contains(host, ":") {
		if parsed.Scheme == "wss" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	httpURL := fmt.Sprintf("http://%s/json/list", host)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(httpURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var targets []map[string]interface{}
	if err := json.Unmarshal(body, &targets); err != nil {
		return "", err
	}

	for _, t := range targets {
		if typ, _ := t["type"].(string); typ == "page" {
			if id, ok := t["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("no existing page target found")
}

func NewClientFromManager(m *browser.Manager) (*Client, error) {
	return NewClient(m.CDPURL())
}

func (c *Client) Ctx() context.Context {
	return c.browserCtx
}

func (c *Client) Close() error {
	return nil
}

var DefaultRunTimeout = 30 * time.Second

func (c *Client) Run(timeout time.Duration, tasks ...chromedp.Action) error {
	if timeout <= 0 {
		timeout = DefaultRunTimeout
	}
	taskCtx, cancel := context.WithTimeout(c.browserCtx, timeout)
	defer cancel()

	if err := chromedp.Run(taskCtx, tasks...); err != nil {
		return fmt.Errorf("cdp run error: %w", err)
	}
	return nil
}
