CREATE TABLE `tbl_file`
(
    `id`        int(11) not null auto_increment,
    `file_sha1` CHAR(40)      not null DEFAULT '' COMMENT '文件hash',
    `file_name` VARCHAR(256)  not null DEFAULT '' COMMENT '文件名',
    `file_size` BIGINT(20) DEFAULT '0' COMMENT '文件大小',
    `file_addr` VARCHAR(1024) not null default '' comment '文件存储位置',
    `create_at` datetime               default NOW() comment '创建日期',
    `update_at` datetime               default NOW() on update CURRENT_TIMESTAMP () COMMENT '更新日期',
    `status`    int(11) not null DEFAULT '0' COMMENT '状态(可用/禁用/已删除)',
    `ext1`      int(11) DEFAULT '0' COMMENT '1',
    `ext2`      text comment '2',
    PRIMARY KEY (`id`),
    unique key `idx_file_hash` (`file_sha1`),
    key         `idx_status` (`status`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;