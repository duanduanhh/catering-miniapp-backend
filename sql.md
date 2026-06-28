# SQL

## callback_history 回拨记录表

```sql
CREATE TABLE `callback_history` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL DEFAULT 0 COMMENT '发起回拨的用户ID',
  `purpose_id` bigint NOT NULL DEFAULT 0 COMMENT '关联岗位ID',
  `purpose_type` int NOT NULL DEFAULT 0 COMMENT '业务类型',
  `purpose_user_id` bigint NOT NULL DEFAULT 0 COMMENT '被回拨方用户ID',
  `purpose_user_name` varchar(64) NOT NULL DEFAULT '' COMMENT '被回拨方姓名',
  `purpose_user_phone` varchar(32) NOT NULL DEFAULT '' COMMENT '被回拨方手机号',
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_purpose_user_id` (`purpose_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='回拨记录表';
```
