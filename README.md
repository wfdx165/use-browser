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
use-browser close
```

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
| `screenshot [path]` | Take screenshot (--full, --format) |
| `eval <js>` | Execute JavaScript |

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
| `tab new [url]` | Open new tab |
| `frame <sel\|main>` | Switch iframe |
| `network route <url>` | Intercept requests |
| `dialog accept` | Accept dialog |
| `batch "cmd1" "cmd2"` | Batch execution |
| `set viewport <w> <h>` | Set viewport |
| `set offline on\|off` | Toggle offline |
| `console` | View console messages |

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

## Architecture

```
cmd/browser/     # CLI entry point
internal/
  cli/           # Cobra commands (open, click, snapshot, ...)
  cdp/           # CDP client wrapper (chromedp)
  browser/       # Chrome launcher, connector, discover, manager
  snapshot/      # DOM tree builder + formatting
  selector/      # Ref/CSS selector parsing
  config/        # Config loading (file + env + flags)
```

## License

MIT
