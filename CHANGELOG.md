# XBoard Go 更新日志

## [未发布] - 2023-12-10

### 新增功能

#### 用户组系统
- ✅ 添加用户组模型 (`internal/model/user_group.go`)
- ✅ 用户组服务层 (`internal/service/user_group.go`)
- ✅ 用户组数据访问层 (`internal/repository/user_group.go`)
- ✅ 用户组管理 API (`internal/handler/user_group.go`)
- ✅ 用户组数据库迁移 (`migrations/003_create_user_group.sql`)

#### 套餐升级功能
- ✅ 套餐添加 `upgrade_group_id` 字段
- ✅ 购买套餐后自动升级用户组
- ✅ 订单完成时处理用户组升级逻辑

#### 本地安装脚本
- ✅ 本地安装脚本 (`local-install.sh`)
  - 开发环境安装
  - 生产环境安装（Docker）
  - 编译二进制文件
  - 数据库迁移
  - 前端构建

#### 数据库迁移工具
- ✅ 迁移管理脚本 (`migrate.sh`)
  - 执行迁移 (up)
  - 回滚迁移 (down)
  - 查看状态 (status)
  - 自动迁移 (auto)
  - 重置数据库 (reset)
  - 创建迁移文件 (create)

#### 数据库升级工具
- ✅ 升级脚本 (`upgrade.sh`)
  - 自动备份数据库
  - 检查数据完整性
  - 执行增量迁移
  - 验证升级结果
  - 保留所有现有数据

### 改进

#### 用户组设计优化
- ✅ 简化用户组模型，移除不必要的默认值字段
- ✅ 明确用户组职责：权限控制和套餐可见性
- ✅ 流量、速度、设备限制由套餐决定

#### 订阅生成优化
- ✅ 修复 `GetAvailableServersForUser` 返回类型
- ✅ 支持根据用户组获取可访问节点
- ✅ 自动构建 `ServerInfo` 包含用户密码

#### Makefile 增强
- ✅ 添加安装相关命令
- ✅ 添加迁移相关命令
- ✅ 添加前端构建命令
- ✅ 添加帮助信息

### 文档

#### 新增文档
- ✅ `docs/user-group-design.md` - 用户组设计说明
- ✅ `docs/local-installation.md` - 本地安装指南
- ✅ `docs/database-migration.md` - 数据库迁移指南
- ✅ `docs/upgrade-guide.md` - 数据库升级指南
- ✅ `QUICK_INSTALL.md` - 快速安装指南
- ✅ `MIGRATION_GUIDE.md` - 迁移快速参考
- ✅ `CHANGELOG.md` - 更新日志

#### 更新文档
- ✅ `README.md` - 添加快速开始和迁移说明

### 修复

#### 编译错误
- ✅ 修复 `model.Plan` 缺少 `UpgradeGroupID` 字段
- ✅ 修复 `GetAvailableServersForUser` 返回类型不匹配
- ✅ 修复 `cmd/migrate/main.go` 未使用的导入

#### 模型完善
- ✅ 添加 `UserGroup` 到自动迁移列表
- ✅ `Plan` 模型添加 `UpgradeGroupID` 字段
- ✅ `UserGroupService` 添加 `ServerService` 依赖

### 数据库变更

#### 新增表
```sql
-- v2_user_group 用户组表
CREATE TABLE v2_user_group (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL,
    description TEXT,
    server_ids JSON,
    plan_ids JSON,
    sort INT DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
```

#### 修改表
```sql
-- v2_user 添加 group_id 字段
ALTER TABLE v2_user ADD COLUMN group_id BIGINT DEFAULT 1;

-- v2_plan 添加 upgrade_group_id 字段
ALTER TABLE v2_plan ADD COLUMN upgrade_group_id BIGINT;
```

### 迁移文件

- `migrations/001_add_host_id_to_servers.sql` - 添加 host_id 字段
- `migrations/002_add_user_fields.sql` - 添加用户字段
- `migrations/003_create_user_group.sql` - 创建用户组表
- `migrations/003_create_user_group_rollback.sql` - 回滚用户组
- `migrations/004_simplify_user_group.sql` - 简化用户组

### 安装和升级

#### 全新安装

```bash
# 开发环境
bash local-install.sh dev

# 生产环境
bash local-install.sh prod
```

#### 升级现有数据库

```bash
# 自动升级（推荐）
bash upgrade.sh

# 手动升级
bash migrate.sh up
```

### 破坏性变更

#### 用户组字段废弃

以下字段已标记为废弃，但保留用于向后兼容：
- `UserGroup.DefaultTransferEnable`
- `UserGroup.DefaultSpeedLimit`
- `UserGroup.DefaultDeviceLimit`

**原因：** 流量、速度、设备限制应该由套餐决定，不应该在用户组中设置。

**迁移建议：**
- 将流量配置移到套餐中
- 用户组只负责权限控制

### API 变更

#### 新增 API

**用户组管理**
```
GET    /api/v2/admin/user-groups           # 获取用户组列表
POST   /api/v2/admin/user-group            # 创建用户组
GET    /api/v2/admin/user-group/:id        # 获取用户组详情
PUT    /api/v2/admin/user-group/:id        # 更新用户组
DELETE /api/v2/admin/user-group/:id        # 删除用户组
```

**流量管理**
```
GET    /api/v2/admin/traffic/stats         # 获取流量统计
GET    /api/v2/admin/traffic/warnings      # 获取流量预警用户
POST   /api/v2/admin/traffic/reset/:id     # 重置用户流量
POST   /api/v2/admin/traffic/reset-all     # 重置所有用户流量
```

#### 修改 API

**套餐管理**
- `POST /api/v2/admin/plan` - 添加 `upgrade_group_id` 参数
- `PUT /api/v2/admin/plan/:id` - 添加 `upgrade_group_id` 参数

**用户组管理**
- `POST /api/v2/admin/user-group` - 移除 `default_transfer_enable` 等参数
- `PUT /api/v2/admin/user-group/:id` - 移除 `default_transfer_enable` 等参数

### 依赖更新

无依赖更新。

### 已知问题

- 流量统计采用平均分配算法，单用户流量可能不够精确
- 详见 `docs/traffic-limitation.md`

### 下一步计划

- [ ] 添加用户组权限管理（更细粒度）
- [ ] 优化流量统计算法
- [ ] 添加流量日志记录
- [ ] 添加流量分析报表
- [ ] 前端界面完善

### 贡献者

- 感谢所有贡献者的支持

### 许可证

MIT License

---

## 如何升级

### 从旧版本升级

如果你已经有旧版本的数据库：

```bash
# 1. 备份数据库
mysqldump -u root -p xboard > backup_$(date +%Y%m%d).sql

# 2. 拉取最新代码
git pull origin main

# 3. 运行升级脚本
bash upgrade.sh

# 4. 配置用户组
# 参考 docs/user-group-design.md

# 5. 重启服务
docker compose restart
```

### 全新安装

```bash
# 克隆项目
git clone https://github.com/ZYHUO/xboard-go.git
cd xboard-go

# 运行安装脚本
bash local-install.sh
```

详细说明请查看：
- [升级指南](docs/upgrade-guide.md)
- [安装指南](docs/local-installation.md)
- [快速开始](QUICK_INSTALL.md)
