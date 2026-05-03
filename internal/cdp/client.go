package cdp

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/wfdx165/use-browser/internal/browser"
)

type Client struct {
	allocCtx context.Context
	allocCancel context.CancelFunc
	browserCtx context.Context
	browserCancel context.CancelFunc
}

func NewClient(cdpURL string) (*Client, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), cdpURL)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	return &Client{
		allocCtx:     allocCtx,
		allocCancel:  allocCancel,
		browserCtx:   browserCtx,
		browserCancel: browserCancel,
	}, nil
}

func NewClientFromManager(m *browser.Manager) (*Client, error) {
	return NewClient(m.CDPURL())
}

func (c *Client) Ctx() context.Context {
	return c.browserCtx
}

func (c *Client) Close() error {
	if c.browserCancel != nil {
		c.browserCancel()
	}
	if c.allocCancel != nil {
		c.allocCancel()
	}
	return nil
}

func (c *Client) Run(timeout time.Duration, tasks ...chromedp.Action) error {
	taskCtx, cancel := context.WithTimeout(c.browserCtx, timeout)
	defer cancel()

	if err := chromedp.Run(taskCtx, tasks...); err != nil {
		return fmt.Errorf("cdp run error: %w", err)
	}
	return nil
}
