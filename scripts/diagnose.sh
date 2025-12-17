#!/bin/bash

# dashGO 快速诊断脚本
# 用于排查 "Service temporarily unavailable" 错误

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}dashGO 诊断工具${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查是否在正确的目录
if [ ! -f "docker-compose.yaml" ]; then
    log_error "未找到 docker-compose.yaml"
    log_info "请在 dashGO 安装目录运行此脚本"
    exit 1
fi

# 1. 检查容器状态
echo "1. 检查容器状态..."
echo ""
docker-compose ps
echo ""

# 检查每个容器
for container in dashgo dashgo-nginx dashgo-redis; do
    if docker ps | grep -q $container; then
        log_info "$container 容器运行中"
    else
        log_error "$container 容器未运行"
        echo "  查看日志: docker-compose logs $container"
    fi
done
echo ""

# 2. 检查 dashGO 服务
echo "2. 检查 dashGO 服务..."
echo ""

if docker ps | grep -q dashgo; then
    # 检查服务健康状态
    if docker exec dashgo wget -q -O- http://localhost:8080/health 2>/dev/null; then
        log_info "dashGO 服务响应正常"
    else
        log_error "dashGO 服务无响应"
        echo ""
        echo "最近的错误日志:"
        docker-compose logs dashgo --tail=20 | grep -i error
    fi
else
    log_error "dashGO 容器未运行"
    echo ""
    echo "最近的日志:"
    docker-compose logs dashgo --tail=30
fi
echo ""

# 3. 检查 Nginx
echo "3. 检查 Nginx..."
echo ""

if docker ps | grep -q dashgo-nginx; then
    # 测试 Nginx 配置
    if docker exec dashgo-nginx nginx -t 2>&1 | grep -q "successful"; then
        log_info "Nginx 配置正确"
    else
        log_error "Nginx 配置错误"
        docker exec dashgo-nginx nginx -t
    fi
    
    # 检查 Nginx 能否连接到 dashGO
    if docker exec dashgo-nginx wget -q -O- http://dashgo:8080/health 2>/dev/null; then
        log_info "Nginx 可以连接到 dashGO"
    else
        log_error "Nginx 无法连接到 dashGO"
        echo "  可能的原因: 网络问题或 dashGO 服务未启动"
    fi
else
    log_error "Nginx 容器未运行"
fi
echo ""

# 4. 检查 Redis
echo "4. 检查 Redis..."
echo ""

if docker ps | grep -q dashgo-redis; then
    if docker exec dashgo-redis redis-cli ping 2>/dev/null | grep -q "PONG"; then
        log_info "Redis 运行正常"
    else
        log_warn "Redis 可能需要密码"
        # 尝试使用密码
        if [ -f ".env" ]; then
            REDIS_PASS=$(grep REDIS_PASSWORD .env | cut -d'=' -f2)
            if [ -n "$REDIS_PASS" ]; then
                if docker exec dashgo-redis redis-cli -a "$REDIS_PASS" ping 2>/dev/null | grep -q "PONG"; then
                    log_info "Redis 运行正常（使用密码）"
                fi
            fi
        fi
    fi
else
    log_error "Redis 容器未运行"
fi
echo ""

# 5. 检查网络
echo "5. 检查容器网络..."
echo ""

NETWORK=$(docker network ls | grep dashgo | awk '{print $2}' | head -1)
if [ -n "$NETWORK" ]; then
    log_info "网络: $NETWORK"
    
    # 检查容器是否在网络中
    for container in dashgo dashgo-nginx dashgo-redis; do
        if docker network inspect $NETWORK 2>/dev/null | grep -q $container; then
            log_info "$container 在网络中"
        else
            log_warn "$container 不在网络中"
        fi
    done
else
    log_error "未找到 dashGO 网络"
fi
echo ""

# 6. 检查端口
echo "6. 检查端口占用..."
echo ""

for port in 8080 80 443; do
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        log_info "端口 $port 正在监听"
    else
        log_warn "端口 $port 未监听"
    fi
done
echo ""

# 7. 检查配置文件
echo "7. 检查配置文件..."
echo ""

if [ -f "configs/config.yaml" ]; then
    log_info "配置文件存在: configs/config.yaml"
    
    # 检查数据库配置
    if grep -q "driver: sqlite" configs/config.yaml; then
        log_info "数据库类型: SQLite"
        if [ -f "data/dashgo.db" ]; then
            log_info "数据库文件存在"
            ls -lh data/dashgo.db
        else
            log_error "数据库文件不存在: data/dashgo.db"
        fi
    elif grep -q "driver: mysql" configs/config.yaml; then
        log_info "数据库类型: MySQL"
    fi
else
    log_error "配置文件不存在: configs/config.yaml"
fi
echo ""

# 8. 常见错误检查
echo "8. 检查常见错误..."
echo ""

# 检查数据库错误
if docker-compose logs dashgo 2>&1 | grep -qi "database"; then
    log_warn "发现数据库相关日志"
    docker-compose logs dashgo | grep -i database | tail -5
    echo ""
fi

# 检查连接错误
if docker-compose logs dashgo 2>&1 | grep -qi "connection refused\|cannot connect"; then
    log_error "发现连接错误"
    docker-compose logs dashgo | grep -i "connection refused\|cannot connect" | tail -5
    echo ""
fi

# 检查权限错误
if docker-compose logs dashgo 2>&1 | grep -qi "permission denied"; then
    log_error "发现权限错误"
    docker-compose logs dashgo | grep -i "permission denied" | tail -5
    echo ""
fi

# 9. 建议的修复步骤
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}建议的修复步骤${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 根据诊断结果给出建议
if ! docker ps | grep -q dashgo; then
    echo "1. dashGO 容器未运行，尝试重启:"
    echo "   docker-compose restart dashgo"
    echo ""
fi

if docker-compose logs dashgo 2>&1 | grep -qi "database"; then
    echo "2. 数据库问题，尝试重新初始化:"
    echo "   docker-compose restart dashgo"
    echo "   docker-compose logs dashgo -f"
    echo ""
fi

if ! docker exec dashgo-nginx wget -q -O- http://dashgo:8080/health 2>/dev/null; then
    echo "3. Nginx 无法连接到 dashGO，尝试重建网络:"
    echo "   docker-compose down"
    echo "   docker-compose up -d"
    echo ""
fi

echo "4. 查看实时日志:"
echo "   docker-compose logs -f"
echo ""

echo "5. 完整重启:"
echo "   docker-compose down"
echo "   docker-compose up -d"
echo ""

echo "详细故障排查指南: TROUBLESHOOTING.md"
echo ""
