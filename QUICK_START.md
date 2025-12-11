# XBoard-Go 快速开始

## 一键安装

```bash
bash setup.sh
```

## 菜单选项

```
1) 全新安装 (本地开发)      - SQLite/MySQL，适合开发测试
2) 安装到现有 MySQL 数据库  - 生产环境推荐
3) 升级现有数据库           - 保留数据，升级结构
4) 修复迁移问题             - 修复 host_id/sold_count 字段
5) 查看迁移状态             - 查看已执行的迁移
6) 生成配置文件             - 生成 config.yaml
```

## 常用场景

### 场景1：本地开发（SQLite）

```bash
bash setup.sh
# 选择 1 -> 选择 1 (SQLite)
./xboard-server
```

访问：http://localhost:8080

### 场景2：生产环境（MySQL）

```bash
# 1. 创建数据库
mysql -u root -p -e "CREATE DATABASE xboard;"

# 2. 运行安装脚本
bash setup.sh
# 选择 2
# 输入数据库信息

# 3. 启动服务
./xboard-server
```

### 场景3：升级现有系统

```bash
bash setup.sh
# 选择 3
```

### 场景4：修复保存失败问题

```bash
bash setup.sh
# 选择 4
```

## 默认管理员

- 邮箱：`admin@example.com`
- 密码：`admin123456`

## 配置文件

位置：`configs/config.yaml`

**重要**：必须使用 `driver` 而不是 `type`

```yaml
database:
  driver: "mysql"  # ✅ 正确
  # type: "mysql"  # ❌ 错误
```

## 详细文档

查看完整文档：`README_SETUP.md`

## 问题排查

### 保存节点失败

```bash
bash setup.sh
# 选择 4 (修复迁移问题)
```

### 配置文件错误

```bash
bash setup.sh
# 选择 6 (重新生成配置)
```

### 查看迁移状态

```bash
bash setup.sh
# 选择 5
```

## 手动操作

### 编译

```bash
make build
```

### 运行迁移

```bash
./migrate -action up
```

### 查看状态

```bash
./migrate -action status
```

## 支持

- GitHub: https://github.com/ZYHUO/xboard-go
- Issues: https://github.com/ZYHUO/xboard-go/issues
