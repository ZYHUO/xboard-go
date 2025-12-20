#!/bin/bash

# 前端测试运行脚本
# 用于运行所有前端测试

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}运行前端测试${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 Node.js
if ! command -v node &>/dev/null; then
    echo -e "${RED}错误: 未安装 Node.js${NC}"
    exit 1
fi

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}安装依赖...${NC}"
    npm install
fi

# 运行测试
echo -e "${YELLOW}运行测试...${NC}"
npm run test:run

echo -e "${GREEN}✓ 测试完成${NC}"
