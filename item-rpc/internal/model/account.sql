-- 用户账户
CREATE TABLE account (
                         id BIGINT AUTO_INCREMENT,
                         balance BIGINT NOT NULL DEFAULT 0 comment '余额',
                         freeze_amount BIGINT NOT NULL DEFAULT 0 comment '冻结金额',
                         version BIGINT NOT NULL DEFAULT 0 comment '版本号',
                         updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP comment '更新时间',
                         PRIMARY KEY (id)
) ENGINE=InnoDB COLLATE utf8mb4_general_ci COMMENT 'account table';

-- 支付流水
CREATE TABLE pay_txn (
                         id BIGINT AUTO_INCREMENT,
                         user_id BIGINT NOT NULL DEFAULT 0 comment '用户ID',
                         order_id BIGINT NOT NULL DEFAULT 0 comment '订单ID',
                         amount BIGINT NOT NULL DEFAULT 0 comment '金额',
                         status TINYINT NOT NULL DEFAULT 0 comment '状态', -- 0:init,1:success,2:fail
                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP comment '创建时间',
                         UNIQUE KEY uk_user_order(user_id, order_id),
                             PRIMARY KEY (id)
) ENGINE=InnoDB COLLATE utf8mb4_general_ci COMMENT 'pay_txn table';

-- TCC 分支表（记录每个分支的状态，确保幂等）
CREATE TABLE tcc_branch (
                            id BIGINT AUTO_INCREMENT,
                            xid VARCHAR(64) NOT NULL DEFAULT '' comment '全局事务ID',
                            branch VARCHAR(64) NOT NULL DEFAULT '' comment '分支事务ID',
                            state TINYINT NOT NULL DEFAULT 0 comment '状态',  -- 0:TRY, 1:CONFIRM, 2:CANCEL
                            payload JSON NULL DEFAULT '' comment '透传必要业务字段',
                            expire_at TIMESTAMP NULL DEFAULT '' comment 'TRY 过期时间（防悬挂）',
                            UNIQUE KEY uk_branch(xid, branch),
                            PRIMARY KEY (id)
) ENGINE=InnoDB COLLATE utf8mb4_general_ci COMMENT 'tcc_branch table';