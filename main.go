package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	ProgramName    = "frps"
	Version        = "1.0.7"
	ProgramDir     = "/usr/local/frps"
	ConfigFile     = "frps.toml"
	InitScript     = "/etc/init.d/frps"
	InitScriptURL  = "https://raw.githubusercontent.com/mvscode/frps-onekey/master/frps.init"
	
	GiteeDownloadURL    = "https://gitee.com/mvscode/frps-onekey/releases/download"
	GithubDownloadURL   = "https://github.com/fatedier/frp/releases/download" 
	GiteeLatestAPI      = "https://gitee.com/api/v5/repos/mvscode/frps-onekey/releases/latest"
	GithubLatestAPI     = "https://api.github.com/repos/fatedier/frp/releases/latest"
	UpdateCheckURL      = "https://raw.githubusercontent.com/mvscode/frps-onekey/master/install-frps.sh"
)

// Config 存储配置信息
type Config struct {
	BindPort         int    `json:"bind_port"`
	DashboardPort    int    `json:"dashboard_port"`
	VhostHTTPPort    int    `json:"vhost_http_port"`
	VhostHTTPSPort   int    `json:"vhost_https_port"`
	DashboardUser    string `json:"dashboard_user"`
	DashboardPwd     string `json:"dashboard_pwd"`
	Token            string `json:"token"`
	SubdomainHost    string `json:"subdomain_host"`
	MaxPoolCount     int    `json:"max_pool_count"`
	LogLevel         string `json:"log_level"`
	LogMaxDays       int    `json:"log_max_days"`
	LogFile          string `json:"log_file"`
	TCPMux           bool   `json:"tcp_mux"`
	KCPBindPort      int    `json:"kcp_bind_port"`
	QuicBindPort     int    `json:"quic_bind_port"`
	TransportProtocol bool  `json:"transport_protocol"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS          string
	Version     string
	Arch        string
	Is64Bit     bool
	FrpsArch    string
	FrpsVersion string
}

// Release GitHub/Gitee API 响应结构
type Release struct {
	TagName string `json:"tag_name"`
}

// FrpsManager 主管理器
type FrpsManager struct {
	Config     *Config
	SystemInfo *SystemInfo
	Colors     map[string]*color.Color
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	manager := NewFrpsManager()
	
	switch os.Args[1] {
	case "install":
		manager.Install()
	case "uninstall":
		manager.Uninstall()
	case "update":
		manager.Update()
	case "config":
		manager.ConfigEdit()
	case "start":
		manager.Start()
	case "stop":
		manager.Stop()
	case "restart":
		manager.Restart()
	case "status":
		manager.Status()
	case "version":
		manager.ShowVersion()
	default:
		showUsage()
	}
}

// NewFrpsManager 创建新的管理器实例
func NewFrpsManager() *FrpsManager {
	manager := &FrpsManager{
		Config:     &Config{},
		SystemInfo: &SystemInfo{},
		Colors: map[string]*color.Color{
			"red":    color.New(color.FgRed, color.Bold),
			"green":  color.New(color.FgGreen, color.Bold),
			"yellow": color.New(color.FgYellow, color.Bold),
			"blue":   color.New(color.FgBlue, color.Bold),
			"pink":   color.New(color.FgMagenta, color.Bold),
		},
	}
	
	manager.detectSystemInfo()
	return manager
}

// showBanner 显示程序横幅
func (fm *FrpsManager) showBanner() {
	fmt.Println()
	fmt.Println("+------------------------------------------------------------+")
	fmt.Println("|    frps for Linux Server, Author Clang, Mender MvsCode    |")
	fmt.Println("|      A tool to auto-compile & install frps on Linux       |")
	fmt.Println("+------------------------------------------------------------+")
	fmt.Println()
}

// detectSystemInfo 检测系统信息
func (fm *FrpsManager) detectSystemInfo() {
	fm.SystemInfo.Arch = runtime.GOARCH
	
	// 检测操作系统
	if content, err := os.ReadFile("/etc/os-release"); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "CentOS") {
			fm.SystemInfo.OS = "CentOS"
		} else if strings.Contains(contentStr, "Ubuntu") {
			fm.SystemInfo.OS = "Ubuntu"
		} else if strings.Contains(contentStr, "Debian") {
			fm.SystemInfo.OS = "Debian"
		} else if strings.Contains(contentStr, "Red Hat") {
			fm.SystemInfo.OS = "RHEL"
		} else if strings.Contains(contentStr, "Rocky") {
			fm.SystemInfo.OS = "Rocky"
		} else if strings.Contains(contentStr, "AlmaLinux") {
			fm.SystemInfo.OS = "AlmaLinux"
		}
	}
	
	// 设置架构信息
	switch runtime.GOARCH {
	case "amd64":
		fm.SystemInfo.Is64Bit = true
		fm.SystemInfo.FrpsArch = "amd64"
	case "386":
		fm.SystemInfo.Is64Bit = false
		fm.SystemInfo.FrpsArch = "386"
	case "arm64":
		fm.SystemInfo.Is64Bit = true
		fm.SystemInfo.FrpsArch = "arm64"
	case "arm":
		fm.SystemInfo.Is64Bit = false
		fm.SystemInfo.FrpsArch = "arm"
	case "mips":
		fm.SystemInfo.Is64Bit = false
		fm.SystemInfo.FrpsArch = "mips"
	case "mips64":
		fm.SystemInfo.Is64Bit = true
		fm.SystemInfo.FrpsArch = "mips64"
	case "mips64le":
		fm.SystemInfo.Is64Bit = true
		fm.SystemInfo.FrpsArch = "mips64le"
	case "mipsle":
		fm.SystemInfo.Is64Bit = false
		fm.SystemInfo.FrpsArch = "mipsle"
	case "riscv64":
		fm.SystemInfo.Is64Bit = true
		fm.SystemInfo.FrpsArch = "riscv64"
	default:
		fm.SystemInfo.FrpsArch = "amd64"
	}
}

// checkRoot 检查是否为root用户
func (fm *FrpsManager) checkRoot() bool {
	if os.Geteuid() != 0 {
		fm.Colors["red"].Println("错误：此脚本必须以root用户运行！")
		return false
	}
	return true
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Println("frps 管理工具")
	fmt.Println("使用方法: frps-onekey {install|uninstall|update|config|start|stop|restart|status|version}")
	fmt.Println()
	fmt.Println("命令说明:")
	fmt.Println("  install   - 安装 frps")
	fmt.Println("  uninstall - 卸载 frps")
	fmt.Println("  update    - 更新 frps")
	fmt.Println("  config    - 编辑配置文件")
	fmt.Println("  start     - 启动 frps 服务")
	fmt.Println("  stop      - 停止 frps 服务")
	fmt.Println("  restart   - 重启 frps 服务")
	fmt.Println("  status    - 查看 frps 状态")
	fmt.Println("  version   - 显示版本信息")
} 