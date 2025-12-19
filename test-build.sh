#!/bin/bash

# 简化的 Alpine 调试版本编译测试

set -e

echo "🔧 测试 Alpine 调试版本编译..."

cd agent

echo "📋 检查关键文件..."
if [ ! -f "security_unix.go" ]; then
    echo "❌ security_unix.go 文件缺失"
    exit 1
fi

if [ ! -f "security.go" ]; then
    echo "❌ security.go 文件缺失"  
    exit 1
fi

echo "✅ 关键文件检查通过"

echo "🚀 尝试编译..."

# 方法1: 使用完整文件列表
echo "方法1: 显式文件列表编译"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=test" \
    -o test-debug-1 \
    main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
    alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
    update_checker.go security.go security_unix.go

if [ $? -eq 0 ]; then
    echo "✅ 方法1 编译成功"
    ls -lh test-debug-1
    rm -f test-debug-1
else
    echo "❌ 方法1 编译失败"
fi

echo ""

# 方法2: 使用 Go 模块
echo "方法2: Go 模块编译"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags unix -o test-debug-2 .

if [ $? -eq 0 ]; then
    echo "✅ 方法2 编译成功"
    ls -lh test-debug-2
    rm -f test-debug-2
else
    echo "❌ 方法2 编译失败，这是正常的，因为我们没有完整的 Go 模块设置"
fi

echo ""
echo "🎉 测试完成！"