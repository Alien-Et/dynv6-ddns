以下是为提供的 Go 脚本编写的 **GitHub README** 中文版，作为面向开源社区的使用说明书。它遵循开源项目的标准惯例，提供清晰的安装、使用和贡献指南，适合中文用户。

---

# Dynv6 动态 DNS 更新工具

一个基于 Go 的轻量级命令行工具，用于在 [dynv6.com](https://dynv6.com) 上更新动态 DNS 记录。该工具会定期检查指定网络接口的公网 IP 地址，并使用 dynv6 API 更新配置域名的 DNS 记录。支持 IPv4 和 IPv6，可选通过 Telegram 发送状态通知。

## 功能

- **动态 DNS 更新**：自动更新 dynv6 上多个域名的 IPv4 和/或 IPv6 记录。
- **交互式配置**：首次运行时引导用户创建配置文件。
- **网络接口支持**：从指定网络接口获取公网 IP。
- **IP 类型灵活性**：支持 IPv4、IPv6 或双栈更新。
- **Telegram 通知**：可选通过 Telegram 发送更新状态通知。
- **重试机制**：更新失败时使用指数退避重试。
- **日志记录**：日志同时输出到文件（`dynv6.log`）和终端。
- **跨平台**：支持 Linux、macOS、Windows 以及 Android 上的 Termux。

## 前提条件

- **Go**：1.16 或更高版本（用于编译工具）。
- **dynv6 账户**：需要从 [dynv6.com](https://dynv6.com) 获取 API 令牌。
- **网络接口**：设备需具有带公网 IP 的网络接口。
- **Telegram（可选）**：用于通知的 Telegram 机器人令牌和聊天 ID。

## 安装

1. **克隆仓库**：
   ```bash
   git clone https://github.com/yourusername/dynv6-updater.git
   cd dynv6-updater
   ```

2. **编译程序**：
   ```bash
   go build -o dynv6 main.go
   ```

3. **运行程序**：
   ```bash
   ./dynv6
   ```

   首次运行时，程序会提示您交互式创建配置文件（`config.json`）。

## 使用方法

### 首次运行
如果首次运行或缺少 `config.json` 文件，程序将引导您交互式创建配置文件。您需要提供以下信息：

- **域名**：您的 dynv6 域名（例如 `example.dns.navy,test.dns.navy`）。
- **API 令牌**：从 [dynv6.com Zones 页面](https://dynv6.com/zones) 获取的 API 令牌。
- **更新间隔**：IP 检查的时间间隔（秒，建议 300）。
- **网络接口**：要监控的网络接口（例如 `wlan0`、`eth0`）。
- **IP 类型**：`ipv4`、`ipv6` 或 `dual`（默认：`ipv6`）。
- **Telegram 机器人令牌**（可选）：用于发送通知。
- **Telegram 聊天 ID**（可选）：用于接收通知。

示例交互：
```
请输入 dynv6 域名（多个域名用英文逗号分隔，例如 example.dns.navy,test.dns.navy）
example.dns.navy
请输入 dynv6 API 令牌（从 dynv6.com 的 Zones 页面获取）
your_api_token
请输入更新间隔（秒，建议 300，即 5 分钟）
300
请输入网络接口（根据设备网络选择）
- 说明：可用接口：wlan0, eth0
wlan0
请输入 IP 类型（可选值：ipv4, ipv6, dual）
ipv6
请输入 Telegram 机器人令牌（可选，留空禁用通知）
your_bot_token
请输入 Telegram 聊天 ID（可选，留空禁用通知）
your_chat_id
```

配置将保存到程序所在目录的 `config.json` 文件中。

### 配置文件
程序使用 `config.json` 文件，结构如下：

```json
{
    "全局设置": {
        "更新间隔秒": 300,
        "网络接口": "wlan0",
        "IP类型": "ipv6"
    },
    Polymind "域名列表": {
        "域名": ["example.dns.navy"],
        "令牌": "your_api_token"
    },
    "通知设置": {
        "Telegram机器人令牌": "your_bot_token",
        "Telegram聊天ID": "your_chat_id"
    }
}
```

- **全局设置**：
  - `更新间隔秒`：IP 更新间隔（秒，例如 `300` 表示 5 分钟）。
  - `网络接口`：网络接口名称（例如 `wlan0`、`eth0`）。
  - `IP类型`：要更新的 IP 类型（`ipv4`、`ipv6` 或 `dual`）。
- **域名列表**：
  - `域名`：dynv6 域名列表（例如 `["example.dns.navy", "test.dns.navy"]`）。
  - `令牌`：dynv6 API 令牌。
- **通知设置**：
  - `Telegram机器人令牌`：Telegram 机器人令牌（可选，留空禁用）。
  - `Telegram聊天ID`：Telegram 聊天 ID（可选，留空禁用）。

您可以在初始设置后手动编辑 `config.json` 来调整配置。

### 运行程序
- **前台运行**：
  ```bash
  ./dynv6
  ```
  日志将显示在终端并保存到 `dynv6.log`。

- **后台运行（例如 Linux 或 Termux）**：
  ```bash
  nohup ./dynv6 &
  ```
  在 Termux 上，需确保设备保持唤醒：
  ```bash
  termux-wake-lock && nohup ./dynv6 &
  ```

### 日志
日志保存在程序所在目录的 `dynv6.log` 文件中，同时输出到终端。日志内容包括：
- 配置文件加载详情。
- IP 地址检测结果。
- DNS 更新成功或失败信息。
- Telegram 通知状态。

## 示例工作流程
1. 运行 `./dynv6`。
2. 按照提示创建 `config.json`。
3. 程序开始监控指定网络接口，每 300 秒（或配置的间隔）更新 DNS 记录。
4. 如果配置了 Telegram，会为每次更新或错误发送通知。

## 故障排除
- **配置文件错误**：
  - 确保 `config.json` 是有效的 JSON 格式且包含所有必需字段。
  - 验证 dynv6 API 令牌是否正确。
- **网络接口问题**：
  - 使用 `ifconfig` 或 `ip addr` 查看可用接口。
  - 确保指定接口具有公网 IP。
- **无公网 IP**：
  - 检查设备是否位于 NAT 后或缺少公网 IP。
  - 确认 `IP类型` 与网络设置匹配。
- **Telegram 通知问题**：
  - 确保机器人令牌和聊天 ID 正确。
  - 通过 Telegram API 手动测试机器人消息发送。

## 贡献
欢迎贡献代码！贡献步骤如下：

1. Fork 本仓库。
2. 创建功能分支（`git checkout -b feature/your-feature`）。
3. 提交更改（`git commit -m "添加你的功能"`）。
4. 推送到分支（`git push origin feature/your-feature`）。
5. 提交 Pull Request。

请确保代码遵循现有风格并包含适当的测试。

### 开发设置
- 安装 Go 1.16 或更高版本。
- 使用 `go fmt` 格式化代码。
- 在 `test` 目录中添加测试（即将支持）。
- 本地编译和测试：`go build && ./dynv6`。

## 许可证
本项目采用 MIT 许可证，详情见 [LICENSE](LICENSE) 文件。

## 致谢
- 感谢 [dynv6.com](https://dynv6.com) 提供免费的动态 DNS 服务。
- 灵感来源于对简单、跨平台 DDNS 客户端的需求。

## 联系方式
如有问题或功能建议，请在 GitHub 上提交 Issue，或联系维护者：[your.email@example.com](mailto:your.email@example.com)。

---

### 实现注意事项
- 将克隆命令中的 `yourusername` 替换为您的实际 GitHub 用户名。
- 在仓库中添加 `LICENSE` 文件（例如 MIT 许可证），以确保项目符合开源要求。
- 如果添加了测试或其他功能，请在“贡献”部分更新具体说明。
- 确保仓库包含 `main.go` 文件和必要的资源（例如示例 `config.json`）。
- 建议添加 `.gitignore` 文件，排除 `dynv6.log` 和 `dynv6` 二进制文件。

此 README 为中文用户提供了清晰的开源项目使用指南。如需帮助设置 GitHub 仓库或添加其他内容，请告知！