# Weex 交易员带单监控系统

一个基于 Go 语言开发的轻量级交易员带单活动实时监控系统，帮助用户及时获取目标交易员的开仓和平仓动态。

## 🚀 特性

- **实时监控**: 定时监控交易员的开仓平仓动态
- **Web 管理界面**: 基于 Vue.js + Element Plus 的简洁管理界面
- **通知推送**: 支持 Webhook 通知推送
- **数据持久化**: 完整的订单历史记录和通知日志
- **容器化部署**: 支持 Docker 一键部署
- **高性能**: 基于 Gin 框架和 GORM，性能优异

## 🏗️ 技术架构

- **后端**: Go + Gin + GORM
- **数据库**: MySQL 8.0
- **前端**: Vue.js 3 + Element Plus
- **容器化**: Docker + Docker Compose
- **配置管理**: Viper

## 📋 项目结构

```
weex-watchdog/
├── docker-compose.yml          # Docker Compose配置
├── Dockerfile                  # Docker构建文件
├── go.mod                     # Go模块定义
├── main.go                    # 主程序入口
├── config/
│   └── config.yaml           # 配置文件
├── internal/
│   ├── api/
│   │   ├── handler/          # API处理器
│   │   ├── middleware/       # 中间件
│   │   └── router.go         # 路由配置
│   ├── model/                # 数据模型
│   ├── service/              # 业务逻辑
│   └── repository/           # 数据访问层
├── pkg/
│   ├── database/             # 数据库连接
│   ├── logger/               # 日志组件
│   └── notification/         # 通知服务
├── web/
│   ├── static/               # 静态文件
│   └── templates/            # 前端模板
└── scripts/
    └── init.sql             # 数据库初始化脚本
```

## 🛠️ 快速开始

### 环境要求

- Docker & Docker Compose
- Go 1.21+ (本地开发)

### 一键部署

1. 克隆项目

```bash
git clone <repository-url>
cd weex-watchdog
```

2. 启动服务

```bash
docker-compose up -d
```

3. 查看日志

```bash
docker-compose logs -f app
```

4. 访问管理界面

```
http://localhost:8080
```

### 本地开发

1. 安装依赖

```bash
go mod download
```

2. 启动数据库服务

```bash
docker-compose up -d mysql redis
```

3. 运行应用

```bash
go run main.go
```

## 📊 API 文档

### 交易员管理

- `GET /api/v1/traders` - 获取交易员列表
- `POST /api/v1/traders` - 添加交易员
- `PUT /api/v1/traders/:id` - 更新交易员信息
- `DELETE /api/v1/traders/:id` - 删除交易员
- `POST /api/v1/traders/:id/toggle` - 启用/禁用监控

### 订单管理

- `GET /api/v1/orders` - 获取订单历史
- `GET /api/v1/orders/statistics` - 获取统计数据

### 通知管理

- `GET /api/v1/notifications` - 获取通知记录
- `POST /api/v1/notifications/test` - 测试通知

### 健康检查

- `GET /health` - 健康检查接口

## ⚙️ 配置说明

配置文件位于 `config/config.yaml`：

```yaml
server:
  port: 8080 # 服务端口
  mode: debug # 运行模式

database:
  host: mysql # 数据库主机
  port: 3306 # 数据库端口
  user: weex_user # 数据库用户
  password: weex_password # 数据库密码
  dbname: weex_monitor # 数据库名

weex:
  api_url: "https://..." # Weex API地址
  timeout: 30s # API超时时间
  retry_times: 3 # 重试次数

monitor:
  default_interval: 30s # 默认监控间隔
  max_goroutines: 100 # 最大协程数

notification:
  webhook_url: "" # Webhook通知地址
  timeout: 10s # 通知超时时间
```

## 🔧 环境变量

支持通过环境变量覆盖配置：

- `WEEX_DATABASE_HOST` - 数据库主机
- `WEEX_DATABASE_USER` - 数据库用户
- `WEEX_DATABASE_PASSWORD` - 数据库密码
- `WEEX_DATABASE_DBNAME` - 数据库名
- `WEEX_NOTIFICATION_WEBHOOK_URL` - Webhook 地址

## 📝 使用说明

1. **添加交易员**: 在 Web 界面中添加需要监控的交易员 ID
2. **配置监控**: 设置监控间隔和启用状态
3. **查看订单**: 实时查看交易员的开仓平仓记录
4. **通知设置**: 配置 Webhook 地址接收通知推送

## 🔔 通知格式

新开仓通知：

```json
{
  "type": "NEW_ORDER",
  "message": "🆕 新开仓提醒\n交易员：xxx\n订单ID：xxx\n方向：LONG\n杠杆：10x\n数量：1000\n开仓价：50000\n时间：2024-01-01 12:00:00",
  "data": { ... }
}
```

平仓通知：

```json
{
  "type": "ORDER_CLOSED",
  "message": "❌ 平仓提醒\n交易员：xxx\n订单ID：xxx\n方向：LONG\n杠杆：10x\n数量：1000\n开仓价：50000\n平仓时间：2024-01-01 13:00:00",
  "data": { ... }
}
```

## 🐛 故障排除

### 常见问题

1. **数据库连接失败**

   - 检查数据库是否启动
   - 验证连接配置

2. **API 调用失败**

   - 检查网络连接
   - 验证 API 地址配置

3. **通知发送失败**
   - 检查 Webhook 地址配置
   - 验证网络连接

### 日志查看

```bash
# 查看应用日志
docker-compose logs -f app

# 查看数据库日志
docker-compose logs -f mysql

# 查看所有服务日志
docker-compose logs -f
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
