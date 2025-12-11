# XBoard-Go 安装和升级指南

## 快速开始

使用一键脚本安装、升级或修复：

```bash
bash setup.sh
```

## 功能菜单

```
1) 全新安装 (本地开发)      - 适合开发者，支持 SQLite/MySQL
2) 安装到现有 MySQL 数据库  - 适合生产环境
3) 升级现有数据库           - 保留数据，只升级结构
4) 修复迁移问题             - 修复 host_id 和 sold_count 字段
5) 查看迁移状态             - 查看已执行的迁移
6) 生成配置文件             - 生成 config.yaml
0) 退出
```

---

## 详细说明

### 1. 全新安装 (本地开发)

适合：本地开发、测试环境

**步骤**：
1. 选择数据库类型（SQLite 或 MySQL）
2. 输入数据库信息（MySQL 需要）
3. 自动生成配置文件
4. 编译项目
5. 运行数据库迁移

**SQLite 示例**：
```bash
bash setup.sh
# 选择 1 -> 选择 1 (SQLite)
```

**MySQL 示例**：
```bash
bash setup.sh
# 选择 1 -> 选择 2 (MySQL)
# 输入数据库信息
```

**启动服务**：
```bash
./xboard-server
```

**默认管理员**：
- 邮箱：`admin@example.com`
- 密码：`admin123456`

---

### 2. 安装到现有 MySQL 数据库

适合：生产环境、已有 MySQL 数据库

**前提条件**：
- MySQL 5.7+ 或 MariaDB 10.2+
- 数据库已创建
- 用户有足够权限

**步骤**：
1. 输入数据库连接信息
2. 测试数据库连接
3. 生成配置文件
4. 编译项目
5. 运行数据库迁移

**示例**：
```bash
bash setup.sh
# 选择 2
# 输入：
#   主机: localhost
#   端口: 3306
#   数据库名: xboard
#   用户名: root
#   密码: ******
```

---

### 3. 升级现有数据库

适合：已安装 XBoard-Go，需要升级到新版本

**特点**：
- ✅ 保留所有数据
- ✅ 只升级数据库结构
- ✅ 自动备份（MySQL）
- ✅ 安全可靠

**步骤**：
1. 检查配置文件
2. 读取数据库信息
3. 备份数据库（MySQL）
4. 运行新的迁移

**示例**：
```bash
bash setup.sh
# 选择 3
# 确认升级
```

**备份位置**：
```
backups/backup_before_upgrade_20241211_153045.sql
```

---

### 4. 修复迁移问题

适合：遇到以下问题时使用

**问题症状**：
- ❌ 保存节点时报错："host_id field not found"
- ❌ 套餐列表不显示库存信息
- ❌ 迁移记录中有 `_rollback.sql` 文件

**修复内容**：
1. 清理错误的迁移记录（`_rollback.sql`）
2. 检查并添加 `host_id` 字段（v2_server 表）
3. 检查并添加 `sold_count` 字段（v2_plan 表）
4. 初始化 `sold_count` 数据

**示例**：
```bash
bash setup.sh
# 选择 4
```

**验证修复**：
```sql
-- 检查 v2_server 表
DESCRIBE v2_server;
-- 应该看到 host_id 字段

-- 检查 v2_plan 表
DESCRIBE v2_plan;
-- 应该看到 sold_count 字段

-- 检查迁移记录
SELECT * FROM migrations;
-- 不应该有 _rollback.sql 文件
```

---

### 5. 查看迁移状态

查看已执行的数据库迁移

**示例**：
```bash
bash setup.sh
# 选择 5
```

**输出示例**：
```
已执行的迁移:
  ✓ 001_add_host_id_to_servers.sql
  ✓ 002_add_user_fields.sql
  ✓ 003_create_user_group.sql
  ✓ 004_simplify_user_group.sql
  ✓ 005_add_plan_sold_count.sql

待执行的迁移:
  (无)
```

---

### 6. 生成配置文件

生成或重新生成 `configs/config.yaml`

**示例**：
```bash
bash setup.sh
# 选择 6
# 选择数据库类型
# 输入数据库信息
```

**配置文件位置**：
```
configs/config.yaml
```

---

## 配置文件说明

### MySQL 配置

```yaml
app:
  name: "XBoard"
  mode: "release"  # debug 或 release
  listen: ":8080"

database:
  driver: "mysql"  # 必须是 driver，不是 type
  host: "localhost"
  port: 3306
  database: "xboard"
  username: "root"
  password: "your_password"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-random-secret-key"
  expire_hour: 24

node:
  token: "your-node-token"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

admin:
  email: "admin@example.com"
  password: "admin123456"
```

### SQLite 配置

```yaml
database:
  driver: "sqlite"
  database: "xboard.db"
```

---

## 常见问题

### Q1: 保存节点时报错 "host_id field not found"

**原因**：数据库表中缺少 `host_id` 字段

**解决**：
```bash
bash setup.sh
# 选择 4 (修复迁移问题)
```

### Q2: 配置文件使用 `type` 还是 `driver`？

**答案**：必须使用 `driver`

```yaml
# ❌ 错误
database:
  type: "mysql"

# ✅ 正确
database:
  driver: "mysql"
```

### Q3: 如何备份数据库？

**MySQL**：
```bash
mysqldump -u root -p xboard > backup.sql
```

**SQLite**：
```bash
cp xboard.db xboard.db.backup
```

### Q4: 如何恢复备份？

**MySQL**：
```bash
mysql -u root -p xboard < backup.sql
```

**SQLite**：
```bash
cp xboard.db.backup xboard.db
```

### Q5: 迁移失败怎么办？

1. 查看错误信息
2. 检查数据库权限
3. 运行修复脚本：
```bash
bash setup.sh
# 选择 4
```

### Q6: 如何清理所有数据重新开始？

**MySQL**：
```sql
DROP DATABASE xboard;
CREATE DATABASE xboard;
```

**SQLite**：
```bash
rm xboard.db
```

然后重新运行安装脚本。

---

## 手动迁移

如果脚本无法使用，可以手动执行迁移：

### 1. 编译迁移工具

```bash
cd cmd/migrate
go build -o ../../migrate
cd ../..
```

### 2. 运行迁移

```bash
./migrate -action up
```

### 3. 查看状态

```bash
./migrate -action status
```

### 4. 自动迁移（不推荐）

```bash
./migrate -action auto
```

---

## 数据库结构

### 核心表

- `v2_user` - 用户表
- `v2_plan` - 套餐表
- `v2_order` - 订单表
- `v2_server` - 节点表
- `v2_host` - 主机表
- `v2_server_node` - 节点实例表
- `v2_user_group` - 用户组表
- `migrations` - 迁移记录表

### 新增字段

#### v2_server.host_id
- 类型：`BIGINT NULL`
- 用途：节点绑定到主机（自动部署）
- 索引：`idx_server_host_id`

#### v2_plan.sold_count
- 类型：`INT NOT NULL DEFAULT 0`
- 用途：套餐已售出数量（库存管理）
- 索引：`idx_plan_capacity`

---

## 架构说明

### 节点-主机关系

```
Server (逻辑节点)
  ├─ host_id → 绑定到 Host（可选，用于自动部署）
  └─ 用于生成订阅链接

Host (物理主机)
  ├─ 运行 sing-box
  └─ 通过 Agent 与面板通信

ServerNode (节点实例)
  ├─ 运行在 Host 上
  └─ 可以绑定到 Server（继承配置）
```

### 套餐-库存管理

```
Plan (套餐)
  ├─ capacity_limit: 最大可售数量
  ├─ sold_count: 已售出数量
  ├─ CanPurchase(): 检查是否可购买
  └─ GetRemainingCount(): 获取剩余数量
```

---

## 开发指南

### 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/ZYHUO/xboard-go.git
cd xboard-go

# 2. 安装依赖
go mod download

# 3. 运行安装脚本
bash setup.sh
# 选择 1 (全新安装)
# 选择 1 (SQLite)

# 4. 启动服务
./xboard-server

# 5. 访问
# 前端: http://localhost:8080
# API: http://localhost:8080/api/v2
```

### 创建新迁移

```bash
# 1. 创建迁移文件
touch migrations/006_your_migration.sql

# 2. 编写 SQL
cat > migrations/006_your_migration.sql <<EOF
-- 你的 SQL 语句
ALTER TABLE v2_user ADD COLUMN new_field VARCHAR(255);
EOF

# 3. 运行迁移
./migrate -action up
```

**注意**：不要创建 `_rollback.sql` 文件，迁移系统会跳过它们。

---

## 生产部署

### Docker Compose

```yaml
version: '3'
services:
  xboard:
    image: xboard-go:latest
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs
      - ./data:/app/data
    environment:
      - CONFIG_PATH=/app/configs/config.yaml
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: xboard
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

### Systemd 服务

```ini
[Unit]
Description=XBoard Go Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=xboard
WorkingDirectory=/opt/xboard
ExecStart=/opt/xboard/xboard-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## 更新日志

### v1.0.0 (2024-12-11)

**新功能**：
- ✅ 节点可以绑定到主机（自动部署）
- ✅ 套餐购买数量限制（库存管理）
- ✅ 用户组权限管理（简化设计）
- ✅ 一键安装/升级/修复脚本

**修复**：
- ✅ 配置文件字段名称（`driver` 替代 `type`）
- ✅ 迁移系统跳过 rollback 文件
- ✅ Plan 模型添加 `sold_count` 字段
- ✅ Server 模型添加 `host_id` 字段

**文档**：
- ✅ 完整的安装指南
- ✅ 数据库迁移指南
- ✅ 架构设计文档
- ✅ API 文档

---

## 支持

- GitHub: https://github.com/ZYHUO/xboard-go
- Issues: https://github.com/ZYHUO/xboard-go/issues
- Discussions: https://github.com/ZYHUO/xboard-go/discussions

---

## 许可证

MIT License
