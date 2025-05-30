package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config 定义程序的配置结构，Caddy 风格
type Config struct {
	GlobalSettings struct {
		UpdateIntervalSeconds int    `json:"更新间隔秒"`
		NetworkInterface     string `json:"网络接口"`
		IPType               string `json:"IP类型"`
	} `json:"全局设置"`
	DomainList struct {
		Domains []string `json:"域名"`
		Token   string   `json:"令牌"`
	} `json:"域名列表"`
	NotificationSettings struct {
		TelegramBotToken string `json:"Telegram机器人令牌"`
		TelegramChatID   string `json:"Telegram聊天ID"`
	} `json:"通知设置"`
}

// loadConfig 加载配置文件
func loadConfig(filePath string) (Config, bool, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, true, nil // 文件不存在，需引导用户输入
		}
		return config, false, fmt.Errorf("无法读取配置文件 %s: %v", filePath, err)
	}
	if len(data) == 0 {
		return config, true, fmt.Errorf("配置文件 %s 为空", filePath)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return config, false, fmt.Errorf("无法解析配置文件 %s: %v", filePath, err)
	}
	return config, false, nil
}

// isRunningInBackground 检测是否在后台运行
func isRunningInBackground() bool {
	// 检查标准输入是否为终端
	stat, err := os.Stdin.Stat()
	if err != nil {
		return true
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// getAvailableInterfaces 获取设备可用网络接口
func getAvailableInterfaces() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("无法获取网络接口: %v", err)
	}
	var available []string
	for _, i := range interfaces {
		if i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 {
			continue
		}
		available = append(available, i.Name)
	}
	if len(available) == 0 {
		return nil, fmt.Errorf("未找到可用网络接口")
	}
	return available, nil
}

// createConfigFromInput 交互式引导用户输入配置
func createConfigFromInput(filePath string) (Config, error) {
	if isRunningInBackground() {
		fmt.Fprintln(os.Stderr, "错误：首次运行需在前台运行以输入配置，请使用 './dynv6'（不要加 '&'）")
		os.Exit(1)
	}

	var config Config
	scanner := bufio.NewScanner(os.Stdin)

	// 打印文件路径以供调试
	log.Printf("将要写入的配置文件路径: %s", filePath)

	// 引导输入域名
	fmt.Println("请输入 dynv6 域名（多个域名用英文逗号分隔，例如 example.dns.navy,test.dns.navy）")
	fmt.Println("- 说明：输入你的 dynv6 域名，例如 example.dns.navy，多个域名用英文逗号分隔")
	fmt.Println("- 直接回车使用默认值 example.dns.navy（需后续修改）")
	for {
		scanner.Scan()
		domains := strings.TrimSpace(scanner.Text())
		if domains == "" {
			config.DomainList.Domains = []string{"example.dns.navy"}
			break
		}
		config.DomainList.Domains = strings.Split(domains, ",")
		for i, d := range config.DomainList.Domains {
			config.DomainList.Domains[i] = strings.TrimSpace(d)
		}
		if len(config.DomainList.Domains) > 0 && config.DomainList.Domains[0] != "" {
			break
		}
		fmt.Println("错误：域名不能为空，请重新输入")
	}

	// 引导输入令牌
	fmt.Println("\n请输入 dynv6 API 令牌（从 dynv6.com 的 Zones 页面获取）")
	fmt.Println("- 说明：在 dynv6 网站登录，进入 Zones，复制你的 API 令牌")
	fmt.Println("- 直接回车使用默认值 your_token_here（需后续修改）")
	scanner.Scan()
	config.DomainList.Token = strings.TrimSpace(scanner.Text())
	if config.DomainList.Token == "" {
		config.DomainList.Token = "your_token_here"
	}

	// 引导输入更新间隔
	fmt.Println("\n请输入更新间隔（秒，建议 300，即 5 分钟）")
	fmt.Println("- 说明：程序检查和更新 IP 的时间间隔，单位为秒，建议 300")
	fmt.Println("- 直接回车使用默认值 300")
	for {
		scanner.Scan()
		intervalStr := strings.TrimSpace(scanner.Text())
		if intervalStr == "" {
			config.GlobalSettings.UpdateIntervalSeconds = 300
			break
		}
		if _, err := fmt.Sscanf(intervalStr, "%d", &config.GlobalSettings.UpdateIntervalSeconds); err == nil {
			if config.GlobalSettings.UpdateIntervalSeconds > 0 {
				break
			}
		}
		fmt.Println("错误：更新间隔必须为正整数，请重新输入")
	}

	// 引导输入网络接口
	fmt.Println("\n请输入网络接口（根据设备网络选择）")
	interfaces, err := getAvailableInterfaces()
	if err != nil {
		fmt.Printf("警告：无法获取网络接口列表: %v\n", err)
		fmt.Println("- 说明：输入网络接口名称，例如 wlan0, rmnet0")
	} else {
		fmt.Printf("- 说明：可用接口：%s\n", strings.Join(interfaces, ", "))
	}
	fmt.Println("- 直接回车使用默认值 wlan0")
	for {
		scanner.Scan()
		config.GlobalSettings.NetworkInterface = strings.TrimSpace(scanner.Text())
		if config.GlobalSettings.NetworkInterface == "" {
			config.GlobalSettings.NetworkInterface = "wlan0"
		}
		if _, err := net.InterfaceByName(config.GlobalSettings.NetworkInterface); err == nil {
			break
		}
		fmt.Printf("错误：网络接口 %s 不存在，请重新输入\n", config.GlobalSettings.NetworkInterface)
		if len(interfaces) > 0 {
			fmt.Printf("- 可用接口：%s\n", strings.Join(interfaces, ", "))
		}
	}

	// 引导输入 IP 类型
	fmt.Println("\n请输入 IP 类型（可选值：ipv4, ipv6, dual）")
	fmt.Println("- 说明：ipv4（仅 IPv4）、ipv6（仅 IPv6）、dual（双栈，IPv4 和 IPv6），建议根据设备支持选择")
	fmt.Println("- 直接回车使用默认值 ipv6")
	for {
		scanner.Scan()
		config.GlobalSettings.IPType = strings.TrimSpace(scanner.Text())
		if config.GlobalSettings.IPType == "" {
			config.GlobalSettings.IPType = "ipv6"
		}
		if config.GlobalSettings.IPType == "ipv4" || config.GlobalSettings.IPType == "ipv6" || config.GlobalSettings.IPType == "dual" {
			break
		}
		fmt.Println("错误：IP 类型必须为 ipv4, ipv6 或 dual，请重新输入")
	}

	// 引导输入 Telegram 机器人令牌
	fmt.Println("\n请输入 Telegram 机器人令牌（可选，留空禁用通知）")
	fmt.Println("- 说明：Telegram 机器人令牌，用于发送更新通知，留空禁用")
	fmt.Println("- 直接回车留空")
	scanner.Scan()
	config.NotificationSettings.TelegramBotToken = strings.TrimSpace(scanner.Text())

	// 引导输入 Telegram 聊天 ID
	fmt.Println("\n请输入 Telegram 聊天 ID（可选，留空禁用通知）")
	fmt.Println("- 说明：Telegram 聊天 ID，与机器人令牌一起使用，留空禁用")
	fmt.Println("- 直接回车留空")
	scanner.Scan()
	config.NotificationSettings.TelegramChatID = strings.TrimSpace(scanner.Text())

	// 打印配置内容以供调试
	log.Printf("配置结构体内容: %+v", config)

	// 生成纯 JSON 配置文件
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Printf("无法序列化配置: %v", err)
		return config, fmt.Errorf("无法序列化配置: %v", err)
	}
	log.Printf("生成的 JSON 内容:\n%s", string(configJSON))

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Printf("无法创建配置文件目录 %s: %v", filepath.Dir(filePath), err)
		return config, fmt.Errorf("无法创建配置文件目录 %s: %v", filepath.Dir(filePath), err)
	}

	// 写入文件
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("无法创建配置文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法创建配置文件 %s: %v", filePath, err)
	}
	defer file.Close()

	if _, err := file.Write(configJSON); err != nil {
		log.Printf("无法写入配置文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法写入配置文件 %s: %v", filePath, err)
	}
	if err := file.Sync(); err != nil {
		log.Printf("无法同步配置文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法同步配置文件 %s: %v", filePath, err)
	}

	// 验证文件内容（不再打印到日志）
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("无法读取新生成的文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法读取新生成的文件 %s: %v", filePath, err)
	}
	if len(data) == 0 {
		log.Printf("配置文件 %s 为空", filePath)
		return config, fmt.Errorf("配置文件 %s 为空", filePath)
	}

	// 重新加载配置文件以验证
	var loadedConfig Config
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		log.Printf("无法解析新生成的配置文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法解析新生成的配置文件 %s: %v", filePath, err)
	}

	// 简洁的写入确认
	log.Printf("配置文件已成功写入: %s", filePath)
	fmt.Printf("\n配置文件已生成: %s\n", filePath)
	fmt.Println("程序将继续在前台运行，日志同时保存至 dynv6.log")
	fmt.Println("下次启动指令：")
	fmt.Println("- 前台运行：./dynv6")
	fmt.Println("- 后台运行：termux-wake-lock && nohup ./dynv6 &")

	// 重新加载 config.json 以确保使用文件内容
	config, _, err = loadConfig(filePath)
	if err != nil {
		log.Printf("无法加载新生成的配置文件 %s: %v", filePath, err)
		return config, fmt.Errorf("无法加载新生成的配置文件 %s: %v", filePath, err)
	}
	return config, nil
}

// initLogger 初始化日志，输出到文件和终端
func initLogger(logFile string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("无法打开日志文件 %s: %v", logFile, err)
	}
	// 日志同时输出到文件和终端
	log.SetOutput(io.MultiWriter(file, os.Stdout))
	return nil
}

// sendTelegramNotification 发送 Telegram 通知
func sendTelegramNotification(botToken, chatID, message string) error {
	if botToken == "" || chatID == "" {
		return nil // 未配置 Telegram，跳过
	}
	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		botToken, chatID, url.QueryEscape(message))
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("发送 Telegram 通知失败: %v", err)
	}
	defer resp.Body.Close()
	return nil
}

// getPublicIPs 获取指定网络接口的公网 IPv4 和 IPv6 地址
func getPublicIPs(networkInterface, ipType string) (ipv4, ipv6 string, err error) {
	var iface *net.Interface
	if networkInterface != "" {
		iface, err = net.InterfaceByName(networkInterface)
		if err != nil {
			return "", "", fmt.Errorf("无法找到网络接口 %s: %v", networkInterface, err)
		}
	} else {
		interfaces, err := net.Interfaces()
		if err != nil {
			return "", "", fmt.Errorf("无法获取网络接口: %v", err)
		}
		for _, i := range interfaces {
			if i.Flags&net.FlagLoopback != 0 {
				continue
			}
			if i.Flags&net.FlagUp == 0 {
				continue
			}
			iface = &i
			break
		}
	}
	if iface == nil {
		return "", "", fmt.Errorf("未找到可用网络接口")
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", "", fmt.Errorf("无法获取接口 %s 的地址: %v", iface.Name, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		if ipType == "ipv4" || ipType == "dual" {
			if ip.To4() != nil && !ip.IsLoopback() && !ip.IsPrivate() {
				ipv4 = ip.String()
			}
		}
		if ipType == "ipv6" || ipType == "dual" {
			if ip.To4() == nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
				ipv6 = ip.String()
			}
		}
	}

	if (ipType == "ipv4" && ipv4 == "") || (ipType == "ipv6" && ipv6 == "") || (ipType == "dual" && ipv4 == "" && ipv6 == "") {
		return "", "", fmt.Errorf("未找到符合要求的公网 IP 地址（接口: %s, IP类型: %s）", networkInterface, ipType)
	}
	return ipv4, ipv6, nil
}

// updateDynv6 更新 dynv6 的 IP 记录
func updateDynv6(ipv4, ipv6, domain, token string) error {
	urlStr := fmt.Sprintf("http://dynv6.com/api/update?hostname=%s&token=%s", domain, token)
	if ipv6 != "" {
		urlStr += fmt.Sprintf("&ipv6=%s", ipv6)
	}
	if ipv4 != "" {
		urlStr += fmt.Sprintf("&ipv4=%s", ipv4)
	}
	resp, err := http.Get(urlStr)
	if err != nil {
		return fmt.Errorf("发送更新请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if !strings.Contains(string(body), "updated") {
		return fmt.Errorf("更新失败，响应: %s", string(body))
	}

	log.Printf("成功更新域名 %s: IPv4=%s, IPv6=%s", domain, ipv4, ipv6)
	return nil
}

// updateWithRetry 带重试机制的更新
func updateWithRetry(ipv4, ipv6, domain, token string, maxRetries int, baseDelay time.Duration) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := updateDynv6(ipv4, ipv6, domain, token)
		if err == nil {
			return nil
		}
		log.Printf("域名 %s 第 %d/%d 次尝试失败: %v", domain, attempt, maxRetries, err)
		time.Sleep(baseDelay * time.Duration(1<<uint(attempt-1)))
	}
	return fmt.Errorf("域名 %s 尝试 %d 次后失败", domain, maxRetries)
}

func main() {
	// 获取程序所在目录，确保日志和配置文件在同级目录
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取程序路径: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	logFile := filepath.Join(exeDir, "dynv6.log")
	configFile := filepath.Join(exeDir, "config.json")

	// 打印配置文件路径以供调试
	log.Printf("配置文件路径: %s", configFile)

	// 初始化日志，输出到文件和终端
	if err := initLogger(logFile); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 加载配置文件
	config, needConfig, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 如果需要配置，引导用户输入
	if needConfig {
		config, err = createConfigFromInput(configFile)
		if err != nil {
			log.Fatalf("生成配置文件失败: %v", err)
		}
	}

	// 验证配置
	if len(config.DomainList.Domains) == 0 {
		log.Fatal("配置文件中未指定任何域名")
	}
	if config.DomainList.Token == "" || config.DomainList.Token == "your_token_here" {
		log.Fatal("请在 config.json 中设置有效的 dynv6 令牌")
	}
	if config.GlobalSettings.NetworkInterface == "" {
		log.Fatal("网络接口不能为空")
	}
	if config.GlobalSettings.IPType != "ipv4" && config.GlobalSettings.IPType != "ipv6" && config.GlobalSettings.IPType != "dual" {
		log.Fatal("IP类型必须为 ipv4, ipv6 或 dual")
	}

	log.Println("启动 dynv6 IP 更新程序...")
	更新间隔 := time.Duration(config.GlobalSettings.UpdateIntervalSeconds) * time.Second
	最后IPs := make(map[string]string) // 缓存每个域名的最后 IPv6

	for {
		// 获取公网 IP
		ipv4, ipv6, err := getPublicIPs(config.GlobalSettings.NetworkInterface, config.GlobalSettings.IPType)
		if err != nil {
			log.Printf("获取 IP 失败: %v", err)
			sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("获取 IP 失败: %v", err))
			time.Sleep(更新间隔)
			continue
		}

		// 遍历域名列表
		for _, domain := range config.DomainList.Domains {
			// 检查 IPv6 是否变化
			if config.GlobalSettings.IPType == "ipv6" || config.GlobalSettings.IPType == "dual" {
				if ipv6 == 最后IPs[domain] {
					log.Printf("域名 %s 的 IPv6 未变化: %s，跳过更新", domain, ipv6)
					continue
				}
			}

			// 更新 dynv6
			err = updateWithRetry(ipv4, ipv6, domain, config.DomainList.Token, 3, 10*time.Second)
			if err != nil {
				log.Printf("更新域名 %s 失败: %v", domain, err)
				sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("更新域名 %s 失败: %v", domain, err))
			} else {
				最后IPs[domain] = ipv6
				sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("成功更新域名 %s: IPv4=%s, IPv6=%s", domain, ipv4, ipv6))
			}
		}

		time.Sleep(更新间隔)
	}
}