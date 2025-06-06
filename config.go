package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// generateConfigFile 生成 frps 配置文件
func (fm *FrpsManager) generateConfigFile() error {
	configPath := filepath.Join(ProgramDir, ConfigFile)
	
	logFile := fm.Config.LogFile
	if logFile == "/dev/null" {
		logFile = "console"
	}
	
	configContent := fmt.Sprintf(`
bindAddr = "0.0.0.0"
bindPort = %d

# udp port used for kcp protocol, it can be same with 'bindPort'.
# if not set, kcp is disabled in frps.
kcpBindPort = %d

# udp port used for quic protocol.
# if not set, quic is disabled in frps.
quicBindPort = %d

# Specify which address proxy will listen for, default value is same with bindAddr
# proxyBindAddr = "127.0.0.1"

# quic protocol options
# transport.quic.keepalivePeriod = 10
# transport.quic.maxIdleTimeout = 30
# transport.quic.maxIncomingStreams = 100000

# Heartbeat configure, it's not recommended to modify the default value
# The default value of heartbeatTimeout is 90. Set negative value to disable it.
transport.heartbeatTimeout = 90

# Pool count in each proxy will keep no more than maxPoolCount.
transport.maxPoolCount = %d

# If tcp stream multiplexing is used, default is true
transport.tcpMux = %s

# Specify keep alive interval for tcp mux.
# only valid if tcpMux is true.
# transport.tcpMuxKeepaliveInterval = 30

# tcpKeepalive specifies the interval between keep-alive probes for an active network connection between frpc and frps.
# If negative, keep-alive probes are disabled.
# transport.tcpKeepalive = 7200

# transport.tls.force specifies whether to only accept TLS-encrypted connections. By default, the value is false.
# transport.tls.force = false

# transport.tls.certFile = "server.crt"
# transport.tls.keyFile = "server.key"
# transport.tls.trustedCaFile = "ca.crt"

# If you want to support virtual host, you must set the http port for listening (optional)
# Note: http port and https port can be same with bindPort
vhostHTTPPort = %d
vhostHTTPSPort = %d

# Response header timeout(seconds) for vhost http server, default is 60s
# vhostHTTPTimeout = 60

# tcpmuxHTTPConnectPort specifies the port that the server listens for TCP
# HTTP CONNECT requests. If the value is 0, the server will not multiplex TCP
# requests on one single port. If it's not - it will listen on this value for
# HTTP CONNECT requests. By default, this value is 0.
# tcpmuxHTTPConnectPort = 1337

# If tcpmuxPassthrough is true, frps won't do any update on traffic.
# tcpmuxPassthrough = false

# Configure the web server to enable the dashboard for frps.
# dashboard is available only if webServerport is set.
webServer.addr = "0.0.0.0"
webServer.port = %d
webServer.user = "%s"
webServer.password = "%s"
# webServer.tls.certFile = "server.crt"
# webServer.tls.keyFile = "server.key"
# dashboard assets directory(only for debug mode)
# webServer.assetsDir = "./static"

# Enable golang pprof handlers in dashboard listener.
# Dashboard port must be set first
# webServer.pprofEnable = false

# enablePrometheus will export prometheus metrics on webServer in /metrics api.
# enablePrometheus = true

# console or real logFile path like ./frps.log
log.to = "%s"
# trace, debug, info, warn, error
log.level = "%s"
log.maxDays = %d
# disable log colors when log.to is console, default is false
# log.disablePrintColor = false

# DetailedErrorsToClient defines whether to send the specific error (with debug info) to frpc. By default, this value is true.
# detailedErrorsToClient = true

# auth.method specifies what authentication method to use authenticate frpc with frps.
# If "token" is specified - token will be read into login message.
# If "oidc" is specified - OIDC (Open ID Connect) token will be issued using OIDC settings. By default, this value is "token".
auth.method = "token"

# auth.additionalScopes specifies additional scopes to include authentication information.
# Optional values are HeartBeats, NewWorkConns.
# auth.additionalScopes = ["HeartBeats", "NewWorkConns"]

# auth token
auth.token = "%s"

# userConnTimeout specifies the maximum time to wait for a work connection.
# userConnTimeout = 10

# Max ports can be used for each client, default value is 0 means no limit
# maxPortsPerClient = 0

# If subDomainHost is not empty, you can set subdomain when type is http or https in frpc's configure file
# When subdomain is test, the host used by routing is test.frps.com
subDomainHost = "%s"

# custom 404 page for HTTP requests
# custom404Page = "/path/to/404.html"

# specify udp packet size, unit is byte. If not set, the default value is 1500.
# This parameter should be same between client and server.
# It affects the udp and sudp proxy.
# udpPacketSize = 1500

# Retention time for NAT hole punching strategy data.
# natholeAnalysisDataReserveHours = 168

# ssh tunnel gateway
# If you want to enable this feature, the bindPort parameter is required, while others are optional.
# By default, this feature is disabled. It will be enabled if bindPort is greater than 0.
# sshTunnelGateway.bindPort = 2200
# sshTunnelGateway.privateKeyFile = "/home/frp-user/.ssh/id_rsa"
# sshTunnelGateway.autoGenPrivateKeyPath = ""
# sshTunnelGateway.authorizedKeysFile = "/home/frp-user/.ssh/authorized_keys"
`,
		fm.Config.BindPort,
		fm.Config.KCPBindPort,
		fm.Config.QuicBindPort,
		fm.Config.MaxPoolCount,
		strconv.FormatBool(fm.Config.TCPMux),
		fm.Config.VhostHTTPPort,
		fm.Config.VhostHTTPSPort,
		fm.Config.DashboardPort,
		fm.Config.DashboardUser,
		fm.Config.DashboardPwd,
		logFile,
		fm.Config.LogLevel,
		fm.Config.LogMaxDays,
		fm.Config.Token,
		fm.Config.SubdomainHost,
	)

	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// downloadInitScript 下载初始化脚本
func (fm *FrpsManager) downloadInitScript() error {
	fm.Colors["green"].Println("正在下载初始化脚本...")
	
	cmd := exec.Command("wget", "-q", InitScriptURL, "-O", InitScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载初始化脚本失败: %v", err)
	}

	// 设置执行权限
	if err := os.Chmod(InitScript, 0755); err != nil {
		return fmt.Errorf("设置脚本权限失败: %v", err)
	}

	return nil
}

// setupService 设置服务开机启动
func (fm *FrpsManager) setupService() error {
	fm.Colors["green"].Println("正在设置服务开机启动...")
	
	var cmd *exec.Cmd
	switch fm.SystemInfo.OS {
	case "CentOS", "RHEL", "Rocky", "AlmaLinux":
		cmd = exec.Command("chkconfig", "--add", ProgramName)
	case "Ubuntu", "Debian":
		cmd = exec.Command("update-rc.d", "-f", ProgramName, "defaults")
	default:
		return fmt.Errorf("不支持的操作系统: %s", fm.SystemInfo.OS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("设置开机启动失败: %v", err)
	}

	// 创建软链接
	linkPath := "/usr/bin/" + ProgramName
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除旧的软链接失败: %v", err)
	}
	
	if err := os.Symlink(InitScript, linkPath); err != nil {
		return fmt.Errorf("创建软链接失败: %v", err)
	}

	return nil
}

// startService 启动服务
func (fm *FrpsManager) startService() error {
	fm.Colors["green"].Println("正在启动 frps 服务...")
	
	cmd := exec.Command(InitScript, "start")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	// 检查服务是否启动成功
	if fm.isInstalled() {
		fm.Colors["green"].Println("frps 服务启动成功。")
		return nil
	} else {
		return fmt.Errorf("frps 服务启动失败")
	}
}

// showInstallationSummary 显示安装总结
func (fm *FrpsManager) showInstallationSummary(serverIP string) {
	fmt.Println()
	fm.Colors["green"].Println("┌─────────────────────────────────────────┐")
	fm.Colors["green"].Println("│   frp service started successfully.     │")
	fm.Colors["green"].Println("└─────────────────────────────────────────┘")
	fm.Colors["green"].Println("┌─────────────────────────────────────────┐")
	fm.Colors["green"].Println("│  Installation completed successfully.   │")
	fm.Colors["green"].Println("└─────────────────────────────────────────┘")
	fmt.Println()
	
	fmt.Println("恭喜，frps 安装完成！")
	fmt.Println("================================================")
	fm.Colors["green"].Printf("服务器IP          : %s\n", serverIP)
	fm.Colors["green"].Printf("绑定端口          : %d\n", fm.Config.BindPort)
	fm.Colors["green"].Printf("vhost http 端口   : %d\n", fm.Config.VhostHTTPPort)
	fm.Colors["green"].Printf("vhost https 端口  : %d\n", fm.Config.VhostHTTPSPort)
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
	fmt.Println("================================================")
	
	dashboardURL := fmt.Sprintf("http://%s:%d/", fm.Config.SubdomainHost, fm.Config.DashboardPort)
	fm.Colors["green"].Printf("frps Dashboard    : %s\n", dashboardURL)
	fm.Colors["green"].Printf("Dashboard 端口    : %d\n", fm.Config.DashboardPort)
	fm.Colors["green"].Printf("Dashboard 用户    : %s\n", fm.Config.DashboardUser)
	fm.Colors["green"].Printf("Dashboard 密码    : %s\n", fm.Config.DashboardPwd)
	fmt.Println("================================================")
	fmt.Println()
	
	fmt.Print("frps 状态管理: ")
	fm.Colors["pink"].Print("frps")
	fmt.Print(" {")
	fm.Colors["green"].Print("start|stop|restart|status|config|version")
	fmt.Println("}")
	fmt.Println("示例:")
	fmt.Print("  启动: ")
	fm.Colors["pink"].Print("frps")
	fmt.Print(" ")
	fm.Colors["green"].Println("start")
	fmt.Print("  停止: ")
	fm.Colors["pink"].Print("frps")
	fmt.Print(" ")
	fm.Colors["green"].Println("stop")
	fmt.Print("  重启: ")
	fm.Colors["pink"].Print("frps")
	fmt.Print(" ")
	fm.Colors["green"].Println("restart")
} 