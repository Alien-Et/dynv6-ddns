# Dynv6 DDNS 更新器

![GitHub](https://img.shields.io/github/license/Alien-Et/dynv6-ddns)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Alien-Et/dynv6-ddns)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/Alien-Et/dynv6-ddns)

这是一个基于 Go 语言开发的动态 DNS (DDNS) 更新工具，用于自动更新 [dynv6.com](https://dynv6.com) 的 IP 记录。项目提供了一个简洁的 Web 界面，允许用户配置域名、API 令牌、网络接口、IP 类型（IPv4、IPv6 或双栈）以及更新间隔。配置数据存储在 SQLite 数据库中，程序会定期检查公网 IP 并更新 DNS 记录，并支持通过 Telegram 发送通知。

## 功能特性

- **自动 IP 更新**：定期检测公网 IPv4 和/或 IPv6 地址，并更新 dynv6.com 的 DNS 记录。
- **Web 配置界面**：通过直观的 Web 界面配置和管理 DDNS 参数。
- **SQLite 数据库**：持久化存储配置数据，支持动态加载和更新。
- **Telegram 通知**：支持通过 Telegram 机器人发送更新成功或失败的通知。
- **重试机制**：更新失败时自动重试（最多 3 次，指数退避）。
- **日志记录**：详细记录运行状态、错误和更新结果到日志文件（`dynv6.log`）和标准输出。
- **多域名支持**：支持同时更新多个 dynv6 域名。

## 技术栈

- **后端**：Go（使用 `net/http` 提供 Web 服务，`mattn/go-sqlite3` 进行数据库操作）
- **前端**：HTML、Tailwind CSS、JavaScript
- **数据库**：SQLite
- **依赖**：`github.com/mattn/go-sqlite3`

## 安装

### 前提条件

- Go 1.16 或更高版本
- SQLite 环境（由 `mattn/go-sqlite3` 自动处理）
- 可选：Telegram 机器人令牌和聊天 ID（用于通知功能）
- 可访问 [dynv6.com](https://dynv6.com) 的 API

### 安装步骤

1. 克隆仓库：
   ```bash
   git clone https://github.com/Alien-Et/dynv6-ddns.git
   cd dynv6-ddns
   ```

2. 安装依赖：
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

3. 构建项目：
   ```bash
   go build -o dynv6-ddns
   ```

4. 确保 `static` 目录包含前端文件：
   - 仓库已包含 `static/index.html` 和 `static/script.js`。
   - `index.html` 使用 Tailwind CSS CDN，无需额外安装。

5. 运行程序：
   ```bash
   ./dynv6-ddns
   ```

## 使用方法

1. 启动程序后，访问 `http://localhost:8080` 打开 Web 配置界面。
2. 在 Web 界面中输入以下配置：
   - **域名**：要更新的 dynv6 域名（多个域名用逗号分隔，例如 `example.dns.navy,test.dns.navy`）。
   - **dynv6 API 令牌**：从 [dynv6.com](https://dynv6.com) 的账户设置中获取。
   - **更新间隔**：IP 检查和更新的时间间隔（秒，建议 300）。
   - **网络接口**：用于获取公网 IP 的网络接口（例如 `wlan0`、`eth0`，留空使用默认接口）。
   - **IP 类型**：选择 `IPv4`、`IPv6` 或 `dual`（双栈）。
   - **Telegram 通知**（可选）：输入 Telegram 机器人令牌和聊天 ID 以启用通知。
3. 点击“保存配置”，程序将开始定期检查 IP 并更新 dynv6 的 DNS 记录。
4. 查看日志文件（`dynv6.log`）或 Telegram 通知以监控更新状态。

## 项目结构

```
dynv6-ddns/
├── main.go                 # 主程序，包含后端逻辑
├── static/
│   ├── index.html          # Web 配置界面
│   └── script.js           # 前端 JavaScript 逻辑
├── config.db               # SQLite 数据库（运行时生成）
├── dynv6.log               # 日志文件（运行时生成）
├── go.mod                  # Go 模块依赖
├── go.sum                  # Go 模块校验文件
└── README.md               # 项目自述文件
```

## 配置说明

- **域名**：必须为有效的 dynv6 域名，支持多个域名（逗号分隔）。
- **API 令牌**：从 dynv6.com 的账户设置中获取，必填。
- **网络接口**：指定网络接口名称（如 `wlan0`、`eth0`），或留空使用默认接口。
- **IP 类型**：
  - `IPv4`：仅更新 IPv4 记录。
  - `IPv6`：仅更新 IPv6 记录。
  - `dual`：同时更新 IPv4 和 IPv6 记录。
- **更新间隔**：建议设置为 300 秒（5 分钟）以避免频繁请求。
- **Telegram 通知**：需要有效的 Telegram 机器人令牌和聊天 ID，留空禁用通知。

## 日志

- 日志文件存储在运行目录的 `dynv6.log` 中，记录程序运行状态、错误和更新结果。
- 日志同时输出到标准输出，便于实时监控和调试。

## 故障排查

- **无法启动 Web 服务器**：检查端口 `8080` 是否被占用（`netstat -tuln | grep 8080`）。
- **IP 获取失败**：
  - 确保指定的网络接口名称正确（运行 `ifconfig` 或 `ip addr` 查看）。
  - 检查网络是否正常连接。
  - 若留空网络接口，确保系统有可用的非回环网络接口。
- **更新失败**：
  - 验证 dynv6 API 令牌和域名是否正确。
  - 检查 dynv6.com 的 API 响应（日志中包含响应内容）。
- **数据库错误**：确保程序有权限在运行目录创建/写入 `config.db`。
- **Telegram 通知失败**：
  - 验证 Telegram 机器人令牌和聊天 ID 是否正确。
  - 检查网络是否可以访问 `api.telegram.org`。
- **Web 界面加载失败**：确保 `static` 目录包含 `index.html` 和 `script.js`。

## 部署建议

- **作为服务运行**：
  - 使用 `systemd` 或 `supervisord` 将程序设置为系统服务，确保开机自启。
  - 示例 `systemd` 服务文件：
    ```ini
    [Unit]
    Description=Dynv6 DDNS Updater
    After=network.target

    [Service]
    ExecStart=/path/to/dynv6-ddns
    WorkingDirectory=/path/to/dynv6-ddns
    Restart=always

    [Install]
    WantedBy=multi-user.target
    ```

- **防火墙**：确保 `8080` 端口对本地或所需网络开放（例如 `ufw allow 8080`）。

## 贡献

欢迎为项目贡献代码或提出建议！请遵循以下步骤：

1. Fork 仓库。
2. 创建特性分支（`git checkout -b feature/your-feature`）。
3. 提交更改（`git commit -m "Add your feature"`）。
4. 推送到分支（`git push origin feature/your-feature`）。
5. 创建 Pull Request。

请在提交前确保代码通过 `gofmt` 格式化，并添加必要的测试。

## 许可证

本项目采用 MIT 许可证。详情见 [LICENSE](LICENSE) 文件。

## 致谢

- [dynv6.com](https://dynv6.com)：提供免费的动态 DNS 服务。
- [Tailwind CSS](https://tailwindcss.com)：用于快速构建响应式 Web 界面。
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)：提供 Go 的 SQLite 驱动支持。

## 联系

如有问题或建议，请在 [GitHub Issues](https://github.com/Alien-Et/dynv6-ddns/issues) 中反馈，或联系作者 [Alien-Et](https://github.com/Alien-Et)。