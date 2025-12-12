# XBoard Docker 重启脚本 (Windows PowerShell)
# 用于快速重启和查看日志

$ErrorActionPreference = "Stop"

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Green }
function Write-Error { Write-Host "[ERROR] $args" -ForegroundColor Red }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }

# 检查配置文件
if (-not (Test-Path "configs/config.yaml")) {
    Write-Error "配置文件不存在: configs/config.yaml"
    Write-Info "正在从示例文件创建..."
    
    if (Test-Path "configs/config.example.yaml") {
        Copy-Item "configs/config.example.yaml" "configs/config.yaml"
        Write-Info "已创建配置文件，请根据需要修改"
    } else {
        Write-Error "示例配置文件也不存在！"
        exit 1
    }
}

# 创建必要的目录
if (-not (Test-Path "data")) {
    New-Item -ItemType Directory -Path "data" | Out-Null
}

Write-Info "停止现有容器..."
docker compose down

$rebuild = Read-Host "是否重新构建镜像? [y/N]"
if ($rebuild -eq "y" -or $rebuild -eq "Y") {
    Write-Info "重新构建镜像..."
    docker compose build --no-cache
}

Write-Info "启动服务..."
docker compose up -d

Write-Info "等待服务启动..."
Start-Sleep -Seconds 5

Write-Info "查看容器状态..."
docker compose ps

Write-Host ""
Write-Info "查看 xboard 日志 (Ctrl+C 退出)..."
Write-Host ""
docker compose logs -f xboard
