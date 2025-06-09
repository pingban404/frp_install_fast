# frp_install_fast

一个用 Go 语言重构的 frps 一键安装管理工具，原始脚本来自 [mvscode/frps-onekey](https://github.com/mvscode/frps-onekey)。

## 功能特性

- 🚀 一键安装 frps 服务
- 🔧 自动配置系统服务
- 📊 支持多种 Linux 发行版
- 🌐 支持多架构（amd64, arm64, mips 等）
- 🎛️ 交互式配置向导
- 📈 支持 GitHub 和 Gitee 下载源
- 🔄 支持在线更新
- 📝 完整的服务管理功能

## 支持的系统

- CentOS 7+
- Ubuntu 16.04+
- Debian 9+
- RHEL 7+
- Rocky Linux 8+
- AlmaLinux 8+

## 支持的架构

- x86_64 (amd64)
- i386
- arm64
- arm
- mips
- mips64
- mips64le
- mipsle
- riscv64

## 安装方法

### 方法 1: 直接下载运行

```bash
# 下载 amd64 版本
wget https://github.com/username/frps-onekey/releases/latest/download/frps-onekey-linux-amd64.tar.gz
tar -xzf frps-onekey-linux-amd64.tar.gz
chmod +x frps-onekey-linux-amd64
sudo mv frps-onekey-linux-amd64 /usr/local/bin/frps-onekey

# 安装 frps
sudo frps-onekey install
```

### 方法 2: 从源码编译

```bash
# 克隆仓库
git clone https://github.com/username/frps-onekey.git
cd frps-onekey

# 安装依赖
go mod download

# 编译
go build -o frps-onekey .

# 安装到系统
sudo mv frps-onekey /usr/local/bin/

# 运行安装
sudo frps-onekey install
```

### 方法 3: 使用构建脚本

```bash
# 构建所有平台版本
chmod +x build.sh
./build.sh

# 构建结果在 build/ 目录下
```

## 使用方法

```bash
frps-onekey {install|uninstall|update|config|start|stop|restart|status|version}
```

### 命令说明

- `install` - 安装 frps 服务
- `uninstall` - 卸载 frps 服务
- `update` - 更新 frps 到最新版本
- `config` - 编辑配置文件
- `start` - 启动 frps 服务
- `stop` - 停止 frps 服务
- `restart` - 重启 frps 服务
- `status` - 查看 frps 运行状态
- `version` - 显示版本信息

## 安装示例

```bash
# 以 root 用户运行安装
sudo frps-onekey install
```

安装过程中会要求您配置以下参数：

- **绑定端口** (默认: 5443)
- **HTTP 端口** (默认: 80) 
- **HTTPS 端口** (默认: 443)
- **Dashboard 端口** (默认: 6443)
- **Dashboard 用户名** (默认: admin)
- **Dashboard 密码** (随机生成)
- **Token** (随机生成)
- **子域名主机** (自动获取服务器IP)
- **日志级别** (默认: info)
- **其他高级选项**

## 配置文件

安装完成后，配置文件位于：`/usr/local/frps/frps.toml`

可以使用以下命令编辑：

```bash
sudo frps-onekey config
```

## 服务管理

```bash
# 启动服务
sudo frps-onekey start

# 停止服务
sudo frps-onekey stop

# 重启服务
sudo frps-onekey restart

# 查看状态
sudo frps-onekey status
```

## 更新

```bash
# 更新到最新版本
sudo frps-onekey update
```

## 卸载

```bash
# 完全卸载 frps
sudo frps-onekey uninstall
```

## 目录结构

```
/usr/local/frps/           # frps 安装目录
├── frps                   # frps 可执行文件
├── frps.toml             # 配置文件
└── frps.log              # 日志文件

/etc/init.d/frps          # 系统服务脚本
/usr/bin/frps             # 服务管理命令软链接
```

## 开发

### 项目结构

```
frps-onekey/
├── main.go              # 主程序入口
├── install.go           # 安装相关功能
├── config.go           # 配置文件管理
├── service.go          # 服务管理功能
├── go.mod              # Go 模块定义
├── build.sh            # 构建脚本
└── README.md           # 说明文档
```

### 编译依赖

- Go 1.19+
- github.com/fatih/color (用于彩色输出)

### 本地开发

```bash
# 克隆项目
git clone https://github.com/username/frps-onekey.git
cd frps-onekey

# 安装依赖
go mod download

# 运行
go run . install
```

## 与原版差异

相比于原始的 bash 脚本版本，Go 版本具有以下优势：

1. **更好的错误处理** - 更详细的错误信息和异常处理
2. **跨平台兼容** - 单个二进制文件支持多架构
3. **更快的启动速度** - 编译后的二进制文件执行更快
4. **更好的代码组织** - 模块化的代码结构，易于维护
5. **内置依赖** - 不依赖外部工具，除了系统基本命令
6. **类型安全** - Go 的类型系统提供更好的代码安全性

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

感谢原始项目 [mvscode/frps-onekey](https://github.com/mvscode/frps-onekey) 提供的思路和实现。







 

 
  
