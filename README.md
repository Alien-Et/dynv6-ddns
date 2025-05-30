### README.md

```markdown
# Dynv6 IP Updater

[Dynv6 IP Updater](https://github.com/Alien-Et/dynv6-ddns) is a Go command-line tool for dynamically updating IPv4 and/or IPv6 addresses for domains hosted on [dynv6.com](https://dynv6.com). It supports multi-domain management, customizable update intervals, network interface selection, and Telegram notifications, running on Linux, Windows, macOS, and Android (Termux).

[中文 README (Chinese)](README.zh-CN.md)

## Features
- Dynamically updates IPv4 and/or IPv6 addresses for multiple dynv6 domains.
- Configurable update intervals (seconds) and network interfaces.
- Sends Telegram notifications for update success or failure.
- Interactive configuration on first run, generating `config.json`.
- Logs to `dynv6.log` and terminal for debugging.
- Supports foreground and background execution (Linux/Android with `nohup`).
- Detects IPv6 changes to avoid unnecessary updates.
- Retries failed updates 3 times with exponential backoff.

## Installation and Compilation

### Requirements
- [Go](https://golang.org) 1.16 or higher.
- [Git](https://git-scm.com) for cloning the repository.
- Android requires [Termux](https://f-droid.org/en/packages/com.termux/) (download from F-Droid).
- Android Root users can run in any `/data` directory or package as a Magisk module.

### Clone Repository
```bash
git clone https://github.com/Alien-Et/dynv6-ddns.git
cd dynv6-ddns
```

### Platform Compilation
1. **Linux**:
   ```bash
   sudo apt update && sudo apt install golang  # Debian/Ubuntu
   sudo dnf install golang  # Fedora
   go build -o dynv6
   ```
   Output: `dynv6` executable.

2. **Windows**:
   - Install Go: [Go Download](https://golang.org/dl/).
   ```bash
   go build -o dynv6.exe
   ```
   Output: `dynv6.exe` executable.

3. **macOS**:
   ```bash
   brew install go
   go build -o dynv6
   ```
   Output: `dynv6` executable.

4. **Android (Termux)**:
   - Install Termux (F-Droid or GitHub version).
   ```bash
   pkg update && pkg install golang git
   git clone https://github.com/Alien-Et/dynv6-ddns.git
   cd dynv6-ddns
   go build -o dynv6
   termux-setup-storage
   ```
   Output: `dynv6` executable. `termux-setup-storage` ensures storage permissions.
   - **Root Users**: Run in any `/data` directory by moving `dynv6` and `config.json` to the target directory.
   - **Magisk Module**: Advanced users can package the program as a Magisk module for auto-start and persistent execution (requires custom module scripting).

## Usage
1. **First Run**:
   Run to generate `config.json`:
   ```bash
   ./dynv6  # Linux/macOS/Android
   dynv6.exe  # Windows
   ```
   Follow prompts to input:
   - Domains (comma-separated, e.g., `example.dns.navy,test.dns.navy`).
   - dynv6 API token (from [dynv6.com Zones](https://dynv6.com/zones)).
   - Update interval (seconds, recommended 300).
   - Network interface (e.g., `wlan0`, `rmnet0` for Linux/Android; `Ethernet` for Windows).
   - IP type (`ipv4`, `ipv6`, or `dual`).
   - Telegram bot token and chat ID (optional, leave empty to disable).

2. **Subsequent Runs**:
   - **Foreground**:
     ```bash
     ./dynv6  # Linux/macOS/Android
     dynv6.exe  # Windows
     ```
   - **Background**:
     ```bash
     termux-wake-lock && nohup ./dynv6 &  # Android (Termux)/Linux
     ```
     Windows: Use Task Scheduler for background tasks.

3. **Configuration File**:
   - Location: `config.json` in the same directory as the executable.
   - Edit: Modify `config.json` to adjust settings.

4. **Logs**:
   - Saved to `dynv6.log` in the executable directory.
   - Also output to the terminal for real-time monitoring.

## Configuration
`config.json` uses a Caddy-style JSON structure. Example:
```json
{
    "GlobalSettings": {
        "UpdateIntervalSeconds": 300,
        "NetworkInterface": "wlan0",
        "IPType": "ipv6"
    },
    "DomainList": {
        "Domains": ["example.dns.navy", "test.dns.navy"],
        "Token": "your_token_here"
    },
    "NotificationSettings": {
        "TelegramBotToken": "bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
        "TelegramChatID": "123456789"
    }
}
```

### Configuration Details
- **GlobalSettings**:
  - `UpdateIntervalSeconds`: IP check interval (seconds, recommended 300, i.e., 5 minutes).
  - `NetworkInterface`: Interface name, e.g., `wlan0` (Android/Linux), `Ethernet` (Windows).
  - `IPType`: `ipv4` (IPv4 only), `ipv6` (IPv6 only), or `dual` (both).
- **DomainList**:
  - `Domains`: List of dynv6 domains, e.g., `example.dns.navy`.
  - `Token`: dynv6 API token from [dynv6.com Zones](https://dynv6.com/zones).
- **NotificationSettings**:
  - `TelegramBotToken`: Telegram bot token (from BotFather), leave empty to disable.
  - `TelegramChatID`: Chat ID (from bot interaction), leave empty to disable.

## Notes
- Updates only when IPv6 changes to reduce unnecessary requests.
- Ensure a valid dynv6 API token and existing network interface (e.g., `wlan0`, `rmnet0` on Android).
- Android/Termux: Use `termux-wake-lock` to prevent process termination.
- Failed updates retry 3 times with exponential backoff (starting at 10 seconds).
- Android: Ensure storage permissions with `termux-setup-storage` for config and logs.
- Windows: Background execution requires Task Scheduler or service setup.
- Android Root: Deploy to any `/data` directory or create a Magisk module for auto-start.

## License
MIT License, see [LICENSE](LICENSE).
```