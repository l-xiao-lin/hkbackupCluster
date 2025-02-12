/*
Navicat MySQL Data Transfer

Source Server         : 10.0.0.180
Source Server Version : 50744
Source Host           : 10.0.0.180:3306
Source Database       : demo

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2025-02-11 17:05:33
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for package_operations
-- ----------------------------
DROP TABLE IF EXISTS `package_operations`;
CREATE TABLE `package_operations` (
                                      `id` int(11) NOT NULL AUTO_INCREMENT,
                                      `task_id` varchar(64) NOT NULL,
                                      `job_name` varchar(64) NOT NULL,
                                      `build_number` int(11) DEFAULT NULL,
                                      `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:待打包,1:成功;2:失败,3:退回',
                                      `host` varchar(128) NOT NULL,
                                      `common` varchar(255) DEFAULT NULL,
                                      `diff` varchar(255) DEFAULT NULL,
                                      `rm_rulepackage` tinyint(1) NOT NULL DEFAULT '0',
                                      `pkg_name` varchar(11) DEFAULT NULL,
                                      `update_jbossconf` tinyint(1) NOT NULL DEFAULT '0',
                                      `src_path` varchar(64) NOT NULL,
                                      `update_sdkconf` tinyint(1) NOT NULL DEFAULT '0',
                                      `update_security` tinyint(1) NOT NULL DEFAULT '0',
                                      `is_sql_exec` tinyint(1) NOT NULL DEFAULT '0' COMMENT '0:不需要停服跑sql; 1:需要停服跑sql',
                                      `is_package` tinyint(1) NOT NULL DEFAULT '0',
                                      `package_time` timestamp NULL DEFAULT NULL,
                                      `scheduled_time` timestamp NULL DEFAULT NULL,
                                      `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                      `canary_status` tinyint(4) DEFAULT NULL COMMENT '1:需要灰度;2:取消灰度',
                                      PRIMARY KEY (`id`),
                                      UNIQUE KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=175 DEFAULT CHARSET=utf8mb4;
