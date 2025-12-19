# dashGO 完整编译脚本 (PowerShell 版本)
# 用于在 Windows 上编译 Dashboard 和全架构的 Agent
#
# 前置要求:
#   - Go >= 1.22 (用于编译 Server 和 Agent)
#   - Node.js >= 18.0.0 (用于编译前端)
#   - npm (随 Node.js 安装)
#
# 使用方法:
#   .\build-all.ps1 -Target all              # 构建所有组件
#   .\build-all.ps1 -Target frontend         # 仅构建前端
#   $env:RUN_TESTS="true"; .\build-all.ps1   # 构建前运行测试
#
# 详细文档:
#   - 前端测试设置: web/TESTING_SETUP.md
#   - 构建文档: BUILD.md

param(
    [string]$Target = "all",
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

# 版本信息
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"
$GIT_COMMIT = try { git rev-parse --short HEAD } catch { "unknown" }

# 输出目录
$OUTPUT_DIR = "dist"
$AGENT_OUTPUT_DIR = "$OUTPUT_DIR/agent"
$SERVER_OUTPUT_DIR = "$OUTPUT_DIR/server"
$WEB_OUTPUT_DIR = "$OUTPUT_DIR/web"

Write-Host "========================================" -ForegroundColor Green
Write-Host "XBoard 完整编译脚本" -ForegroundColor Green
Write-Host "版本: $Version" -ForegroundColor Green
Write-Host "构建时间: $BUILD_TIME" -ForegroundColor Green
Write-Host "Git Commit: $GIT_COMMIT" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# 清理旧的构建文件
function Clean {
    Write-Host "清理旧的构建文件..." -ForegroundColor Yellow
    if (Test-Path $OUTPUT_DIR) {
        Remove-Item -Recurse -Force $OUTPUT_DIR
    }
    New-Item -ItemType Directory -Force -Path $AGENT_OUTPUT_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $SERVER_OUTPUT_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $WEB_OUTPUT_DIR | Out-Null
    Write-Host "✓ 清理完成" -ForegroundColor Green
}

# 构建前端
function Build-Frontend {
    Write-Host "开始构建前端..." -ForegroundColor Yellow
    
    # 检查 Node.js 是否安装
    try {
        $nodeVersion = node --version
        Write-Host "  Node.js 版本: $nodeVersion" -ForegroundColor Green
        
        # 检查版本号
        $versionNumber = [int]($nodeVersion -replace 'v(\d+)\..*', '$1')
        if ($versionNumber -lt 18) {
            Write-Host "  警告: Node.js 版本过低 (当前: $nodeVersion, 推荐: >= 18.0.0)" -ForegroundColor Yellow
            Write-Host "  可能会遇到兼容性问题" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Host "错误: 未安装 Node.js" -ForegroundColor Red
        Write-Host "请访问 https://nodejs.org/ 安装 Node.js (>= 18.0.0)" -ForegroundColor Yellow
        Write-Host "或参考 web/TESTING_SETUP.md 获取详细安装指南" -ForegroundColor Yellow
        throw "Node.js 未安装"
    }
    
    # 检查并安装依赖
    if (-not (Test-Path "web/node_modules")) {
        Write-Host "安装前端依赖..." -ForegroundColor Yellow
        Push-Location web
        try {
            npm install
            Write-Host "✓ 依赖安装完成" -ForegroundColor Green
        }
        catch {
            Write-Host "依赖安装失败" -ForegroundColor Red
            Pop-Location
            throw
        }
        Pop-Location
    }
    else {
        Write-Host "✓ 依赖已存在，跳过安装" -ForegroundColor Green
    }
    
    # 运行测试（可选）
    if ($env:RUN_TESTS -eq "true") {
        Write-Host "运行前端测试..." -ForegroundColor Yellow
        Push-Location web
        try {
            npm run test:run
        }
        catch {
            Write-Host "警告: 测试失败，但继续构建" -ForegroundColor Yellow
        }
        Pop-Location
    }
    
    # 构建前端
    Write-Host "构建前端应用..." -ForegroundColor Yellow
    Push-Location web
    try {
        npm run build
    }
    catch {
        Write-Host "前端构建失败" -ForegroundColor Red
        Pop-Location
        throw
    }
    Pop-Location
    
    # 复制构建产物
    if (Test-Path "web/dist") {
        Copy-Item -Recurse -Force web/dist $WEB_OUTPUT_DIR/
        Write-Host "✓ 前端构建完成" -ForegroundColor Green
        
        # 显示构建产物大小
        $distSize = (Get-ChildItem -Recurse web/dist | Measure-Object -Property Length -Sum).Sum / 1MB
        Write-Host ("  构建产物大小: {0:N2} MB" -f $distSize) -ForegroundColor Green
    }
    else {
        Write-Host "错误: 构建产物目录不存在" -ForegroundColor Red
        throw "构建产物目录不存在"
    }
}

# 构建 Server (Dashboard)
function Build-Server {
    Write-Host "开始构建 Server..." -ForegroundColor Yellow
    
    $ldflags = "-s -w -X main.Version=$Version -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
    
    # Linux amd64
    Write-Host "构建 Server (Linux amd64)..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "$SERVER_OUTPUT_DIR/xboard-server-linux-amd64" ./cmd/server
    
    # Linux arm64
    Write-Host "构建 Server (Linux arm64)..." -ForegroundColor Yellow
    $env:GOARCH = "arm64"
    go build -ldflags="$ldflags" -o "$SERVER_OUTPUT_DIR/xboard-server-linux-arm64" ./cmd/server
    
    # Windows amd64
    Write-Host "构建 Server (Windows amd64)..." -ForegroundColor Yellow
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "$SERVER_OUTPUT_DIR/xboard-server-windows-amd64.exe" ./cmd/server
    
    # macOS amd64
    Write-Host "构建 Server (macOS amd64)..." -ForegroundColor Yellow
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "$SERVER_OUTPUT_DIR/xboard-server-darwin-amd64" ./cmd/server
    
    # macOS arm64
    Write-Host "构建 Server (macOS arm64)..." -ForegroundColor Yellow
    $env:GOARCH = "arm64"
    go build -ldflags="$ldflags" -o "$SERVER_OUTPUT_DIR/xboard-server-darwin-arm64" ./cmd/server
    
    Write-Host "✓ Server 构建完成" -ForegroundColor Green
}

# 构建 Agent (全架构)
function Build-Agent {
    Write-Host "开始构建 Agent (全架构)..." -ForegroundColor Yellow
    
    Push-Location agent
    
    $ldflags = "-s -w -X main.Version=$Version -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
    
    # Linux amd64
    Write-Host "构建 Agent (Linux amd64)..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-linux-amd64" .
    
    # Linux arm64
    Write-Host "构建 Agent (Linux arm64)..." -ForegroundColor Yellow
    $env:GOARCH = "arm64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-linux-arm64" .
    
    # Linux 386
    Write-Host "构建 Agent (Linux 386)..." -ForegroundColor Yellow
    $env:GOARCH = "386"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-linux-386" .
    
    # Windows amd64
    Write-Host "构建 Agent (Windows amd64)..." -ForegroundColor Yellow
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-windows-amd64.exe" .
    
    # Windows 386
    Write-Host "构建 Agent (Windows 386)..." -ForegroundColor Yellow
    $env:GOARCH = "386"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-windows-386.exe" .
    
    # macOS amd64
    Write-Host "构建 Agent (macOS amd64)..." -ForegroundColor Yellow
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-darwin-amd64" .
    
    # macOS arm64
    Write-Host "构建 Agent (macOS arm64)..." -ForegroundColor Yellow
    $env:GOARCH = "arm64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-darwin-arm64" .
    
    # FreeBSD amd64
    Write-Host "构建 Agent (FreeBSD amd64)..." -ForegroundColor Yellow
    $env:GOOS = "freebsd"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/xboard-agent-freebsd-amd64" .
    
    Pop-Location
    
    Write-Host "✓ Agent 构建完成" -ForegroundColor Green
}

# 构建 Alpine 调试版本 Agent
function Build-AgentDebug {
    Write-Host "开始构建 Alpine 调试版本 Agent..." -ForegroundColor Yellow
    
    Push-Location agent
    
    $ldflags = "-s -w -X main.Version=$Version -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
    
    # Linux amd64 调试版本
    Write-Host "构建 Agent Debug (Linux amd64)..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/dashgo-agent-debug-linux-amd64" `
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go `
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go `
        update_checker.go security.go security_unix.go
    
    # Linux arm64 调试版本
    Write-Host "构建 Agent Debug (Linux arm64)..." -ForegroundColor Yellow
    $env:GOARCH = "arm64"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/dashgo-agent-debug-linux-arm64" `
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go `
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go `
        update_checker.go security.go security_unix.go
    
    # Linux 386 调试版本
    Write-Host "构建 Agent Debug (Linux 386)..." -ForegroundColor Yellow
    $env:GOARCH = "386"
    go build -ldflags="$ldflags" -o "../$AGENT_OUTPUT_DIR/dashgo-agent-debug-linux-386" `
        main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go `
        alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go `
        update_checker.go security.go security_unix.go
    
    # 复制诊断脚本
    Write-Host "复制诊断脚本..." -ForegroundColor Yellow
    Copy-Item debug-alpine.sh "../$AGENT_OUTPUT_DIR/"
    
    Pop-Location
    
    Write-Host "✓ Alpine 调试版本 Agent 构建完成" -ForegroundColor Green
}

# 构建 Migrate 工具
function Build-Migrate {
    Write-Host "开始构建 Migrate 工具..." -ForegroundColor Yellow
    
    # Linux amd64
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags="-s -w" -o "$SERVER_OUTPUT_DIR/migrate-linux-amd64" ./cmd/migrate
    
    # Linux arm64
    $env:GOARCH = "arm64"
    go build -ldflags="-s -w" -o "$SERVER_OUTPUT_DIR/migrate-linux-arm64" ./cmd/migrate
    
    Write-Host "✓ Migrate 工具构建完成" -ForegroundColor Green
}

# 生成校验和
function Generate-Checksums {
    Write-Host "生成校验和文件..." -ForegroundColor Yellow
    
    # Agent 校验和
    Push-Location "$AGENT_OUTPUT_DIR"
    Get-ChildItem | Get-FileHash -Algorithm SHA256 | ForEach-Object {
        "$($_.Hash.ToLower())  $($_.Path | Split-Path -Leaf)"
    } | Out-File -Encoding ASCII SHA256SUMS
    Pop-Location
    
    # Server 校验和
    Push-Location "$SERVER_OUTPUT_DIR"
    Get-ChildItem | Get-FileHash -Algorithm SHA256 | ForEach-Object {
        "$($_.Hash.ToLower())  $($_.Path | Split-Path -Leaf)"
    } | Out-File -Encoding ASCII SHA256SUMS
    Pop-Location
    
    Write-Host "✓ 校验和文件生成完成" -ForegroundColor Green
}

# 创建版本信息文件
function Create-VersionInfo {
    Write-Host "创建版本信息文件..." -ForegroundColor Yellow
    
    @"
XBoard Build Information
========================

Version: $Version
Build Time: $BUILD_TIME
Git Commit: $GIT_COMMIT

Server Binaries:
- xboard-server-linux-amd64
- xboard-server-linux-arm64
- xboard-server-windows-amd64.exe
- xboard-server-darwin-amd64
- xboard-server-darwin-arm64

Agent Binaries:
- xboard-agent-linux-amd64
- xboard-agent-linux-arm64
- xboard-agent-linux-386
- xboard-agent-windows-amd64.exe
- xboard-agent-windows-386.exe
- xboard-agent-darwin-amd64
- xboard-agent-darwin-arm64
- xboard-agent-freebsd-amd64

Tools:
- migrate-linux-amd64
- migrate-linux-arm64

Frontend:
- web/dist (Vue.js build)
"@ | Out-File -Encoding UTF8 "$OUTPUT_DIR/VERSION.txt"
    
    Write-Host "✓ 版本信息文件创建完成" -ForegroundColor Green
}

# 显示构建结果
function Show-Results {
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "构建完成！" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "构建产物位置: $OUTPUT_DIR/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Agent 文件:" -ForegroundColor Yellow
    Get-ChildItem $AGENT_OUTPUT_DIR | Format-Table Name, Length -AutoSize
    Write-Host ""
    Write-Host "Server 文件:" -ForegroundColor Yellow
    Get-ChildItem $SERVER_OUTPUT_DIR | Format-Table Name, Length -AutoSize
    Write-Host ""
    Write-Host "所有二进制文件已生成到 $OUTPUT_DIR 目录" -ForegroundColor Green
}

# 主函数
switch ($Target.ToLower()) {
    "clean" {
        Clean
    }
    "frontend" {
        Clean
        Build-Frontend
    }
    "server" {
        Clean
        Build-Server
        Build-Migrate
    }
    "agent" {
        Clean
        Build-Agent
    }
    "agent-debug" {
        Clean
        Build-AgentDebug
    }
    "all" {
        Clean
        Build-Frontend
        Build-Server
        Build-Agent
        Build-Migrate
        Generate-Checksums
        Create-VersionInfo
        Show-Results
    }
    "all-debug" {
        Clean
        Build-Frontend
        Build-Server
        Build-Agent
        Build-AgentDebug
        Build-Migrate
        Generate-Checksums
        Create-VersionInfo
        Show-Results
    }
    default {
        Write-Host "未知参数: $Target" -ForegroundColor Red
        Write-Host ""
        Write-Host "用法: .\build-all.ps1 [-Target <target>] [-Version <version>]"
        Write-Host ""
        Write-Host "  -Target clean       - 仅清理构建文件"
        Write-Host "  -Target frontend    - 仅构建前端"
        Write-Host "  -Target server      - 仅构建 Server"
        Write-Host "  -Target agent       - 仅构建 Agent (全架构)"
        Write-Host "  -Target agent-debug - 仅构建 Alpine 调试版本 Agent"
        Write-Host "  -Target all         - 构建所有组件 (默认)"
        Write-Host "  -Target all-debug   - 构建所有组件 + Alpine 调试版本"
        Write-Host ""
        Write-Host "  -Version <ver>    - 指定版本号 (默认: 1.0.0)"
        Write-Host ""
        Write-Host "环境变量:"
        Write-Host "  `$env:RUN_TESTS   - 构建前运行测试 (设置为 'true' 启用)"
        Write-Host ""
        Write-Host "示例:"
        Write-Host "  .\build-all.ps1 -Target all                    # 构建所有组件"
        Write-Host "  .\build-all.ps1 -Target all -Version 2.0.0     # 指定版本号构建"
        Write-Host "  `$env:RUN_TESTS='true'; .\build-all.ps1         # 构建前运行测试"
        exit 1
    }
}
