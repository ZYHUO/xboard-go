# sing-box 集成指南

本文档介绍如何将 sing-box 作为后端节点与 XBoard Go 面板对接。

## 概述

XBoard Go 通过 sing-box 的 SSMAPI (Server Side Management API) 与节点通信，实现：

- 用户同步：自动将面板用户同步到节点
- 流量统计：实时获取用户流量数据
- 节点状态：监控节点运行状态

## SSMAPI 接口

sing-box 的 SSMAPI 提供以下接口：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /server/v1/ | 获取服务器信息 |
| GET | /server/v1/users | 获取用户列表 |
| POST | /server/v1/users | 添加用户 |
| GET | /server/v1/users/{username} | 获取单个用户 |
| PUT | /server/v1/users/{username} | 更新用户 |
| DELETE | /server/v1/users/{username} | 删除用户 |
| GET | /server/v1/stats | 获取流量统计 |
| GET | /server/v1/stats?clear=true | 获取并清空流量统计 |

## 节点配置

### 1. 安装 sing-box

```bash
# 从源码编译（推荐，包含所有功能）
git clone https://github.com/sagernet/sing-box
cd sing-box
go build -tags "with_quic,with_grpc,with_wireguard,with_utls,with_reality_server,with_clash_api" ./cmd/sing-box
```

### 2. 配置文件示例

创建 `/etc/sing-box/config.json`：

```json
{
  "log": {
    "level": "info",
    "timestamp": true
  },
  "inbounds": [
    {
      "type": "shadowsocks",
      "tag": "ss-in",
      "listen": "::",
      "listen_port": 8388,
      "method": "2022-blake3-aes-128-gcm",
      "password": "YOUR_SERVER_KEY"
    },
    {
      "type": "vmess",
      "tag": "vmess-in",
      "listen": "::",
      "listen_port": 10086,
      "users": []
    },
    {
      "type": "trojan",
      "tag": "trojan-in",
      "listen": "::",
      "listen_port": 443,
      "users": [],
      "tls": {
        "enabled": true,
        "certificate_path": "/path/to/cert.pem",
        "key_path": "/path/to/key.pem"
      }
    }
  ],
  "outbounds": [
    {"type": "direct", "tag": "direct"}
  ],
  "services": [
    {
      "type": "ssmapi",
      "tag": "ssmapi",
      "listen": "::",
      "listen_port": 9000,
      "servers": {
        "/shadowsocks": "ss-in",
        "/vmess": "vmess-in",
        "/trojan": "trojan-in"
      },
      "cache_path": "/var/lib/sing-box/ssmapi.cache"
    }
  ]
}
```

### 3. 启动节点

```bash
./sing-box run -c /etc/sing-box/config.json
```

## 面板配置

### 1. 添加节点

在管理后台添加节点时，配置以下字段：

- **类型**: 选择对应的协议类型（shadowsocks/vmess/vless/trojan 等）
- **地址**: 节点服务器地址
- **端口**: 协议监听端口
- **SSMAPI URL**: `http://节点IP:9000/协议类型`

例如：
- Shadowsocks: `http://1.2.3.4:9000/shadowsocks`
- VMess: `http://1.2.3.4:9000/vmess`
- Trojan: `http://1.2.3.4:9000/trojan`

### 2. 协议设置

在节点的 `protocol_settings` JSON 中配置：

```json
{
  "ssmapi_url": "http://1.2.3.4:9000/shadowsocks",
  "ssmapi_token": "optional_bearer_token",
  "cipher": "2022-blake3-aes-128-gcm"
}
```

## 支持的协议

| 协议 | inbound 类型 | 说明 |
|------|-------------|------|
| Shadowsocks | shadowsocks | 支持 2022 新协议 |
| VMess | vmess | V2Ray VMess |
| VLESS | vless | 支持 Reality |
| Trojan | trojan | Trojan 协议 |
| Hysteria2 | hysteria2 | QUIC 协议 |
| TUIC | tuic | QUIC 协议 |
| AnyTLS | anytls | TLS 伪装 |

## 用户同步机制

1. 面板定期（默认 60 秒）从数据库获取可用用户
2. 对比节点上的用户列表
3. 添加新用户、删除过期用户、更新密码变更的用户
4. 用户名使用 UUID，密码根据协议类型生成

## 流量统计机制

1. 面板定期（默认 60 秒）调用 `/stats?clear=true` 接口
2. 获取每个用户的上传/下载流量
3. 应用节点倍率后更新用户流量
4. 记录到统计表用于报表

## 安全建议

### 1. 限制 SSMAPI 访问

使用防火墙限制 SSMAPI 端口只允许面板服务器访问：

```bash
# iptables
iptables -A INPUT -p tcp --dport 9000 -s 面板IP -j ACCEPT
iptables -A INPUT -p tcp --dport 9000 -j DROP

# ufw
ufw allow from 面板IP to any port 9000
```

### 2. 启用 TLS

在 SSMAPI 配置中启用 TLS：

```json
{
  "type": "ssmapi",
  "listen_port": 9000,
  "servers": {...},
  "tls": {
    "enabled": true,
    "certificate_path": "/path/to/cert.pem",
    "key_path": "/path/to/key.pem"
  }
}
```

### 3. 使用认证令牌

可以在面板配置中设置 Bearer Token，并在节点端验证。

## 故障排查

### 1. 检查 SSMAPI 是否正常

```bash
curl http://节点IP:9000/shadowsocks/server/v1/
# 应返回: {"server":"sing-box x.x.x","apiVersion":"v1"}
```

### 2. 检查用户列表

```bash
curl http://节点IP:9000/shadowsocks/server/v1/users
# 应返回: {"users":[...]}
```

### 3. 检查流量统计

```bash
curl http://节点IP:9000/shadowsocks/server/v1/stats
# 应返回流量统计数据
```

### 4. 查看日志

```bash
# sing-box 日志
journalctl -u sing-box -f

# 面板日志
tail -f /var/log/xboard/xboard.log
```

## 多协议配置示例

一个节点可以同时运行多个协议：

```json
{
  "inbounds": [
    {"type": "shadowsocks", "tag": "ss", "listen_port": 8388, ...},
    {"type": "vmess", "tag": "vmess", "listen_port": 10086, ...},
    {"type": "trojan", "tag": "trojan", "listen_port": 443, ...}
  ],
  "services": [
    {
      "type": "ssmapi",
      "listen_port": 9000,
      "servers": {
        "/ss": "ss",
        "/vmess": "vmess",
        "/trojan": "trojan"
      }
    }
  ]
}
```

在面板中分别添加三个节点，SSMAPI URL 分别为：
- `http://节点IP:9000/ss`
- `http://节点IP:9000/vmess`
- `http://节点IP:9000/trojan`
