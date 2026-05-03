# use-browser

Browser automation CLI for AI agents, written in Go.

Connects to an existing Chrome browser via Chrome DevTools Protocol (CDP) or launches a new one automatically.

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

## Requirements

- Go 1.22+
- Google Chrome or Chromium (auto-detected from common paths)

## Quick Start

```bash
use-browser open https://example.com
use-browser snapshot -i                # Interactive elements only
use-browser click a                    # Click first link
use-browser fill "#search" "hello"     # Fill input
use-browser screenshot page.png        # Save screenshot
use-browser screenshot - | base64      # Output binary to stdout
use-browser tab new https://google.com # Open new tab
use-browser console                    # Open DevTools console (headed)
use-browser find role button click     # Find by semantic locator
use-browser is visible "#submit"       # Check element state
use-browser keyboard type "hello"      # Type without selector
use-browser batch "open url" "click"   # Batch execution
```

## Browser Persistence

By default, `use-browser` keeps the Chrome instance running after the command exits. The next command will automatically reconnect to the same browser instead of launching a new one. Chrome state (PID and CDP URL) is saved to `~/.use-browser/pids/`.

```bash
# First run: starts Chrome
use-browser open https://example.com

# Second run: reconnects to the same Chrome
use-browser snapshot

# To manually stop the browser:
killall "Google Chrome"
```

Each new Chrome instance uses a temporary user-data directory (`/tmp/use-browser-*`) to avoid profile lock conflicts.

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
| `console` | Open DevTools console (F12) in headed mode |

## Options

| Flag | Env | Description |
|------|-----|-------------|
| `--cdp <port\|url>` | `USE_BROWSER_CDP` | Connect via CDP |
| `--auto-connect` | `USE_BROWSER_AUTO_CONNECT` | Auto-discover Chrome |
| `--headed` | `USE_BROWSER_HEADED` | Show browser window |
| `--executable-path` | `USE_BROWSER_EXECUTABLE_PATH` | Custom Chrome path |
| `--proxy` | `USE_BROWSER_PROXY` | Proxy server |
| `--user-agent` | `USE_BROWSER_USER_AGENT` | Custom UA |
| `--json` | — | JSON output |
| `--config` | `USE_BROWSER_CONFIG` | Config file path |

## Configuration

Create `~/.use-browser/config.json` or `./use-browser.json`:

```json
{
  "headed": true,
  "executablePath": "/path/to/chrome",
  "proxy": "http://localhost:8080"
}
```

Priority: CLI flags > env vars > project config > user config > defaults.

## License

MIT
