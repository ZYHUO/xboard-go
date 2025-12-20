# 前端测试运行脚本 (PowerShell)
# 用于运行所有前端测试

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Green
Write-Host "运行前端测试" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# 检查 Node.js
try {
    $nodeVersion = node --version
    Write-Host "Node.js 版本: $nodeVersion" -ForegroundColor Green
}
catch {
    Write-Host "错误: 未安装 Node.js" -ForegroundColor Red
    exit 1
}

# 检查依赖
if (-not (Test-Path "node_modules")) {
    Write-Host "安装依赖..." -ForegroundColor Yellow
    npm install
}

# 运行测试
Write-Host "运行测试..." -ForegroundColor Yellow
npm run test:run

Write-Host "✓ 测试完成" -ForegroundColor Green
