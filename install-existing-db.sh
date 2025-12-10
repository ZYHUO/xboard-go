#!/bin/bash

# XBoard 现有数据库安装脚本
# 适用于已有 MySQL 数据库的情况
# 用法: bash install-existing-db.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${PURPLE}[SUCCESS]${NC} $1"; }

clear
echo -e "${CYAN}"
cat << 'EOF'
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗      ║
║   ╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗     ║
║    ╚███╔╝ ██████╔╝██║   ██║███████║██████╔╝██║  ██║     ║
║    ██╔██╗ ██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║     ║
║   ██╔╝ ██╗██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝     ║
║   ╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝      ║
║                                                           ║
║              现有数据库安装向导                           ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

echo ""
log_info "此脚本将帮助你配置现有的 MySQL 数据库"
echo ""

# 步骤 1: 获取数据库信息
log_info "步骤 1/4: 输入数据库信息"
echo ""

read -p "数据库主机 [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "数据库端口 [3306]: " DB_PORT
DB_PORT=${DB_PORT:-3306}

read -p "数据库用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}

read -sp "数据库密码: " DB_PASS
echo ""

read -p "数据库名称: " DB_NAME

if [ -z "$DB_NAME" ]; then
    log_error "数据库名称不能为空"
    exit 1
fi

# 步骤 2: 测试连接
echo ""
log_info "步骤 2/4: 测试数据库连接"

if ! command -v mysql &>/dev/null; then
    log_error "未检测到 mysql 命令"
    log_info "请先安装 MySQL 客户端"
    exit 1
fi

if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME;" 2>/dev/null; then
    log_success "数据库连接成功"
else
    log_error "数据库连接失败，请检查信息是否正确"
    exit 1
fi

# 步骤 3: 创建配置文件
echo ""
log_info "步骤 3/4: 创建配置文件"

# 生成 JWT Secret
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "change-this-secret-key-in-production")

# 创建配置目录
mkdir -p configs

# 创建配置文件
cat > configs/config.yaml << EOF
app:
  name: "XBoard"
  url: "http://localhost:8080"
  debug: false
  jwt_secret: "${JWT_SECRET}"
  
server:
  host: "0.0.0.0"
  port: 8080
  
database:
  type: "mysql"
  host: "${DB_HOST}"
  port: ${DB_PORT}
  username: "${DB_USER}"
  password: "${DB_PASS}"
  database: "${DB_NAME}"
  
redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  
mail:
  driver: "smtp"
  host: "smtp.gmail.com"
  port: 587
  username: ""
  password: ""
  encryption: "tls"
  from_address: ""
  from_name: "XBoard"
  
telegram:
  bot_token: ""
  
subscribe:
  single_mode: false
EOF

log_success "配置文件已创建: configs/config.yaml"

# 步骤 4: 执行数据库迁移
echo ""
log_info "步骤 4/4: 执行数据库迁移"
echo ""

# 检查是否需要清理迁移记录
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT 1 FROM migrations LIMIT 1;" 2>/dev/null; then
    log_warn "检测到已有迁移记录"
    read -p "是否清理迁移记录重新开始? [y/N]: " clean_migrations
    
    if [ "$clean_migrations" = "y" ] || [ "$clean_migrations" = "Y" ]; then
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "TRUNCATE TABLE migrations;" 2>/dev/null || true
        log_info "迁移记录已清理"
    fi
fi

echo ""
log_info "开始执行迁移..."
echo ""

# 执行迁移
if bash migrate.sh up; then
    log_success "数据库迁移完成"
else
    log_error "数据库迁移失败"
    exit 1
fi

# 显示结果
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                                                           ║"
echo "║                  ✓ 安装完成！                             ║"
echo "║                                                           ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

echo "配置信息:"
echo "  数据库: $DB_NAME"
echo "  主机: $DB_HOST:$DB_PORT"
echo "  配置文件: configs/config.yaml"
echo ""

echo "下一步:"
echo "  1. 配置用户组权限"
echo "     mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p $DB_NAME"
echo "     UPDATE v2_user_group SET server_ids = '[1,2,3]' WHERE id = 2;"
echo "     UPDATE v2_user_group SET plan_ids = '[1,2,3]' WHERE id = 2;"
echo ""
echo "  2. 启动服务"
echo "     go run ./cmd/server -config configs/config.yaml"
echo "     或"
echo "     docker compose up -d"
echo ""
echo "  3. 访问后台"
echo "     http://localhost:8080/admin"
echo ""

log_info "详细文档: docs/user-group-design.md"
echo ""
