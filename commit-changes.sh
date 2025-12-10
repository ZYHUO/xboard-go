#!/bin/bash

# 提交所有更改到 GitHub
# 用法: bash commit-changes.sh

echo "准备提交更改到 GitHub..."

# 添加所有新文件和修改
git add .

# 显示将要提交的文件
echo ""
echo "将要提交的文件:"
git status --short

echo ""
read -p "确认提交这些更改? [Y/n]: " confirm

if [ "$confirm" = "n" ] || [ "$confirm" = "N" ]; then
    echo "已取消"
    exit 0
fi

# 提交
git commit -m "feat: 添加用户组系统和数据库迁移工具

新增功能:
- 用户组管理系统
- 套餐升级功能
- 数据库迁移工具 (migrate.sh)
- 数据库升级工具 (upgrade.sh, upgrade-mysql.sh)
- 本地安装脚本 (local-install.sh)
- 现有数据库安装脚本 (install-existing-db.sh)

改进:
- 简化用户组设计
- 优化订阅生成逻辑
- 增强 Makefile 命令

文档:
- 用户组设计说明
- 数据库迁移指南
- 数据库升级指南
- 本地安装指南
- 快速安装指南
- 更新日志

修复:
- 修复 Plan 模型缺少 UpgradeGroupID 字段
- 修复订阅生成类型不匹配
- 修复未使用的导入
"

echo ""
echo "提交完成！"
echo ""

# 推送到 GitHub
read -p "是否推送到 GitHub? [Y/n]: " push

if [ "$push" != "n" ] && [ "$push" != "N" ]; then
    echo "推送到 GitHub..."
    git push origin main
    echo ""
    echo "✓ 已推送到 GitHub"
else
    echo "跳过推送，稍后可以手动推送:"
    echo "  git push origin main"
fi

echo ""
echo "完成！"
