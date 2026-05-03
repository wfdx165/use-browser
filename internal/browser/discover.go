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

func DiscoverChrome() ([]DiscoverResult, error) {
	var results []DiscoverResult

	if r := discoverDevToolsPort(); r != nil {
		results = append(results, *r)
	}

	for _, port := range commonPorts {
		if r := probePort(port); r != nil {
			results = append(results, *r)
		}
	}

	return results, nil
}

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
