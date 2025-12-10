# 迁移系统修复说明

## 问题发现

用户报告"保存失败"，经过检查发现是数据库迁移的问题。

## 根本原因

迁移系统会执行所有 `.sql` 文件，包括 `_rollback.sql` 文件，导致：

1. 执行 `001_add_host_id_to_servers.sql` - 添加 `host_id` 字段 ✅
2. 执行 `001_add_host_id_to_servers_rollback.sql` - 删除 `host_id` 字段 ❌

结果：`host_id` 字段被删除了，导致保存节点时失败！

## 迁移文件列表

```
migrations/
├── 001_add_host_id_to_servers.sql          <- UP (添加字段)
├── 001_add_host_id_to_servers_rollback.sql <- DOWN (删除字段) ⚠️ 会被执行！
├── 002_add_user_fields.sql
├── 003_create_user_group.sql
├── 003_create_user_group_rollback.sql      <- DOWN ⚠️ 会被执行！
├── 004_simplify_user_group.sql
├── 004_simplify_user_group_rollback.sql    <- DOWN ⚠️ 会被执行！
└── 005_add_plan_sold_count.sql
```

## 修复方案

### 方案1：修改迁移系统（已实施）

修改 `cmd/migrate/main.go`，跳过 `_rollback.sql` 文件：

```go
// 过滤并排序 SQL 文件（跳过 rollback 文件）
var sqlFiles []string
for _, f := range files {
    if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && !strings.Contains(f.Name(), "_rollback") {
        sqlFiles = append(sqlFiles, f.Name())
    }
}
```

### 方案2：删除 rollback 文件（推荐）

Rollback 文件不应该放在 migrations 目录中，应该：

1. 删除所有 `_rollback.sql` 文件
2. 或者移动到单独的目录 `migrations/rollback/`

## 修复步骤

### 1. 重新编译迁移工具

```bash
cd cmd/migrate
go build -o ../../migrate
```

### 2. 清理迁移记录

```bash
# 连接到数据库
mysql -u root -p your_database

# 删除错误的迁移记录
DELETE FROM migrations WHERE name LIKE '%_rollback%';

# 退出
EXIT;
```

### 3. 检查表结构

```bash
# 检查 v2_server 表是否有 host_id 字段
mysql -u root -p your_database -e "DESCRIBE v2_server;"
```

如果没有 `host_id` 字段，手动添加：

```sql
ALTER TABLE `v2_server` 
ADD COLUMN `host_id` BIGINT NULL DEFAULT NULL COMMENT '绑定的主机ID' AFTER `parent_id`;

ALTER TABLE `v2_server` 
ADD INDEX `idx_server_host_id` (`host_id`);
```

### 4. 重新运行迁移

```bash
./migrate -action up
```

## 验证

### 1. 检查迁移记录

```bash
mysql -u root -p your_database -e "SELECT * FROM migrations ORDER BY id;"
```

应该只看到：
```
001_add_host_id_to_servers.sql
002_add_user_fields.sql
003_create_user_group.sql
004_simplify_user_group.sql
005_add_plan_sold_count.sql
```

**不应该**看到任何 `_rollback.sql` 文件！

### 2. 检查表结构

```bash
mysql -u root -p your_database -e "DESCRIBE v2_server;" | grep host_id
```

应该看到：
```
host_id | bigint | YES | MUL | NULL |
```

### 3. 测试保存节点

在前端创建或编辑节点，选择绑定主机，保存应该成功。

## 长期解决方案

### 选项1：删除 rollback 文件

```bash
cd migrations
rm *_rollback.sql
```

### 选项2：移动到单独目录

```bash
mkdir -p migrations/rollback
mv migrations/*_rollback.sql migrations/rollback/
```

### 选项3：使用标准迁移格式

使用 `+migrate Up/Down` 标记，将 up 和 down 放在同一个文件中：

```sql
-- +migrate Up
ALTER TABLE `v2_server` ADD COLUMN `host_id` BIGINT NULL;

-- +migrate Down
ALTER TABLE `v2_server` DROP COLUMN `host_id`;
```

## 相关文件

```
modified:   cmd/migrate/main.go
new file:   MIGRATION_FIX.md
```

## 总结

- ✅ 修复了迁移系统，跳过 rollback 文件
- ⏳ 需要用户清理迁移记录并重新运行
- ⏳ 建议删除或移动 rollback 文件
- ⏳ 考虑使用标准迁移格式
