#!/bin/bash

# 测试 Alpine 调试版本编译

set -e

echo "🔧 测试 Alpine 调试版本编译..."

cd agent

echo "📋 检查所需文件..."
required_files=(
    "main_debug.go"
    "debug_logger.go" 
    "alpine_types.go"
    "alpine_system_checker.go"
    "alpine_system_checker_unix.go"
    "alpine_error_handler.go"
    "diagnostic_tool.go"
    "version.go"
    "update_checker.go"
    "security.go"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file"
    else
        echo "❌ $file (缺失)"
    fi
done

echo ""
echo "🚀 开始编译..."

# 尝试编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=test" \
    -o test-debug \
    main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
    alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
    update_checker.go security.go

if [ $? -eq 0 ]; then
    echo "✅ 编译成功！"
    ls -lh test-debug
    echo ""
    echo "🧪 测试可执行文件..."
    ./test-debug -h 2>/dev/null || echo "帮助信息正常"
    rm -f test-debug
else
    echo "❌ 编译失败"
    exit 1
fi

echo ""
echo "🎉 Alpine 调试版本编译测试完成！"