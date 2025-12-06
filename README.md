# XBoard Go

XBoard 的 Go 语言重写版本，后端节点使用 sing-box server。

## ✨ 特性

- 🚀 Go 语言重写，高性能低资源占用
- 🎨 Vue3 + TypeScript 现代化前端，马卡龙配色主题
- 📦 支持 sing-box、Clash、Surge、Quantumult X 等多种订阅格式
- 🔐 支持 AnyTLS、SS2022、VMess、VLESS、Trojan 等协议
- 📧 邮件通知系统
- 🤖 Telegram Bot 集成
- 🎫 完整工单系统
- 💾 兼容原版 XBoard 数据库，可平滑迁移

## 快速开始

### 使用 Docker Compose（推荐）

```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑 config.yaml 配置数据库等信息
docker-compose up -d
```

### 手动部署

```bash
# 后端
go mod tidy
go build -o xboard cmd/server/main.go
./xboard -config configs/config.yaml

# 前端
cd web
npm install
npm run build
```

## 文档

详细文档请参考 [docs/README.md](docs/README.md)

## License

MIT
