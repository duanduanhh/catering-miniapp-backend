-- 付费 SKU 初始化（MySQL 8.0）
-- 执行前先运行数据库迁移；本脚本仅重建 payment_sku 及其变更记录。

SET NAMES utf8mb4;
START TRANSACTION;

SET @operator := 'system_seed';
SET @created_at := CURRENT_TIMESTAMP(3);

DELETE FROM payment_sku_change_log;
DELETE FROM payment_sku;

INSERT INTO payment_product (product_code, name, selection_mode, purchase_notice, create_at, update_at)
VALUES
  ('job_top', '岗位置顶', 2, '置顶权益支付成功后立即生效，具体时长以所选 SKU 为准。', @created_at, @created_at),
  ('contact_voucher', '联系券', 2, '联系券绑定当前账号，不可转让；支付成功后自动到账。', @created_at, @created_at),
  ('paid_refresh', '付费刷新', 1, '付费刷新支付成功后立即提升当前信息的排序时间。', @created_at, @created_at),
  ('rent_publish', '招租发布', 1, '支付成功后招租信息自动发布。', @created_at, @created_at)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  selection_mode = VALUES(selection_mode),
  purchase_notice = VALUES(purchase_notice),
  update_at = VALUES(update_at);

CREATE TEMPORARY TABLE payment_sku_seed (
  product_code VARCHAR(64) NOT NULL,
  sku_code VARCHAR(64) NOT NULL,
  name VARCHAR(100) NOT NULL,
  subtitle VARCHAR(200) NOT NULL,
  badge VARCHAR(50) NOT NULL,
  price_cents BIGINT NOT NULL,
  original_price_cents BIGINT NOT NULL,
  benefit_config JSON NOT NULL,
  sale_rule JSON NOT NULL,
  promotion_config JSON NULL,
  sort INT NOT NULL,
  PRIMARY KEY (sku_code)
);

INSERT INTO payment_sku_seed VALUES
  ('job_top', 'job_top_1d', '置顶1天', '置顶1天，提升曝光', '', 500, 0, JSON_OBJECT('top_hours', 24), JSON_OBJECT('max_purchase_per_user', 0), JSON_OBJECT('first_purchase_price_cents', 390, 'first_purchase_scope', 'product', 'subtitle', '新用户专享', 'badge', '首单特惠', 'virtual_product_id', 'job_top_1d_first'), 320),
  ('job_top', 'job_top_3d', '置顶3天', '每天仅需3.33元', '推荐', 1000, 0, JSON_OBJECT('top_hours', 72, 'gift_contact_vouchers', 2), JSON_OBJECT('max_purchase_per_user', 0), NULL, 310),
  ('job_top', 'job_top_7d', '置顶7天', '每天仅需2.86元', '', 2000, 0, JSON_OBJECT('top_hours', 168, 'gift_contact_vouchers', 5), JSON_OBJECT('max_purchase_per_user', 0), NULL, 300),
  ('contact_voucher', 'contact_voucher_2', '2张联系券', '限购1次', '', 150, 0, JSON_OBJECT('contact_vouchers', 2), JSON_OBJECT('max_purchase_per_user', 1), NULL, 260),
  ('contact_voucher', 'contact_voucher_5', '5张联系券', '购买最多', '最畅销', 500, 0, JSON_OBJECT('contact_vouchers', 5), JSON_OBJECT('max_purchase_per_user', 0), NULL, 250),
  ('contact_voucher', 'contact_voucher_10', '10张联系券', '每张0.8元', '8折', 800, 1000, JSON_OBJECT('contact_vouchers', 10), JSON_OBJECT('max_purchase_per_user', 0), NULL, 240),
  ('contact_voucher', 'contact_voucher_20', '20张联系券', '每张0.7元', '7折', 1400, 2000, JSON_OBJECT('contact_vouchers', 20), JSON_OBJECT('max_purchase_per_user', 0), NULL, 230),
  ('contact_voucher', 'contact_voucher_50', '50张联系券', '每张0.6元', '6折', 3000, 5000, JSON_OBJECT('contact_vouchers', 50), JSON_OBJECT('max_purchase_per_user', 0), NULL, 220),
  ('contact_voucher', 'contact_voucher_100', '100张联系券', '每张0.5元', '5折', 5000, 10000, JSON_OBJECT('contact_vouchers', 100), JSON_OBJECT('max_purchase_per_user', 0), NULL, 210),
  ('paid_refresh', 'paid_refresh_1', '刷新1次', '立即刷新，提升曝光', '惊爆价', 200, 400, JSON_OBJECT('refresh_times', 1), JSON_OBJECT('max_purchase_per_user', 0), NULL, 200),
  ('rent_publish', 'rent_publish_1', '发布招租信息', '发布1条招租信息', '', 1800, 0, JSON_OBJECT('rent_publish_times', 1), JSON_OBJECT('max_purchase_per_user', 0), NULL, 100);

INSERT INTO payment_sku (product_id, sku_code, name, subtitle, badge, price_cents, original_price_cents, benefit_config, sale_rule, promotion_config, status, sort, version, created_by, updated_by, create_at, update_at)
SELECT product.id, seed.sku_code, seed.name, seed.subtitle, seed.badge, seed.price_cents, seed.original_price_cents, seed.benefit_config, seed.sale_rule, seed.promotion_config, 1, seed.sort, 1, @operator, @operator, @created_at, @created_at
FROM payment_sku_seed seed
JOIN payment_product product ON product.product_code = seed.product_code;

DROP TEMPORARY TABLE payment_sku_seed;
COMMIT;
