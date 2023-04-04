/*
 Navicat Premium Data Transfer

 Source Server         : openserver
 Source Server Type    : MySQL
 Source Server Version : 100322
 Source Host           : 127.0.0.1:3306
 Source Schema         : testevents

 Target Server Type    : MySQL
 Target Server Version : 100322
 File Encoding         : 65001

 Date: 29/01/2022 09:26:28
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for demos
-- ----------------------------
DROP TABLE IF EXISTS `demos`;
CREATE TABLE `demos`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `faceitId` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `team1` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `team2` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `mapId` int NULL DEFAULT NULL,
  `parsedAt` timestamp(0) NULL DEFAULT NULL,
  `startedAt` timestamp(0) NULL DEFAULT NULL,
  `endedAt` timestamp(0) NULL DEFAULT NULL,
  `path` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_demos_maps`(`mapId`) USING BTREE,
  CONSTRAINT `fk_demos_maps` FOREIGN KEY (`mapId`) REFERENCES `info_maps` (`id`) ON DELETE NO ACTION ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of demos
-- ----------------------------

-- ----------------------------
-- Table structure for events_bomb
-- ----------------------------
DROP TABLE IF EXISTS `events_bomb`;
CREATE TABLE `events_bomb`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `demoId` bigint NULL DEFAULT NULL,
  `isDefusing` bit(1) NULL DEFAULT NULL,
  `playerSteamId` bigint UNSIGNED NULL DEFAULT NULL,
  `playerName` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `playerFaceitLevel` int NULL DEFAULT NULL,
  `playerHealth` smallint NULL DEFAULT NULL,
  `aliveCT` smallint NULL DEFAULT NULL,
  `aliveT` smallint NULL DEFAULT NULL,
  `flags` int NULL DEFAULT NULL,
  `timeRound` int NULL DEFAULT NULL,
  `timeSecond` double NULL DEFAULT NULL,
  `timeFrame` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of events_bomb
-- ----------------------------

-- ----------------------------
-- Table structure for events_kill
-- ----------------------------
DROP TABLE IF EXISTS `events_kill`;
CREATE TABLE `events_kill`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `demoId` int NULL DEFAULT NULL,
  `killerSteamId` bigint UNSIGNED NULL DEFAULT NULL,
  `killerName` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `killerHealth` smallint NULL DEFAULT NULL,
  `killerFaceitLevel` int NULL DEFAULT NULL,
  `victimSteamId` bigint UNSIGNED NULL DEFAULT NULL,
  `victimName` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  `victimFaceitLevel` int NULL DEFAULT NULL,
  `weaponId` int NULL DEFAULT NULL,
  `timeRound` int NULL DEFAULT NULL,
  `timeSecond` double NULL DEFAULT NULL,
  `timeFrame` int NULL DEFAULT NULL,
  `aliveCT` smallint NULL DEFAULT NULL,
  `aliveT` smallint NULL DEFAULT NULL,
  `flags` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `fk_eventkill_weapon`(`weaponId`) USING BTREE,
  CONSTRAINT `fk_eventkill_weapon` FOREIGN KEY (`weaponId`) REFERENCES `info_weapons` (`id`) ON DELETE NO ACTION ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of events_kill
-- ----------------------------

-- ----------------------------
-- Table structure for info_maps
-- ----------------------------
DROP TABLE IF EXISTS `info_maps`;
CREATE TABLE `info_maps`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of info_maps
-- ----------------------------
INSERT INTO `info_maps` VALUES (1, 'de_mirage');
INSERT INTO `info_maps` VALUES (2, 'de_dust2');
INSERT INTO `info_maps` VALUES (3, 'de_vertigo');
INSERT INTO `info_maps` VALUES (4, 'de_train');
INSERT INTO `info_maps` VALUES (5, 'de_inferno');
INSERT INTO `info_maps` VALUES (6, 'de_nuke');
INSERT INTO `info_maps` VALUES (7, 'de_overpass');
INSERT INTO `info_maps` VALUES (8, 'de_ancient');

-- ----------------------------
-- Table structure for info_weapons
-- ----------------------------
DROP TABLE IF EXISTS `info_weapons`;
CREATE TABLE `info_weapons`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 40 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of info_weapons
-- ----------------------------
INSERT INTO `info_weapons` VALUES (1, 'AK-47');
INSERT INTO `info_weapons` VALUES (2, 'AUG');
INSERT INTO `info_weapons` VALUES (3, 'AWP');
INSERT INTO `info_weapons` VALUES (4, 'CZ75 Auto');
INSERT INTO `info_weapons` VALUES (5, 'Desert Eagle');
INSERT INTO `info_weapons` VALUES (6, 'Dual Berettas');
INSERT INTO `info_weapons` VALUES (7, 'FAMAS');
INSERT INTO `info_weapons` VALUES (8, 'Five-SeveN');
INSERT INTO `info_weapons` VALUES (9, 'G3SG1');
INSERT INTO `info_weapons` VALUES (10, 'Galil AR');
INSERT INTO `info_weapons` VALUES (11, 'Glock-18');
INSERT INTO `info_weapons` VALUES (12, 'HE Grenade');
INSERT INTO `info_weapons` VALUES (13, 'Incendiary Grenade');
INSERT INTO `info_weapons` VALUES (14, 'Knife');
INSERT INTO `info_weapons` VALUES (15, 'M249');
INSERT INTO `info_weapons` VALUES (16, 'M4A1');
INSERT INTO `info_weapons` VALUES (17, 'M4A4');
INSERT INTO `info_weapons` VALUES (18, 'MAC-10');
INSERT INTO `info_weapons` VALUES (19, 'MAG-7');
INSERT INTO `info_weapons` VALUES (20, 'Molotov');
INSERT INTO `info_weapons` VALUES (21, 'MP5-SD');
INSERT INTO `info_weapons` VALUES (22, 'MP7');
INSERT INTO `info_weapons` VALUES (23, 'MP9');
INSERT INTO `info_weapons` VALUES (24, 'Negev');
INSERT INTO `info_weapons` VALUES (25, 'Nova');
INSERT INTO `info_weapons` VALUES (26, 'P2000');
INSERT INTO `info_weapons` VALUES (27, 'P250');
INSERT INTO `info_weapons` VALUES (28, 'P90');
INSERT INTO `info_weapons` VALUES (29, 'PP-Bizon');
INSERT INTO `info_weapons` VALUES (30, 'R8 Revolver');
INSERT INTO `info_weapons` VALUES (31, 'Sawed-Off');
INSERT INTO `info_weapons` VALUES (32, 'SCAR-20');
INSERT INTO `info_weapons` VALUES (33, 'SG 553');
INSERT INTO `info_weapons` VALUES (34, 'SSG 08');
INSERT INTO `info_weapons` VALUES (35, 'Tec-9');
INSERT INTO `info_weapons` VALUES (36, 'UMP-45');
INSERT INTO `info_weapons` VALUES (37, 'USP-S');
INSERT INTO `info_weapons` VALUES (38, 'XM1014');
INSERT INTO `info_weapons` VALUES (39, 'Zeus x27');

SET FOREIGN_KEY_CHECKS = 1;
