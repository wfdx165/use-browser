package browser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type DiscoverResult struct {
	Port      int    `json:"port"`
	WebSocket string `json:"webSocketDebuggerUrl"`
	Version   string `json:"browser"`
	PID       int    `json:"pid"`
}

var commonPorts = []int{9222, 9223, 9229, 9333}

// browserProfiles defines the user data directories for different CDP-compatible browsers
type browserProfile struct {
	name    string
	paths   map[string][]string // os -> possible profile directories
}

var browserProfiles = []browserProfile{
	{
		name: "Chrome",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/Google/Chrome"},
			"linux":  {".config/google-chrome", ".config/chrome"},
			"windows": {"Google/Chrome/User Data"},
		},
	},
	{
		name: "Edge",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/Microsoft Edge"},
			"linux":  {".config/microsoft-edge", ".config/microsoft-edge-stable"},
			"windows": {"Microsoft/Edge/User Data"},
		},
	},
	{
		name: "Brave",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/BraveSoftware/Brave-Browser"},
			"linux":  {".config/BraveSoftware/Brave-Browser", ".config/brave"},
			"windows": {"BraveSoftware/Brave-Browser/User Data"},
		},
	},
	{
		name: "Opera",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/com.operasoftware.Opera"},
			"linux":  {".config/opera", ".var/app/com.opera.Opera/config/opera"},
			"windows": {"Opera Software/Opera Stable"},
		},
	},
	{
		name: "Vivaldi",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/Vivaldi"},
			"linux":  {".config/vivaldi", ".config/vivaldi-stable"},
			"windows": {"Vivaldi/User Data"},
		},
	},
	{
		name: "Chromium",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/Chromium"},
			"linux":  {".config/chromium", ".config/chrome-unstable"},
			"windows": {"Chromium/User Data"},
		},
	},
	{
		name: "Arc",
		paths: map[string][]string{
			"darwin": {"Library/Application Support/Arc"},
		},
	},
}

// DiscoverBrowser searches for running CDP-compatible browsers
// This function is kept for backward compatibility, use DiscoverBrowsers for comprehensive search
func DiscoverBrowser() ([]DiscoverResult, error) {
	return DiscoverBrowsers()
}

// DiscoverBrowsers searches for running CDP-compatible browsers on common ports
func DiscoverBrowsers() ([]DiscoverResult, error) {
	var results []DiscoverResult
	foundPorts := make(map[int]bool)

	// First, try to discover from DevToolsActivePort files
	for _, result := range discoverDevToolsPorts() {
		if !foundPorts[result.Port] {
			results = append(results, result)
			foundPorts[result.Port] = true
		}
	}

	// Then probe common ports
	for _, port := range commonPorts {
		if foundPorts[port] {
			continue
		}
		if r := probePort(port); r != nil {
			results = append(results, *r)
			foundPorts[port] = true
		}
	}

	return results, nil
}

// Deprecated: Use DiscoverBrowsers instead
func DiscoverChrome() ([]DiscoverResult, error) {
	return DiscoverBrowsers()
}

func discoverDevToolsPorts() []DiscoverResult {
	var results []DiscoverResult
	home, _ := os.UserHomeDir()

	for _, profile := range browserProfiles {
		if paths, ok := profile.paths[runtime.GOOS]; ok {
			for _, profilePath := range paths {
				filePath := profilePath
				if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
					filePath = filepath.Join(home, profilePath, "DevToolsActivePort")
				} else if runtime.GOOS == "windows" {
					filePath = filepath.Join(os.Getenv("LOCALAPPDATA"), profilePath, "DevToolsActivePort")
				}

				if result := readDevToolsPortFile(filePath); result != nil {
					results = append(results, *result)
				}
			}
		}
	}

	return results
}

func readDevToolsPortFile(filePath string) *DiscoverResult {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return nil
	}

	port, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil
	}

	return probePort(port)
}

// Deprecated: Use discoverDevToolsPorts instead
func discoverDevToolsPort() *DiscoverResult {
	var filePath string

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, "Library/Application Support/Google/Chrome/DevToolsActivePort")
	case "linux":
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, ".config/google-chrome/DevToolsActivePort")
	case "windows":
		filePath = filepath.Join(os.Getenv("LOCALAPPDATA"), "Google/Chrome/User Data/DevToolsActivePort")
	}

	if filePath == "" {
		return nil
	}

	return readDevToolsPortFile(filePath)
}

func ProbePort(port int) *DiscoverResult {
	return probePort(port)
}

func probePort(port int) *DiscoverResult {
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil
	}

	wsURL, _ := info["webSocketDebuggerUrl"].(string)
	browser, _ := info["Browser"].(string)
	version, _ := info["Protocol-Version"].(string)

	return &DiscoverResult{
		Port:      port,
		WebSocket: wsURL,
		Version:   fmt.Sprintf("%s (protocol %s)", browser, version),
	}
}
