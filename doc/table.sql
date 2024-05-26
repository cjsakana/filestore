CREATE TABLE `tbl_file`
(
    `id`        int(11)       not null auto_increment,
    `file_sha1` CHAR(40)      not null DEFAULT '' COMMENT '文件hash',
    `file_name` VARCHAR(256)  not null DEFAULT '' COMMENT '文件名',
    `file_size` BIGINT(20)             DEFAULT '0' COMMENT '文件大小',
    `file_addr` VARCHAR(1024) not null default '' comment '文件存储位置',
    `create_at` datetime               default NOW() comment '创建日期',
    `update_at` datetime               default NOW() on update CURRENT_TIMESTAMP() COMMENT '更新日期',
    `status`    int(11)       not null DEFAULT '0' COMMENT '状态(可用/禁用/已删除)',
    `ext1`      int(11)                DEFAULT '0' COMMENT '1',
    `ext2`      text comment '2',
    PRIMARY KEY (`id`),
    unique key `idx_file_hash` (`file_sha1`),
    key `idx_status` (`status`)
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

CREATE TABLE `tbl_user`
(
    `id`              int(11)      not null auto_increment,
    `user_name`       VARCHAR(64)  not null DEFAULT '' COMMENT '用户名',
    `user_pwd`        VARCHAR(256) not null DEFAULT '' COMMENT '用户encoded密码（加密过的）',
    `email`           VARCHAR(64)           default '' COMMENT '邮箱',
    `phone`           VARCHAR(128)          default '' comment '手机号',
    `email_validated` tinyint(1)            default 0 comment '邮箱是否已验证',
    `phone_validated` tinyint(1)            default 0 COMMENT '注册日期',
    `signup_at`       datetime              DEFAULT current_timestamp COMMENT '注册日期',
    `last_active`     datetime              DEFAULT current_timestamp on update current_timestamp COMMENT '最后活跃时间',
    `profile`         VARCHAR(256)          default '' comment '用户属性',
    `status`          int(11)      not null DEFAULT '0' COMMENT '用户状态(启用/禁用/锁定/标记删除)',
    PRIMARY KEY (`id`),
    unique key `idx_phone` (`phone`),
    key `idx_status` (`status`)
) ENGINE = INNODB
  DEFAULT CHARSET = utf8mb4;

CREATE TABLE `tbl_user_token`
(
    `id`         int(11)     NOT NULL AUTO_INCREMENT,
    `user_name`  varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
    `user_token` char(40)    NOT NULL DEFAULT '' COMMENT '用户登录token',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`user_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

# 用户文件表
CREATE TABLE `tbl_user_file`
(
    `id`          int(11)     not null auto_increment,
    `user_name`   VARCHAR(64) not null comment '哪个用户上传的',
    `file_sha1`   VARCHAR(64) not null DEFAULT '' COMMENT '文件hash',
    `file_size`   VARCHAR(20)          default '0' COMMENT '文件大小',
    `file_name`   VARCHAR(256)         default '' comment '文件名',
    `upload_at`   datetime             DEFAULT current_timestamp COMMENT '上传时间',
    `last_update` datetime             DEFAULT current_timestamp on update current_timestamp COMMENT '最后修改时间',
    `status`      int(11)     not null DEFAULT '0' COMMENT '文件状态(0正常/1已删除/2禁用)',
    PRIMARY KEY (`id`),
    unique key `idx_user_file` (`user_name`,`file_sha1`),
    key `idx_status` (`status`),
    key `idx_user_id` (`user_name`)
) ENGINE = INNODB
  DEFAULT CHARSET = utf8mb4;