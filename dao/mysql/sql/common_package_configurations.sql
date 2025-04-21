-- demo.common_package_configurations definition

CREATE TABLE `common_package_configurations` (
                                                 `id` int(11) NOT NULL AUTO_INCREMENT,
                                                 `task_id` varchar(64) NOT NULL,
                                                 `service_name` varchar(64) NOT NULL,
                                                 `config_action` varchar(64) NOT NULL,
                                                 `build_number` int(11) DEFAULT NULL,
                                                 `config_content` text,
                                                 `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                                 `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                                 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;