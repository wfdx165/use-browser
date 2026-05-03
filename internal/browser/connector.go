package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wfdx165/use-browser/internal/config"
)

type Connector struct {
	config *config.Config
	cdpURL string
}

func NewConnector(cfg *config.Config) *Connector {
	return &Connector{
		config: cfg,
	}
}

func (c *Connector) Connect(ctx context.Context) (string, error) {
	cdpInput := c.config.CDP
	if cdpInput == "" {
		return "", fmt.Errorf("no CDP address specified, use --cdp flag")
	}

	if strings.HasPrefix(cdpInput, "ws://") || strings.HasPrefix(cdpInput, "wss://") {
		c.cdpURL = cdpInput
		return c.cdpURL, nil
	}

	port := cdpInput
	if !strings.Contains(port, ":") {
		port = "127.0.0.1:" + port
	}

	httpURL := fmt.Sprintf("http://%s/json/version", port)

	timeout := time.Duration(c.config.DefaultTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", httpURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to browser at %s: %w", httpURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var versionInfo map[string]interface{}
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		return "", fmt.Errorf("failed to parse version info: %w", err)
	}

	wsURL, ok := versionInfo["webSocketDebuggerUrl"].(string)
	if !ok || wsURL == "" {
		return "", fmt.Errorf("no webSocketDebuggerUrl in version response")
	}

	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse websocket URL: %w", err)
	}

	if parsedURL.Host == "" || parsedURL.Host == "localhost:0" {
		parsedURL.Host = strings.Split(port, "/")[0]
	}

	c.cdpURL = parsedURL.String()
	return c.cdpURL, nil
}

func (c *Connector) CDPURL() string {
	return c.cdpURL
}
