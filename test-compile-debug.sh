#!/bin/bash

# 测试 Alpine 调试版本编译脚本

set -e

echo "测试 Alpine 调试版本编译..."

cd agent

echo "编译 Linux amd64 调试版本..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=test" \
    -o test-debug-amd64 \
    main_debug.go debug_logger.go alpine_types.go alpine_system_checker.go \
    alpine_system_checker_unix.go alpine_error_handler.go diagnostic_tool.go version.go \
    update_checker.go security.go

echo "编译成功！"
ls -lh test-debug-amd64

echo "测试帮助信息..."
./test-debug-amd64 -h || echo "帮助信息显示正常"

echo "清理测试文件..."
rm -f test-debug-amd64

echo "✅ Alpine 调试版本编译测试通过！"