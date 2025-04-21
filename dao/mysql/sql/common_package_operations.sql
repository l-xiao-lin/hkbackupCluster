-- demo.common_package_operations definition

CREATE TABLE `common_package_operations` (
                                             `id` int(11) NOT NULL AUTO_INCREMENT,
                                             `task_id` varchar(64) NOT NULL,
                                             `service_name` varchar(64) NOT NULL,
                                             `has_configuration` tinyint(1) NOT NULL DEFAULT '0',
                                             `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:待打包,1:成功;2:失败,3:退回',
                                             `build_number` int(11) DEFAULT NULL,
                                             `open_schema` tinyint(1) DEFAULT '0',
                                             `package_time` timestamp NULL DEFAULT NULL,
                                             `scheduled_time` timestamp NULL DEFAULT NULL,
                                             `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                             `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                             PRIMARY KEY (`id`),
                                             UNIQUE KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4;