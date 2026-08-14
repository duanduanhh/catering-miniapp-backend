# 求职模拟数据导入

脚本：`import_job_seekers.go`
数据源：`【整理后】求职数据.xlsx`

脚本直接连接数据库写入 `user`、`job` 表，**不调用 HTTP 接口**。

Excel 可以包含多个工作表；为避免误导入，`-sheet` 为必填参数。当前求职文件包含：

- `2028.8.7`：11 条数据；
- `2026.8.6`：11 条数据；
- `Sheet3`：空表。

两张数据表的列名略有差异：`2028.8.7` 使用“薪资下限/薪资上限”且没有 `uid`，脚本会用其 `id` 作为稳定的来源用户标识；`2026.8.6` 使用 `money_min/money_max` 和 `uid`。两者均可导入。

例如，先探索 `2028.8.7` 表（预演，不写库）：

```bash
go run ./scripts/mock/import_job_seekers.go \
  -conf config/test.yml \
  -sheet '2028.8.7'
```

## 导入内容

- Excel 每行生成一条求职信息：`biz_type=2`、`status=1`；
- `岗位类别`、行政区、地址、自我介绍、薪资、联系人和电话会映射到 `job` 表；
- `create_at`、`update_at`、`refresh_time` 均使用脚本执行时的当前时间；
- Excel 的 `user` 包含 `DL` 时，联系人会生成稳定的 `餐饮人 + 6 位字母数字` 昵称；
- 重复执行不会新增重复岗位，而是更新同一用户下、相同求职描述的岗位。

## 招聘数据导入

招聘数据源为 `【整理后】招聘数据.xlsx`，使用 `import_job_recruits.sh`。它会导入
`biz_type=1` 的招聘岗位，并映射企业名称、岗位要求、工作内容、经纬度和招聘人数。

招聘脚本已兼容 `2026.8.6` 和 `2026.8.7` 两种导出格式。`2026.8.7` 没有 `id` 时会使用
`uid` 作为来源记录标识；同时支持“工作要求”“薪资下限/薪资上限”“纬度”“法人代表”“注册资产”
等字段名。两张招聘表的坐标标签与实际数值方向相反，脚本会自动校正后写入 `longitude`、`latitude`。
`注册日期` 同时支持 Excel 数字日期和“2003年09月02日”这类中文日期；空值或“无”会使用脚本执行当天作为企业注册日期。
Excel 的“基础保障”“薪酬福利”“考勤休假”列分别写入岗位 `basic_protection`、`salary_benefits`、`attendance_leave`；三者均不是必填项，留空时对应字段也为空。

每条招聘数据会先按“导入用户 ID + 统一社会信用代码”查询 `enterprise`：

- 企业存在：复用企业并写入岗位 `enterprise_id`；
- 企业不存在：创建一条已认证企业，再写入岗位 `enterprise_id`。

先预演一条：

```bash
bash scripts/mock/import_job_recruits.sh \
  -conf config/test.yml \
  -sheet '工作表名称' \
  -limit 1
```

写入测试库全部招聘数据：

```bash
bash scripts/mock/import_job_recruits.sh \
  -conf config/test.yml \
  -sheet '工作表名称' \
  -execute
```

## 测试环境

测试环境会按 Excel 的 `uid` 创建或复用模拟用户，再将岗位关联给该用户。

先预演，不会连接或写入数据库：

```bash
go run ./scripts/mock/import_job_seekers.go \
  -conf config/test.yml \
  -sheet '2028.8.7'
```

仅写入第一条，用于验证：

```bash
go run ./scripts/mock/import_job_seekers.go \
  -conf config/test.yml \
  -sheet '2028.8.7' \
  -limit 1 \
  -execute
```

写入全部数据：

```bash
go run ./scripts/mock/import_job_seekers.go \
  -conf config/test.yml \
  -sheet '2028.8.7' \
  -execute
```

测试模式仅接受 `test.yml`，且 DSN 必须包含 `test`，避免误写生产库。

## 生产环境

生产模式不会创建模拟用户。必须指定一个已存在的目标用户，所有导入岗位都会关联给该用户。

以下示例将全部数据关联到 `user_id=971`：

```bash
go run ./scripts/mock/import_job_seekers.go \
  -conf config/prod.yml \
  -production \
  -sheet '2028.8.7' \
  -target-user-id 971 \
  -confirm-production-user-id 971 \
  -execute
```

`-confirm-production-user-id` 必须与 `-target-user-id` 相同；这是生产写入的二次确认。脚本会先校验目标用户存在，才会开始导入。

## 常用参数

| 参数 | 作用 |
| --- | --- |
| `-file` | 指定 Excel 文件路径 |
| `-sheet` | 要导入的 Excel 工作表名称，必填 |
| `-conf` | 数据库配置文件，默认 `config/test.yml` |
| `-limit` | 最多处理的记录数，`0` 表示全部 |
| `-execute` | 实际写入数据库；未传时仅预演 |
| `-production` | 开启生产导入模式 |
| `-target-user-id` | 生产导入关联的既有用户 ID |
