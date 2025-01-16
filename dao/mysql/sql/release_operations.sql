/*
Navicat MySQL Data Transfer

Source Server         : 10.0.0.180
Source Server Version : 50744
Source Host           : 10.0.0.180:3306
Source Database       : autoDeploy

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2024-12-23 17:14:41
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for release_operations
-- ----------------------------
DROP TABLE IF EXISTS `release_operations`;
CREATE TABLE `release_operations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` varchar(64) NOT NULL,
  `build_number` int(11) DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:待打包,1:成功;2:失败',
  `host` varchar(128) NOT NULL,
  `rm_rulepackage` tinyint(1) NOT NULL DEFAULT '0',
  `pkg_name` varchar(64) DEFAULT NULL,
  `scheduled_time` timestamp  NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4;
