# frps-onekey Makefile

BINARY_NAME=frps-onekey
VERSION=1.0.7
BUILD_DIR=build
MAIN_FILE=main.go

# Go 编译器设置
GO=go
GOFLAGS=-ldflags="-s -w"

# 支持的平台
PLATFORMS=linux/amd64 linux/386 linux/arm64 linux/arm linux/mips linux/mips64 linux/mips64le linux/mipsle linux/riscv64

.PHONY: all build clean install deps test lint format help

# 默认目标
all: build

# 构建当前平台版本
build:
	@echo "构建 $(BINARY_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "构建完成: $(BINARY_NAME)"

# 构建所有平台版本
build-all: clean
	@echo "构建所有平台版本..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output_name="$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		echo "构建 $$GOOS/$$GOARCH..."; \
		env GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$$output_name .; \
		if [ $$? -eq 0 ]; then \
			cd $(BUILD_DIR) && tar -czf $$output_name.tar.gz $$output_name && rm $$output_name && cd ..; \
			echo "✓ $$GOOS/$$GOARCH 构建完成"; \
		else \
			echo "✗ $$GOOS/$$GOARCH 构建失败"; \
			exit 1; \
		fi; \
	done
	@echo "所有平台构建完成！构建文件位于: $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# 安装依赖
deps:
	@echo "安装依赖..."
	$(GO) mod download
	$(GO) mod tidy

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# 安装到系统
install: build
	@echo "安装 $(BINARY_NAME) 到系统..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "安装完成！现在可以使用 'frps-onekey' 命令。"

# 卸载
uninstall:
	@echo "从系统卸载 $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成！"

# 运行测试
test:
	@echo "运行测试..."
	$(GO) test -v ./...

# 代码格式化
format:
	@echo "格式化代码..."
	$(GO) fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "请安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 发布版本
release: build-all
	@echo "创建发布版本 $(VERSION)..."
	@echo "发布文件位于: $(BUILD_DIR)/"

# 开发模式运行
run:
	@echo "开发模式运行..."
	$(GO) run . $(ARGS)

# 查看版本
version:
	@echo "$(BINARY_NAME) 版本: $(VERSION)"

# 帮助信息
help:
	@echo "frps-onekey Makefile 使用说明"
	@echo ""
	@echo "可用命令:"
	@echo "  build      - 构建当前平台版本"
	@echo "  build-all  - 构建所有平台版本"
	@echo "  deps       - 安装依赖"
	@echo "  clean      - 清理构建文件"
	@echo "  install    - 构建并安装到系统"
	@echo "  uninstall  - 从系统卸载"
	@echo "  test       - 运行测试"
	@echo "  format     - 格式化代码"
	@echo "  lint       - 代码检查"
	@echo "  release    - 创建发布版本"
	@echo "  run        - 开发模式运行 (使用 ARGS=参数)"
	@echo "  version    - 显示版本信息"
	@echo "  help       - 显示此帮助信息"
	@echo ""
	@echo "示例:"
	@echo "  make build                    # 构建当前平台"
	@echo "  make build-all               # 构建所有平台"
	@echo "  make run ARGS=install        # 开发模式运行安装命令"
	@echo "  make install                 # 构建并安装到系统" 