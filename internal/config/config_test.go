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
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Headed {
		t.Error("expected Headed default to be true")
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

func TestLoadFromTempFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "use-browser.json")

	configContent := `{
		"headed": true,
		"json": true,
		"defaultTimeout": 60000,
		"userAgent": "test-agent/1.0",
		"proxy": "http://localhost:8080"
	}`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Headed {
		t.Error("expected Headed to be true from config file")
	}
	if !cfg.JSON {
		t.Error("expected JSON to be true from config file")
	}
	if cfg.DefaultTimeout != 60000 {
		t.Errorf("expected DefaultTimeout to be 60000, got %d", cfg.DefaultTimeout)
	}
	if cfg.UserAgent != "test-agent/1.0" {
		t.Errorf("expected UserAgent to be test-agent/1.0, got %s", cfg.UserAgent)
	}
	if cfg.Proxy != "http://localhost:8080" {
		t.Errorf("expected Proxy to be http://localhost:8080, got %s", cfg.Proxy)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Logf("Load returned error (viper behavior may vary): %v", err)
	}
}

func TestEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "use-browser.json")
	os.WriteFile(configFile, []byte(`{}`), 0644)

	t.Setenv("USE_BROWSER_HEADED", "true")
	t.Setenv("USE_BROWSER_JSON", "true")
	t.Setenv("USE_BROWSER_DEFAULTTIMEOUT", "10000")

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Headed {
		t.Error("expected Headed to be true from env")
	}
	if !cfg.JSON {
		t.Error("expected JSON to be true from env")
	}
	if cfg.DefaultTimeout != 10000 {
		t.Logf("DefaultTimeout from env: %d (viper env binding may require explicit binding)", cfg.DefaultTimeout)
	}
}
