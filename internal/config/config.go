package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Headed             bool   `mapstructure:"headed"`
	CDP                string `mapstructure:"cdp"`
	AutoConnect        bool   `mapstructure:"autoConnect"`
	ExecutablePath     string `mapstructure:"executablePath"`
	State              string `mapstructure:"state"`
	Proxy              string `mapstructure:"proxy"`
	ProxyBypass        string `mapstructure:"proxyBypass"`
	IgnoreHTTPS        bool   `mapstructure:"ignoreHttpsErrors"`
	AllowFileAccess    bool   `mapstructure:"allowFileAccess"`
	UserAgent          string `mapstructure:"userAgent"`
	ColorScheme        string `mapstructure:"colorScheme"`
	DownloadPath       string `mapstructure:"downloadPath"`
	ScreenshotDir      string `mapstructure:"screenshotDir"`
	ScreenshotFormat   string `mapstructure:"screenshotFormat"`
	ScreenshotQuality  int    `mapstructure:"screenshotQuality"`
	Annotate           bool   `mapstructure:"annotate"`
	JSON               bool   `mapstructure:"json"`
	Verbose            bool   `mapstructure:"verbose"`
	Debug              bool   `mapstructure:"debug"`
	ConfigPath         string `mapstructure:"config"`
	DefaultTimeout     int    `mapstructure:"defaultTimeout"`
	AllowedDomains     string `mapstructure:"allowedDomains"`
	MaxOutput          int    `mapstructure:"maxOutput"`
	NoAutoDialog       bool   `mapstructure:"noAutoDialog"`
	ContentBoundaries  bool   `mapstructure:"contentBoundaries"`
}

func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".use-browser"
	}
	return filepath.Join(home, ".use-browser")
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("USE_BROWSER")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Note: Config file loading is disabled. Use environment variables only.
	// Example: USE_BROWSER_CDP=9222 use-browser open https://example.com

	cfg := &Config{
		Headed:             v.GetBool("headed"),
		CDP:                v.GetString("cdp"),
		AutoConnect:        v.GetBool("autoConnect"),
		ExecutablePath:     v.GetString("executablePath"),
		State:              v.GetString("state"),
		Proxy:              v.GetString("proxy"),
		ProxyBypass:        v.GetString("proxyBypass"),
		IgnoreHTTPS:        v.GetBool("ignoreHttpsErrors"),
		AllowFileAccess:    v.GetBool("allowFileAccess"),
		UserAgent:          v.GetString("userAgent"),
		ColorScheme:        v.GetString("colorScheme"),
		DownloadPath:       v.GetString("downloadPath"),
		ScreenshotDir:      v.GetString("screenshotDir"),
		ScreenshotFormat:   v.GetString("screenshotFormat"),
		ScreenshotQuality:  v.GetInt("screenshotQuality"),
		Annotate:           v.GetBool("annotate"),
		JSON:               v.GetBool("json"),
		Verbose:            v.GetBool("verbose"),
		Debug:              v.GetBool("debug"),
		DefaultTimeout:     v.GetInt("defaultTimeout"),
		AllowedDomains:     v.GetString("allowedDomains"),
		MaxOutput:          v.GetInt("maxOutput"),
		NoAutoDialog:       v.GetBool("noAutoDialog"),
		ContentBoundaries:  v.GetBool("contentBoundaries"),
	}

	// Set defaults if not provided via env
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 25000
	}
	if cfg.ScreenshotFormat == "" {
		cfg.ScreenshotFormat = "png"
	}
	if cfg.ScreenshotQuality == 0 {
		cfg.ScreenshotQuality = 100
	}
	if cfg.MaxOutput == 0 {
		cfg.MaxOutput = 50000
	}
	// Always headed (no headless mode)
	cfg.Headed = true

	return cfg, nil
}
