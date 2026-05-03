package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wfdx165/use-browser/internal/config"
)

type Launcher struct {
	config      *config.Config
	cmd         *exec.Cmd
	cdpURL      string
	cdpPort     int
	mu          sync.Mutex
	userDataDir string
}

func NewLauncher(cfg *config.Config) *Launcher {
	return &Launcher{
		config: cfg,
	}
}

func (l *Launcher) Launch(ctx context.Context) (cdpURL string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if savedPID, savedCDP, err := l.loadState(); err == nil {
		if processAlive(savedPID) && cdpReachable(savedCDP) {
			l.cdpURL = savedCDP
			return savedCDP, nil
		}
		l.removeState()
	}

	execPath, err := l.resolveExecPath()
	if err != nil {
		return "", err
	}

	userDataDir, err := os.MkdirTemp("", "use-browser-*")
	if err != nil {
		return "", fmt.Errorf("failed to create user data dir: %w", err)
	}
	l.userDataDir = userDataDir

	l.cdpPort, err = freePort()
	if err != nil {
		return "", fmt.Errorf("failed to choose remote debugging port: %w", err)
	}

	args := l.buildArgs()

	l.killExistingChrome()

	l.cmd = exec.Command(execPath, args...)

	if err := l.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start chrome: %w", err)
	}

	cdpURL, err = l.waitForCDPFromPort(ctx, l.cdpPort)
	if err != nil {
		l.cmd.Process.Kill()
		return "", err
	}

	l.cdpURL = cdpURL
	l.saveState(l.cmd.Process.Pid, cdpURL)
	return cdpURL, nil
}

func (l *Launcher) killExistingChrome() {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("killall", "-9", "Google Chrome", "Chrome").Run()
	case "linux":
		exec.Command("pkill", "-9", "chrome").Run()
		exec.Command("pkill", "-9", "chromium").Run()
	case "windows":
		exec.Command("taskkill", "/F", "/IM", "chrome.exe").Run()
	}

	time.Sleep(500 * time.Millisecond)
}

func (l *Launcher) Close() error {
	l.removeState()
	if l.cmd != nil && l.cmd.Process != nil {
		l.cmd.Process.Kill()
		l.cmd.Wait()
	}
	if l.userDataDir != "" {
		os.RemoveAll(l.userDataDir)
	}
	return nil
}

func (l *Launcher) SavedCDPURL() (string, bool) {
	savedPID, savedCDP, err := l.loadState()
	if err != nil {
		return "", false
	}
	if processAlive(savedPID) && cdpReachable(savedCDP) {
		return savedCDP, true
	}
	l.removeState()
	return "", false
}

func (l *Launcher) PID() int {
	if l.cmd == nil || l.cmd.Process == nil {
		return 0
	}
	return l.cmd.Process.Pid
}

func (l *Launcher) CDPURL() string {
	return l.cdpURL
}

func (l *Launcher) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", l.cdpPort),
		fmt.Sprintf("--user-data-dir=%s", l.userDataDir),
		"--disable-background-tab-crashing",
		"--disable-default-apps",
		"--disable-extensions",
		"--disable-sync",
		"--disable-translate",
		"--no-first-run",
		"--no-pings",
		"--window-size=1280,720",
	}

	if !l.config.Headed {
		args = append(args, "--headless=new")
	}

	if l.config.IgnoreHTTPS {
		args = append(args, "--ignore-certificate-errors")
	}

	if l.config.AllowFileAccess {
		args = append(args, "--allow-file-access-from-files", "--allow-file-access")
	}

	if l.config.UserAgent != "" {
		args = append(args, "--user-agent="+l.config.UserAgent)
	}

	if l.config.Proxy != "" {
		args = append(args, "--proxy-server="+l.config.Proxy)
	}

	return args
}

func (l *Launcher) resolveExecPath() (string, error) {
	if l.config.ExecutablePath != "" {
		return l.config.ExecutablePath, nil
	}

	if path := findSystemChrome(); path != "" {
		return path, nil
	}

	return "", fmt.Errorf("chrome not found, specify --executable-path")
}

func (l *Launcher) waitForCDPFromPort(ctx context.Context, port int) (string, error) {
	timeout := time.Duration(l.config.DefaultTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if cdpURL, ok := cdpURLForPort(ctx, port); ok {
			return cdpURL, nil
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timed out waiting for chrome CDP")
		case <-ticker.C:
		}
	}
}

func findSystemChrome() string {
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "linux":
		paths := []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
			"/snap/bin/chromium",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "windows":
		paths := []string{
			filepath.Join(os.Getenv("PROGRAMFILES"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google/Chrome/Application/chrome.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chrome"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium-browser"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium"); err == nil {
		return path
	}

	return ""
}

func (l *Launcher) saveState(pid int, cdpURL string) error {
	pidDir := filepath.Join(config.DefaultConfigDir(), "pids")
	os.MkdirAll(pidDir, 0755)

	os.WriteFile(filepath.Join(pidDir, "default.pid"), []byte(fmt.Sprintf("%d", pid)), 0644)
	os.WriteFile(filepath.Join(pidDir, "default.cdp"), []byte(cdpURL), 0644)
	return nil
}

func (l *Launcher) loadState() (pid int, cdpURL string, err error) {
	pidDir := filepath.Join(config.DefaultConfigDir(), "pids")

	data, err := os.ReadFile(filepath.Join(pidDir, "default.pid"))
	if err != nil {
		return 0, "", err
	}
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &pid)

	cdpData, err := os.ReadFile(filepath.Join(pidDir, "default.cdp"))
	if err != nil {
		return 0, "", err
	}

	return pid, strings.TrimSpace(string(cdpData)), nil
}

func (l *Launcher) removeState() {
	pidDir := filepath.Join(config.DefaultConfigDir(), "pids")
	os.Remove(filepath.Join(pidDir, "default.pid"))
	os.Remove(filepath.Join(pidDir, "default.cdp"))
}

func processAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	switch runtime.GOOS {
	case "windows":
		err = process.Signal(nil)
	default:
		err = process.Signal(syscall.Signal(0))
	}
	return err == nil
}

func cdpReachable(cdpURL string) bool {
	parsed, err := url.Parse(cdpURL)
	if err != nil || parsed.Host == "" {
		return false
	}

	scheme := "http"
	if parsed.Scheme == "wss" {
		scheme = "https"
	}
	versionURL := url.URL{Scheme: scheme, Host: parsed.Host, Path: "/json/version"}

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(versionURL.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false
	}
	wsURL, _ := info["webSocketDebuggerUrl"].(string)
	return wsURL != ""
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected listener address %s", listener.Addr())
	}
	return addr.Port, nil
}

func cdpURLForPort(ctx context.Context, port int) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://127.0.0.1:%d/json/version", port), nil)
	if err != nil {
		return "", false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", false
	}
	wsURL, _ := info["webSocketDebuggerUrl"].(string)
	return wsURL, wsURL != ""
}
