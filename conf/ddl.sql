CREATE TABLE `trade_state`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `trade_id`       bigint        NOT NULL COMMENT '交易ID',
    `trade_scene`    bigint        NOT NULL COMMENT '交易场景',
    `from_wallet_id` bigint        NOT NULL COMMENT '扣款钱包ID',
    `from_amount`    bigint        NOT NULL COMMENT '扣款金额',
    `to_wallets`     varchar(5000) NOT NULL COMMENT '收款钱包信息列表',
    `coin_type`      int           NOT NULL COMMENT '虚拟币类型',
    `status`         int           NOT NULL COMMENT '状态 1-进行中 2-回滚中 3-半成功 4-成功 5-已回滚',
    `comment`        varchar(1000) NOT NULL COMMENT '备注',
    `created_at`     bigint        NOT NULL COMMENT '创建时间',
    `updated_at`     bigint        NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_trade_state (trade_id, trade_scene),
    KEY              `status_index` (`status`)
) COMMENT '交易状态表';

CREATE TABLE `trade_record`
(
    `id`           bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `wallet_id`    bigint        NOT NULL COMMENT '钱包ID',
    `trade_id`     bigint        NOT NULL COMMENT '交易ID',
    `trade_scene`  int           NOT NULL COMMENT '交易场景',
    `trade_type`   int           NOT NULL COMMENT '交易类型',
    `trade_status` int           NOT NULL COMMENT '交易状态 1-正常 2-已回滚 3-空回滚',
    `amount`       bigint        NOT NULL COMMENT '扣款金额',
    `coin_type`    int           NOT NULL COMMENT '虚拟币类型',
    `change_type`  int           NOT NULL COMMENT '变动类型，消费着的trade_scene 收益者的add_type',
    `comment`      varchar(1000) NOT NULL COMMENT '备注',
    `created_at`   bigint        NOT NULL COMMENT '创建时间',
    `updated_at`   bigint        NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_trade_record (wallet_id, trade_id, coin_type, trade_scene, trade_type, change_type)
) COMMENT '记录表';

CREATE TABLE `wallet_bag`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `wallet_id`  bigint NOT NULL COMMENT '钱包ID',
    `amount`     bigint NOT NULL COMMENT '支付余额',
    `coin_type`  int    NOT NULL COMMENT '虚拟币类型',
    `created_at` bigint NOT NULL COMMENT '创建时间',
    `updated_at` bigint NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_wallet (wallet_id, coin_type)
) COMMENT '钱包表';