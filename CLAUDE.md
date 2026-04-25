# 项目记忆：catering-miniapp-backend

## 项目简介
餐饮行业招聘小程序后端服务。提供岗位发布/筛选/置顶、联系凭证购买、微信登录与支付等功能。

## 技术栈
- **语言/框架**: Go 1.24 + Gin v1.11
- **ORM**: GORM v1.31 (MySQL)
- **依赖注入**: Google Wire v0.7
- **配置**: Viper，通过 `-conf` 指定，默认 `config/local.yml`
- **认证**: JWT (HS256)，放在 `token` header（非 Authorization Bearer）
- **ID 生成**: Sonyflake，订单号 base62 编码
- **微信支付**: wechatpay-apiv3/wechatpay-go (JSAPI)
- **对象存储**: Aliyun OSS SDK v3，懒初始化
- **日志**: Uber zap + lumberjack 文件轮转
- **API 文档**: swaggo/swag (Swagger)
- **脚手架**: Nunu (go-nunu/nunu-layout-advanced)

## 目录结构
```
api/v1/          # DTO + 错误码，HandleSuccess/HandleError 统一响应
cmd/server/      # 入口 main.go + wire/wire.go（Wire DI 注册）
cmd/task/        # cron 任务入口
cmd/migration/   # GORM AutoMigrate 入口
config/          # local/test/prod.yml + apiclient_key.pem（微信支付私钥）
deploy/          # Dockerfile（多阶段）+ docker-compose（MySQL:3380, Redis:6350）
docs/            # Swagger 生成文档
internal/
  handler/       # Gin 控制器
  service/       # 业务逻辑（interface + impl）
  repository/    # 数据访问（GORM interface + impl）
  model/         # GORM 实体
  router/        # 路由注册（每个域一文件）
  middleware/    # StrictAuth / NoStrictAuth / CORS / 请求日志
  server/        # HTTP/Job/Migration server 适配器
  job/           # 后台 Job（Kafka 占位，暂未实现）
  task/          # cron 任务
pkg/
  app/           # 应用生命周期（多 server 启停 + 信号处理）
  jwt/           # JWT 生成与解析
  wechatpay/     # 微信支付封装（prepay + 回调解密 + RSA 签名）
  sid/           # Sonyflake ID 生成器
  log/           # zap + context 传播
test/            # 单元/集成测试 + mockgen mocks
```

## 核心架构模式
- **4层架构**: handler → service → repository → model，禁止跨层调用
- **接口驱动**: service/repository 均为 Go interface，便于 mockgen 测试
- **Context 事务传播**: `Repository.Transaction(ctx, fn)` 将 tx 注入 context；子调用通过 `r.DB(ctx)` 透明获取
- **统一响应**: `{code, message, trace_id, data}`，成功 HTTP 200 + code:0
- **软删除**: Job 用 status 枚举（1活跃/2关闭/3禁用/4删除），ContactHistory 用 deleted flag

## 认证
- JWT 在 `token` header（小写），claims 含 `UserId`(string) + `Openid`(微信 openid)
- Token 默认不过期（ExpiresAt=0）
- `StrictAuth` 强制校验，`NoStrictAuth` 可选解析（允许匿名）

## 微信支付流程
1. 客户端调付款接口 → 后端创建 orders + order_item（status=pending）
2. 后端调 JSAPI prepay → 返回 RSA 签名参数给小程序
3. 小程序调 `wx.requestPayment()`
4. 微信异步回调 `POST /wechat/pay/notify` → AES-GCM 解密 → 校验金额 → 更新订单 → 执行业务逻辑

## API 路由概览

### 公开（无需认证）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /wechat/user/register | 小程序注册 |
| POST | /wechat/user/login | 小程序登录 |
| POST | /wechat/pay/notify | 微信支付回调 |
| POST | /jobs/list | 岗位列表（过滤+分页） |
| GET  | /positions/all | 所有职位分类 |
| GET  | /close_reasons | 岗位关闭原因枚举 |

### 需认证（token header）
| 路径前缀 | 主要功能 |
|----------|----------|
| /user/* | 用户信息、地理位置、订单历史、邀请记录 |
| /jobs/* | 岗位增删改、刷新、置顶（付费）、我的岗位 |
| /collect/* | 岗位收藏/取消/列表 |
| /contact_history/* | 联系记录（双向软删除） |
| /contact_voucher/* | 联系凭证购买/消耗/记录 |
| /feedback/* | 用户反馈提交/列表 |
| /img/upload | 图片上传到 OSS |

## 数据库表
`user`, `job`, `orders`, `order_item`, `collect`, `contact_history`, `contact_voucher_history`, `feedback`, `position_category`, `position_subcategory`

> 注意：时间字段用 `create_at`/`update_at`（非 GORM 默认 `created_at`/`updated_at`）；金额用自定义 `Decimal` 类型。

## 岗位排序模式（QueryType）
- `1`: 置顶优先（top_start/end_time 有效）+ refresh_time DESC
- `2`: 地理距离排序（经纬度欧氏距离）
- `3` / 默认: 最新优先（create_at DESC）

## 付费产品类型（order_item.product_type）
- `1`: 岗位置顶
- `2`: 联系凭证
- `3`: 付费刷新

## 常用 Make 命令
```bash
make bootstrap    # docker-compose up + migration + 启动 server
make build        # 编译当前平台 ./bin/server
make build-linux  # 交叉编译 linux/amd64（CGO_ENABLED=0）
make swag         # 生成 Swagger 文档
make mock         # 重新生成 mock 文件
make test         # 运行测试 + 覆盖率
make docker       # 构建并推送镜像到华为云 SWR
```

## 部署
- Docker 镜像: `swr.cn-north-1.myhuaweicloud.com/catering-cyxx/miniapp-backend`
- 容器端口: `8000`，时区: `Asia/Shanghai`
- 本地开发: `docker-compose` 提供 MySQL（3380）+ Redis（6350）
