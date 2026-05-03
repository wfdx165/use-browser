# use-browser

适用于 AI 代理的浏览器自动化工具，使用 Go 编写。

通过 Chrome DevTools Protocol (CDP) 连接到现有的浏览器，或自动启动新浏览器。支持 Chrome、Edge、Brave、Opera、Vivaldi、Chromium 和 Arc。

## 安装

```bash
go install github.com/wfdx165/use-browser/cmd/browser@latest
```

或从源码构建：

```bash
git clone https://github.com/wfdx165/use-browser
cd use-browser
go build -o use-browser ./cmd/browser
```

## 环境要求

- Go 1.26+
- 任意支持 CDP 的浏览器（自动从常见路径检测）：
  - Google Chrome
  - Microsoft Edge
  - Brave Browser
  - Opera
  - Vivaldi
  - Chromium
  - Arc (macOS)

## 快速开始

```bash
use-browser open https://example.com
use-browser snapshot -i                # 仅显示可交互元素
use-browser click a                    # 点击第一个链接
use-browser fill "#search" "hello"     # 填充输入框
use-browser screenshot page.png        # 保存截图
use-browser screenshot - | base64      # 截图输出二进制到标准输出
use-browser tab new https://google.com # 打开新标签页
use-browser console                    # 打开 DevTools 控制台
use-browser find role button click     # 通过语义定位器查找元素
use-browser is visible "#submit"       # 检查元素状态
use-browser keyboard type "hello"      # 无选择器直接输入
use-browser batch "open url" "click"   # 批量执行
```

## 浏览器持久化

默认情况下，`use-browser` 在命令执行完毕后不会关闭浏览器实例。下次执行命令时会自动重连到同一个浏览器，而无需重新启动。浏览器状态（PID 和 CDP URL）保存在 `~/.use-browser/pids/` 目录中。

```bash
# 首次运行：启动浏览器
use-browser open https://example.com

# 再次运行：自动重连到已运行的浏览器
use-browser snapshot

# 手动关闭浏览器：
killall "Google Chrome"  # 或其他浏览器名称
```

每次新启动的浏览器实例都会使用临时用户数据目录（`/tmp/use-browser-*`），以避免配置文件锁定冲突。

## 命令

### 核心命令

| 命令 | 说明 |
|---------|-------------|
| `open [url]` | 启动浏览器，可选导航到指定 URL |
| `close` | 关闭浏览器 |
| `snapshot [-i] [-c] [-d N]` | 无障碍树快照，使用 `@e1`、`@e2` 引用 |
| `click <sel>` | 通过 CSS 选择器点击元素 |
| `fill <sel> <text>` | 清空并填充输入框 |
| `type <sel> <text>` | 通过键盘事件输入 |
| `press <key>` | 按下键盘按键 |
| `screenshot [path\|-]` | 截图（支持 --full、--format）。使用 `-` 输出到标准输出 |
| `eval <js>` | 执行 JavaScript |
| `batch "cmd1" "cmd2"` | 批量执行多个命令 |

### 语义定位

| 命令 | 说明 |
|---------|-------------|
| `find <strategy> <value> [action]` | 通过 role/text/label 等语义属性查找元素 |
| `find role button click --name "X"` | 通过 role 查找按钮，可选按名称过滤 |
| `find text "Sign In" click` | 通过文本内容查找 |
| `find label "Email" fill "x"` | 通过关联 label 查找 |
| `is visible <sel>` | 检查元素是否可见 |
| `is enabled <sel>` | 检查元素是否可用 |
| `is checked <sel>` | 检查复选框是否选中 |

### 键盘输入

| 命令 | 说明 |
|---------|-------------|
| `keyboard type <text>` | 在当前焦点处输入文本 |
| `keyboard inserttext <text>` | 直接插入文本（无键盘事件） |

### 导航

| 命令 | 说明 |
|---------|-------------|
| `back` / `forward` / `reload` | 历史记录导航 |
| `scroll <dir> [px]` | 滚动页面 |
| `scrollintoview <sel>` | 滚动元素到可视区域 |

### 获取信息

| 命令 | 说明 |
|---------|-------------|
| `get text <sel>` | 获取元素文本 |
| `get html <sel>` | 获取 innerHTML |
| `get value <sel>` | 获取输入值 |
| `get title` | 获取页面标题 |
| `get url` | 获取当前 URL |
| `get count <sel>` | 统计匹配元素数量 |
| `get attr <sel> <name>` | 获取属性值 |

### 交互

| 命令 | 说明 |
|---------|-------------|
| `hover <sel>` | 悬停元素 |
| `select <sel> <val>` | 选择下拉选项 |
| `check / uncheck <sel>` | 勾选/取消勾选复选框 |
| `highlight <sel>` | 高亮元素 |

### 等待

| 命令 | 说明 |
|---------|-------------|
| `wait <sel>` | 等待元素出现 |
| `wait <ms>` | 等待毫秒数 |
| `wait --text "..."` | 等待文本出现 |
| `wait --load <state>` | 等待加载状态 |

### 状态管理

| 命令 | 说明 |
|---------|-------------|
| `cookies` | 列出 Cookie |
| `cookies set <k> <v>` | 设置 Cookie |
| `cookies clear` | 清除 Cookie |
| `storage local [key]` | 管理 localStorage |
| `storage session [key]` | 管理 sessionStorage |

### 高级功能

| 命令 | 说明 |
|---------|-------------|
| `tab new [url]` | 打开新标签页（真正创建新标签页，而非在当前页跳转） |
| `tab [n]` | 切换到第 n 个标签页 |
| `tab close [n]` | 关闭标签页（当前或指定索引） |
| `tab list` | 列出所有标签页 |
| `frame <sel\|main>` | 切换 iframe |
| `network route <url>` | 拦截请求 |
| `dialog accept` | 接受对话框 |
| `batch "cmd1" "cmd2"` | 批量执行 |
| `set viewport <w> <h>` | 设置视口 |
| `set offline on\|off` | 切换离线模式 |
| `console` | 按 F12 打开 DevTools 控制台 |

## 选项

配置通过 **CLI 标志** 或 **环境变量** 完成（不支持配置文件）。

| 标志 | 环境变量 | 说明 |
|------|-----|-------------|
| `--cdp <port\|url>` | `USE_BROWSER_CDP` | 通过 CDP 端口或 WebSocket URL 连接 |
| `--auto-connect` | `USE_BROWSER_AUTO_CONNECT` | 自动发现运行中的浏览器 |
| `--executable-path` | `USE_BROWSER_EXECUTABLE_PATH` | 自定义浏览器可执行文件路径 |
| `--proxy` | `USE_BROWSER_PROXY` | 代理服务器 |
| `--user-agent` | `USE_BROWSER_USER_AGENT` | 自定义 User-Agent |
| `--json` | — | JSON 输出 |

### 环境变量示例

```bash
# 通过 CDP 连接到现有浏览器
USE_BROWSER_CDP=9222 use-browser open https://example.com

# 使用特定浏览器
USE_BROWSER_EXECUTABLE_PATH=/usr/bin/brave use-browser open https://example.com

# 多个选项
USE_BROWSER_CDP=9222 USE_BROWSER_PROXY=http://localhost:8080 use-browser open https://example.com
```

优先级：CLI 标志 > 环境变量 > 默认值

## 许可证

MIT
