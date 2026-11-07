-- SKU 规则收敛：移除旧“购买用户”限制，所有用户均可购买。
-- 执行前请备份；本脚本不处理微信虚拟支付道具价格配置。

ALTER TABLE `payment_sku`
  ADD COLUMN IF NOT EXISTS `promotion_config` JSON NULL COMMENT '促销配置：首购价格、资格范围与命中文案' AFTER `sale_rule`;

START TRANSACTION;

UPDATE `payment_sku`
SET `promotion_config` = JSON_OBJECT(
  'first_purchase_price_cents', CAST(JSON_UNQUOTE(JSON_EXTRACT(`sale_rule`, '$.first_purchase_price_cents')) AS UNSIGNED),
  'first_purchase_scope', JSON_UNQUOTE(JSON_EXTRACT(`sale_rule`, '$.first_purchase_scope')),
  'subtitle', '新用户专享',
  'badge', '首单特惠',
  'virtual_product_id', `virtual_product_id`
)
WHERE JSON_EXTRACT(`sale_rule`, '$.first_purchase_price_cents') IS NOT NULL
  AND CAST(JSON_UNQUOTE(JSON_EXTRACT(`sale_rule`, '$.first_purchase_price_cents')) AS UNSIGNED) > 0;

-- 历史首购 SKU 曾把“新用户专享 / 首单特惠”写进普通展示字段；恢复为普通文案。
UPDATE `payment_sku`
SET
  `subtitle` = CASE WHEN `subtitle` = '新用户专享' THEN '置顶1天，提升曝光' ELSE `subtitle` END,
  `badge` = CASE WHEN `badge` IN ('首单特惠', '新用户专享') THEN '' ELSE `badge` END
WHERE `sku_code` IN ('job_top_1d', 'job_top_1d_new')
  AND JSON_EXTRACT(`sale_rule`, '$.first_purchase_price_cents') IS NOT NULL;

-- 注意：迁移会将历史 virtual_product_id 视为“首购微信道具 ID”。
-- 请在后台将该 SKU 的“微信道具ID”改为常规售价对应的道具，再核对“首购微信道具ID”。

UPDATE `payment_sku`
SET `sale_rule` = JSON_REMOVE(`sale_rule`, '$.first_purchase_price_cents', '$.first_purchase_scope')
WHERE JSON_CONTAINS_PATH(`sale_rule`, 'one', '$.first_purchase_price_cents', '$.first_purchase_scope');

UPDATE `payment_sku`
SET `sale_rule` = JSON_REMOVE(`sale_rule`, '$.audience')
WHERE JSON_CONTAINS_PATH(`sale_rule`, 'one', '$.audience');

COMMIT;

-- first_top_status 仅属于旧的“置顶首购专属 SKU”判断，代码已不再使用。
-- 确认没有旧客户端依赖 /user/info.first_top_status 后再执行：
-- ALTER TABLE `user` DROP COLUMN `first_top_status`;
