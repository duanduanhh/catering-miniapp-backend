# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目简介
餐饮行业招聘小程序后端服务。提供岗位发布/筛选/置顶、联系凭证购买、微信登录与支付、企业认证（营业执照 OCR）等功能。

## 技术栈
- **语言/框架**: Go 1.24 + Gin v1.11
- **ORM**: GORM v1.31 (MySQL)
- **依赖注入**: Google Wire v0.7
- **配置**: Viper，通过 `-conf` 指定，默认 `config/local.yml`
- **认证**: JWT (HS256)，放在 `token` header（非 Authorization Bearer）
- **ID 生成**: Sonyflake，订单号 base62 编码
- **微信支付**: wechatpay-apiv3/wechatpay-go (JSAPI)
- **对象存储**: Aliyun OSS SDK v3，懒初始化
- **OCR**: 阿里云 ocr-20191230 SDK，懒初始化，复用 OSS AK/SK
- **日志**: Uber zap + lumberjack 文件轮转
- **API 文档**: swaggo/swag (Swagger)
- **脚手架**: Nunu (go-nunu/nunu-layout-advanced)

## 常用命令

```bash
make bootstrap    # docker-compose up + migration + 启动 server
make build        # 编译当前平台 ./bin/server
make build-linux  # 交叉编译 linux/amd64（CGO_ENABLED=0）
make swag         # 生成 Swagger 文档（修改接口后必须运行）
make mock         # 重新生成 mock 文件（修改 service/repository interface 后运行）
make test         # 运行测试 + 覆盖率报告
make docker       # 构建并推送镜像到华为云 SWR

# Wire 重新生成（新增/修改依赖后必须运行）
cd cmd/server/wire && wire
```

## 核心架构

### 四层架构（严格分层，禁止跨层调用）
```
handler → service → repository → model
```
- **model** (`internal/model/`) — GORM 实体，只定义字段和 `TableName()`，不含逻辑
- **repository** (`internal/repository/`) — 只做 DB 操作，接口 + 实现分开，实现用 `r.DB(ctx)` 获取连接（支持事务传播）
- **service** (`internal/service/`) — 业务逻辑，依赖 repository interface，不直接操作 DB
- **handler** (`internal/handler/`) — 解析请求、调 service、返回响应，不含业务逻辑

禁止：handler 直接调 repository；service 调 handler；跨域 handler 相互调用。

### Wire 依赖注入
所有依赖通过 `cmd/server/wire/wire.go` 中的 `repositorySet`、`serviceSet`、`handlerSet`、`wechatPaySet` 注册，`wire_gen.go` 自动生成，**不要手动编辑 wire_gen.go**。

两个特殊 Provider（因为构造函数返回 error，不能直接放入 wire.NewSet）：
- `PayService` → `NewPayServiceProvider`
- `WechatPayClient` → `NewWechatPayClientProvider`（放在 `wechatPaySet`）

`EnterpriseService` 不依赖 repository，直接接受 `*viper.Viper` 和 `*log.Logger`，在 `serviceSet` 中注册。

### Context 事务传播
```go
// service 层发起事务
s.tm.Transaction(ctx, func(ctx context.Context) error {
    // 子调用透明获取 tx
    return s.xxxRepo.Create(ctx, ...)
})

// repository 层透明获取连接（普通或事务）
func (r *xxxRepository) Create(ctx context.Context, ...) error {
    return r.DB(ctx).Create(...).Error
}
```

### 统一响应格式
```json
{ "code": 0, "message": "ok", "trace_id": "...", "data": {} }
```
- 成功：`v1.HandleSuccess(ctx, data)`，data 为 nil 时传 `nil`（自动转为 `{}`）
- 失败：`v1.HandleError(ctx, httpCode, v1.ErrXxx, detail)`
- 错误码在 `api/v1/errors.go` 用 `newError(code, msg)` 声明；业务域错误码段：企业模块 2001-2004

### 认证
路由只有两种：
- **StrictAuth**：`r.Group("/").Use(middleware.StrictAuth(...))` — 未登录返回 401
- **noAuthRouter**：`r.Group("/")` 无中间件 — 匿名可访问

> `middleware.NoStrictAuth` 已定义但当前路由中**未使用**，不要用它注册路由。

Handler 内获取用户信息：
```go
userID := GetUserIdFromCtx(ctx)  // int64，未登录为 0
openid := GetOpenidFromCtx(ctx)  // string，未登录为 ""
```

## 数据库约定
- 时间字段：`create_at` / `update_at`（**非** GORM 默认的 `created_at`/`updated_at`）
- 金额：自定义 `model.Decimal` 类型，转分调用 `.ToCents()`，构造用 `model.NewDecimalFromFloat64()`
- 多值字段（图片、标签等）：存 CSV 字符串；读取用 `splitCSV(field)` → `[]string{}`（空时返回空数组）
- 软删除：Job 用 `status` 枚举（1活跃/2关闭/3禁用/4删除）；ContactHistory 用 `deleted` flag
- ID：`int64`，用 `s.sid.GenUint64()` 生成
- 新表需在 `internal/server/migration.go` 的 `AutoMigrate()` 调用中追加

## 新增业务域的完整步骤
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

## 微信支付流程
1. 客户端调付款接口 → 后端创建 `orders` + `order_item`（status=pending）
2. 后端调 `payService.BuildPayParams()` → JSAPI prepay → 返回签名参数
3. 小程序调 `wx.requestPayment()`
4. 微信回调 `POST /wechat/pay/notify` → AES-GCM 解密 → 校验金额 → 更新订单 → 执行业务逻辑

新增付费产品类型时，在 `order_item.product_type` 枚举和 `service/order.go` 的 `HandlePayNotify` switch 中同步添加。

付费产品类型：`1`=岗位置顶，`2`=联系凭证，`3`=付费刷新

## OCR 配置
`config/local.yml` 中 `ocr.access_key_id` 为空时自动 fallback 到 `oss.access_key_id`（同账号复用）。OCR 客户端懒初始化，mutex 保护，与 OSS 客户端模式一致。

## Handler 编写规范

### Swagger 注释（每个方法必须有）
```go
// MethodName godoc
// @Summary 一句话描述
// @Description 详细说明，枚举值含义
// @Tags 模块名
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.XxxRequest true "params"
// @Success 200 {object} v1.XxxResponse
// @Router /path [post]
```

### 错误处理模式
```go
if err != nil {
    h.logger.WithContext(ctx).Error("xyzService.Method error", zap.Error(err))
    v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
    return
}
```
service 错误映射：`ErrForbidden`→403，`ErrUserNotFound`→404，参数校验失败→400，其余→500

### Patch 语义（Update 接口）
数组字段用 `*string` 存拼接值，`nil`=不修改，`""`=清空：
```go
if req.PhotoURLs != nil {
    joined := strings.Join(req.PhotoURLs, ",")
    input.PhotoURLs = &joined
}
```

### HTTP 方法约定
- **POST**：写操作 + 复杂查询（带请求体）
- **GET**：简单无副作用查询
- **DELETE**：软删除操作

## 路由概览

### 公开（无需认证）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /wechat/user/register | 小程序注册 |
| POST | /wechat/user/login | 小程序登录 |
| POST | /wechat/pay/notify | 微信支付回调 |
| POST | /jobs/list | 岗位列表（过滤+分页） |
| POST | /home/top | 首页置顶区 |
| POST | /home/feed | 首页信息流 |
| GET  | /positions/all | 所有职位分类 |
| GET  | /close_reasons | 岗位关闭原因枚举 |
| GET  | /enterprise/detail | 企业详情（公开） |

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

## 部署
- Docker 镜像: `swr.cn-north-1.myhuaweicloud.com/catering-cyxx/miniapp-backend`
- 容器端口: `8000`，时区: `Asia/Shanghai`
- 本地开发: `docker-compose` 提供 MySQL（3380）+ Redis（6350）
