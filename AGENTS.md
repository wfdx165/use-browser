# AGENTS.md

## Build & Run

```
go build -o use-browser ./cmd/browser   # binary in repo root (gitignored)
go install github.com/wfdx165/use-browser/cmd/browser@latest
go test ./...
```

Go **1.26**, deps: `chromedp`, `cobra`, `viper`.

## Config Priority (highest → lowest)

1. CLI flags (`--headed`, `--cdp`, `--auto-connect`, …)
2. Env vars (`USE_BROWSER_HEADED`, `USE_BROWSER_CDP`, …)
3. `./use-browser.json` (project-level)
4. `~/.use-browser/config.json` (user-level)

Viper env key replacer: `-` → `_`. Only keys matching `Config` struct fields are bound automatically.

## Browser Startup Flow (`internal/browser/manager.go`)

`Manager.Start()` follows this order:

1. **`cfg.CDP != ""`** → `Connector.Connect()` — connects to explicit port/WS URL.
2. **`cfg.AutoConnect`** → `DiscoverChrome()` — reads `DevToolsActivePort` file, then probes ports `{9222, 9223, 9229, 9333}`.
3. **Default** → `Launcher.Launch()` — spawns new Chrome with `--headless=new`, ephemeral profile dir, `--remote-debugging-port=0`.

Headed mode (`--headed`) skips `--headless=new` but still launches a *new* instance; it does **not** attach to a running Chrome.

## CLI Pattern

Every command:
1. Checks `cfg != nil` (set by `PersistentPreRunE` in `root.go`).
2. Creates `browser.NewManager(cfg)` → `manager.Start(ctx)` → `defer manager.Close()`.
3. Creates `cdp.NewClientFromManager(manager)` → `defer client.Close()`.
4. Runs `client.Run(timeout, actions...)`.

**Implication**: Each command is a self-contained session. `manager.Close()` kills the launcher-owned browser process. Commands chained via `batch` each get their own manager/connection (batch is currently a no-op skeleton — prints commands but doesn't execute them).

## Ref / Selector System

- Snapshot produces refs like `@e1`, `@e2`, … via a counter in JS.
- `selector.IsRef("@e3")` → true if `@e` + digits.
- `RefResolver` maps `@eN` ↔ CDP node IDs (not currently wired into click/fill — those use raw CSS selectors via `chromedp`).

## Snapshot

Two JS implementations:
- `internal/snapshot/tree.go:buildJS()` — minified inline (used at runtime via `chromedp.Evaluate`).
- `internal/snapshot/snapshot_builder.js` — readable source, **not** embedded. Keep them in sync if modifying the walker.

Output format: tree with `[e1] role "name" value="..."` lines, connector characters `├─` / `└─` / `│`.

## Key Files

| Path | Purpose |
|------|---------|
| `cmd/browser/main.go` | Entry point, calls `cli.Execute()` |
| `internal/cli/root.go` | Root command, flag declarations, `applyFlagOverrides` |
| `internal/cli/open.go` | `open` command — creates manager, navigates |
| `internal/browser/manager.go` | Connect/Discover/Launch decision logic |
| `internal/browser/launcher.go` | Spawns Chrome, parses CDP URL from stderr |
| `internal/browser/connector.go` | HTTP→WS upgrade for explicit CDP port |
| `internal/browser/discover.go` | `DevToolsActivePort` + port probing |
| `internal/cdp/client.go` | chromedp remote allocator wrapper |
| `internal/config/config.go` | `Config` struct + viper loading |
| `internal/snapshot/tree.go` | DOM walker + tree builder |
| `internal/snapshot/format.go` | Text/JSON output formatting |
| `internal/selector/ref.go` | `@eN` ref parsing |

## Empty / Stub Directories

`internal/annotate/`, `internal/diff/`, `internal/session/`, `internal/exec/`, `js/` — directories exist but contain no Go files. Placeholder for future features.

## Version

`pkg/version/version.go` has static constants. `Commit` and `Date` are `"unknown"` unless set via `-ldflags` at build time (not currently configured in any build script).
