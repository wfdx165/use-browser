package cdp

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
	"github.com/wfdx165/use-browser/internal/browser"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(cdpURL string) (*Client, error) {
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), cdpURL)
	return &Client{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func NewClientFromManager(m *browser.Manager) (*Client, error) {
	return NewClient(m.CDPURL())
}

func (c *Client) Ctx() context.Context {
	return c.ctx
}

func (c *Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *Client) Run(ctx context.Context, tasks ...chromedp.Action) error {
	if ctx == nil {
		ctx = c.ctx
	}

	if err := chromedp.Run(ctx, tasks...); err != nil {
		return fmt.Errorf("cdp run error: %w", err)
	}
	return nil
}
