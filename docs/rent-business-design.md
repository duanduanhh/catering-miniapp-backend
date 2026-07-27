# 招租业务技术方案

> 版本：v1.0
> 日期：2026-07-08
> 状态：待评审 → 待开发

## 1. 背景与目标

在现有"招聘/求职"小程序后端上，新增**招租（找店铺）**业务域。核心差异：

- 招租**不设发布上限**，但**发布需付费**（一条一次性收费）
- 招租**不支持置顶**，支持付费刷新
- 找店铺页面无区域筛选、无置顶区
- 支持"店铺面积区间"和"转让费"筛选
- 列表/详情/收藏/联系凭证/举报/反馈/关闭/刷新等能力**复用现有链路**

`biz_type = 3` 代表招租，代码里已预埋常量 [api/v1/job.go:237](../api/v1/job.go#L237) `BizTypeRent = 3`。

## 2. 关键决策（已确认）

| # | 决策项 | 结论 |
|---|--------|------|
| 1 | 转让费存储 | `transfer_fee_type`(0=无/1=固定/2=面议) + `transfer_fee_amount`(int) 两列 |
| 2 | 转让说明字段 | 新列 `rent_detail.transfer_desc`，不复用 `job.description` |
| 3 | 联系人/联系电话 | 招租表单包含，存 `job.contact` / `job.contact_person_name` |
| 4 | 付费发布流程 | 预建 `job.status=5(待支付)` + `rent_detail` + `order`，回调激活为 active |
| 5 | 待支付订单展示 | "我的发布"列表**显示**，标为"待支付"，支持继续支付 |
| 6 | 发布价格来源 | `config/*.yml` 配置，不做后台可视化管理 |
| 7 | 超时清理 | 24h 未支付的 pending job 由定时任务清理 |
| 8 | 招租置顶 | **不支持**。`/jobs/top` 收到 biz_type=3 返回错误 |
| 9 | 招租付费刷新 | **支持**。招租付费点：发布 + 联系凭证 + 付费刷新 |
| 10 | 面积区间边界 | 下闭上开 `[min, max)` |
| 11 | `/home/top?type=3` | 返回空数组（招租无置顶）；前端约定不调 |
| 12 | 关闭原因 | 沿用 [handler/job.go:478-485](../internal/handler/job.go#L478-L485) 已写好的 5 条 |

## 3. 数据模型

### 3.1 复用 `job` 表（不改结构，仅扩状态枚举）

| 字段 | 招租时含义 |
|------|-----------|
| `biz_type = 3` | 业务类型标识 |
| `contact` / `contact_person_name` | 联系电话 / 联系人 |
| `address` / `address_detail` / `longitude` / `latitude` | 店面地址 |
| `first/second/third/four_area_*` | 区域信息（发布时用） |
| `photo_urls` | CSV，图片 URL 列表 |
| `status` | 1=active / 2=用户关闭 / 3=禁用 / 4=删除 / **5=待支付（新增）** |
| `refresh_time` | 用于"推荐/最新"排序 |
| `close_reason` / `close_time` | 关闭原因、关闭时间 |
| `create_at` / `update_at` | 时间戳 |

**招租时留空的 job 字段**：`positions` / `company_name` / `salary_min/max` / `basic_protection` / `salary_benefits` / `attendance_leave` / `work_content` / `recruit_num` / `enterprise_id` / `top_start_time` / `top_end_time` / `paid_refresh_time`。

### 3.2 新增 `rent_detail` 表

一对一扩展表，`job_id` 做主键关联 `job.id`。

```sql
CREATE TABLE rent_detail (
  job_id              BIGINT      NOT NULL PRIMARY KEY,
  monthly_rent        INT         NOT NULL DEFAULT 0 COMMENT '月租金（元），0~999999',
  area_size           INT         NOT NULL DEFAULT 0 COMMENT '店面面积（㎡），0~999999',
  transfer_fee_type   TINYINT     NOT NULL DEFAULT 0 COMMENT '0=无 1=固定金额 2=面议',
  transfer_fee_amount INT         NOT NULL DEFAULT 0 COMMENT '转让费金额（元）',
  transfer_desc       VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '转让说明',
  create_at           DATETIME    NOT NULL,
  update_at           DATETIME    NOT NULL,
  KEY idx_area (area_size),
  KEY idx_transfer (transfer_fee_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**索引说明**：`area_size` 和 `transfer_fee_type` 用于列表筛选场景 JOIN 后的 WHERE 过滤。

### 3.3 状态枚举扩展

[internal/model/job.go](../internal/model/job.go)：
```go
const (
    JobStatusActive        JobStatus = 1
    JobStatusUserClosed    JobStatus = 2
    JobStatusAdminDisabled JobStatus = 3
    JobStatusDeleted       JobStatus = 4
    JobStatusPendingPay    JobStatus = 5 // 新增：招租待支付
)
```

所有对外列表（`/jobs/list`、`/home/*`、`/collect/*`）默认过滤 `status = 1`，天然屏蔽 pending。"我的发布"列表根据 status 展示"待支付"标记。

### 3.4 付费产品类型扩展

[internal/service/order.go](../internal/service/order.go) 或 `order_item.product_type`：
```
1 = 岗位置顶
2 = 联系凭证
3 = 付费刷新
4 = 发布招租  ← 新增
```

## 4. 付费发布流程

```
① 前端填表 → POST /jobs/rent/pre_publish
       │
       ▼
② 后端事务（tm.Transaction）:
     - INSERT job (biz_type=3, status=5 待支付, refresh_time=NULL)
     - INSERT rent_detail
     - INSERT order (product_type=4, ref_id=job_id, amount=config.rent.publish_price)
     - 调 payService.BuildPayParams
     - 返回 { job_id, order_id, pay_params }
       │
       ▼
③ 前端 wx.requestPayment(pay_params)
       │
       ├── 用户点"我再想想" / 支付失败 → job 仍为 status=5，用户可继续支付
       │
       └── 支付成功
              │
              ▼
④ 微信回调 POST /wechat/pay/notify
     HandlePayNotify → switch product_type:
       case 4:  // 发布招租
         - UPDATE job SET status=1, refresh_time=NOW() WHERE id=ref_id AND status=5
         - UPDATE order SET status=paid
       │
       ▼
⑤ 前端支付回调成功 → 跳转 /jobs/detail?id=<job_id>
```

**幂等性**：回调 case 4 的 UPDATE 带 `status=5` 条件，重复回调不会重复激活。

**超时清理**：新增 `internal/task/rent_pending_cleanup.go` 定时任务，每小时扫描 `status=5 AND create_at < NOW() - INTERVAL 24 HOUR` 的 job，软删除（status=4）并同步关闭其关联 order。

## 5. API 设计

### 5.1 新增：预发布招租（付费入口）

**`POST /jobs/rent/pre_publish`** — 需 StrictAuth

Request：
```json
{
  "contact_person_name": "张三",
  "contact": "13800000000",
  "monthly_rent": 5000,
  "area_size": 80,
  "address": "北京市朝阳区xxx",
  "address_detail": "1号楼",
  "longitude": 116.4,
  "latitude": 39.9,
  "first_area_id": 1,
  "first_area_des": "北京",
  "second_area_id": 1,
  "second_area_des": "朝阳区",
  "third_area_id": 0,
  "third_area_des": "",
  "four_area_id": 0,
  "four_area_des": "",
  "transfer_fee_type": 1,
  "transfer_fee_amount": 50000,
  "transfer_desc": "旺铺人流量大...",
  "photo_urls": ["https://...", "..."]
}
```

Response：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "job_id": 12345,
    "order_id": 67890,
    "pay_params": {
      "timeStamp": "...",
      "nonceStr": "...",
      "package": "...",
      "signType": "RSA",
      "paySign": "..."
    }
  }
}
```

校验规则：
- `monthly_rent`: `required,min=1,max=999999`
- `area_size`: `required,min=1,max=999999`
- `transfer_fee_type`: `required,oneof=0 1 2`
- `transfer_fee_amount`: 当 `transfer_fee_type=1` 时必须 `> 0`；其他情况忽略
- `transfer_desc`: `required,max=1000`
- `photo_urls`: 至少 1 张，最多 4 张
- `address` / `longitude` / `latitude` / `contact` / `contact_person_name`: 必填

### 5.2 复用现有接口（新增 biz_type=3 支持）

| 接口 | 说明 |
|------|------|
| `POST /jobs/list` | filter.biz_type=3；filter 新增 `area_size_range` / `transfer_fee_flag`；返回项含 `rent_detail` |
| `POST /jobs/info` | biz_type=3 时补 `rent_detail` |
| `POST /jobs/update` | biz_type=3 时同步更新 `rent_detail`（Patch 语义） |
| `POST /jobs/my` | 支持透传 biz_type=3；返回项含 status（含 5=待支付） |
| `POST /jobs/close` / `reopen` / `delete` / `refresh` | 透传，无改动 |
| `POST /jobs/create` | **拒绝 biz_type=3**（招租必须走 pre_publish） |
| `POST /jobs/top` | **拒绝 biz_type=3** |
| `POST /home/top`,`/home/feed` | type=3 支持；top 返回空数组 |
| `GET /close_reasons?type=3` | 已实现 |
| `POST /contact_feedback/*` | 已支持 |
| `POST /collect` / `cancel_collect` | 已支持 CollectTypeRent=3 |
| `POST /contact_voucher/*` | 无改动 |
| `POST /report/*` | 已支持 |

### 5.3 筛选枚举

**面积区间**（下闭上开 `[min, max)`）：
```go
const (
    AreaSizeRangeUnder15   = 1 // [0, 15)
    AreaSizeRange15To30    = 2 // [15, 30)
    AreaSizeRange30To50    = 3 // [30, 50)
    AreaSizeRange50To100   = 4 // [50, 100)
    AreaSizeRange100To200  = 5 // [100, 200)
    AreaSizeRangeOver200   = 6 // [200, +∞)
)
```

**转让费**：
```go
const (
    TransferFeeFlagHas = 1 // 有转让费 (transfer_fee_type > 0)
    TransferFeeFlagNo  = 2 // 无转让费 (transfer_fee_type = 0)
)
```

### 5.4 DTO 变更

**`api/v1/rent.go`**（新增文件）：
```go
type RentDetailDTO struct {
    MonthlyRent       int    `json:"monthly_rent"`
    AreaSize          int    `json:"area_size"`
    TransferFeeType   int    `json:"transfer_fee_type"`
    TransferFeeAmount int    `json:"transfer_fee_amount"`
    TransferDesc      string `json:"transfer_desc"`
}

type RentPrePublishRequest struct { /* 见 5.1 */ }
type RentPrePublishResponseData struct {
    JobID     int64                `json:"job_id"`
    OrderID   int64                `json:"order_id"`
    PayParams payservice.PayParams `json:"pay_params"`
}
```

**`api/v1/job.go`** 修改：
- `JobFilter` 增加 `AreaSizeRange int` / `TransferFeeFlag int`
- `JobListItem` 增加 `RentDetail *RentDetailDTO`
- `JobMyItem` 增加 `RentDetail *RentDetailDTO`（可选，看前端展示需要）

## 6. 分层实现细节

### 6.1 Model 层

**新增** [internal/model/rent_detail.go](../internal/model/rent_detail.go)：
```go
type TransferFeeType int
const (
    TransferFeeNone       TransferFeeType = 0
    TransferFeeFixed      TransferFeeType = 1
    TransferFeeNegotiable TransferFeeType = 2
)

type RentDetail struct {
    JobID             int64           `gorm:"primaryKey;column:job_id"`
    MonthlyRent       int             `gorm:"column:monthly_rent"`
    AreaSize          int             `gorm:"column:area_size"`
    TransferFeeType   TransferFeeType `gorm:"column:transfer_fee_type"`
    TransferFeeAmount int             `gorm:"column:transfer_fee_amount"`
    TransferDesc      string          `gorm:"column:transfer_desc"`
    CreateAt          time.Time       `gorm:"column:create_at"`
    UpdateAt          time.Time       `gorm:"column:update_at"`
}
func (RentDetail) TableName() string { return "rent_detail" }
```

**修改** [internal/model/job.go](../internal/model/job.go)：追加 `JobStatusPendingPay = 5`。

**修改** [internal/server/migration.go](../internal/server/migration.go)：`AutoMigrate` 追加 `&model.RentDetail{}`。

### 6.2 Repository 层

**新增** [internal/repository/rent_detail.go](../internal/repository/rent_detail.go)：
```go
type RentDetailRepository interface {
    Create(ctx context.Context, d *model.RentDetail) error
    Update(ctx context.Context, d *model.RentDetail) error
    GetByJobID(ctx context.Context, jobID int64) (*model.RentDetail, error)
    GetByJobIDs(ctx context.Context, ids []int64) (map[int64]*model.RentDetail, error)
    DeleteByJobID(ctx context.Context, jobID int64) error
}
```

**修改** [internal/repository/job.go](../internal/repository/job.go)：
- `ListQuery` 增加 `AreaSizeRange int` / `TransferFeeFlag int`
- 当 `BizType==3` 且以上任一非零时，`LEFT JOIN rent_detail rd ON rd.job_id = job.id`，追加 WHERE：
  - `AreaSizeRange`：翻译为 `rd.area_size >= ? AND rd.area_size < ?`（>200 只保留下界）
  - `TransferFeeFlag=1`：`rd.transfer_fee_type > 0`
  - `TransferFeeFlag=2`：`rd.transfer_fee_type = 0`
- 默认列表过滤增加 `job.status != 5`（避免待支付露出）
- "我的发布"查询保留 status=5

### 6.3 Service 层

**修改** [internal/service/job.go](../internal/service/job.go)：

新增字段：
```go
type jobService struct {
    *Service
    jobRepository                   repository.JobRepository
    userRepository                  repository.UserRepository
    contactVoucherHistoryRepository repository.ContactVoucherHistoryRepository
    rentDetailRepository            repository.RentDetailRepository   // 新增
    orderRepository                 repository.OrderRepository        // 新增（若尚无）
    payService                      PayService                        // 新增（若尚无）
    conf                            *viper.Viper                      // 读 rent.publish_price
}
```

新增方法：
```go
func (s *jobService) PrePublishRent(ctx context.Context, userID int64, input RentPrePublishInput) (*RentPrePublishResult, error) {
    // 1. 校验 input
    // 2. tm.Transaction:
    //    - 生成 sonyflake id
    //    - 创建 job (status=5, biz_type=3, refresh_time=nil)
    //    - 创建 rent_detail
    //    - 创建 order (product_type=4, ref_id=job.ID, amount=conf.GetFloat64("rent.publish_price"))
    //    - 调 payService.BuildPayParams
    //    - 返回 { JobID, OrderID, PayParams }
}
```

`Create()`：入口拒绝 `biz_type=3`（返回业务错误）
`Top()`：入口拒绝 `biz_type=3`
`Update()`：如果目标 job 是 biz_type=3，同步更新 rent_detail（Patch 语义）
`GetInfo()`：如果 job.biz_type=3，补取 rent_detail 拼到响应
`List()`：批量收集结果里 biz_type=3 的 job_id，`rentDetailRepository.GetByJobIDs` 补取，装配到 `JobListItem.RentDetail`
`My()`：同上；`status=5` 的记录允许展示

**修改** [internal/service/order.go](../internal/service/order.go)：
```go
// HandlePayNotify switch
case ProductTypePublishRent: // = 4
    err := s.tm.Transaction(ctx, func(ctx context.Context) error {
        // 只在 status=5 时激活，保证幂等
        if err := s.jobRepository.ActivatePendingRent(ctx, orderItem.RefID); err != nil {
            return err
        }
        return s.orderRepository.MarkPaid(ctx, order.ID)
    })
```

**新增定时任务** [internal/task/rent_pending_cleanup.go](../internal/task/rent_pending_cleanup.go)：
```go
// 每小时执行：清理超过 24h 的 pending 招租
// UPDATE job SET status=4 WHERE biz_type=3 AND status=5 AND create_at < NOW() - INTERVAL 24 HOUR
// 同步 order 表关闭对应 order
```
挂到 `jobSet`（如现有 task 结构），依赖注入到 `serverSet`。

### 6.4 Handler 层

**修改** [internal/handler/job.go](../internal/handler/job.go)：

- 新增 `PrePublishRent(ctx *gin.Context)` — 参照现有 `Create` 结构
- `Create` 校验分支追加 biz_type=3 → 返回 400 "rent must use /jobs/rent/pre_publish"
- `Top` 追加 biz_type=3 → 返回 400
- Swagger `@Description` 批量更新为 "1=招聘 2=求职 3=招租"

### 6.5 Router 层

**修改** [internal/router/job.go](../internal/router/job.go)：
```go
strictAuth.POST("/jobs/rent/pre_publish", deps.JobHandler.PrePublishRent)
```

### 6.6 Wire 依赖

**修改** [cmd/server/wire/wire.go](../cmd/server/wire/wire.go)：
- `repositorySet` 追加 `repository.NewRentDetailRepository`
- `jobService` 构造函数签名变化后 wire 会自动重新生成

执行：
```bash
cd cmd/server/wire && wire
```

### 6.7 配置

**修改** [config/local.yml](../config/local.yml) 及其他环境配置：
```yaml
rent:
  publish_price: 9.9   # 元
```

### 6.8 Admin 后台

**修改** [internal/service/admin_job.go](../internal/service/admin_job.go) / [api/v1/admin_job.go](../api/v1/admin_job.go)：
- 列表接口增加 biz_type=3 的筛选和展示
- 详情返回 `rent_detail`
- 举报/反馈列表 UI 增加 "招租" 筛选项

管理后台前端（[../catering-miniapp-admin-frontend/](../../catering-miniapp-admin-frontend/)）需要单独适配。

## 7. 改动清单

### 新增文件
- [internal/model/rent_detail.go](../internal/model/rent_detail.go)
- [internal/repository/rent_detail.go](../internal/repository/rent_detail.go)
- [api/v1/rent.go](../api/v1/rent.go)
- [internal/task/rent_pending_cleanup.go](../internal/task/rent_pending_cleanup.go)

### 修改文件
- [internal/model/job.go](../internal/model/job.go) — 新增 `JobStatusPendingPay`
- [internal/server/migration.go](../internal/server/migration.go) — 追加 `&model.RentDetail{}`
- [api/v1/job.go](../api/v1/job.go) — Filter 加字段、ListItem 加 `RentDetail`
- [internal/repository/job.go](../internal/repository/job.go) — ListQuery 加字段、JOIN rent_detail、默认过滤 pending
- [internal/service/job.go](../internal/service/job.go) — 新增 `PrePublishRent`、`Create/Top` 拒绝 biz_type=3、`List/GetInfo/Update` 补 rent_detail
- [internal/service/order.go](../internal/service/order.go) — `HandlePayNotify` 新增 case 4
- [internal/handler/job.go](../internal/handler/job.go) — 新增 `PrePublishRent`、Swagger 文案
- [internal/router/job.go](../internal/router/job.go) — 挂新路由
- [cmd/server/wire/wire.go](../cmd/server/wire/wire.go) — 注册 `NewRentDetailRepository`
- [config/local.yml](../config/local.yml) 及其他环境配置 — 新增 `rent.publish_price`
- Admin 相关：[api/v1/admin_job.go](../api/v1/admin_job.go)、[internal/service/admin_job.go](../internal/service/admin_job.go)、[internal/handler/admin_job.go](../internal/handler/admin_job.go)

### 不改动
- 收藏 [internal/service/collect.go](../internal/service/collect.go)（已支持 CollectTypeRent=3）
- 联系记录 [internal/service/contact_history.go](../internal/service/contact_history.go)
- 联系凭证 [internal/service/cost_history.go](../internal/service/cost_history.go)
- 反馈 [internal/service/contact_feedback.go](../internal/service/contact_feedback.go)
- 举报 [internal/service/report.go](../internal/service/report.go)
- 企业模块 [internal/service/enterprise.go](../internal/service/enterprise.go)

## 8. 实施步骤

按依赖顺序：

1. **数据层**
   - 建 [internal/model/rent_detail.go](../internal/model/rent_detail.go)
   - 追加 `JobStatusPendingPay` 到 [internal/model/job.go](../internal/model/job.go)
   - 追加 AutoMigrate
   - `go run ./cmd/migration -conf config/local.yml` 验证建表

2. **Repository 层**
   - 建 [internal/repository/rent_detail.go](../internal/repository/rent_detail.go)
   - 修改 [internal/repository/job.go](../internal/repository/job.go)：ListQuery、JOIN、pending 过滤、`ActivatePendingRent` 方法

3. **Service 层**
   - 修改 [internal/service/job.go](../internal/service/job.go)：注入 rent/order/pay 依赖，实现 `PrePublishRent`，改造 `Create/Top/Update/List/GetInfo/My`
   - 修改 [internal/service/order.go](../internal/service/order.go)：`HandlePayNotify` case 4
   - `make mock` 重新生成 mock

4. **Wire**
   - 修改 [cmd/server/wire/wire.go](../cmd/server/wire/wire.go)
   - `cd cmd/server/wire && wire` 重新生成

5. **API / Handler / Router**
   - 建 [api/v1/rent.go](../api/v1/rent.go)
   - 修改 [api/v1/job.go](../api/v1/job.go)
   - 修改 [internal/handler/job.go](../internal/handler/job.go)：`PrePublishRent`、Swagger
   - 修改 [internal/router/job.go](../internal/router/job.go)
   - `make swag` 重新生成文档

6. **配置**
   - `rent.publish_price` 追加到 [config/local.yml](../config/local.yml) 及各环境

7. **定时任务**
   - [internal/task/rent_pending_cleanup.go](../internal/task/rent_pending_cleanup.go)
   - Wire 注册

8. **Admin 后台**
   - 列表/详情增加 rent 支持

9. **测试**
   - handler 单测（`test/server/handler/`）
   - service 单测（`test/server/service/`）
   - repository 单测（`test/server/repository/`）
   - `make test`

10. **联调**
    - 微信 JSAPI 支付回调流程端到端验证
    - pending 清理任务本地手工触发验证

## 9. 风险与注意点

- **回调幂等**：`ActivatePendingRent` 必须带 `status=5` WHERE 条件，避免重复回调重置 refresh_time
- **图片校验**：`PrePublishRent` 里同样要跑 `validatePhotoURLs`，避免脏数据落库
- **pending 期间用户能否编辑？**：暂定**允许**用户在支付成功前重复调 pre_publish（每次创建新 pending 记录，旧的由 24h 清理），或复用同一条（需在接口里加 job_id 可选参数）。**建议第一版：每次新建，简单**
- **金额校验**：回调里校验 `order.amount == config.rent.publish_price` 一致，防篡改
- **列表 JOIN 性能**：`rent_detail` 上 `area_size` 和 `transfer_fee_type` 建索引；数据量增大后如成为瓶颈，考虑把这两列冗余到 job 表
- **status=5 数据泄漏**：所有对外查询必须显式 `status = 1`，禁止用 `status != 4`

## 10. 未来扩展点（不在本期）

- 招租的付费刷新
- 招租的付费置顶
- 招租发布价格阶梯（VIP 用户折扣、多次发布折扣）
- 招租的更多筛选维度（转让费金额区间、租金区间、装修状态等）
