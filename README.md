Below is a **GitHub README** for the provided Go script, designed to serve as a usage guide for the open-source community. It follows standard conventions for open-source projects, providing clear instructions, configuration details, and contribution guidelines.

---

# Dynv6 Dynamic DNS Updater

A lightweight Go-based command-line tool to update dynamic DNS records on [dynv6.com](https://dynv6.com). This tool periodically checks the public IP address of a specified network interface and updates the DNS records for configured domains using the dynv6 API. It supports both IPv4 and IPv6, with optional Telegram notifications for status updates.

## Features

- **Dynamic DNS Updates**: Automatically updates IPv4 and/or IPv6 records for multiple domains on dynv6.
- **Interactive Configuration**: Guides users to create a configuration file on first run.
- **Network Interface Support**: Selects public IPs from a specified network interface.
- **IP Type Flexibility**: Supports IPv4, IPv6, or dual-stack updates.
- **Telegram Notifications**: Sends update status notifications via Telegram (optional).
- **Retry Mechanism**: Retries failed updates with exponential backoff.
- **Logging**: Outputs logs to both a file (`dynv6.log`) and the terminal.
- **Cross-Platform**: Runs on Linux, macOS, Windows, and platforms like Termux on Android.

## Prerequisites

- **Go**: Version 1.16 or later (required to build the tool).
- **dynv6 Account**: An account with API token from [dynv6.com](https://dynv6.com).
- **Network Interface**: A device with a network interface that has a public IP address.
- **Telegram (Optional)**: A Telegram bot token and chat ID for notifications.

## Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/yourusername/dynv6-updater.git
   cd dynv6-updater
   ```

2. **Build the Binary**:
   ```bash
   go build -o dynv6 main.go
   ```

3. **Run the Program**:
   ```bash
   ./dynv6
   ```

   On first run, the program will prompt you to create a configuration file (`config.json`) interactively.

## Usage

### First Run
When you run the program for the first time (or if `config.json` is missing), it will guide you through an interactive setup to create the configuration file. You will be prompted to provide:

- **Domains**: Your dynv6 domains (e.g., `example.dns.navy,test.dns.navy`).
- **API Token**: Your dynv6 API token from the [Zones page](https://dynv6.com/zones).
- **Update Interval**: Time between IP checks (in seconds, default: 300).
- **Network Interface**: The network interface to monitor (e.g., `wlan0`, `eth0`).
- **IP Type**: `ipv4`, `ipv6`, or `dual` (default: `ipv6`).
- **Telegram Bot Token** (optional): For sending notifications.
- **Telegram Chat ID** (optional): For receiving notifications.

Example interaction:
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

The configuration is saved to `config.json` in the same directory as the executable.

### Configuration File
The program uses a `config.json` file with the following structure:

```json
{
    "全局设置": {
        "更新间隔秒": 300,
        "网络接口": "wlan0",
        "IP类型": "ipv6"
    },
    "域名列表": {
        "域名": ["example.dns.navy"],
        "令牌": "your_api_token"
    },
    "通知设置": {
        "Telegram机器人令牌": "your_bot_token",
        "Telegram聊天ID": "your_chat_id"
    }
}
```

- **全局设置**:
  - `更新间隔秒`: Update interval in seconds (e.g., `300` for 5 minutes).
  - `网络接口`: Network interface name (e.g., `wlan0`, `eth0`).
  - `IP类型`: IP type to update (`ipv4`, `ipv6`, or `dual`).
- **域名列表**:
  - `域名`: List of dynv6 domains (e.g., `["example.dns.navy", "test.dns.navy"]`).
  - `令牌`: dynv6 API token.
- **通知设置**:
  - `Telegram机器人令牌`: Telegram bot token (optional, leave empty to disable).
  - `Telegram聊天ID`: Telegram chat ID (optional, leave empty to disable).

You can manually edit `config.json` to update settings after the initial setup.

### Running the Program
- **Foreground**:
  ```bash
  ./dynv6
  ```
  Logs are displayed in the terminal and saved to `dynv6.log`.

- **Background (e.g., on Linux or Termux)**:
  ```bash
  nohup ./dynv6 &
  ```
  For Termux, ensure the device stays awake:
  ```bash
  termux-wake-lock && nohup ./dynv6 &
  ```

### Logs
Logs are written to `dynv6.log` in the same directory as the executable. They include:
- Configuration loading details.
- IP address detection results.
- DNS update successes or failures.
- Telegram notification status.

## Example Workflow
1. Run `./dynv6` for the first time.
2. Follow the interactive prompts to create `config.json`.
3. The program starts monitoring the specified network interface and updates DNS records every 300 seconds (or as configured).
4. If configured, Telegram notifications are sent for each update or error.

## Troubleshooting
- **Configuration File Errors**:
  - Ensure `config.json` is valid JSON and contains all required fields.
  - Verify the dynv6 API token is correct.
- **Network Interface Issues**:
  - Run `ifconfig` or `ip addr` to list available interfaces.
  - Ensure the specified interface has a public IP.
- **No Public IP**:
  - Check if the device is behind a NAT or lacks a public IP.
  - Verify the `IP类型` matches your network setup.
- **Telegram Notifications**:
  - Ensure the bot token and chat ID are correct.
  - Test the bot by sending a message manually via the Telegram API.

## Contributing
Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m "Add your feature"`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a Pull Request.

Please ensure your code follows the existing style and includes appropriate tests.

### Development Setup
- Install Go 1.16+.
- Use `go fmt` to format code.
- Add tests in the `test` directory (coming soon).
- Build and test locally: `go build && ./dynv6`.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments
- Thanks to [dynv6.com](https://dynv6.com) for providing a free dynamic DNS service.
- Inspired by the need for a simple, cross-platform DDNS client.

## Contact
For issues or feature requests, open an issue on GitHub or contact the maintainer at [your.email@example.com](mailto:your.email@example.com).

---

### Notes for Implementation
- Replace `yourusername` in the clone command with your actual GitHub username.
- Add a `LICENSE` file to the repository (e.g., MIT License) to make it truly open-source.
- If you add tests or additional features, update the "Contributing" section with specific instructions.
- Ensure the repository includes the `main.go` file and any necessary assets (e.g., sample `config.json`).
- Consider adding a `.gitignore` file to exclude `dynv6.log` and `dynv6` binary.

This README provides a clear, community-friendly guide for using and contributing to your project. Let me know if you need help setting up the GitHub repository or additional sections!