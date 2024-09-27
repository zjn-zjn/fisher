CREATE TABLE `state`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `transfer_id`    bigint        NOT NULL COMMENT '转移ID',
    `transfer_scene` bigint        NOT NULL COMMENT '转移场景',
    `from_bags`      varchar(5000) NOT NULL COMMENT '收款背包信息列表',
    `to_bags`        varchar(5000) NOT NULL COMMENT '收款背包信息列表',
    `item_type`      int           NOT NULL COMMENT '物品类型',
    `status`         int           NOT NULL COMMENT '状态 1-进行中 2-回滚中 3-半成功 4-成功 5-已回滚',
    `comment`        varchar(1000) NOT NULL COMMENT '备注',
    `created_at`     bigint        NOT NULL COMMENT '创建时间',
    `updated_at`     bigint        NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_state (transfer_id, transfer_scene),
    KEY              `status_updated_at_index` (`status`, `updated_at`)
) COMMENT '转移状态表';

CREATE TABLE `record`
(
    `id`              bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `bag_id`          bigint        NOT NULL COMMENT '背包ID',
    `transfer_id`     bigint        NOT NULL COMMENT '转移ID',
    `transfer_scene`  int           NOT NULL COMMENT '转移场景',
    `transfer_type`   int           NOT NULL COMMENT '转移类型',
    `transfer_status` int           NOT NULL COMMENT '转移状态 1-正常 2-已回滚 3-空回滚',
    `amount`          bigint        NOT NULL COMMENT '变动金额',
    `item_type`       int           NOT NULL COMMENT '物品类型',
    `change_type`     int           NOT NULL COMMENT '变动类型',
    `comment`         varchar(1000) NOT NULL COMMENT '备注',
    `created_at`      bigint        NOT NULL COMMENT '创建时间',
    `updated_at`      bigint        NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_record (bag_id, transfer_id, item_type, transfer_scene, transfer_type, change_type)
) COMMENT '记录表';

CREATE TABLE `bag`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `bag_id`     bigint NOT NULL COMMENT '背包ID',
    `amount`     bigint NOT NULL COMMENT '物品数量',
    `item_type`  int    NOT NULL COMMENT '物品类型',
    `created_at` bigint NOT NULL COMMENT '创建时间',
    `updated_at` bigint NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    unique index uk_bag (bag_id, item_type)
) COMMENT '背包表';