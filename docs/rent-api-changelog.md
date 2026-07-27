# 招租业务接口变更文档

> 版本：2026-07-10
> 范围：新增招租业务（biz_type=3）付费发布 + 找店铺列表 + 详情/我的岗位扩展
> 联调环境：test → `https://new-test.canyinxinxi.com`
> 认证：小程序侧接口均使用 `token` header（沿用现有 JWT，无变更）

---

## 一、总览

| 类型 | 数量 | 说明 |
|------|------|------|
| 新增接口 | 1 | `POST /jobs/rent/pre_publish` 招租付费预下单 |
| 变更接口 | 9 | 列表/更新/详情/我的岗位/首页置顶/首页信息流/发布/置顶/付费刷新/关闭原因/支付回调 增补招租相关字段与规则 |
| 新增枚举 | 3 | biz_type=3、product_type=4、job.status=5 |
| 新增字段（响应） | 1 | `rent_detail` 挂在 job 列表项与详情上 |
| 新增字段（请求） | 2 | JobFilter 追加 `area_size_range`、`transfer_fee_flag` |

**核心约束**
- 招租(biz_type=3) **必须**走 `/jobs/rent/pre_publish` 付费发布，通用 `/jobs/create` 收到 biz_type=3 直接返回 400
- 招租 **不支持** 置顶（`/jobs/top`），支持付费刷新（`/jobs/refresh/pay`）
- 招租 job 在支付成功前 `status=5(待支付)`，微信回调后由后端翻转为 `status=1(active)`，前端拉取列表/详情时若命中 status=5 视为"未生效"（列表本身已过滤，仅"我的岗位"会看到）

---

## 二、新增枚举

### 2.1 `biz_type` 内容类型（沿用扩展）

| 值 | 含义 | 备注 |
|----|------|------|
| 0 | 全部 | 仅筛选/查询用 |
| 1 | 招聘 | 无变更 |
| 2 | 求职 | 无变更 |
| **3** | **招租** | **新增** |

### 2.2 `job.status`（新增 5）

| 值 | 含义 |
|----|------|
| 1 | active（正常展示） |
| 2 | 用户关闭 |
| 3 | 管理员禁用 |
| 4 | 已删除 |
| **5** | **待支付（招租专用，未支付前不进入列表）** |

### 2.3 订单商品类型 `product_type`（新增 4）

| 值 | 含义 |
|----|------|
| 1 | 岗位置顶 |
| 2 | 联系凭证 |
| 3 | 付费刷新 |
| **4** | **发布招租** |

### 2.4 转让费类型 `transfer_fee_type`

| 值 | 含义 | 说明 |
|----|------|------|
| 0 | 无转让费 | `transfer_fee_amount` 忽略 |
| 1 | 固定金额 | `transfer_fee_amount` 必填且 > 0 |
| 2 | 面议 | `transfer_fee_amount` 忽略 |

### 2.5 店铺面积区间 `area_size_range`（列表筛选用）

| 值 | 区间(㎡) |
|----|---------|
| 0 | 不限 |
| 1 | <15 |
| 2 | [15, 30) |
| 3 | [30, 50) |
| 4 | [50, 100) |
| 5 | [100, 200) |
| 6 | ≥200 |

### 2.6 转让费筛选 `transfer_fee_flag`

| 值 | 含义 |
|----|------|
| 0 | 不限 |
| 1 | 有转让费（含固定 + 面议） |
| 2 | 无转让费 |

---

## 三、新增接口

### 3.1 招租付费预下单

**`POST /jobs/rent/pre_publish`**（需登录，token header）

一次调用完成：预建 job(status=5) + rent_detail + order，返回微信支付参数供小程序调 `wx.requestPayment()`。支付成功后后端回调将 job 翻转为 active。发布价格由服务端 `rent.publish_price` 配置（当前 9.9 元），**客户端不传价格**。

**Request Body**

```json
{
  "positions": "临街50㎡奶茶店转让",
  "address": "北京市朝阳区望京SOHO",
  "address_detail": "T3座 1层-108",
  "longitude": 116.478,
  "latitude": 39.997,
  "contact": "13812345678",
  "contact_person_name": "张老板",
  "description": "商圈成熟，人流稳定，可餐饮可零售",
  "photo_urls": ["https://.../a.jpg", "https://.../b.jpg"],
  "first_area_id": 110000,
  "first_area_des": "北京市",
  "second_area_id": 110105,
  "second_area_des": "朝阳区",
  "third_area_id": 0,
  "third_area_des": "",
  "four_area_id": 0,
  "four_area_des": "",
  "monthly_rent": 15000,
  "area_size": 50,
  "transfer_fee_type": 1,
  "transfer_fee_amount": 80000,
  "transfer_desc": "含全套设备"
}
```

**字段约束**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| positions | string | ✅ | 招租标题 |
| address | string | ✅ | 详细地址 |
| longitude/latitude | float | ✅ | 经纬度，不能为 0 |
| contact | string | ✅ | 联系电话 |
| contact_person_name | string | ✅ | 联系人 |
| photo_urls | []string | ✅ | 至少 1 张，最多 4 张 |
| first_area_des/second_area_des | string | ✅ | 一级/二级区划 |
| monthly_rent | int | ✅ | 月租(元)，>0 |
| area_size | int | ✅ | 店铺面积(㎡)，>0 |
| transfer_fee_type | int | ✅ | 0/1/2 |
| transfer_fee_amount | int | 当 type=1 | 转让费(元)，type=1 时必须 >0 |
| transfer_desc | string | 否 | 转让说明 |
| description | string | 否 | 补充描述 |

**Response**

```json
{
  "code": 0,
  "message": "ok",
  "trace_id": "...",
  "data": {
    "job_id": 1234567890,
    "order_id": 987654321,
    "order_no": "RENT202607081205129999",
    "amount": 9.9,
    "pay_params": {
      "timeStamp": "1720422312",
      "nonceStr": "abcd1234",
      "package": "prepay_id=wx...",
      "signType": "RSA",
      "paySign": "..."
    }
  }
}
```

**错误码**

| HTTP | code | 场景 |
|------|------|------|
| 400 | 400 | 参数缺失/photo_urls 数量非法/transfer_fee_type=1 但金额<=0/area_size<=0/monthly_rent<=0 |
| 401 | 401 | 未登录或 openid 缺失 |
| 500 | 500 | 服务端错误 |

---

## 四、变更接口

### 4.1 `POST /jobs/list` 岗位列表（含找店铺）

**变更点：** `filter` 增加两个招租专属筛选字段，仅当 `filter.biz_type=3` 时生效；响应项每条 job 可能带 `rent_detail`。

**Request Body（新增字段加粗）**

```json
{
  "request_id": "abc",
  "query_type": 1,
  "filter": {
    "biz_type": 3,
    "positions": "",
    "first_area_id": 110000,
    "second_area_id": 110105,
    "salary_min": 0,
    "salary_max": 0,
    "basic_protection": [],
    "salary_benefits": [],
    "attendance_leave": [],
    "longitude": 116.47,
    "latitude": 39.99,
    "area_size_range": 3,
    "transfer_fee_flag": 1
  },
  "page_num": 1,
  "page_size": 10
}
```

| 新增字段 | 类型 | 说明 |
|----------|------|------|
| filter.area_size_range | int | 见 2.5 枚举；仅 biz_type=3 生效 |
| filter.transfer_fee_flag | int | 见 2.6 枚举；仅 biz_type=3 生效 |

**Response（新增字段加粗）**

```json
{
  "code": 0,
  "data": {
    "total": 42,
    "jobs": [
      {
        "id": 123,
        "biz_type": 3,
        "positions": "临街50㎡奶茶店转让",
        "address": "...",
        "photo_urls": ["..."],
        "status": 1,
        "create_at": "2026-07-08 12:05:12.000",
        "rent_detail": {
          "monthly_rent": 15000,
          "area_size": 50,
          "transfer_fee_type": 1,
          "transfer_fee_amount": 80000,
          "transfer_desc": "含全套设备"
        }
      }
    ]
  }
}
```

- 非招租条目（biz_type=1/2）**不返回** `rent_detail` 字段（`omitempty`）
- 找店铺页面三个 Tab 建议：
  - "推荐" → `query_type=1`
  - "附近" → `query_type=2`（须传 longitude/latitude）
  - "最新" → `query_type=3`
- 找店铺页 **不再展示** 首页置顶区，直接拉 `/jobs/list` 即可

### 4.2 `POST /jobs/update` 修改岗位信息

**变更点：** 招租岗位（`biz_type=3`）可在通用更新接口里传 `rent_detail`，用于更新月租、面积、转让费信息。普通招聘/求职岗位传 `rent_detail` 会返回 400。

**Request Body（招租扩展示例）**

```json
{
  "id": 1234567890,
  "positions": "临街50㎡奶茶店转让",
  "address": "北京市朝阳区望京SOHO",
  "rent_detail": {
    "monthly_rent": 15000,
    "area_size": 50,
    "transfer_fee_type": 1,
    "transfer_fee_amount": 80000,
    "transfer_desc": "含全套设备"
  }
}
```

**字段约束**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| rent_detail.monthly_rent | int | 传 rent_detail 时必填 | 月租(元)，>0 |
| rent_detail.area_size | int | 传 rent_detail 时必填 | 店铺面积(㎡)，>0 |
| rent_detail.transfer_fee_type | int | 传 rent_detail 时必填 | 0=无转让费，1=固定金额，2=面议 |
| rent_detail.transfer_fee_amount | int | type=1 时必填 | 固定转让费金额(元)，type=0/2 时后端置 0 |
| rent_detail.transfer_desc | string | 否 | 转让说明 |

### 4.3 `POST /jobs/info` 岗位详情

**变更点：** 响应体新增 `rent_detail`（招租时）；status=5 的招租条目仍可通过详情查看（用户自己"我的-待支付"入口点进去）。

响应结构同 4.1 单条 `JobListItem`，字段一致。

### 4.4 `POST /jobs/my` 我发布的岗位

**变更点：** `JobMyItem` 新增 `rent_detail` 字段（招租条目）。**招租条目会包含 `status=5(待支付)`**，前端需在"我的岗位"列表根据 status 区分展示：

| status | 展示 | 操作 |
|--------|------|------|
| 5 | 【待支付】灰色卡片 | 显示"继续支付"或"取消" |
| 1 | 正常 | 展示、关闭 |
| 2 | 已关闭 | 重开 |
| 3 | 已禁用 | 联系客服 |

> 注：目前不提供"继续支付"接口（未支付 job 30 分钟后自动清理）；如需，请提前提需求。

### 4.5 `POST /home/top` 首页置顶区、`POST /home/feed` 首页信息流

**变更点：** 响应项支持 `rent_detail`（招租条目）。传参不变。

**招租不参与置顶** → `/home/top` 里不会出现 biz_type=3 的条目（除非未来产品放开）；`/home/feed` 中招租按刷新/发布时间正常混排（如需只看招租，请用 `/jobs/list` 传 `filter.biz_type=3`）。

### 4.6 `POST /jobs/create` 通用发布

**变更点：** biz_type=3 直接返回 400，业务错误信息：`rent must be published via /jobs/rent/pre_publish`。前端招租入口不要走此接口。

### 4.7 `POST /jobs/top` 岗位置顶、`POST /jobs/refresh/pay` 付费刷新

**变更点：** 当 `job_id` 对应的 job 是招租（biz_type=3）时，置顶接口返回 400：`rent does not support top`。前端招租详情/我的岗位卡片 **不要** 展示置顶按钮，但应展示付费刷新按钮。

### 4.8 `GET /close_reasons?type=3` 关闭原因（招租）

**已存在，无 API 结构变更**。type=3 返回招租对应关闭原因：

```json
{
  "code": 0,
  "data": {
    "type": 3,
    "reasons": ["租出去了","暂时不租了","效果不好，没人联系我","信息有误，需要重新发布","其他原因"]
  }
}
```

### 4.9 `POST /wechat/pay/notify` 微信支付回调

**变更点（前端无感知）：** 服务端 `product_type=4` 分支已实现，回调后招租 job 状态由 5 → 1，refresh_time 更新为当前时刻。前端可用 `POST /orders/status` 主动查询订单状态判断支付是否成功。
