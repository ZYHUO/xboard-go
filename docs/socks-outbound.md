# SOCKS 出口配置

## 功能说明

允许主机通过 SOCKS5 代理出站，而不是直接使用 VPS 的 IP 地址。这在以下场景非常有用：

1. **IP 隐藏**：隐藏 VPS 的真实 IP
2. **多出口**：不同主机使用不同的出口 IP
3. **链式代理**：通过其他代理服务器出站
4. **地区切换**：使用不同地区的出口 IP

## 配置方式

### 1. 通过 API 配置

**更新主机配置**：
```bash
PUT /api/v2/admin/host/:id
```

**请求示例**：
```json
{
  "socks_outbound": "socks5://user:pass@proxy.example.com:1080"
}
```

### 2. 通过前端界面配置

在主机管理界面，编辑主机时添加 SOCKS 出口配置。

## 配置格式

### 基本格式

```
socks5://host:port
```

**示例**：
```
socks5://proxy.example.com:1080
```

### 带认证

```
socks5://username:password@host:port
```

**示例**：
```
socks5://myuser:mypass@proxy.example.com:1080
```

### 支持的格式

- `socks5://host:port`
- `socks5://user:pass@host:port`
- `socks://host:port` (自动识别为 SOCKS5)

## 工作原理

### 默认行为（不配置 SOCKS 出口）

```
用户 → VPS (sing-box) → 直接出站 → 目标网站
                ↓
            使用 VPS IP
```

### 配置 SOCKS 出口后

```
用户 → VPS (sing-box) → SOCKS5 代理 → 目标网站
                ↓            ↓
            VPS IP      代理服务器 IP
                        (对外显示的 IP)
```

## sing-box 配置示例

### 不配置 SOCKS 出口

```json
{
  "outbounds": [
    {"type": "direct", "tag": "direct"},
    {"type": "block", "tag": "block"}
  ],
  "route": {
    "final": "direct"
  }
}
```

### 配置 SOCKS 出口

```json
{
  "outbounds": [
    {
      "type": "socks",
      "tag": "socks-out",
      "server": "proxy.example.com",
      "server_port": 1080,
      "username": "myuser",
      "password": "mypass",
      "version": "5"
    },
    {"type": "direct", "tag": "direct"},
    {"type": "block", "tag": "block"}
  ],
  "route": {
    "rules": [
      {"ip_is_private": true, "outbound": "block"}
    ],
    "final": "socks-out"  // 默认使用 SOCKS 出口
  }
}
```

## 使用场景

### 场景1：隐藏 VPS IP

**问题**：VPS IP 可能被目标网站识别和封禁

**解决**：
```json
{
  "socks_outbound": "socks5://proxy.example.com:1080"
}
```

所有流量通过代理服务器出站，目标网站只能看到代理服务器的 IP。

### 场景2：多出口负载均衡

**问题**：单个 VPS 需要支持多个出口 IP

**解决**：
- 主机1：`socks5://proxy1.example.com:1080`
- 主机2：`socks5://proxy2.example.com:1080`
- 主机3：`socks5://proxy3.example.com:1080`

不同主机使用不同的出口 IP，实现负载均衡。

### 场景3：链式代理

**问题**：需要通过多层代理访问目标

**解决**：
```
用户 → VPS (sing-box) → SOCKS5 代理1 → SOCKS5 代理2 → 目标
```

配置 VPS 使用 SOCKS5 代理1，代理1 再配置使用代理2。

### 场景4：地区切换

**问题**：需要使用特定地区的 IP

**解决**：
- 美国主机：`socks5://us-proxy.example.com:1080`
- 日本主机：`socks5://jp-proxy.example.com:1080`
- 香港主机：`socks5://hk-proxy.example.com:1080`

## API 文档

### 更新主机配置

**端点**：`PUT /api/v2/admin/host/:id`

**请求参数**：
```json
{
  "name": "主机名称（可选）",
  "socks_outbound": "SOCKS5 代理地址（可选）"
}
```

**响应示例**：
```json
{
  "data": {
    "id": 1,
    "name": "香港主机1",
    "token": "xxx",
    "ip": "1.2.3.4",
    "socks_outbound": "socks5://proxy.example.com:1080",
    "status": 1,
    "created_at": 1702345678,
    "updated_at": 1702345678
  }
}
```

### 获取主机列表

**端点**：`GET /api/v2/admin/hosts`

**响应示例**：
```json
{
  "data": [
    {
      "id": 1,
      "name": "香港主机1",
      "socks_outbound": "socks5://proxy.example.com:1080",
      "status": 1
    },
    {
      "id": 2,
      "name": "美国主机1",
      "socks_outbound": null,
      "status": 1
    }
  ]
}
```

## 前端界面

### 主机编辑表单

```jsx
<FormItem label="SOCKS 出口代理" name="socks_outbound">
  <Input 
    placeholder="socks5://user:pass@host:port"
    addonBefore="🌐"
  />
  <p className="text-xs text-gray-500 mt-1">
    可选。配置后，所有流量将通过此 SOCKS5 代理出站。
    <br />
    格式：socks5://[user:pass@]host:port
  </p>
</FormItem>
```

### 主机列表显示

```jsx
<Table>
  <Column title="主机名称" dataIndex="name" />
  <Column 
    title="出口方式" 
    render={(record) => {
      if (record.socks_outbound) {
        return (
          <Tag color="blue">
            🌐 SOCKS 代理
          </Tag>
        );
      }
      return (
        <Tag color="green">
          📡 直接出站
        </Tag>
      );
    }}
  />
</Table>
```

## 注意事项

### 1. 性能影响

使用 SOCKS 代理会增加一跳，可能影响速度：
- 延迟增加：+10-50ms（取决于代理服务器位置）
- 带宽限制：受代理服务器带宽限制

### 2. 代理服务器要求

- 必须支持 SOCKS5 协议
- 必须允许 TCP 连接
- 建议使用高带宽、低延迟的代理服务器

### 3. 安全性

- 使用认证（用户名/密码）保护代理服务器
- 不要使用公共代理服务器
- 定期更换代理服务器密码

### 4. 故障处理

如果代理服务器不可用：
- sing-box 会尝试连接失败
- 用户无法访问
- 建议配置监控和告警

**解决方案**：
1. 移除 SOCKS 出口配置，恢复直接出站
2. 更换可用的代理服务器
3. 配置多个主机，分散风险

## 测试

### 1. 测试代理连接

```bash
# 使用 curl 测试 SOCKS5 代理
curl -x socks5://user:pass@proxy.example.com:1080 https://api.ipify.org

# 应该返回代理服务器的 IP，而不是 VPS IP
```

### 2. 测试 sing-box 配置

```bash
# 1. 配置 SOCKS 出口
PUT /api/v2/admin/host/1
{
  "socks_outbound": "socks5://proxy.example.com:1080"
}

# 2. 获取配置
GET /api/v2/admin/host/1/config

# 3. 验证配置中包含 SOCKS outbound
# 应该看到：
# "outbounds": [
#   {
#     "type": "socks",
#     "tag": "socks-out",
#     "server": "proxy.example.com",
#     "server_port": 1080
#   }
# ]

# 4. Agent 会自动应用新配置
```

### 3. 验证出口 IP

```bash
# 连接到节点后，访问 IP 查询服务
curl https://api.ipify.org

# 应该返回代理服务器的 IP
```

## 数据库

### 表结构

```sql
ALTER TABLE `v2_host` 
ADD COLUMN `socks_outbound` TEXT NULL 
COMMENT 'SOCKS5 出口代理地址，格式：socks5://user:pass@host:port';
```

### 迁移文件

`migrations/006_add_socks_outbound_to_host.sql`

### 查询示例

```sql
-- 查看所有配置了 SOCKS 出口的主机
SELECT id, name, socks_outbound 
FROM v2_host 
WHERE socks_outbound IS NOT NULL;

-- 更新主机的 SOCKS 出口
UPDATE v2_host 
SET socks_outbound = 'socks5://proxy.example.com:1080' 
WHERE id = 1;

-- 移除 SOCKS 出口配置
UPDATE v2_host 
SET socks_outbound = NULL 
WHERE id = 1;
```

## 常见问题

### Q1: 配置后不生效？

**检查**：
1. Agent 是否在线
2. 配置是否正确
3. 代理服务器是否可访问

**解决**：
```bash
# 查看 Agent 日志
journalctl -u xboard-agent -f

# 测试代理连接
curl -x socks5://proxy.example.com:1080 https://google.com
```

### Q2: 速度变慢？

**原因**：代理服务器带宽或延迟问题

**解决**：
1. 更换更快的代理服务器
2. 使用地理位置更近的代理
3. 移除 SOCKS 配置，恢复直接出站

### Q3: 如何移除 SOCKS 配置？

**方法1**：设置为空字符串
```json
{
  "socks_outbound": ""
}
```

**方法2**：设置为 null
```json
{
  "socks_outbound": null
}
```

### Q4: 支持 HTTP 代理吗？

**当前**：只支持 SOCKS5

**未来**：可能支持 HTTP/HTTPS 代理

### Q5: 可以配置多个出口吗？

**当前**：每个主机只能配置一个 SOCKS 出口

**替代方案**：
- 创建多个主机，每个配置不同的出口
- 在代理服务器端实现负载均衡

## 相关文档

- [主机管理](local-installation.md)
- [节点配置](server-host-binding.md)
- [sing-box 文档](https://sing-box.sagernet.org/)
