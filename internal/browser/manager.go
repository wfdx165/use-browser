package browser

import (
	"context"
	"fmt"
	"os"

	"github.com/wfdx165/use-browser/internal/config"
)

type Manager struct {
	config   *config.Config
	launcher *Launcher
	connector *Connector
	cdpURL   string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	if m.config.CDP != "" {
		m.connector = NewConnector(m.config)
		cdpURL, err := m.connector.Connect(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to chrome: %w", err)
		}
		m.cdpURL = cdpURL
		return nil
	}

	if m.config.AutoConnect {
		results, err := DiscoverChrome()
		if err != nil || len(results) == 0 {
			return fmt.Errorf("no running chrome found, start chrome with --remote-debugging-port or remove --auto-connect")
		}

		result := results[0]
		m.config.CDP = result.WebSocket
		m.connector = NewConnector(m.config)
		m.cdpURL = result.WebSocket
		return nil
	}

	m.launcher = NewLauncher(m.config)
	if cdpURL, ok := m.launcher.SavedCDPURL(); ok {
		m.cdpURL = cdpURL
		return nil
	}

	results, err := DiscoverChrome()
	if err == nil && len(results) > 0 {
		result := results[0]
		m.config.CDP = result.WebSocket
		m.connector = NewConnector(m.config)
		m.cdpURL = result.WebSocket
		return nil
	}

	fmt.Fprintln(os.Stderr, "WARN: Could not find running Chrome with debugging port. Starting a new browser instance. To reuse your current Chrome, restart it with --remote-debugging-port=9222.")

	cdpURL, err := m.launcher.Launch(ctx)
	if err != nil {
		return fmt.Errorf("failed to launch chrome: %w", err)
	}
	m.cdpURL = cdpURL
	return nil
}

func (m *Manager) CDPURL() string {
	return m.cdpURL
}

func (m *Manager) Close() error {
	return nil
}

func (m *Manager) IsSelfLaunched() bool {
	return m.launcher != nil
}
