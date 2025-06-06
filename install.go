package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Install 安装 frps
func (fm *FrpsManager) Install() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	// 检查是否已经安装
	if fm.isInstalled() {
		fm.Colors["green"].Println("frps 已经安装并正在运行。")
		fmt.Print("是否要重新安装 frps? (y/n): ")
		
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		
		if choice != "y" && choice != "yes" {
			fm.Colors["yellow"].Println("跳过安装。")
			return
		}
	}

	fm.Colors["green"].Println("开始安装 frps...")
	
	// 安装依赖包
	if err := fm.installDependencies(); err != nil {
		fm.Colors["red"].Printf("安装依赖包失败: %v\n", err)
		return
	}

	// 选择下载源
	downloadSource := fm.selectDownloadSource()
	
	// 获取最新版本
	if err := fm.getLatestVersion(downloadSource); err != nil {
		fm.Colors["red"].Printf("获取最新版本失败: %v\n", err)
		return
	}

	// 获取服务器IP
	serverIP := fm.getServerIP()
	fm.Colors["green"].Printf("服务器IP: %s\n", serverIP)

	// 收集用户配置
	if err := fm.collectUserConfig(serverIP); err != nil {
		fm.Colors["red"].Printf("收集配置失败: %v\n", err)
		return
	}

	// 显示配置确认
	fm.showConfigConfirmation(serverIP)
	
	fmt.Print("按任意键继续安装...或按 Ctrl+C 取消: ")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	// 执行安装
	if err := fm.performInstall(downloadSource); err != nil {
		fm.Colors["red"].Printf("安装失败: %v\n", err)
		return
	}

	fm.Colors["green"].Println("frps 安装完成！")
	fm.showInstallationSummary(serverIP)
}

// isInstalled 检查是否已安装
func (fm *FrpsManager) isInstalled() bool {
	// 检查进程是否运行
	cmd := exec.Command("pgrep", "-x", ProgramName)
	return cmd.Run() == nil
}

// installDependencies 安装依赖包
func (fm *FrpsManager) installDependencies() error {
	var installCmd []string
	
	switch fm.SystemInfo.OS {
	case "CentOS", "RHEL", "Rocky", "AlmaLinux":
		installCmd = []string{"yum", "install", "-y", "wget", "psmisc", "net-tools", "curl"}
	case "Ubuntu", "Debian":
		// 先更新包列表
		if err := exec.Command("apt-get", "-y", "update").Run(); err != nil {
			return fmt.Errorf("更新包列表失败: %v", err)
		}
		installCmd = []string{"apt-get", "-y", "install", "wget", "psmisc", "net-tools", "curl"}
	default:
		return fmt.Errorf("不支持的操作系统: %s", fm.SystemInfo.OS)
	}

	fm.Colors["green"].Println("正在安装依赖包...")
	cmd := exec.Command(installCmd[0], installCmd[1:]...)
	return cmd.Run()
}

// selectDownloadSource 选择下载源
func (fm *FrpsManager) selectDownloadSource() int {
	fmt.Println()
	fm.Colors["pink"].Println("请选择 frps 下载源:")
	fmt.Println("[1]. gitee")
	fmt.Println("[2]. github (默认)")
	
	fmt.Print("请选择 (1, 2 或 exit，默认[github]): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))
	
	switch choice {
	case "1", "gitee":
		fmt.Println("-----------------------------------")
		fm.Colors["yellow"].Println("       您选择了: gitee")
		fmt.Println("-----------------------------------")
		return 1
	case "exit":
		os.Exit(1)
		return 0 // This line will never be reached, but Go requires it
	default:
		fmt.Println("-----------------------------------")
		fm.Colors["yellow"].Println("       您选择了: github")
		fmt.Println("-----------------------------------")
		return 2
	}
}

// getLatestVersion 获取最新版本
func (fm *FrpsManager) getLatestVersion(source int) error {
	fm.Colors["green"].Println("正在获取最新版本...")
	
	var apiURL string
	if source == 1 {
		apiURL = GiteeLatestAPI
	} else {
		apiURL = GithubLatestAPI
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	// 移除版本号前的 'v'
	fm.SystemInfo.FrpsVersion = strings.TrimPrefix(release.TagName, "v")
	fm.Colors["green"].Printf("找到最新版本: %s\n", fm.SystemInfo.FrpsVersion)
	
	return nil
}

// getServerIP 获取服务器公网IP
func (fm *FrpsManager) getServerIP() string {
	fm.Colors["green"].Println("正在获取服务器IP...")
	
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		fm.Colors["yellow"].Println("获取IP失败，使用默认值")
		return "127.0.0.1"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "127.0.0.1"
	}

	return strings.TrimSpace(string(body))
}

// collectUserConfig 收集用户配置
func (fm *FrpsManager) collectUserConfig(serverIP string) error {
	fmt.Println()
	fm.Colors["red"].Println("————————————————————————————————————————————")
	fm.Colors["red"].Println("     请输入您的服务器设置:")
	fm.Colors["red"].Println("————————————————————————————————————————————")

	// 收集各项配置
	fm.Config.BindPort = fm.inputPort("bind_port", 5443)
	fm.Config.VhostHTTPPort = fm.inputPort("vhost_http_port", 80)
	fm.Config.VhostHTTPSPort = fm.inputPort("vhost_https_port", 443)
	fm.Config.DashboardPort = fm.inputPort("dashboard_port", 6443)
	
	fm.Config.DashboardUser = fm.inputString("dashboard_user", "admin")
	fm.Config.DashboardPwd = fm.inputString("dashboard_pwd", fm.generateRandomString(8))
	fm.Config.Token = fm.inputString("token", fm.generateRandomString(16))
	fm.Config.SubdomainHost = fm.inputString("subdomain_host", serverIP)
	
	fm.Config.MaxPoolCount = fm.inputNumber("max_pool_count", 5, 50)
	fm.Config.LogLevel = fm.selectLogLevel()
	fm.Config.LogMaxDays = fm.inputNumber("log_max_days", 3, 15)
	
	fm.Config.LogFile = fm.selectLogFile()
	fm.Config.TCPMux = fm.selectBoolOption("tcp_mux", true)
	
	transportProtocol := fm.selectBoolOption("transport protocol support", true)
	fm.Config.TransportProtocol = transportProtocol
	
	if transportProtocol {
		fm.Config.KCPBindPort = fm.inputPort("kcp_bind_port", fm.Config.BindPort)
		fm.Config.QuicBindPort = fm.inputPort("quic_bind_port", fm.Config.VhostHTTPSPort)
	}

	return nil
}

// inputPort 输入端口号
func (fm *FrpsManager) inputPort(name string, defaultValue int) int {
	for {
		fmt.Printf("请输入 %s [1-65535] (默认: %d): ", name, defaultValue)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			if fm.checkPort(defaultValue) {
				return defaultValue
			}
			continue
		}
		
		port, err := strconv.Atoi(input)
		if err != nil || port < 1 || port > 65535 {
			fm.Colors["red"].Println("输入错误！请输入正确的端口号。")
			continue
		}
		
		if fm.checkPort(port) {
			return port
		}
	}
}

// checkPort 检查端口是否被占用
func (fm *FrpsManager) checkPort(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fm.Colors["red"].Printf("错误：端口 %d 已被占用\n", port)
		return false
	}
	ln.Close()
	return true
}

// inputString 输入字符串
func (fm *FrpsManager) inputString(name, defaultValue string) string {
	fmt.Printf("请输入 %s (默认: %s): ", name, defaultValue)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		return defaultValue
	}
	return input
}

// inputNumber 输入数字
func (fm *FrpsManager) inputNumber(name string, defaultValue, maxValue int) int {
	for {
		fmt.Printf("请输入 %s [1-%d] (默认: %d): ", name, maxValue, defaultValue)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			return defaultValue
		}
		
		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > maxValue {
			fm.Colors["red"].Println("输入错误！请输入正确的数字。")
			continue
		}
		
		return num
	}
}

// selectLogLevel 选择日志级别
func (fm *FrpsManager) selectLogLevel() string {
	fmt.Println("请选择 log_level:")
	fmt.Println("1: info (默认)")
	fmt.Println("2: warn")
	fmt.Println("3: error")
	fmt.Println("4: debug")
	fmt.Println("5: trace")
	fmt.Println("-------------------------")
	
	fmt.Print("请选择 (1-5，默认[1]): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "2", "warn":
		return "warn"
	case "3", "error":
		return "error"
	case "4", "debug":
		return "debug"
	case "5", "trace":
		return "trace"
	default:
		return "info"
	}
}

// selectLogFile 选择日志文件
func (fm *FrpsManager) selectLogFile() string {
	fmt.Println("请选择 log_file:")
	fmt.Println("1: enable (默认)")
	fmt.Println("2: disable")
	fmt.Println("-------------------------")
	
	fmt.Print("请选择 (1, 2，默认[1]): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	if choice == "2" || strings.ToLower(choice) == "disable" {
		return "/dev/null"
	}
	// 使用绝对路径而不是相对路径，确保日志文件在正确的目录
	return filepath.Join(ProgramDir, "frps.log")
}

// selectBoolOption 选择布尔选项
func (fm *FrpsManager) selectBoolOption(name string, defaultValue bool) bool {
	defaultStr := "enable"
	if !defaultValue {
		defaultStr = "disable"
	}
	
	fmt.Printf("请选择 %s:\n", name)
	fmt.Println("1: enable (默认)")
	fmt.Println("2: disable")
	fmt.Println("-------------------------")
	
	fmt.Printf("请选择 (1, 2，默认[%s]): ", defaultStr)
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	if choice == "2" || strings.ToLower(choice) == "disable" {
		return false
	}
	return true
}

// generateRandomString 生成随机字符串
func (fm *FrpsManager) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// showConfigConfirmation 显示配置确认
func (fm *FrpsManager) showConfigConfirmation(serverIP string) {
	fmt.Println()
	fmt.Println("============== 请检查您的输入 ==============")
	fm.Colors["green"].Printf("服务器IP          : %s\n", serverIP)
	fm.Colors["green"].Printf("绑定端口          : %d\n", fm.Config.BindPort)
	fm.Colors["green"].Printf("vhost http 端口   : %d\n", fm.Config.VhostHTTPPort)
	fm.Colors["green"].Printf("vhost https 端口  : %d\n", fm.Config.VhostHTTPSPort)
	fm.Colors["green"].Printf("Dashboard 端口    : %d\n", fm.Config.DashboardPort)
	fm.Colors["green"].Printf("Dashboard 用户    : %s\n", fm.Config.DashboardUser)
	fm.Colors["green"].Printf("Dashboard 密码    : %s\n", fm.Config.DashboardPwd)
	fm.Colors["green"].Printf("Token            : %s\n", fm.Config.Token)
	fm.Colors["green"].Printf("子域名主机        : %s\n", fm.Config.SubdomainHost)
	fm.Colors["green"].Printf("TCP多路复用       : %t\n", fm.Config.TCPMux)
	fm.Colors["green"].Printf("最大连接池        : %d\n", fm.Config.MaxPoolCount)
	fm.Colors["green"].Printf("日志级别          : %s\n", fm.Config.LogLevel)
	fm.Colors["green"].Printf("日志保存天数      : %d\n", fm.Config.LogMaxDays)
	fm.Colors["green"].Printf("传输协议支持      : %t\n", fm.Config.TransportProtocol)
	if fm.Config.TransportProtocol {
		fm.Colors["green"].Printf("KCP 绑定端口     : %d\n", fm.Config.KCPBindPort)
		fm.Colors["green"].Printf("QUIC 绑定端口    : %d\n", fm.Config.QuicBindPort)
	}
	fmt.Println("=============================================")
	fmt.Println()
}

// performInstall 执行安装
func (fm *FrpsManager) performInstall(downloadSource int) error {
	// 创建程序目录
	if err := os.MkdirAll(ProgramDir, 0755); err != nil {
		return fmt.Errorf("创建程序目录失败: %v", err)
	}

	// 切换到程序目录
	if err := os.Chdir(ProgramDir); err != nil {
		return fmt.Errorf("切换目录失败: %v", err)
	}

	// 生成配置文件
	if err := fm.generateConfigFile(); err != nil {
		return fmt.Errorf("生成配置文件失败: %v", err)
	}

	// 下载并安装 frps 二进制文件
	if err := fm.downloadAndInstallBinary(downloadSource); err != nil {
		return fmt.Errorf("下载安装二进制文件失败: %v", err)
	}

	// 下载并安装初始化脚本
	if err := fm.downloadInitScript(); err != nil {
		return fmt.Errorf("下载初始化脚本失败: %v", err)
	}

	// 设置服务开机启动
	if err := fm.setupService(); err != nil {
		return fmt.Errorf("设置服务失败: %v", err)
	}

	// 启动服务
	if err := fm.startService(); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	return nil
}

// downloadAndInstallBinary 下载并安装二进制文件
func (fm *FrpsManager) downloadAndInstallBinary(downloadSource int) error {
	// 检查本地是否已有frps二进制文件（并且不是空文件）
	binaryPath := filepath.Join(ProgramDir, "frps")
	if stat, err := os.Stat(binaryPath); err == nil && stat.Size() > 0 {
		fm.Colors["yellow"].Println("检测到本地已有 frps 二进制文件，跳过下载...")
		
		// 设置权限确保可执行
		if err := os.Chmod(binaryPath, 0755); err != nil {
			fm.Colors["yellow"].Printf("设置权限失败: %v\n", err)
		}
		return nil
	}

	var baseURL string
	if downloadSource == 1 {
		baseURL = GiteeDownloadURL
	} else {
		baseURL = GithubDownloadURL
	}

	filename := fmt.Sprintf("frp_%s_linux_%s.tar.gz", fm.SystemInfo.FrpsVersion, fm.SystemInfo.FrpsArch)
	downloadURL := fmt.Sprintf("%s/v%s/%s", baseURL, fm.SystemInfo.FrpsVersion, filename)

	fm.Colors["green"].Printf("正在下载 %s...\n", filename)
	
	// 使用带进度条的下载
	if err := fm.downloadWithProgress(downloadURL, filename, "下载 frps 二进制文件"); err != nil {
		return err
	}

	fm.Colors["green"].Println("正在解压...")
	
	// 解压文件
	if err := fm.extractTarGz(filename, "."); err != nil {
		return err
	}

	// 移动二进制文件
	extractedDir := fmt.Sprintf("frp_%s_linux_%s", fm.SystemInfo.FrpsVersion, fm.SystemInfo.FrpsArch)
	srcPath := filepath.Join(extractedDir, "frps")
	dstPath := filepath.Join(ProgramDir, "frps")
	
	if err := os.Rename(srcPath, dstPath); err != nil {
		return err
	}

	// 设置权限
	if err := os.Chmod(dstPath, 0755); err != nil {
		return err
	}

	// 清理临时文件
	os.Remove(filename)
	os.RemoveAll(extractedDir)

	return nil
}

// extractTarGz 解压 tar.gz 文件
func (fm *FrpsManager) extractTarGz(filename, destDir string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			
			if _, err := io.Copy(file, tr); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	}

	return nil
} 