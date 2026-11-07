# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目简介

餐饮行业招聘小程序后端服务。功能：岗位发布/筛选/置顶（付费）、联系券购买、微信登录与小程序虚拟支付、企业认证（营业执照 OCR）。

## 技术栈

- Go 1.24 + Gin v1.11，脚手架：Nunu (nunu-layout-advanced)
- ORM：GORM v1.31（MySQL，也支持 Postgres/SQLite）
- 依赖注入：Google Wire v0.7
- 配置：Viper；认证：JWT HS256，放在 `token` header（非 Authorization Bearer）
- ID：Sonyflake；虚拟支付：微信小程序 `wx.requestVirtualPayment`；OSS/OCR：阿里云 SDK（懒初始化）
- 日志：zap + lumberjack；API 文档：swaggo/swag
- 管理后台前端：同级独立项目 `../catering-miniapp-admin-frontend/`，Vue 3 + Vite + Element Plus

## 常用命令

```bash
# 工具安装（首次）
make init             # 安装 wire / mockgen / swag

# 本地开发
make bootstrap        # docker-compose up + 数据库迁移 + nunu run ./cmd/server

# 只启动依赖（MySQL:3380, Redis:6350）
cd deploy/docker-compose && docker compose up -d

# 单独运行 server
go run ./cmd/server -conf config/local.yml

# 数据库迁移
go run ./cmd/migration -conf config/local.yml

# 生成 Swagger（改接口后必须运行）
make swag             # 等价于: swag init -g cmd/server/main.go -o ./docs

# Wire 重新生成（新增/修改依赖后必须运行）
cd cmd/server/wire && wire

# 重新生成 mock（改 service/repository interface 后运行）
make mock

# 测试
make test             # 全量测试 + 覆盖率报告（coverage.html）
go test ./test/server/handler/... -v   # 单跑 handler 测试
go test ./test/server/service/... -v   # 单跑 service 测试
go test ./test/server/repository/... -v

# 构建
make build            # 当前平台 → ./bin/server
make build-linux      # linux/amd64，CGO_ENABLED=0

# 后端镜像（只包含 Go 服务 + swag + push 华为云 SWR；前端独立构建/部署）
make docker

# 管理后台前端开发（同级独立项目）
cd ../catering-miniapp-admin-frontend && npm run dev      # dev server，代理 /admin/ → localhost:8000
cd ../catering-miniapp-admin-frontend && npm run build    # 产出 dist/，由前端镜像使用
```

## 核心架构

### 四层分层（严格，禁止跨层调用）

```
handler → service → repository → model
```

- `internal/model/` — GORM 实体，只定义字段和 `TableName()`
- `internal/repository/` — 只做 DB 操作；interface + 实现；实现用 `r.DB(ctx)` 获取连接（透明事务）
- `internal/service/` — 业务逻辑，依赖 repository interface，不直接操作 DB
- `internal/handler/` — 解析请求、调 service、返回响应，不含业务逻辑

禁止：handler 直接调 repository；service 调 handler；跨域 handler 互调。

### Wire 依赖注入

`cmd/server/wire/wire.go` 中有 `repositorySet / serviceSet / handlerSet / jobSet / serverSet`。`wire_gen.go` 自动生成，**不要手动编辑**。

`EnterpriseService` 不依赖 repository，直接接受 `*viper.Viper` 和 `*log.Logger`。

### 事务传播

```go
// service 层发起事务
s.tm.Transaction(ctx, func(ctx context.Context) error {
    return s.xxxRepo.Create(ctx, ...)  // repository 透明获取 tx
})
```

### 统一响应

```json
{ "code": 0, "message": "ok", "trace_id": "...", "data": {} }
```

- 成功：`v1.HandleSuccess(ctx, data)`，data 为 nil 时传 `nil`（自动转为 `{}`）
- 失败：`v1.HandleError(ctx, httpCode, v1.ErrXxx, detail)`
- 错误码在 `api/v1/errors.go`；业务域错误码段：企业模块 2001-2004

### 认证

路由两种：`StrictAuth`（`middleware.StrictAuth`，缺 token 返回 401）和无中间件匿名路由。`NoStrictAuth` 已定义但**当前路由中未使用**，不要用它注册路由。

Handler 内获取用户信息：
```go
userID := GetUserIdFromCtx(ctx)  // int64，未登录为 0
openid := GetOpenidFromCtx(ctx)  // string，未登录为 ""
```

### 配置加载优先级

`APP_CONF` 环境变量 > `APP_ENV` 环境变量（`config/<env>.yml`）> `-conf` 启动参数（默认 `config/local.yml`）

## 数据库约定

- 时间字段：`create_at` / `update_at`（**非** GORM 默认的 `created_at`/`updated_at`）
- 金额：`model.Decimal`，转分用 `.ToCents()`，构造用 `model.NewDecimalFromFloat64()`
- 多值字段（图片、标签等）：存 CSV 字符串；读取用 `splitCSV(field)` → `[]string{}`（空时返回空数组）
- 软删除：Job 用 `status` 枚举（1活跃/2关闭/3禁用/4删除）；ContactHistory 用 `deleted` flag
- ID：`int64`，用 `s.sid.GenUint64()` 生成
- 新表需在 `internal/server/migration.go` 的 `AutoMigrate()` 中追加

## 新增业务域步骤

1. `internal/model/xxx.go` — 实体
2. `api/v1/xxx.go` — Request/Response DTO
3. `internal/repository/xxx.go` — interface + 实现
4. `internal/service/xxx.go` — interface + 实现
5. `internal/handler/xxx.go` — Handler struct + 方法（含 Swagger 注释）
6. `internal/router/xxx.go` — `InitXxxRouter` 函数
7. `internal/server/http.go` — 调用 `router.InitXxxRouter`
8. `internal/router/router.go` — 在 `RouterDeps` 中添加 `XxxHandler` 字段
9. `cmd/server/wire/wire.go` — 在对应 Set 中注册 Provider
10. 运行 `cd cmd/server/wire && wire`

## Handler 编写规范

每个方法必须有 Swagger 注释：
```go
// MethodName godoc
// @Summary 一句话描述
// @Tags 模块名
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.XxxRequest true "params"
// @Success 200 {object} v1.XxxResponse
// @Router /path [post]
```

错误处理：
```go
if err != nil {
    h.logger.WithContext(ctx).Error("xyzService.Method error", zap.Error(err))
    v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
    return
}
```

service 错误映射：`ErrForbidden`→403，`ErrUserNotFound`→404，参数校验失败→400，其余→500

Patch 语义（数组字段用 `*string`，`nil`=不修改，`""`=清空）：
```go
if req.PhotoURLs != nil {
    joined := strings.Join(req.PhotoURLs, ",")
    input.PhotoURLs = &joined
}
```

HTTP 方法约定：**POST** 写操作+复杂查询；**GET** 简单无副作用查询；**DELETE** 软删除。

## 小程序虚拟支付流程

1. 业务下单接口创建 `orders` + `order_item`（status=pending），返回 `order_no` 和 `amount_cents`
2. 小程序调用 `POST /payment/virtual/prepare` 获取签名
3. 小程序原样传入参数调用 `wx.requestVirtualPayment()`
4. 微信道具发货推送 `POST /wechat/virtual-payment/notify`，服务端验签、校验订单快照后幂等发放权益
5. 小程序调用 `POST /order/status` 确认最终状态

新增付费产品类型时，在 `order_item.product_type` 枚举和 `service/order.go` 的权益发放 switch 中同步添加。

付费产品类型：`1`=岗位置顶，`2`=联系券，`3`=付费刷新，`4`=招租发布

## 路由概览

### 公开（无需认证）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /wechat/user/register | 小程序注册 |
| POST | /wechat/user/login | 小程序登录 |
| GET/POST | /wechat/virtual-payment/notify | 虚拟支付推送验证与道具发货回调 |
| POST | /jobs/list | 岗位列表（过滤+分页）|
| POST | /home/top | 首页置顶区 |
| POST | /home/feed | 首页信息流 |
| GET  | /positions/all | 所有职位分类 |
| GET  | /close_reasons | 岗位关闭原因枚举 |
| GET  | /enterprise/detail | 企业详情（公开）|

### 需认证（token header）
| 路径前缀 | 主要功能 |
|----------|----------|
| GET /user/info, POST /user/* | 用户信息、地理位置、订单历史、邀请记录 |
| POST /jobs/* | 岗位增删改、刷新、置顶（付费）、我的岗位 |
| POST /collect/* | 岗位收藏/取消/列表 |
| DELETE /contact_history/* | 联系记录软删除 |
| POST /contact_voucher/* | 联系凭证购买/消耗/记录 |
| POST /feedback/* | 用户反馈提交/列表 |
| POST /img/upload | 图片上传到 OSS |
| POST /enterprise/ocr | 营业执照 OCR 识别 |

### 管理后台（/admin/*）
| 路径 | 说明 |
|------|------|
| POST /admin/jobs/list | 岗位列表（支持关键词/状态/类型筛选）|
| POST /admin/jobs/disable | 禁用岗位 |
| POST /admin/jobs/enable | 恢复岗位 |
| POST /admin/jobs/delete | 删除岗位 |

管理后台前端在同级独立项目 `../catering-miniapp-admin-frontend/`（Vue 3 SPA）。后端只提供 `/admin/*` API；前端独立 Docker/nginx 镜像服务 `/admin-ui/`，并代理 `/admin/` 到后端。

## 部署

- 后端 Docker 镜像：`swr.cn-north-1.myhuaweicloud.com/catering-cyxx/miniapp-backend`
- 后端容器端口：`8000`，时区：`Asia/Shanghai`
- 前端独立镜像在 `../catering-miniapp-admin-frontend/deploy/Dockerfile` 构建，nginx 暴露 `80`，服务 `/admin-ui/`
- 本地开发：`docker-compose` 提供 MySQL（3380）+ Redis（6350）
- 运行示例：`docker run -d -p 16789:8000 -e APP_ENV=test --restart=always <image>:<tag>`
