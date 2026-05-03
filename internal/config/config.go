package config

import (
	"fmt"
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

	v.SetDefault("headed", true)
	v.SetDefault("autoConnect", false)
	v.SetDefault("ignoreHttpsErrors", false)
	v.SetDefault("allowFileAccess", false)
	v.SetDefault("screenshotFormat", "png")
	v.SetDefault("screenshotQuality", 100)
	v.SetDefault("annotate", false)
	v.SetDefault("json", false)
	v.SetDefault("verbose", false)
	v.SetDefault("debug", false)
	v.SetDefault("defaultTimeout", 25000)
	v.SetDefault("maxOutput", 50000)
	v.SetDefault("noAutoDialog", false)
	v.SetDefault("contentBoundaries", false)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		userConfigDir := DefaultConfigDir()
		v.AddConfigPath(userConfigDir)
		v.SetConfigName("config")

		wd, _ := os.Getwd()
		v.AddConfigPath(wd)
		v.SetConfigName("use-browser")
	}

	_ = v.ReadInConfig()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
