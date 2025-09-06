CREATE TABLE IF NOT EXISTS stock (
    id        bigint AUTO_INCREMENT,
    sku      BIGINT      NOT NULL DEFAULT 0 COMMENT 'sku',
    qty      BIGINT      NOT NULL DEFAULT 0 COMMENT 'qty',
    version  BIGINT      NOT NULL DEFAULT 0 COMMENT 'version',
    update_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE sku_index (sku)
) ENGINE=InnoDB COLLATE utf8mb4_general_ci COMMENT 'stock table';