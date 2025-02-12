/*
Navicat MySQL Data Transfer

Source Server         : 10.0.0.180
Source Server Version : 50744
Source Host           : 10.0.0.180:3306
Source Database       : demo

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2025-02-11 15:38:54
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for release_xml
-- ----------------------------
DROP TABLE IF EXISTS `release_xml`;
CREATE TABLE `release_xml` (
                               `id` int(11) NOT NULL AUTO_INCREMENT,
                               `task_id` varchar(64) NOT NULL,
                               `job_name` varchar(64) NOT NULL,
                               `src_path` varchar(64) NOT NULL,
                               `host` varchar(64) NOT NULL,
                               `common` varchar(64) DEFAULT NULL,
                               `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '-1:进行中,0:待打包,1:成功;2:失败',
                               `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                               `scheduled_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4;
