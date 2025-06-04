package main

import (
	"database/sql"
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
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config defines the configuration structure
type Config struct {
	GlobalSettings struct {
		UpdateIntervalSeconds int    `json:"updateIntervalSeconds"`
		NetworkInterface      string `json:"networkInterface"`
		IPType                string `json:"ipType"`
	} `json:"globalSettings"`
	DomainList struct {
		Domains []string `json:"domains"`
		Token   string   `json:"token"`
	} `json:"domainList"`
	NotificationSettings struct {
		TelegramBotToken string `json:"telegramBotToken"`
		TelegramChatID   string `json:"telegramChatID"`
	} `json:"notificationSettings"`
}

// initDB initializes the SQLite database
func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库 %s: %v", dbPath, err)
	}

	// Create config table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		update_interval_seconds INTEGER NOT NULL,
		network_interface TEXT NOT NULL,
		ip_type TEXT NOT NULL,
		domains TEXT NOT NULL,
		token TEXT NOT NULL,
		telegram_bot_token TEXT,
		telegram_chat_id TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("无法创建配置表: %v", err)
	}
	log.Printf("数据库初始化成功: %s", dbPath)
	return db, nil
}

// loadConfigFromDB loads configuration from SQLite
func loadConfigFromDB(db *sql.DB) (Config, bool, error) {
	var config Config
	row := db.QueryRow(`
		SELECT update_interval_seconds, network_interface, ip_type, domains, token, telegram_bot_token, telegram_chat_id
		FROM config
		ORDER BY id DESC LIMIT 1`)

	var domainsJSON string
	err := row.Scan(
		&config.GlobalSettings.UpdateIntervalSeconds,
		&config.GlobalSettings.NetworkInterface,
		&config.GlobalSettings.IPType,
		&domainsJSON,
		&config.DomainList.Token,
		&config.NotificationSettings.TelegramBotToken,
		&config.NotificationSettings.TelegramChatID,
	)
	if err == sql.ErrNoRows {
		log.Println("数据库中没有配置")
		return config, false, nil
	}
	if err != nil {
		log.Printf("无法从数据库加载配置: %v", err)
		return config, false, fmt.Errorf("无法从数据库加载配置: %v", err)
	}

	// Parse domains list
	if err := json.Unmarshal([]byte(domainsJSON), &config.DomainList.Domains); err != nil {
		log.Printf("无法解析域名列表: %v", err)
		return config, false, fmt.Errorf("无法解析域名列表: %v", err)
	}

	log.Printf("成功加载配置: %+v", config)
	return config, true, nil
}

// saveConfigToDB saves configuration to SQLite
func saveConfigToDB(db *sql.DB, config Config) error {
	domainsJSON, err := json.Marshal(config.DomainList.Domains)
	if err != nil {
		log.Printf("无法序列化域名列表: %v", err)
		return fmt.Errorf("无法序列化域名列表: %v", err)
	}
	log.Printf("保存配置: %+v, domainsJSON: %s", config, domainsJSON) // Debug log

	// Delete old configuration
	_, err = db.Exec(`DELETE FROM config`)
	if err != nil {
		log.Printf("无法清除旧配置: %v", err)
		return fmt.Errorf("无法清除旧配置: %v", err)
	}

	// Insert new configuration
	_, err = db.Exec(`
		INSERT INTO config (update_interval_seconds, network_interface, ip_type, domains, token, telegram_bot_token, telegram_chat_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		config.GlobalSettings.UpdateIntervalSeconds,
		config.GlobalSettings.NetworkInterface,
		config.GlobalSettings.IPType,
		string(domainsJSON),
		config.DomainList.Token,
		config.NotificationSettings.TelegramBotToken,
		config.NotificationSettings.TelegramChatID,
	)
	if err != nil {
		log.Printf("插入配置失败: %v", err)
		return fmt.Errorf("无法保存配置到数据库: %v", err)
	}
	log.Printf("成功插入配置到数据库: %+v", config)
	return nil
}

// initLogger initializes logging
func initLogger(logFile string) error {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("无法打开日志文件 %s: %v", logFile, err)
	}
	log.SetOutput(io.MultiWriter(file, os.Stdout))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

// sendTelegramNotification sends a Telegram notification
func sendTelegramNotification(botToken, chatID, message string) error {
	if botToken == "" || chatID == "" {
		return nil
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

// getPublicIPs retrieves public IP addresses
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
			if i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 {
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

// updateDynv6 updates dynv6 IP records
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

// updateWithRetry updates with retry mechanism
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

// startUpdateLoop runs the IP update loop
func startUpdateLoop(db *sql.DB, startSignal chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	<-startSignal // Wait for start signal

	for {
		config, exists, err := loadConfigFromDB(db)
		if err != nil {
			log.Printf("加载配置失败: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if !exists {
			log.Printf("数据库中没有配置，等待配置保存")
			time.Sleep(10 * time.Second)
			continue
		}

		updateInterval := time.Duration(config.GlobalSettings.UpdateIntervalSeconds) * time.Second
		lastIPs := make(map[string]string)

		for {
			ipv4, ipv6, err := getPublicIPs(config.GlobalSettings.NetworkInterface, config.GlobalSettings.IPType)
			if err != nil {
				log.Printf("获取 IP 失败: %v", err)
				sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("获取 IP 失败: %v", err))
				time.Sleep(updateInterval)
				continue
			}

			for _, domain := range config.DomainList.Domains {
				if config.GlobalSettings.IPType == "ipv6" || config.GlobalSettings.IPType == "dual" {
					if ipv6 == lastIPs[domain] {
						log.Printf("域名 %s 的 IPv6 未变化: %s，跳过更新", domain, ipv6)
						continue
					}
				}

				err = updateWithRetry(ipv4, ipv6, domain, config.DomainList.Token, 3, 10*time.Second)
				if err != nil {
					log.Printf("更新域名 %s 失败: %v", domain, err)
					sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("更新域名 %s 失败: %v", domain, err))
				} else {
					lastIPs[domain] = ipv6
					sendTelegramNotification(config.NotificationSettings.TelegramBotToken, config.NotificationSettings.TelegramChatID, fmt.Sprintf("成功更新域名 %s: IPv4=%s, IPv6=%s", domain, ipv4, ipv6))
				}
			}

			time.Sleep(updateInterval)
		}
	}
}

func main() {
	// Get program directory
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取程序路径: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	logFile := filepath.Join(exeDir, "dynv6.log")
	dbPath := filepath.Join(exeDir, "config.db")

	// Initialize logger
	if err := initLogger(logFile); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// Initialize database
	db, err := initDB(dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(exeDir, "static")))))

	// Control update loop start
	startSignal := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	// Start update loop (waits for config)
	go startUpdateLoop(db, startSignal, &wg)

	// Start web server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(exeDir, "static", "index.html"))
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			config, exists, err := loadConfigFromDB(db)
			if err != nil {
				log.Printf("GET /api/config 失败: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !exists {
				log.Println("GET /api/config: 数据库中没有配置")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "{}") // Return empty object
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(config); err != nil {
				log.Printf("GET /api/config 编码响应失败: %v", err)
				http.Error(w, "无法编码配置数据", http.StatusInternalServerError)
				return
			}
			log.Println("GET /api/config: 成功返回配置")

		case http.MethodPost:
			var newConfig Config
			if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
				log.Printf("POST /api/config 解码请求失败: %v", err)
				http.Error(w, "无效的配置数据", http.StatusBadRequest)
				return
			}

			// Validate configuration
			if len(newConfig.DomainList.Domains) == 0 {
				log.Println("POST /api/config: 域名列表为空")
				http.Error(w, "域名列表不能为空", http.StatusBadRequest)
				return
			}
			if newConfig.DomainList.Token == "" {
				log.Println("POST /api/config: 令牌为空")
				http.Error(w, "令牌不能为空", http.StatusBadRequest)
				return
			}
			if newConfig.GlobalSettings.NetworkInterface == "" {
				log.Println("POST /api/config: 网络接口为空")
				http.Error(w, "网络接口不能为空", http.StatusBadRequest)
				return
			}
			if newConfig.GlobalSettings.IPType != "ipv4" && newConfig.GlobalSettings.IPType != "ipv6" && newConfig.GlobalSettings.IPType != "dual" {
				log.Println("POST /api/config: 无效的 IP 类型")
				http.Error(w, "IP类型必须为 ipv4, ipv6 或 dual", http.StatusBadRequest)
				return
			}
			if newConfig.GlobalSettings.UpdateIntervalSeconds <= 0 {
				log.Println("POST /api/config: 无效的更新间隔")
				http.Error(w, "更新间隔必须为正整数", http.StatusBadRequest)
				return
			}

			if err := saveConfigToDB(db, newConfig); err != nil {
				log.Printf("POST /api/config 保存配置失败: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Trigger update loop (only on first save)
			select {
			case startSignal <- struct{}{}:
				log.Println("收到配置，启动 IP 更新循环")
			default:
				log.Println("配置更新，IP 更新循环已运行")
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "配置已保存")
			log.Println("POST /api/config: 配置已保存")
		default:
			log.Printf("不支持的请求方法: %s", r.Method)
			http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
		}
	})

	log.Println("启动 Web 服务器，访问 http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("启动 Web 服务器失败: %v", err)
	}

	wg.Wait()
}
