#!/bin/bash

# frps-onekey 构建脚本

set -e

PROGRAM_NAME="frps-onekey"
VERSION="1.0.8"
BUILD_DIR="build"

# 清理构建目录
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}开始构建 ${PROGRAM_NAME}...${NC}"

# 设置Go代理为中国代理
echo -e "${YELLOW}设置Go代理...${NC}"
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 构建不同架构的版本
PLATFORMS=(
    "linux/amd64"
    "linux/386" 
    "linux/arm64"
    "linux/arm"
    "linux/mips"
    "linux/mips64"
    "linux/mips64le"
    "linux/mipsle"
    "linux/riscv64"
)

for platform in "${PLATFORMS[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="${PROGRAM_NAME}-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo -e "${YELLOW}构建 ${GOOS}/${GOARCH}...${NC}"
    
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o ${BUILD_DIR}/${output_name} .
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}构建 ${GOOS}/${GOARCH} 失败${NC}"
        exit 1
    fi
    
    # 压缩二进制文件
    cd ${BUILD_DIR}
    tar -czf ${output_name}.tar.gz ${output_name}
    rm ${output_name}
    cd ..
    
    echo -e "${GREEN}${GOOS}/${GOARCH} 构建完成: ${BUILD_DIR}/${output_name}.tar.gz${NC}"
done

echo -e "${GREEN}所有平台构建完成！${NC}"
echo -e "${GREEN}构建文件位于: ${BUILD_DIR}/${NC}"

# 显示构建结果
echo -e "\n${YELLOW}构建结果:${NC}"
ls -la ${BUILD_DIR}/ 