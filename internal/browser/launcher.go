package browser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wfdx165/use-browser/internal/config"
)

type Launcher struct {
	config   *config.Config
	cmd      *exec.Cmd
	cdpURL   string
	tempDir  string
	pidFile  string
	mu       sync.Mutex
}

func NewLauncher(cfg *config.Config) *Launcher {
	return &Launcher{
		config: cfg,
	}
}

func (l *Launcher) Launch(ctx context.Context) (cdpURL string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

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

	stderrPipe, err := l.cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := l.cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start chrome: %w", err)
	}

	if l.config.Session != "" {
		pidDir := filepath.Join(config.DefaultConfigDir(), "pids")
		os.MkdirAll(pidDir, 0755)
		l.pidFile = filepath.Join(pidDir, l.config.Session+".pid")
		os.WriteFile(l.pidFile, []byte(fmt.Sprintf("%d", l.cmd.Process.Pid)), 0644)
	}

	cdpURL, err = l.waitForCDPFromStderr(ctx, stderrPipe)
	if err != nil {
		l.cmd.Process.Kill()
		return "", err
	}

	l.cdpURL = cdpURL
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
	return l.cdpURL
}

func (l *Launcher) buildArgs() []string {
	args := []string{
		"--remote-debugging-port=0",
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

var wsURLPattern = regexp.MustCompile(`ws://[^\s]+`)

func (l *Launcher) waitForCDPFromStderr(ctx context.Context, stderrPipe io.ReadCloser) (string, error) {
	timeout := time.Duration(l.config.DefaultTimeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "DevTools listening on") {
			urls := wsURLPattern.FindAllString(line, -1)
			if len(urls) > 0 {
				return urls[0], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading chrome stderr: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("timed out waiting for chrome CDP")
	default:
		return "", fmt.Errorf("chrome closed without providing CDP URL")
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
