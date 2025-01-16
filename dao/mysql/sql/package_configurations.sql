/*
Navicat MySQL Data Transfer

Source Server         : 10.0.0.180
Source Server Version : 50744
Source Host           : 10.0.0.180:3306
Source Database       : autoDeploy

Target Server Type    : MYSQL
Target Server Version : 50744
File Encoding         : 65001

Date: 2025-01-08 14:29:33
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for package_configurations
-- ----------------------------
DROP TABLE IF EXISTS `package_configurations`;
CREATE TABLE `package_configurations` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `task_id` varchar(64) NOT NULL,
  `config_type` varchar(64) NOT NULL,
  `config_content` text,
  `config_action` varchar(64) NOT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;
