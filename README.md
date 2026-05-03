# use-browser

Browser automation CLI for AI agents, written in Go.

Connects to an existing browser via Chrome DevTools Protocol (CDP) or launches a new one automatically. Supports Chrome, Edge, Brave, Opera, Vivaldi, Chromium, and Arc.

[中文文档](README_zh.md)

## Installation

```bash
go install github.com/wfdx165/use-browser/cmd/browser@latest
```

Or build from source:

```bash
git clone https://github.com/wfdx165/use-browser
cd use-browser
go build -o use-browser ./cmd/browser
```

## Building

Build for all platforms using the included `build.sh` script:

```bash
# Build all platforms (Linux, macOS, Windows)
./build.sh

# Build specific platform
./build.sh linux-amd64     # Linux x86_64
./build.sh linux-arm64     # Linux ARM64
./build.sh darwin-amd64    # macOS Intel
./build.sh darwin-arm64    # macOS Apple Silicon
./build.sh windows-amd64   # Windows x86_64

# Other commands
./build.sh --clean         # Remove build artifacts
./build.sh --version       # Show version info
./build.sh --help          # Show all available options
```

**Version Auto-Increment**: The script automatically detects the latest Git tag and increments the patch version. For example, if the latest tag is `v0.1.0`, it will build `v0.1.1`.

**Create Git Tag after Build**:
```bash
./build.sh --tag           # Build all and create git tag vX.X.X
```

Build artifacts are placed in `dist/`:
- `use-browser-v0.1.1-linux-amd64.tar.gz`
- `use-browser-v0.1.1-linux-arm64.tar.gz`
- `use-browser-v0.1.1-darwin-amd64.tar.gz`
- `use-browser-v0.1.1-darwin-arm64.tar.gz`
- `use-browser-v0.1.1-windows-amd64.zip`
- `checksums.txt` (SHA256)

Each archive contains:
- Directory: `use-browser-v0.1.1-{platform}/`
- Executable: `use-browser` (Linux/macOS) or `use-browser.exe` (Windows)
- `LICENSE`
- `README.md`

## Requirements

- Go 1.26+
- Any CDP-compatible browser (auto-detected):
  - Google Chrome
  - Microsoft Edge
  - Brave Browser
  - Opera
  - Vivaldi
  - Chromium
  - Arc (macOS)

## Quick Start

```bash
use-browser open https://example.com
use-browser snapshot -i                # Interactive elements only
use-browser click a                    # Click first link
use-browser fill "#search" "hello"     # Fill input
use-browser screenshot page.png        # Save screenshot
use-browser screenshot - | base64      # Output binary to stdout
use-browser tab new https://google.com # Open new tab
use-browser console                    # Open DevTools console
use-browser find role button click     # Find by semantic locator
use-browser is visible "#submit"       # Check element state
use-browser keyboard type "hello"      # Type without selector
use-browser batch "open url" "click"   # Batch execution
```

## Browser Persistence

By default, `use-browser` keeps the browser instance running after the command exits. The next command will automatically reconnect to the same browser instead of launching a new one. Browser state (PID and CDP URL) is saved to `~/.use-browser/pids/`.

```bash
# First run: starts browser
use-browser open https://example.com

# Second run: reconnects to the same browser
use-browser snapshot

# To manually stop the browser:
killall "Google Chrome"  # or your browser name
```

Each new browser instance uses a temporary user-data directory (`/tmp/use-browser-*`) to avoid profile lock conflicts.

## Commands

### Core

| Command | Description |
|---------|-------------|
| `open [url]` | Launch browser, optionally navigate |
| `close` | Close browser |
| `snapshot [-i] [-c] [-d N]` | Accessibility tree with `@e1`, `@e2` refs |
| `click <sel>` | Click element by CSS selector |
| `fill <sel> <text>` | Clear and fill input |
| `type <sel> <text>` | Type with key events |
| `press <key>` | Press keyboard key |
| `screenshot [path\|-]` | Take screenshot (--full, --format). Use `-` for stdout |
| `eval <js>` | Execute JavaScript |
| `batch "cmd1" "cmd2"` | Execute multiple commands |

### Semantic Locators
| Command | Description |
|---------|-------------|
| `find <strategy> <value> [action]` | Find elements by role/text/label/etc |
| `find role button click --name "X"` | Find button by role and optional name |
| `find text "Sign In" click` | Find by text content |
| `find label "Email" fill "x"` | Find by associated label |
| `is visible <sel>` | Check if element is visible |
| `is enabled <sel>` | Check if element is enabled |
| `is checked <sel>` | Check if checkbox is checked |

### Keyboard
| Command | Description |
|---------|-------------|
| `keyboard type <text>` | Type text at current focus |
| `keyboard inserttext <text>` | Insert text without events |

### Navigation
| Command | Description |
|---------|-------------|
| `back` / `forward` / `reload` | History navigation |
| `scroll <dir> [px]` | Scroll page |
| `scrollintoview <sel>` | Scroll element into view |

### Get Info
| Command | Description |
|---------|-------------|
| `get text <sel>` | Get element text |
| `get html <sel>` | Get innerHTML |
| `get value <sel>` | Get input value |
| `get title` | Get page title |
| `get url` | Get current URL |
| `get count <sel>` | Count matching elements |
| `get attr <sel> <name>` | Get attribute value |

### Interaction
| Command | Description |
|---------|-------------|
| `hover <sel>` | Hover element |
| `select <sel> <val>` | Select dropdown option |
| `check / uncheck <sel>` | Check/uncheck checkbox |
| `highlight <sel>` | Highlight element |

### Wait
| Command | Description |
|---------|-------------|
| `wait <sel>` | Wait for element |
| `wait <ms>` | Wait milliseconds |
| `wait --text "..."` | Wait for text |
| `wait --load <state>` | Wait for load state |

### State
| Command | Description |
|---------|-------------|
| `cookies` | List cookies |
| `cookies set <k> <v>` | Set cookie |
| `cookies clear` | Clear cookies |
| `storage local [key]` | Manage localStorage |
| `storage session [key]` | Manage sessionStorage |

### Advanced
| Command | Description |
|---------|-------------|
| `tab new [url]` | Open new tab (creates a real new tab) |
| `tab [n]` | Switch to tab by index |
| `tab close [n]` | Close tab (current or by index) |
| `tab list` | List all tabs |
| `frame <sel\|main>` | Switch iframe |
| `network route <url>` | Intercept requests |
| `dialog accept` | Accept dialog |
| `batch "cmd1" "cmd2"` | Batch execution |
| `set viewport <w> <h>` | Set viewport |
| `set offline on\|off` | Toggle offline |
| `console` | Open DevTools console (F12) |

## Options

Configuration is done via **CLI flags** or **environment variables** (no config files).

| Flag | Env | Description |
|------|-----|-------------|
| `--cdp <port\|url>` | `USE_BROWSER_CDP` | Connect via CDP port or WebSocket URL |
| `--auto-connect` | `USE_BROWSER_AUTO_CONNECT` | Auto-discover running browser |
| `--executable-path` | `USE_BROWSER_EXECUTABLE_PATH` | Custom browser executable path |
| `--proxy` | `USE_BROWSER_PROXY` | Proxy server |
| `--user-agent` | `USE_BROWSER_USER_AGENT` | Custom User-Agent |
| `--json` | — | JSON output |

### Environment Variable Examples

```bash
# Connect to existing browser via CDP
USE_BROWSER_CDP=9222 use-browser open https://example.com

# Use specific browser
USE_BROWSER_EXECUTABLE_PATH=/usr/bin/brave use-browser open https://example.com

# Multiple options
USE_BROWSER_CDP=9222 USE_BROWSER_PROXY=http://localhost:8080 use-browser open https://example.com
```

Priority: CLI flags > Environment variables > Defaults

## License

MIT
