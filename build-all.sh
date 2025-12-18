#!/bin/bash

# dashGO 完整编译脚本
# 用于在 Linux 上编译 Dashboard 和全架构的 Agent
#
# 前置要求:
#   - Go >= 1.22 (用于编译 Server 和 Agent)
#   - Node.js >= 18.0.0 (用于编译前端)
#   - npm (随 Node.js 安装)
#   - Docker (可选，用于交叉编译支持 SQLite 的 Server)
#
# 使用方法:
#   ./build-all.sh all              # 构建所有组件
#   ./build-all.sh frontend         # 仅构建前端
#   RUN_TESTS=true ./build-all.sh   # 构建前运行测试
#
# 详细文档:
#   - 前端测试设置: web/TESTING_SETUP.md
#   - 构建文档: BUILD.md

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 版本信息
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 输出目录
OUTPUT_DIR="dist"
AGENT_OUTPUT_DIR="${OUTPUT_DIR}/agent"
SERVER_OUTPUT_DIR="${OUTPUT_DIR}/server"
WEB_OUTPUT_DIR="${OUTPUT_DIR}/web"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}dashGO 完整编译脚本${NC}"
echo -e "${GREEN}版本: ${VERSION}${NC}"
echo -e "${GREEN}构建时间: ${BUILD_TIME}${NC}"
echo -e "${GREEN}Git Commit: ${GIT_COMMIT}${NC}"
echo -e "${GREEN}========================================${NC}"

# 清理旧的构建文件
clean() {
    echo -e "${YELLOW}清理旧的构建文件...${NC}"
    rm -rf ${OUTPUT_DIR}
    mkdir -p ${AGENT_OUTPUT_DIR}
    mkdir -p ${SERVER_OUTPUT_DIR}
    mkdir -p ${WEB_OUTPUT_DIR}
    echo -e "${GREEN}✓ 清理完成${NC}"
}

# 构建前端
build_frontend() {
    echo -e "${YELLOW}开始构建前端...${NC}"
    
    # 检查 Node.js 是否安装
    if ! command -v node &>/dev/null; then
        echo -e "${RED}错误: 未安装 Node.js${NC}"
        echo -e "${YELLOW}请访问 https://nodejs.org/ 安装 Node.js (>= 18.0.0)${NC}"
        echo -e "${YELLOW}或参考 web/TESTING_SETUP.md 获取详细安装指南${NC}"
        return 1
    fi
    
    # 检查 Node.js 版本
    NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$NODE_VERSION" -lt 18 ]; then
        echo -e "${YELLOW}警告: Node.js 版本过低 (当前: $(node -v), 推荐: >= 18.0.0)${NC}"
        echo -e "${YELLOW}可能会遇到兼容性问题${NC}"
    fi
    
    # 检查并安装依赖
    if [ ! -d "web/node_modules" ]; then
        echo -e "${YELLOW}安装前端依赖...${NC}"
        cd web
        npm install || {
            echo -e "${RED}依赖安装失败${NC}"
            cd ..
            return 1
        }
        cd ..
        echo -e "${GREEN}✓ 依赖安装完成${NC}"
    else
        echo -e "${GREEN}✓ 依赖已存在，跳过安装${NC}"
    fi
    
    # 运行测试（可选）
    if [ "${RUN_TESTS}" = "true" ]; then
        echo -e "${YELLOW}运行前端测试...${NC}"
        cd web
        npm run test:run || {
            echo -e "${YELLOW}警告: 测试失败，但继续构建${NC}"
        }
        cd ..
    fi
    
    # 构建前端
    echo -e "${YELLOW}构建前端应用...${NC}"
    cd web
    npm run build || {
        echo -e "${RED}前端构建失败${NC}"
        cd ..
        return 1
    }
    cd ..
    
    # 复制构建产物
    if [ -d "web/dist" ]; then
        cp -r web/dist ${WEB_OUTPUT_DIR}/
        echo -e "${GREEN}✓ 前端构建完成${NC}"
        
        # 显示构建产物大小
        DIST_SIZE=$(du -sh web/dist | cut -f1)
        echo -e "${GREEN}  构建产物大小: ${DIST_SIZE}${NC}"
    else
        echo -e "${RED}错误: 构建产物目录不存在${NC}"
        return 1
    fi
}

# 使用 Docker 构建 Server (支持交叉编译 + SQLite)
build_server_docker() {
    echo -e "${YELLOW}使用 Docker 构建 Server (支持 SQLite)...${NC}"
    
    if ! command -v docker &>/dev/null; then
        echo -e "${RED}错误: 未安装 Docker${NC}"
        echo -e "${YELLOW}请使用 'build_server' 函数进行本地编译${NC}"
        return 1
    fi
    
    # 构建 Linux amd64
    echo -e "${YELLOW}构建 Server (Linux amd64 with SQLite)...${NC}"
    docker run --rm --platform linux/amd64 \
        -v "$PWD":/app -w /app \
        golang:1.22-alpine sh -c "
        apk add --no-cache gcc musl-dev && \
        CGO_ENABLED=1 go build -trimpath -ldflags='-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}' \
        -o dashgo-server-linux-amd64 ./cmd/server
    "
    mv dashgo-server-linux-amd64 ${SERVER_OUTPUT_DIR}/
    echo -e "${GREEN}✓ amd64 完成${NC}"
    
    # 构建 Linux arm64
    echo -e "${YELLOW}构建 Server (Linux arm64 with SQLite)...${NC}"
    docker run --rm --platform linux/arm64 \
        -v "$PWD":/app -w /app \
        golang:1.22-alpine sh -c "
        apk add --no-cache gcc musl-dev && \
        CGO_ENABLED=1 go build -trimpath -ldflags='-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}' \
        -o dashgo-server-linux-arm64 ./cmd/server
    "
    mv dashgo-server-linux-arm64 ${SERVER_OUTPUT_DIR}/
    echo -e "${GREEN}✓ arm64 完成${NC}"
    
    echo -e "${GREEN}✓ Server 构建完成 (Docker 模式，支持 SQLite)${NC}"
}

# 构建 Server (Dashboard) - 本地编译
build_server() {
    echo -e "${YELLOW}开始构建 Server (本地编译)...${NC}"
    
    # 检测当前架构
    CURRENT_ARCH=$(uname -m)
    
    # 构建 Linux amd64
    echo -e "${YELLOW}构建 Server (Linux amd64)...${NC}"
    if [ "$CURRENT_ARCH" = "x86_64" ]; then
        # 本地架构，启用 CGO 支持 SQLite
        echo -e "${GREEN}  → 启用 CGO (支持 SQLite)${NC}"
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
            -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-amd64 \
            ./cmd/server
    else
        # 交叉编译，禁用 CGO（仅支持 MySQL）
        echo -e "${YELLOW}  → 交叉编译，禁用 CGO (仅支持 MySQL)${NC}"
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-amd64 \
            ./cmd/server
    fi
    
    # 构建 Linux arm64
    echo -e "${YELLOW}构建 Server (Linux arm64)...${NC}"
    if [ "$CURRENT_ARCH" = "aarch64" ] || [ "$CURRENT_ARCH" = "arm64" ]; then
        # 本地架构，启用 CGO 支持 SQLite
        echo -e "${GREEN}  → 启用 CGO (支持 SQLite)${NC}"
        CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build \
            -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-arm64 \
            ./cmd/server
    else
        # 交叉编译，禁用 CGO（仅支持 MySQL）
        echo -e "${YELLOW}  → 交叉编译，禁用 CGO (仅支持 MySQL)${NC}"
        CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
            -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
            -o ${SERVER_OUTPUT_DIR}/dashgo-server-linux-arm64 \
            ./cmd/server
    fi
    
    # 构建 Windows amd64
    echo -e "${YELLOW}构建 Server (Windows amd64)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-windows-amd64.exe \
        ./cmd/server
    
    # 构建 macOS amd64
    echo -e "${YELLOW}构建 Server (macOS amd64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-darwin-amd64 \
        ./cmd/server
    
    # 构建 macOS arm64 (Apple Silicon)
    echo -e "${YELLOW}构建 Server (macOS arm64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ${SERVER_OUTPUT_DIR}/dashgo-server-darwin-arm64 \
        ./cmd/server
    
    echo -e "${GREEN}✓ Server 构建完成${NC}"
}

# 构建 Agent (全架构)
build_agent() {
    echo -e "${YELLOW}开始构建 Agent (全架构)...${NC}"
    
    cd agent
    
    # Linux amd64
    echo -e "${YELLOW}构建 Agent (Linux amd64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-amd64 .
    
    # Linux arm64
    echo -e "${YELLOW}构建 Agent (Linux arm64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-arm64 .
    
    # Linux 386
    echo -e "${YELLOW}构建 Agent (Linux 386)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=386 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-linux-386 .
    
    # Windows amd64
    echo -e "${YELLOW}构建 Agent (Windows amd64)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-windows-amd64.exe .
    
    # Windows 386
    echo -e "${YELLOW}构建 Agent (Windows 386)...${NC}"
    CGO_ENABLED=0 GOOS=windows GOARCH=386 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-windows-386.exe .
    
    # macOS amd64
    echo -e "${YELLOW}构建 Agent (macOS amd64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-darwin-amd64 .
    
    # macOS arm64 (Apple Silicon)
    echo -e "${YELLOW}构建 Agent (macOS arm64)...${NC}"
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-darwin-arm64 .
    
    # FreeBSD amd64
    echo -e "${YELLOW}构建 Agent (FreeBSD amd64)...${NC}"
    CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-freebsd-amd64 .
    
    cd ..
    
    echo -e "${GREEN}✓ Agent 构建完成${NC}"
}

# 构建 Alpine 调试版本 Agent
build_agent_debug() {
    echo -e "${YELLOW}开始构建 Alpine 调试版本 Agent...${NC}"
    
    cd agent
    
    # Linux amd64 调试版本
    echo -e "${YELLOW}构建 Agent Debug (Linux amd64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-debug-linux-amd64 \
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
        update_checker.go security.go
    
    # Linux arm64 调试版本
    echo -e "${YELLOW}构建 Agent Debug (Linux arm64)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-debug-linux-arm64 \
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
        update_checker.go security.go
    
    # Linux 386 调试版本
    echo -e "${YELLOW}构建 Agent Debug (Linux 386)...${NC}"
    CGO_ENABLED=0 GOOS=linux GOARCH=386 go build \
        -trimpath -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o ../${AGENT_OUTPUT_DIR}/dashgo-agent-debug-linux-386 \
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
        update_checker.go security.go
    
    # 复制诊断脚本
    echo -e "${YELLOW}复制诊断脚本...${NC}"
    cp debug-alpine.sh ../${AGENT_OUTPUT_DIR}/
    chmod +x ../${AGENT_OUTPUT_DIR}/debug-alpine.sh
    
    cd ..
    
    echo -e "${GREEN}✓ Alpine 调试版本 Agent 构建完成${NC}"
}

# 构建 Migrate 工具
build_migrate() {
    echo -e "${YELLOW}开始构建 Migrate 工具...${NC}"
    
    # Linux amd64
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath -ldflags="-s -w" \
        -o ${SERVER_OUTPUT_DIR}/migrate-linux-amd64 \
        ./cmd/migrate
    
    # Linux arm64
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
        -trimpath -ldflags="-s -w" \
        -o ${SERVER_OUTPUT_DIR}/migrate-linux-arm64 \
        ./cmd/migrate
    
    echo -e "${GREEN}✓ Migrate 工具构建完成${NC}"
}

# 生成校验和
generate_checksums() {
    echo -e "${YELLOW}生成校验和文件...${NC}"
    
    cd ${OUTPUT_DIR}
    
    # Agent 校验和
    cd agent
    sha256sum * > SHA256SUMS
    cd ..
    
    # Server 校验和
    cd server
    sha256sum * > SHA256SUMS
    cd ..
    
    cd ..
    
    echo -e "${GREEN}✓ 校验和文件生成完成${NC}"
}

# 创建版本信息文件
create_version_info() {
    echo -e "${YELLOW}创建版本信息文件...${NC}"
    
    cat > ${OUTPUT_DIR}/VERSION.txt << EOF
dashGO Build Information
========================

Version: ${VERSION}
Build Time: ${BUILD_TIME}
Git Commit: ${GIT_COMMIT}

Server Binaries:
- dashgo-server-linux-amd64
- dashgo-server-linux-arm64
- dashgo-server-windows-amd64.exe
- dashgo-server-darwin-amd64
- dashgo-server-darwin-arm64

Agent Binaries:
- dashgo-agent-linux-amd64
- dashgo-agent-linux-arm64
- dashgo-agent-linux-386
- dashgo-agent-windows-amd64.exe
- dashgo-agent-windows-386.exe
- dashgo-agent-darwin-amd64
- dashgo-agent-darwin-arm64
- dashgo-agent-freebsd-amd64

Alpine Debug Agent Binaries:
- dashgo-agent-debug-linux-amd64
- dashgo-agent-debug-linux-arm64
- dashgo-agent-debug-linux-386
- debug-alpine.sh (diagnostic script)

Tools:
- migrate-linux-amd64
- migrate-linux-arm64

Frontend:
- web/dist (Vue.js build)
EOF
    
    echo -e "${GREEN}✓ 版本信息文件创建完成${NC}"
}

# 显示构建结果
show_results() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}构建完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}构建产物位置: ${OUTPUT_DIR}/${NC}"
    echo ""
    echo -e "${YELLOW}文件大小:${NC}"
    du -sh ${OUTPUT_DIR}/*
    echo ""
    echo -e "${YELLOW}Agent 文件:${NC}"
    ls -lh ${AGENT_OUTPUT_DIR}/
    echo ""
    echo -e "${YELLOW}Server 文件:${NC}"
    ls -lh ${SERVER_OUTPUT_DIR}/
    echo ""
    echo -e "${GREEN}所有二进制文件已生成到 ${OUTPUT_DIR} 目录${NC}"
}

# 主函数
main() {
    # 检查参数
    case "${1}" in
        clean)
            clean
            ;;
        frontend)
            clean
            build_frontend
            ;;
        server)
            clean
            build_server
            build_migrate
            ;;
        server-docker)
            clean
            build_server_docker
            build_migrate
            ;;
        agent)
            clean
            build_agent
            ;;
        agent-debug)
            clean
            build_agent_debug
            ;;
        all|"")
            clean
            build_frontend
            build_server
            build_agent
            build_migrate
            generate_checksums
            create_version_info
            show_results
            ;;
        all-debug)
            clean
            build_frontend
            build_server
            build_agent
            build_agent_debug
            build_migrate
            generate_checksums
            create_version_info
            show_results
            ;;
        all-docker)
            clean
            build_frontend
            build_server_docker
            build_agent
            build_migrate
            generate_checksums
            create_version_info
            show_results
            ;;
        *)
            echo -e "${RED}未知参数: ${1}${NC}"
            echo ""
            echo "用法: $0 [clean|frontend|server|server-docker|agent|agent-debug|all|all-debug|all-docker]"
            echo ""
            echo "  clean         - 仅清理构建文件"
            echo "  frontend      - 仅构建前端"
            echo "  server        - 仅构建 Server (本地编译，当前架构支持 SQLite)"
            echo "  server-docker - 仅构建 Server (Docker 编译，所有架构支持 SQLite)"
            echo "  agent         - 仅构建 Agent (全架构)"
            echo "  agent-debug   - 仅构建 Alpine 调试版本 Agent"
            echo "  all           - 构建所有组件 (默认，本地编译)"
            echo "  all-debug     - 构建所有组件 + Alpine 调试版本"
            echo "  all-docker    - 构建所有组件 (Docker 编译，推荐)"
            echo ""
            echo "环境变量:"
            echo "  VERSION       - 设置版本号 (默认: 1.0.0)"
            echo "  RUN_TESTS     - 构建前运行测试 (设置为 'true' 启用)"
            echo ""
            echo "示例:"
            echo "  $0 all                    # 构建所有组件"
            echo "  VERSION=2.0.0 $0 all      # 指定版本号构建"
            echo "  RUN_TESTS=true $0 all     # 构建前运行测试"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
