### README.zh-CN.md

```markdown
# Dynv6 IP 更新程序

[Dynv6 IP 更新程序](https://github.com/Alien-Et/dynv6-ddns) 是一个 Go 语言编写的命令行工具，用于动态更新 [dynv6.com](https://dynv6.com) 域名的 IPv4 和/或 IPv6 地址。支持多域名管理、自定义更新间隔、网络接口选择和 Telegram 通知，适用于 Linux、Windows、macOS 和 Android（Termux）等平台。

[English README](README.md)

## 功能
- 动态更新多个 dynv6 域名的 IPv4 和/或 IPv6 地址。
- 支持自定义更新间隔（秒）和网络接口。
- 通过 Telegram 发送更新成功或失败通知。
- 首次运行提供交互式配置，生成 `config.json`。
- 日志同时输出到 `dynv6.log` 和终端，便于调试。
- 支持前台和后台运行（Linux/Android 支持 `nohup`）。
- 检测 IPv6 地址变化，仅在必要时更新。
- 更新失败自动重试 3 次，采用指数退避策略。

## 安装与编译

### 环境要求
- [Go](https://golang.org) 1.16 或更高版本。
- [Git](https://git-scm.com) 用于克隆仓库。
- Android 平台推荐 [Termux](https://f-droid.org/en/packages/com.termux/)（从 F-Droid 下载）。
- Android Root 用户可选择在 `/data` 任意目录运行，或制作 Magisk 模块刷入使用。

### 克隆仓库
```bash
git clone https://github.com/Alien-Et/dynv6-ddns.git
cd dynv6-ddns
```

### 平台编译
1. **Linux**:
   ```bash
   sudo apt update && sudo apt install golang  # Debian/Ubuntu
   sudo dnf install golang  # Fedora
   go build -o dynv6
   ```
   输出：`dynv6` 可执行文件。

2. **Windows**:
   - 下载并安装 Go：[Go 下载](https://golang.org/dl/)。
   ```bash
   go build -o dynv6.exe
   ```
   输出：`dynv6.exe` 可执行文件。

3. **macOS**:
   ```bash
   brew install go
   go build -o dynv6
   ```
   输出：`dynv6` 可执行文件。

4. **Android (Termux)**:
   - 安装 Termux（F-Droid 或 GitHub 版本）。
   ```bash
   pkg update && pkg install golang git
   git clone https://github.com/Alien-Et/dynv6-ddns.git
   cd dynv6-ddns
   go build -o dynv6
   termux-setup-storage
   ```
   输出：`dynv6` 可执行文件。`termux-setup-storage` 确保存储权限。
   - **Root 用户**：可在 `/data` 任意目录运行程序，编译后将 `dynv6` 和 `config.json` 移至目标目录执行。
   - **Magisk 模块**：动手能力强的用户可将程序打包为 Magisk 模块，刷入系统实现开机自启和持久运行（需自行编写模块脚本）。

## 使用方法
1. **首次运行**：
   运行程序生成 `config.json`：
   ```bash
   ./dynv6  # Linux/macOS/Android
   dynv6.exe  # Windows
   ```
   根据提示输入：
   - 域名（多个用逗号分隔，如 `example.dns.navy,test.dns.navy`）。
   - dynv6 API 令牌（从 [dynv6.com Zones](https://dynv6.com/zones) 获取）。
   - 更新间隔（秒，建议 300）。
   - 网络接口（Linux/Android 如 `wlan0`、`rmnet0`，Windows 如 `Ethernet`）。
   - IP 类型（`ipv4`、`ipv6` 或 `dual`）。
   - Telegram 机器人令牌和聊天 ID（可选，留空禁用）。

2. **后续运行**：
   - **前台**：
     ```bash
     ./dynv6  # Linux/macOS/Android
     dynv6.exe  # Windows
     ```
   - **后台**：
     ```bash
     termux-wake-lock && nohup ./dynv6 &  # Android (Termux)/Linux
     ```
     Windows 可通过任务计划程序运行后台任务。

3. **配置文件**：
   - 位置：与可执行文件同目录的 `config.json`。
   - 修改：直接编辑 `config.json` 调整设置。

4. **日志**：
   - 保存至 `dynv6.log`（同目录）。
   - 同时输出到终端，便于实时监控。

## 配置
`config.json` 采用 Caddy 风格 JSON 结构，示例：
```json
{
    "全局设置": {
        "更新间隔秒": 300,
        "网络接口": "wlan0",
        "IP类型": "ipv6"
    },
    "域名列表": {
        "域名": ["example.dns.navy", "test.dns.navy"],
        "令牌": "your_token_here"
    },
    "通知设置": {
        "Telegram机器人令牌": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
        "Telegram聊天ID": "123456789"
    }
}
```

### 配置说明
- **全局设置**：
  - `更新间隔秒`：IP 检查间隔（秒，建议 300，即 5 分钟）。
  - `网络接口`：网络接口名称，如 `wlan0`（Android/Linux）、`Ethernet`（Windows）。
  - `IP类型`：支持 `ipv4`（仅 IPv4）、`ipv6`（仅 IPv6）、`dual`（双栈）。
- **域名列表**：
  - `域名`：dynv6 域名列表，如 `example.dns.navy`。
  - `令牌`：dynv6 API 令牌，从 [dynv6.com Zones](https://dynv6.com/zones) 获取。
- **通知设置**：
  - `Telegram机器人令牌`：Telegram 机器人令牌（BotFather 获取），留空禁用。
  - `Telegram聊天ID`：聊天 ID（与机器人交互获取），留空禁用。

## 注意事项
- 程序仅在 IPv6 地址变化时更新，减少不必要请求。
- 确保 dynv6 API 令牌有效，网络接口存在（Android 常用 `wlan0` 或 `rmnet0`）。
- Android/Termux 需 `termux-wake-lock` 防止进程被系统终止。
- 更新失败自动重试 3 次，间隔采用指数退避（10秒起）。
- 配置文件和日志需存储权限（Android 运行 `termux-setup-storage`）。
- Windows 后台运行需配置任务计划程序或服务。
- Android Root 用户可将程序部署至 `/data`，或制作 Magisk 模块实现开机自启。

## 许可证
MIT 许可证，详见 [LICENSE](LICENSE).
```
