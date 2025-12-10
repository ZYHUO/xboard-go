#!/bin/bash

# 修复迁移问题的快速脚本

set -e

echo "=========================================="
echo "XBoard 迁移修复脚本"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# 读取数据库配置
echo "请输入数据库信息:"
read -p "主机 (默认: localhost): " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "端口 (默认: 3306): " DB_PORT
DB_PORT=${DB_PORT:-3306}

read -p "数据库名: " DB_NAME
if [ -z "$DB_NAME" ]; then
    echo -e "${RED}错误: 数据库名不能为空${NC}"
    exit 1
fi

read -p "用户名 (默认: root): " DB_USER
DB_USER=${DB_USER:-root}

read -sp "密码: " DB_PASS
echo ""

# 测试连接
echo ""
echo "测试数据库连接..."
if ! mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT 1;" &>/dev/null; then
    echo -e "${RED}错误: 无法连接到数据库${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 数据库连接成功${NC}"

# 1. 清理错误的迁移记录
echo ""
echo "步骤 1: 清理错误的迁移记录..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
DELETE FROM migrations WHERE name LIKE '%_rollback%';
EOF
echo -e "${GREEN}✓ 清理完成${NC}"

# 2. 检查并添加 host_id 字段
echo ""
echo "步骤 2: 检查 v2_server 表结构..."
HAS_HOST_ID=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW COLUMNS FROM v2_server LIKE 'host_id';" | wc -l)

if [ "$HAS_HOST_ID" -eq 0 ]; then
    echo -e "${YELLOW}⚠ host_id 字段不存在，正在添加...${NC}"
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE \`v2_server\` 
ADD COLUMN \`host_id\` BIGINT NULL DEFAULT NULL COMMENT '绑定的主机ID' AFTER \`parent_id\`;

ALTER TABLE \`v2_server\` 
ADD INDEX \`idx_server_host_id\` (\`host_id\`);
EOF
    echo -e "${GREEN}✓ host_id 字段已添加${NC}"
else
    echo -e "${GREEN}✓ host_id 字段已存在${NC}"
fi

# 3. 检查并添加 sold_count 字段
echo ""
echo "步骤 3: 检查 v2_plan 表结构..."
HAS_SOLD_COUNT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SHOW COLUMNS FROM v2_plan LIKE 'sold_count';" | wc -l)

if [ "$HAS_SOLD_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}⚠ sold_count 字段不存在，正在添加...${NC}"
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<EOF
ALTER TABLE \`v2_plan\` ADD COLUMN \`sold_count\` INT NOT NULL DEFAULT 0 COMMENT '已售出数量';

UPDATE \`v2_plan\` p 
SET \`sold_count\` = (
    SELECT COUNT(*) 
    FROM \`v2_user\` u 
    WHERE u.\`plan_id\` = p.\`id\`
);

CREATE INDEX \`idx_plan_capacity\` ON \`v2_plan\`(\`capacity_limit\`, \`sold_count\`);
EOF
    echo -e "${GREEN}✓ sold_count 字段已添加${NC}"
else
    echo -e "${GREEN}✓ sold_count 字段已存在${NC}"
fi

# 4. 显示迁移状态
echo ""
echo "步骤 4: 当前迁移状态..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT * FROM migrations ORDER BY id;"

echo ""
echo "=========================================="
echo -e "${GREEN}修复完成！${NC}"
echo "=========================================="
echo ""
echo "下一步:"
echo "1. 重新编译项目: make build"
echo "2. 重启服务: docker compose restart 或 systemctl restart xboard"
echo "3. 测试节点保存功能"
echo ""
