package browser

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wfdx165/use-browser/internal/config"
)

type Launcher struct {
	config  *config.Config
	cmd     *exec.Cmd
	cdpPort int
	tempDir string
	pidFile string
	mu      sync.Mutex
}

func NewLauncher(cfg *config.Config) *Launcher {
	return &Launcher{
		config: cfg,
	}
}

func (l *Launcher) Launch(ctx context.Context) (cdpURL string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	port, err := l.findFreePort()
	if err != nil {
		return "", fmt.Errorf("failed to find free port: %w", err)
	}
	l.cdpPort = port

	if l.config.Session != "" {
		l.tempDir = filepath.Join(config.DefaultConfigDir(), "tmp", l.config.Session)
	} else {
		l.tempDir, err = os.MkdirTemp("", "use-browser-*")
	}
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	args := l.buildArgs()

	execPath, err := l.resolveExecPath()
	if err != nil {
		return "", err
	}

	l.cmd = exec.CommandContext(ctx, execPath, args...)
	l.cmd.Stdout = nil
	l.cmd.Stderr = nil

	if err := l.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start chrome: %w", err)
	}

	if l.config.Session != "" {
		pidDir := filepath.Join(config.DefaultConfigDir(), "pids")
		os.MkdirAll(pidDir, 0755)
		l.pidFile = filepath.Join(pidDir, l.config.Session+".pid")
		os.WriteFile(l.pidFile, []byte(fmt.Sprintf("%d", l.cmd.Process.Pid)), 0644)
	}

	cdpURL, err = l.waitForCDP(ctx, port)
	if err != nil {
		l.cmd.Process.Kill()
		return "", err
	}

	return cdpURL, nil
}

func (l *Launcher) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.pidFile != "" {
		os.Remove(l.pidFile)
	}

	if l.cmd != nil && l.cmd.Process != nil {
		l.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			l.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			l.cmd.Process.Kill()
		}
	}

	if l.tempDir != "" {
		os.RemoveAll(l.tempDir)
	}

	return nil
}

func (l *Launcher) PID() int {
	if l.cmd == nil || l.cmd.Process == nil {
		return 0
	}
	return l.cmd.Process.Pid
}

func (l *Launcher) CDPURL() string {
	return fmt.Sprintf("ws://127.0.0.1:%d", l.cdpPort)
}

func (l *Launcher) buildArgs() []string {
	args := []string{
		fmt.Sprintf("--remote-debugging-port=%d", l.cdpPort),
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

	if l.config.Profile != "" {
		args = append(args, "--user-data-dir="+l.config.Profile)
	} else if l.tempDir != "" {
		args = append(args, "--user-data-dir="+l.tempDir)
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

func (la *Launcher) findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	return port, nil
}

func (l *Launcher) waitForCDP(ctx context.Context, port int) (string, error) {
	timeout := time.Duration(l.config.DefaultTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timed out waiting for chrome CDP on port %d", port)
		case <-ticker.C:
			url := l.tryGetWSURL(port)
			if url != "" {
				return url, nil
			}
		}
	}
}

func (l *Launcher) tryGetWSURL(port int) string {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "webSocketDebuggerUrl") {
			parts := strings.Split(line, `"webSocketDebuggerUrl":"`)
			if len(parts) > 1 {
				url := strings.Split(parts[1], `"`)
				if len(url) > 0 {
					return url[0]
				}
			}
		}
	}
	return ""
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
