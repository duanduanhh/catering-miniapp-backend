-- 付费 SKU 配置表（MySQL 8.0）
-- 仅创建第一步“后台配置收费业务与 SKU”的数据结构，不包含初始化套餐数据。
-- 已存在的生产库请先执行 go run ./cmd/migration -conf <配置文件> 完成历史 payment_package 表迁移；
-- 新环境可直接执行本文件，再执行 seed_payment_packages.sql 初始化测试套餐。

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `payment_product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `product_code` varchar(64) NOT NULL COMMENT '收费业务编码：job_top/contact_voucher/paid_refresh/rent_publish',
  `name` varchar(100) NOT NULL COMMENT '收费业务名称',
  `selection_mode` tinyint NOT NULL COMMENT '规格选择方式：1=单规格，2=多规格',
  `purchase_notice` text COMMENT '小程序购买须知',
  `create_at` datetime(3) NOT NULL COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_product_code` (`product_code`),
  CONSTRAINT `chk_payment_product_selection_mode` CHECK (`selection_mode` IN (1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='付费收费业务';

CREATE TABLE IF NOT EXISTS `payment_sku` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `product_id` bigint NOT NULL COMMENT '所属收费业务 ID',
  `sku_code` varchar(64) NOT NULL COMMENT 'SKU 唯一编码',
  `virtual_product_id` varchar(128) NOT NULL DEFAULT '' COMMENT '微信虚拟支付后台道具 ID',
  `name` varchar(100) NOT NULL COMMENT 'SKU 名称',
  `subtitle` varchar(200) NOT NULL DEFAULT '' COMMENT '副标题',
  `badge` varchar(50) NOT NULL DEFAULT '' COMMENT '角标',
  `price_cents` bigint NOT NULL COMMENT '售价，单位为分',
  `original_price_cents` bigint NOT NULL DEFAULT 0 COMMENT '划线价，单位为分；0=不展示',
  `benefit_config` json NOT NULL COMMENT '权益配置',
  `sale_rule` json NOT NULL COMMENT '限购规则',
  `promotion_config` json DEFAULT NULL COMMENT '促销配置：首购价格、资格范围与命中文案',
  `status` tinyint NOT NULL COMMENT '状态：1=草稿，2=已上架，3=已下架',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序，数值越大越靠前',
  `version` int NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_by` varchar(64) NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by` varchar(64) NOT NULL DEFAULT '' COMMENT '最后修改人',
  `create_at` datetime(3) NOT NULL COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_sku_code` (`sku_code`),
  KEY `idx_payment_sku_product_id` (`product_id`),
  KEY `idx_payment_sku_list` (`status`, `sort`),
  KEY `idx_payment_sku_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='付费 SKU';

CREATE TABLE IF NOT EXISTS `payment_sku_change_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `sku_id` bigint NOT NULL COMMENT 'SKU ID',
  `sku_version` int NOT NULL COMMENT '变更后的 SKU 版本号',
  `action` tinyint NOT NULL COMMENT '操作：1=创建，2=编辑，3=上架，4=下架，5=删除',
  `before_snapshot` json DEFAULT NULL COMMENT '变更前快照',
  `after_snapshot` json DEFAULT NULL COMMENT '变更后快照',
  `change_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '变更原因',
  `operator` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人',
  `create_at` datetime(3) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_payment_sku_log` (`sku_id`, `create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='付费 SKU 变更记录';

-- 订单明细由原订单模块创建。若该表已存在但尚未接入 SKU，请通过迁移程序补齐以下字段：
-- ALTER TABLE `order_item`
--   ADD COLUMN `sku_id` bigint NOT NULL DEFAULT 0,
--   ADD COLUMN `sku_code` varchar(64) NOT NULL DEFAULT '',
--   ADD COLUMN `sku_version` int NOT NULL DEFAULT 0,
--   ADD COLUMN `virtual_product_id_snapshot` varchar(128) NOT NULL DEFAULT '',
--   ADD COLUMN `price_cents_snapshot` bigint NOT NULL DEFAULT 0,
--   ADD COLUMN `benefit_snapshot` json DEFAULT NULL,
--   ADD COLUMN `promotion_snapshot` json DEFAULT NULL COMMENT '订单命中的营销规则快照';
