package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Start 启动 frps 服务
func (fm *FrpsManager) Start() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	if fm.isInstalled() {
		fm.Colors["yellow"].Println("frps 服务已经在运行中。")
		return
	}

	cmd := exec.Command(InitScript, "start")
	if err := cmd.Run(); err != nil {
		fm.Colors["red"].Printf("启动服务失败: %v\n", err)
		return
	}

	if fm.isInstalled() {
		fm.Colors["green"].Println("frps 服务启动成功。")
	} else {
		fm.Colors["red"].Println("frps 服务启动失败。")
	}
}

// Stop 停止 frps 服务
func (fm *FrpsManager) Stop() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	if !fm.isInstalled() {
		fm.Colors["yellow"].Println("frps 服务没有运行。")
		return
	}

	cmd := exec.Command(InitScript, "stop")
	if err := cmd.Run(); err != nil {
		fm.Colors["red"].Printf("停止服务失败: %v\n", err)
		return
	}

	if !fm.isInstalled() {
		fm.Colors["green"].Println("frps 服务停止成功。")
	} else {
		fm.Colors["red"].Println("frps 服务停止失败。")
	}
}

// Restart 重启 frps 服务
func (fm *FrpsManager) Restart() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()

	cmd := exec.Command(InitScript, "restart")
	if err := cmd.Run(); err != nil {
		fm.Colors["red"].Printf("重启服务失败: %v\n", err)
		return
	}

	if fm.isInstalled() {
		fm.Colors["green"].Println("frps 服务重启成功。")
	} else {
		fm.Colors["red"].Println("frps 服务重启失败。")
	}
}

// Status 查看 frps 服务状态
func (fm *FrpsManager) Status() {
	fm.showBanner()

	if fm.isInstalled() {
		fm.Colors["green"].Println("frps 服务正在运行。")
		
		// 显示进程信息
		cmd := exec.Command("ps", "aux")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, ProgramName) && !strings.Contains(line, "grep") {
					fmt.Printf("进程信息: %s\n", line)
				}
			}
		}
		
		// 显示配置文件路径
		configPath := filepath.Join(ProgramDir, ConfigFile)
		if _, err := os.Stat(configPath); err == nil {
			fm.Colors["blue"].Printf("配置文件: %s\n", configPath)
		}
		
		// 显示日志文件路径
		logPath := filepath.Join(ProgramDir, "frps.log")
		if _, err := os.Stat(logPath); err == nil {
			fm.Colors["blue"].Printf("日志文件: %s\n", logPath)
		}
	} else {
		fm.Colors["red"].Println("frps 服务没有运行。")
	}
}

// ConfigEdit 编辑配置文件
func (fm *FrpsManager) ConfigEdit() {
	if !fm.checkRoot() {
		return
	}

	configPath := filepath.Join(ProgramDir, ConfigFile)
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fm.Colors["red"].Println("配置文件不存在！")
		return
	}

	// 使用系统默认编辑器打开配置文件
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // 默认使用 vi
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fm.Colors["red"].Printf("编辑配置文件失败: %v\n", err)
		return
	}

	fm.Colors["green"].Println("配置文件编辑完成。")
	fmt.Print("是否重启 frps 服务以应用新配置？(y/n): ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))
	
	if choice == "y" || choice == "yes" {
		fm.Restart()
	}
}

// ShowVersion 显示版本信息
func (fm *FrpsManager) ShowVersion() {
	fm.showBanner()
	
	fmt.Printf("frps-onekey 版本: %s\n", Version)
	
	// 显示 frps 二进制版本
	binaryPath := filepath.Join(ProgramDir, ProgramName)
	if _, err := os.Stat(binaryPath); err == nil {
		cmd := exec.Command(binaryPath, "--version")
		output, err := cmd.Output()
		if err == nil {
			fmt.Printf("frps 版本: %s", string(output))
		}
	}
	
	fmt.Printf("系统架构: %s\n", fm.SystemInfo.FrpsArch)
	fmt.Printf("操作系统: %s\n", fm.SystemInfo.OS)
}

// Uninstall 卸载 frps
func (fm *FrpsManager) Uninstall() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	// 检查是否已安装
	configPath := filepath.Join(ProgramDir, ConfigFile)
	binaryPath := filepath.Join(ProgramDir, ProgramName)
	
	if _, err := os.Stat(InitScript); os.IsNotExist(err) &&
		_, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fm.Colors["yellow"].Println("frps 没有安装。")
		return
	}

	fmt.Println("============== 卸载 frps ==============")
	fm.Colors["yellow"].Print("您确定要卸载吗？")
	fmt.Print("[Y/N]: ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))
	
	if choice != "y" && choice != "yes" {
		fmt.Println("您选择了 [No]，脚本退出！")
		return
	}

	fmt.Println()
	fmt.Println("您选择了 [Yes]，按任意键继续。")
	reader.ReadString('\n')

	// 停止服务
	if fm.isInstalled() {
		fm.Colors["green"].Println("正在停止 frps 服务...")
		cmd := exec.Command(InitScript, "stop")
		cmd.Run()
	}

	// 移除服务
	fm.Colors["green"].Println("正在移除服务...")
	switch fm.SystemInfo.OS {
	case "CentOS", "RHEL", "Rocky", "AlmaLinux":
		cmd := exec.Command("chkconfig", "--del", ProgramName)
		cmd.Run()
	case "Ubuntu", "Debian":
		cmd := exec.Command("update-rc.d", "-f", ProgramName, "remove")
		cmd.Run()
	}

	// 删除文件
	filesToRemove := []string{
		InitScript,
		"/var/run/" + ProgramName + ".pid",
		"/usr/bin/" + ProgramName,
		ProgramDir,
	}

	for _, file := range filesToRemove {
		if err := os.RemoveAll(file); err != nil {
			fm.Colors["yellow"].Printf("删除 %s 失败: %v\n", file, err)
		} else {
			fm.Colors["green"].Printf("已删除: %s\n", file)
		}
	}

	fm.Colors["green"].Println("frps 卸载成功！")
}

// Update 更新 frps
func (fm *FrpsManager) Update() {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	// 检查是否已安装
	binaryPath := filepath.Join(ProgramDir, ProgramName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fm.Colors["red"].Println("frps 没有安装，请先安装！")
		return
	}

	fmt.Println("============== 更新 frps ==============")
	
	// 获取当前版本
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		fm.Colors["red"].Printf("获取当前版本失败: %v\n", err)
		return
	}
	currentVersion := strings.TrimSpace(string(output))
	fm.Colors["green"].Printf("当前版本: %s\n", currentVersion)

	// 选择下载源并获取最新版本
	downloadSource := fm.selectDownloadSource()
	if err := fm.getLatestVersion(downloadSource); err != nil {
		fm.Colors["red"].Printf("获取最新版本失败: %v\n", err)
		return
	}

	fm.Colors["green"].Printf("最新版本: %s\n", fm.SystemInfo.FrpsVersion)

	// 比较版本
	if strings.Contains(currentVersion, fm.SystemInfo.FrpsVersion) {
		fm.Colors["yellow"].Println("已经是最新版本，无需更新。")
		return
	}

	fm.Colors["green"].Println("发现新版本，开始更新...")

	// 停止服务
	if fm.isInstalled() {
		cmd := exec.Command(InitScript, "stop")
		cmd.Run()
	}

	// 备份当前二进制文件
	backupPath := binaryPath + ".backup"
	if err := exec.Command("cp", binaryPath, backupPath).Run(); err != nil {
		fm.Colors["yellow"].Printf("备份当前版本失败: %v\n", err)
	}

	// 下载新版本
	if err := fm.downloadAndInstallBinary(downloadSource); err != nil {
		fm.Colors["red"].Printf("下载新版本失败: %v\n", err)
		// 恢复备份
		if _, err := os.Stat(backupPath); err == nil {
			exec.Command("mv", backupPath, binaryPath).Run()
		}
		return
	}

	// 更新初始化脚本
	if err := fm.downloadInitScript(); err != nil {
		fm.Colors["yellow"].Printf("更新初始化脚本失败: %v\n", err)
	}

	// 重新设置服务
	fm.setupService()

	// 启动服务
	if err := fm.startService(); err != nil {
		fm.Colors["red"].Printf("启动服务失败: %v\n", err)
		return
	}

	// 删除备份文件
	os.Remove(backupPath)

	// 显示新版本
	cmd = exec.Command(binaryPath, "--version")
	output, err = cmd.Output()
	if err == nil {
		fm.Colors["green"].Printf("更新完成！新版本: %s\n", strings.TrimSpace(string(output)))
	} else {
		fm.Colors["green"].Println("frps 更新成功！")
	}
} 