CREATE TABLE user
(
    id        bigint AUTO_INCREMENT,
    password  varchar(255) NOT NULL DEFAULT '' COMMENT 'The user password',
    mobile    varchar(11)  NOT NULL DEFAULT '' COMMENT 'The mobile phone number',
    email     varchar(50)  NOT NULL DEFAULT '' COMMENT 'The email',
    gender    char(10)     NOT NULL DEFAULT 'male' COMMENT 'gender,male|female|unknown',
    age       varchar(3) NULL DEFAULT '' COMMENT 'The age',
    create_at timestamp NULL,
    update_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE mobile_index (mobile),
    UNIQUE email_index (email),
    PRIMARY KEY (id)
) ENGINE = InnoDB COLLATE utf8mb4_general_ci COMMENT 'user table';