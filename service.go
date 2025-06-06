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

	fmt.Println("============== 编辑配置文件 ==============")
	fm.Colors["blue"].Printf("配置文件路径: %s\n", configPath)
	fmt.Println()

	// 选择编辑器的优先级：nano > vim > EDITOR环境变量 > vi
	var editor string
	var editorName string
	
	// 首先检查是否有 nano
	if _, err := exec.LookPath("nano"); err == nil {
		editor = "nano"
		editorName = "nano"
		fm.Colors["green"].Println("✓ 检测到 nano 编辑器")
	} else if _, err := exec.LookPath("vim"); err == nil {
		editor = "vim"
		editorName = "vim"
		fm.Colors["green"].Println("✓ 检测到 vim 编辑器")
	} else {
		// 没有nano和vim，检查EDITOR环境变量
		editor = os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // 最后默认使用 vi
			editorName = "vi"
		} else {
			editorName = editor
		}
		fm.Colors["yellow"].Printf("! nano/vim 不可用，使用 %s 编辑器\n", editorName)
	}

	fmt.Println()
	fm.Colors["blue"].Printf("使用 %s 打开配置文件...\n", editorName)
	if editorName == "nano" {
		fm.Colors["green"].Println("提示: 按 Ctrl+X 保存并退出")
	} else if editorName == "vim" {
		fm.Colors["green"].Println("提示: 按 :wq 保存并退出")
	} else {
		fm.Colors["green"].Println("提示: 请按照编辑器的帮助说明保存并退出")
	}
	fmt.Println()

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
	binaryPath := filepath.Join(ProgramDir, ProgramName)
	
	_, err1 := os.Stat(InitScript)
	_, err2 := os.Stat(binaryPath)
	if os.IsNotExist(err1) && os.IsNotExist(err2) {
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

// ImportConfig 导入用户指定的配置文件
func (fm *FrpsManager) ImportConfig(configPath string) {
	if !fm.checkRoot() {
		return
	}

	fm.showBanner()
	
	fmt.Println("============== 导入配置文件 ==============")
	fm.Colors["blue"].Printf("指定的配置文件: %s\n", configPath)
	
	// 检查用户指定的配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fm.Colors["red"].Printf("错误：配置文件 %s 不存在！\n", configPath)
		return
	}
	
	// 检查文件是否可读
	file, err := os.Open(configPath)
	if err != nil {
		fm.Colors["red"].Printf("错误：无法读取配置文件 %s: %v\n", configPath, err)
		return
	}
	file.Close()
	
	// 验证配置文件格式（基本检查）
	if err := fm.validateConfigFile(configPath); err != nil {
		fm.Colors["red"].Printf("错误：配置文件格式验证失败: %v\n", err)
		return
	}
	
	fm.Colors["green"].Println("✓ 配置文件验证通过")
	
	// 目标配置文件路径
	targetConfigPath := filepath.Join(ProgramDir, ConfigFile)
	
	// 检查目标目录是否存在，不存在则创建
	if err := os.MkdirAll(ProgramDir, 0755); err != nil {
		fm.Colors["red"].Printf("错误：创建目录失败: %v\n", err)
		return
	}
	
	// 备份现有配置文件（如果存在）
	if _, err := os.Stat(targetConfigPath); err == nil {
		backupPath := targetConfigPath + ".backup." + fmt.Sprintf("%d", os.Getpid())
		if err := fm.copyFile(targetConfigPath, backupPath); err != nil {
			fm.Colors["yellow"].Printf("警告：备份现有配置文件失败: %v\n", err)
		} else {
			fm.Colors["green"].Printf("✓ 已备份现有配置文件到: %s\n", backupPath)
		}
	}
	
	// 复制用户配置文件到目标位置
	if err := fm.copyFile(configPath, targetConfigPath); err != nil {
		fm.Colors["red"].Printf("错误：复制配置文件失败: %v\n", err)
		return
	}
	
	// 设置文件权限
	if err := os.Chmod(targetConfigPath, 0644); err != nil {
		fm.Colors["yellow"].Printf("警告：设置文件权限失败: %v\n", err)
	}
	
	fm.Colors["green"].Printf("✓ 配置文件已成功导入到: %s\n", targetConfigPath)
	
	// 询问是否重启服务
	fmt.Print("是否重启 frps 服务以应用新配置？(y/n): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))
	
	if choice == "y" || choice == "yes" {
		if fm.isInstalled() {
			fm.Restart()
		} else {
			fm.Colors["yellow"].Println("frps 服务未运行，请使用 'frps-onekey start' 启动服务")
		}
	}
	
	fmt.Println("配置文件导入完成！")
}

// validateConfigFile 验证配置文件格式
func (fm *FrpsManager) validateConfigFile(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	configStr := string(content)
	
	// 基本的TOML格式检查
	if !strings.Contains(configStr, "bindPort") && !strings.Contains(configStr, "bind_port") {
		return fmt.Errorf("配置文件缺少必要的 bindPort 或 bind_port 配置")
	}
	
	// 检查文件是否为空
	if len(strings.TrimSpace(configStr)) == 0 {
		return fmt.Errorf("配置文件为空")
	}
	
	// 可以添加更多的验证逻辑
	// 比如检查TOML语法，检查必要的配置项等
	
	return nil
}

// copyFile 复制文件
func (fm *FrpsManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = destFile.ReadFrom(sourceFile)
	return err
} 