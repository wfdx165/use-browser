package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigDir(t *testing.T) {
	dir := DefaultConfigDir()
	if dir == "" {
		t.Fatal("DefaultConfigDir returned empty string")
	}
	if !filepath.IsAbs(dir) && dir != ".use-browser" {
		t.Errorf("expected absolute path or .use-browser, got %s", dir)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	for _, key := range []string{"USE_BROWSER_HEADED", "USE_BROWSER_JSON", "USE_BROWSER_DEFAULTTIMEOUT", "USE_BROWSER_SCREENSHOTFORMAT", "USE_BROWSER_SCREENSHOTQUALITY"} {
		os.Unsetenv(key)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Always headed (no headless mode)
	if !cfg.Headed {
		t.Error("expected Headed default to be true (forced headed mode)")
	}
	if cfg.AutoConnect {
		t.Error("expected AutoConnect default to be false")
	}
	if cfg.JSON {
		t.Error("expected JSON default to be false")
	}
	if cfg.Verbose {
		t.Error("expected Verbose default to be false")
	}
	if cfg.DefaultTimeout != 25000 {
		t.Errorf("expected DefaultTimeout to be 25000, got %d", cfg.DefaultTimeout)
	}
	if cfg.ScreenshotFormat != "png" {
		t.Errorf("expected ScreenshotFormat to be png, got %s", cfg.ScreenshotFormat)
	}
	if cfg.ScreenshotQuality != 100 {
		t.Errorf("expected ScreenshotQuality to be 100, got %d", cfg.ScreenshotQuality)
	}
}

func TestEnvOverride(t *testing.T) {
	// Clear any existing env vars
	for _, key := range []string{"USE_BROWSER_JSON", "USE_BROWSER_DEFAULTTIMEOUT", "USE_BROWSER_USERAGENT", "USE_BROWSER_PROXY"} {
		os.Unsetenv(key)
	}

	// Set environment variables
	t.Setenv("USE_BROWSER_JSON", "true")
	t.Setenv("USE_BROWSER_DEFAULTTIMEOUT", "10000")
	t.Setenv("USE_BROWSER_USERAGENT", "test-agent/1.0")
	t.Setenv("USE_BROWSER_PROXY", "http://localhost:8080")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Headed is always true (forced)
	if !cfg.Headed {
		t.Error("expected Headed to be true (forced headed mode)")
	}
	if !cfg.JSON {
		t.Error("expected JSON to be true from env")
	}
	if cfg.DefaultTimeout != 10000 {
		t.Errorf("expected DefaultTimeout to be 10000, got %d", cfg.DefaultTimeout)
	}
	if cfg.UserAgent != "test-agent/1.0" {
		t.Errorf("expected UserAgent to be test-agent/1.0, got %s", cfg.UserAgent)
	}
	if cfg.Proxy != "http://localhost:8080" {
		t.Errorf("expected Proxy to be http://localhost:8080, got %s", cfg.Proxy)
	}
}

func TestHeadedAlwaysTrue(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("USE_BROWSER_HEADED")

	// Try to set headed to false via environment (should be ignored)
	t.Setenv("USE_BROWSER_HEADED", "false")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Headed should always be true regardless of env setting
	if !cfg.Headed {
		t.Error("expected Headed to always be true (forced headed mode)")
	}
}

func TestCDPAndExecutablePathFromEnv(t *testing.T) {
	// Clear any existing env vars
	for _, key := range []string{"USE_BROWSER_CDP", "USE_BROWSER_EXECUTABLEPATH"} {
		os.Unsetenv(key)
	}

	// Set environment variables
	t.Setenv("USE_BROWSER_CDP", "9222")
	t.Setenv("USE_BROWSER_EXECUTABLEPATH", "/custom/browser/path")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CDP != "9222" {
		t.Errorf("expected CDP to be 9222, got %s", cfg.CDP)
	}
	if cfg.ExecutablePath != "/custom/browser/path" {
		t.Errorf("expected ExecutablePath to be /custom/browser/path, got %s", cfg.ExecutablePath)
	}
}
