# 数据表设计

## 用户表（改造）

```mysql
CREATE TABLE `user` (
  `id` int NOT NULL AUTO_INCREMENT,
  `avatar` longtext COMMENT '用户头像URL',
  `name` longtext COMMENT '用户昵称',
  `sex` int DEFAULT '0' COMMENT '性别：0未知 1男 2女',
  `age` int DEFAULT '0' COMMENT '年龄',
  `birthday` longtext COMMENT '生日',
  `phone` varchar(64) DEFAULT NULL COMMENT '手机号（登录/联系）',
  `wechart_code` varchar(255) DEFAULT NULL COMMENT '微信登录临时code（wx.login获取, 一次性）',
  `wechat_open_id` longtext COMMENT '微信 openId（用户唯一标识）',
  `token` longtext COMMENT '登录态 token',
  `password` longtext COMMENT '用户密码（非微信登录或后台使用）',
  `first_area_id` int DEFAULT '0' COMMENT '一级地区ID（省）',
  `second_area_id` int DEFAULT '0' COMMENT '二级地区ID（市）',
  `third_area_id` int DEFAULT '0' COMMENT '三级地区ID（区/县）',
  `address` longtext COMMENT '详细地址',
  `longitude` double DEFAULT '0' COMMENT '经度',
  `latitude` double DEFAULT '0' COMMENT '纬度',
  `type` int DEFAULT '0' COMMENT '用户类型：0普通用户 1商家 2管理员',
  `status` int DEFAULT '0' COMMENT '用户状态：0正常 1禁用 2注销',
  `integral` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户积分',
  `collect_num` bigint unsigned DEFAULT '0' COMMENT '收藏数量',
  `buy_num` bigint unsigned DEFAULT '0' COMMENT '购买次数',
  `invite_id` bigint DEFAULT '0' COMMENT '邀请人用户ID',
  `invite_num` bigint unsigned DEFAULT '0' COMMENT '成功邀请人数',
  `first_recharge` longtext COMMENT '首次充值标识/记录',
  `total_recharge` double DEFAULT '0' COMMENT '累计充值金额',
  `device_model` longtext COMMENT '设备型号',
  `ip` longtext COMMENT '最近登录IP',
  `contact_voucher_num` int DEFAULT '0' COMMENT '联系券余额',
  `first_top_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '首单状态：0=未触发，1=已触发（首次置顶支付成功后设为1，不可回退）',
  `new_customer_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '新客状态：0=未触发，1=已触发（购买联系券支付成功后设为1，不可回退）',
  `profile_complete_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '信息完善状态：0=未完善，1=已完善（首次编辑个人信息成功后设为1，赠送2张联系券，不可回退）',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4882 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户'
```

## 订单表（新建）

```mysql
CREATE TABLE `orders` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_no` varchar(32) NOT NULL COMMENT '订单号（业务唯一）',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `amount_total` decimal(10,2) NOT NULL COMMENT '订单总金额（应付）',
  `amount_paid` decimal(10,2) NOT NULL DEFAULT 0.00 COMMENT '实付金额',
  `currency` char(3) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '订单状态：1=待支付 2=已支付 3=已取消 4=已退款',
  `pay_channel` varchar(32) DEFAULT NULL COMMENT '支付渠道：wxpay/alipay/stripe等',
  `pay_trade_no` varchar(64) DEFAULT NULL COMMENT '第三方支付单号',
  `paid_at` datetime(3) DEFAULT NULL COMMENT '支付时间',
  `canceled_at` datetime(3) DEFAULT NULL COMMENT '取消时间',
  `refunded_at` datetime(3) DEFAULT NULL COMMENT '退款时间',
  `remark` varchar(255) DEFAULT NULL COMMENT '订单备注（可用于客服）',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status_time` (`user_id`, `status`, `create_at`),
  KEY `idx_status_create_at` (`status`, `create_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='订单主表';


CREATE TABLE `order_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` bigint NOT NULL COMMENT '订单ID（order.id）',
  `product_type` tinyint NOT NULL COMMENT '商品类型：1=置顶套餐 2=联系券套餐 3=付费刷新',
  `title_snapshot` varchar(64) NOT NULL COMMENT '套餐名称快照',
  `unit_price_snapshot` decimal(10,2) NOT NULL COMMENT '单价快照（元）',
  `top_hour` int NOT NULL DEFAULT 0 COMMENT '置顶时长（小时）, 仅product_type=1有效',
  `contact_voucher_num` int NOT NULL DEFAULT 0 COMMENT '联系券数量, 仅product_type=2有效',
  `target_type` tinyint DEFAULT NULL COMMENT '目标内容类型：1=招聘 2=求职, 仅product_type=1/2有效',
  `target_id` bigint DEFAULT NULL COMMENT '目标内容ID（如job_id/resume_id, ,仅product_type=1/2有效）',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_product_type` (`product_type`),
  KEY `idx_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='订单明细表';

```

## 招聘信息表（复用）

```mysql
# 新
CREATE TABLE `job` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `user_id` bigint NOT NULL COMMENT '发布岗位的用户ID, 对应 user.id',
  `positions` varchar(64) NOT NULL COMMENT '岗位名称',
  `company_name` varchar(128) DEFAULT NULL COMMENT '企业名称',
  `longitude` decimal(10,7) DEFAULT NULL COMMENT '岗位所在地经度',
  `latitude` decimal(10,7) DEFAULT NULL COMMENT '岗位所在地纬度',
  `address` varchar(512) DEFAULT NULL COMMENT '岗位详细地址',
  `contact_person_name` varchar(64) NOT NULL COMMENT '联系人昵称',
  `contact` varchar(64) NOT NULL COMMENT '联系方式（手机号）',
  `description` text COMMENT '岗位描述',
  `photo_urls` longtext COMMENT '岗位相关图片URL列表（支持多个, 逗号分割）',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1=生效 2=用户关闭 3=管理员下架 4=删除',
  `close_reason` varchar(255) DEFAULT NULL COMMENT '关闭原因：用户关闭/管理员下架时填写',
  `close_time` datetime(3) DEFAULT NULL COMMENT '关闭时间',
  `first_area_id` int DEFAULT NULL COMMENT '一级地区ID（省）',
  `first_area_des` varchar(64) DEFAULT NULL COMMENT '一级地区名称',
  `second_area_id` int DEFAULT NULL COMMENT '二级地区ID（市）',
  `second_area_des` varchar(64) DEFAULT NULL COMMENT '二级地区名称',
  `third_area_id` int DEFAULT NULL COMMENT '三级地区ID（区/县）',
  `third_area_des` varchar(64) DEFAULT NULL COMMENT '三级地区名称',
  `four_area_id` int DEFAULT NULL COMMENT '四级地区ID（街道/商圈）',
  `four_area_des` varchar(64) DEFAULT NULL COMMENT '四级地区名称',
  `salary_min` int DEFAULT NULL COMMENT '薪资下限（NULL=未知）',
  `salary_max` int DEFAULT NULL COMMENT '薪资上限（NULL=未知）',
  `basic_protection` text COMMENT '基础保障（逗号分隔）',
  `salary_benefits` text COMMENT '薪酬福利（逗号分隔）',
  `attendance_leave` text COMMENT '考勤休假（逗号分隔）',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `refresh_time` datetime(3) DEFAULT NULL COMMENT '刷新时间',
  `top_start_time` datetime(3) DEFAULT NULL COMMENT '置顶开始时间',
  `top_end_time` datetime(3) DEFAULT NULL COMMENT '置顶结束时间',
  `is_top` tinyint NOT NULL DEFAULT '0' COMMENT '是否置顶（冗余字段，便于列表排序/筛选）',
  `biz_type` tinyint NOT NULL DEFAULT '1' COMMENT '业务类型：1=招聘 2=求职',

  PRIMARY KEY (`id`),
  KEY `idx_job_user_id` (`user_id`),
  KEY `idx_status_refresh` (`status`,`refresh_time`),
  KEY `idx_status_four_area_refresh` (`status`,`four_area_id`,`refresh_time`),
  KEY `idx_status_top_refresh` (`status`,`is_top`,`refresh_time`),
  KEY `idx_top_time` (`top_start_time`,`top_end_time`),

  CONSTRAINT `chk_salary_range` CHECK (
    (
      (`salary_min` IS NULL AND `salary_max` IS NULL)
      OR (`salary_min` IS NOT NULL AND `salary_max` IS NULL)
      OR (`salary_min` IS NULL AND `salary_max` IS NOT NULL)
      OR (`salary_min` IS NOT NULL AND `salary_max` IS NOT NULL AND `salary_max` >= `salary_min`)
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='岗位详情表';

```

## 沟通记录表（复用）

```mysql
CREATE TABLE `contact_history` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '联系记录主键ID, 自增',
  `user_id` bigint NOT NULL COMMENT '发起联系的用户ID, 对应 user.id',
  `purpose_id` bigint NOT NULL COMMENT '被联系的业务对象ID（如招聘ID/简历ID/招租ID）',
  `purpose_type` tinyint NOT NULL COMMENT '联系对象类型：1=招聘 2=求职 3=招租',
  `purpose_user_id` bigint DEFAULT NULL COMMENT '被联系用户ID',
  `purpose_user_name` varchar(64) DEFAULT NULL COMMENT '被联系用户昵称',
  `purpose_user_phone` varchar(64) DEFAULT NULL COMMENT '被联系对象手机号（快照）',
  `user_deleted` tinyint NOT NULL DEFAULT '0' COMMENT '发起方是否删除：0-否 1-是',
  `purpose_user_deleted` tinyint NOT NULL DEFAULT '0' COMMENT '被联系方是否删除：0-否 1-是',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '联系创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '联系更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_purpose` (`purpose_type`, `purpose_id`),
  KEY `idx_purpose_user_id` (`purpose_user_id`, `purpose_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='联系记录';
```

## 收藏表（复用）

```mysql
CREATE TABLE `collect` (
  `user_id` int NOT NULL COMMENT '用户ID',
  `content_id` int NOT NULL COMMENT '业务对象ID（如招聘ID/简历ID/招租ID）',
  `type` tinyint NOT NULL COMMENT '业务类型：1=招聘 2=求职 3=招租',
  `status` int DEFAULT '1' COMMENT '状态: 1=生效 2=删除',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`content_id`,`user_id`,`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
```

## 联系券变更表（新建）

```mysql
CREATE TABLE `contact_voucher_history` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `user_id` bigint DEFAULT NULL COMMENT '发起联系的用户ID, 对应 user.id',
  `biz_type` bigint DEFAULT NULL COMMENT '1=充值 2=消费',
  `change_num` int NOT NULL DEFAULT 0 COMMENT '变更数量',
  `last_num` int NOT NULL DEFAULT 0 COMMENT '变更前数量',
  `next_num` int NOT NULL DEFAULT 0 COMMENT '变更后数量',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='联系券变更表';
```

## 岗位列表（新建）

### 一级分类表 

```mysql
CREATE TABLE `position_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_name` varchar(64) NOT NULL COMMENT '分类名称',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序序号',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=生效 0=禁用',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='岗位一级分类表';

INSERT INTO `position_category` (`id`, `category_name`, `sort_order`, `status`) VALUES
(1, '前厅/门店服务', 1, 1),
(2, '后厨-厨师', 2, 1),
(3, '后厨-其他', 3, 1),
(4, '饮品/甜点', 4, 1),
(5, '综合职能支持', 5, 1),
(6, '其他', 6, 1);
```

### 二级分类表

```mysql
CREATE TABLE `position_subcategory` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_id` bigint NOT NULL COMMENT '一级分类ID',
  `subcategory_name` varchar(64) NOT NULL COMMENT '二级分类名称',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序序号',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=生效 0=禁用',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `update_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_sort_order` (`sort_order`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_category_id` FOREIGN KEY (`category_id`) REFERENCES `position_category` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='岗位二级分类表';

-- 前厅/门店服务
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(1, '店长/副店长', 1, 1),
(1, '大堂经理', 2, 1),
(1, '领班/主管', 3, 1),
(1, '服务员/传菜', 4, 1),
(1, '迎宾/接待', 5, 1),
(1, '收银员/前台', 6, 1),
(1, '打包/送餐', 7, 1),
(1, '营业/售卖员', 8, 1);

-- 后厨-厨师
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(2, '厨师长', 1, 1),
(2, '炉子/主厨/炒锅', 2, 1),
(2, '白案/面条/饺子', 3, 1),
(2, '凉菜/冷盘/腊卤', 4, 1),
(2, '点心/烘焙', 5, 1),
(2, '蒸菜/上什', 6, 1),
(2, '汤锅/煲汤/粥水', 7, 1),
(2, '火锅底料/打锅', 8, 1),
(2, '烧烤/烤鸭/穿串', 9, 1),
(2, '西餐厨师', 10, 1),
(2, '日料/寿司/刺身', 11, 1),
(2, '韩料', 12, 1),
(2, '小吃厨师', 13, 1),
(2, '雕刻/姿造', 14, 1);

-- 后厨-其他
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(3, '勤杂工/洗碗工', 1, 1),
(3, '配菜/切配/墩子', 2, 1),
(3, '打荷', 3, 1),
(3, '快餐备餐员', 4, 1),
(3, '后厨学徒/帮厨', 5, 1);

-- 饮品/甜点
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(4, '咖啡师', 1, 1),
(4, '奶茶师', 2, 1),
(4, '茶艺师', 3, 1),
(4, '烘焙/裱花', 4, 1),
(4, '水吧', 5, 1),
(4, '调酒师', 6, 1);

-- 综合职能支持
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(5, '文员/库管', 1, 1),
(5, '人力资源', 2, 1),
(5, '财务/出纳', 3, 1),
(5, '采购', 4, 1),
(5, '运营/营销/销售', 5, 1),
(5, '品控/食安/督导', 6, 1);

-- 其他
INSERT INTO `position_subcategory` (`category_id`, `subcategory_name`, `sort_order`, `status`) VALUES
(6, '其他', 1, 1);
```

## 意见反馈表（新建）

```mysql
CREATE TABLE `feedback` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '提交用户ID',
  `type` tinyint NOT NULL COMMENT '反馈类型：1=产品建议 2=功能问题 3=内容修正 4=其他',
  `content` text NOT NULL COMMENT '详细说明（最多500字）',
  `photo_urls` longtext DEFAULT NULL COMMENT '图片URL，逗号分隔，最多4张',
  `create_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='意见反馈';
```
