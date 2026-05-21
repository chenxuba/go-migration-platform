-- MySQL dump 10.13  Distrib 8.0.46, for Linux (aarch64)
--
-- Host: localhost    Database: ybk_rebuild_edu
-- ------------------------------------------------------
-- Server version	8.0.46

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `approval_history`
--

DROP TABLE IF EXISTS `approval_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_history` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `approval_id` bigint DEFAULT NULL COMMENT '审批ID',
  `step` int DEFAULT NULL COMMENT '审批步骤',
  `approval_person` bigint DEFAULT NULL COMMENT '审批人',
  `approval_time` datetime DEFAULT NULL COMMENT '审批时间',
  `approval_status` int DEFAULT NULL COMMENT '审批状态',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=93 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='审批历史记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `approval_record`
--

DROP TABLE IF EXISTS `approval_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `approval_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `order_id` bigint DEFAULT NULL COMMENT '订单编号',
  `student_id` bigint DEFAULT NULL COMMENT '学生ID',
  `approval_number` varchar(255) DEFAULT NULL COMMENT '审批编号',
  `config_version` int DEFAULT NULL COMMENT '审批版本',
  `applicant` bigint DEFAULT NULL COMMENT '申请人',
  `current_approver` varchar(255) DEFAULT NULL COMMENT '当前审批人',
  `current_step` int DEFAULT NULL COMMENT '审批步骤',
  `approval_type` int DEFAULT NULL COMMENT '审批类型',
  `approval_status` int DEFAULT NULL COMMENT '审批状态',
  `approval_time` datetime DEFAULT NULL COMMENT '申请时间',
  `finish_time` datetime DEFAULT NULL COMMENT '完成时间',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  `initiate_reason` varchar(1000) DEFAULT NULL COMMENT '审批发起时的触发条件快照',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=86 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='审批单主记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_caregiver_invite`
--

DROP TABLE IF EXISTS `assessment_caregiver_invite`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_caregiver_invite` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `ticket` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `draft_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `wechat_url_link` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `mini_program_code_data_url` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `expires_at` datetime DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_caregiver_invite_ticket` (`ticket`),
  KEY `idx_assessment_caregiver_invite_draft` (`inst_id`,`draft_id`,`update_time`),
  KEY `idx_assessment_caregiver_invite_record` (`inst_id`,`record_id`,`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=232 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估照顾者报告邀请表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_draft`
--

DROP TABLE IF EXISTS `assessment_draft`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_draft` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `birth_date` date DEFAULT NULL,
  `assessment_date` date DEFAULT NULL,
  `examiner_id` bigint NOT NULL DEFAULT '0',
  `examiner_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `input_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `progress_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `answered_item_count` int NOT NULL DEFAULT '0',
  `raw_score_count` int NOT NULL DEFAULT '0',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft',
  `submitted_record_id` bigint NOT NULL DEFAULT '0',
  `remark` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_assessment_draft_inst_code_date` (`inst_id`,`assessment_code`,`assessment_date`,`id`),
  KEY `idx_assessment_draft_inst_student` (`inst_id`,`student_id`,`update_time`,`id`),
  KEY `idx_assessment_draft_status` (`inst_id`,`assessment_code`,`status`,`update_time`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=84 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估草稿表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_draft_item_record_value`
--

DROP TABLE IF EXISTS `assessment_draft_item_record_value`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_draft_item_record_value` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `draft_id` bigint NOT NULL DEFAULT '0',
  `item_no` int NOT NULL DEFAULT '0',
  `field_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `value_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_draft_item_record_value` (`draft_id`,`item_no`,`field_key`),
  KEY `idx_assessment_draft_item_record_inst` (`inst_id`,`draft_id`,`item_no`)
) ENGINE=InnoDB AUTO_INCREMENT=777 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估草稿题目记录值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_draft_item_score`
--

DROP TABLE IF EXISTS `assessment_draft_item_score`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_draft_item_score` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `draft_id` bigint NOT NULL DEFAULT '0',
  `item_no` int NOT NULL DEFAULT '0',
  `score` int NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_draft_item_score` (`draft_id`,`item_no`),
  KEY `idx_assessment_draft_item_score_inst` (`inst_id`,`draft_id`,`item_no`)
) ENGINE=InnoDB AUTO_INCREMENT=28192 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估草稿题目得分表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_draft_raw_score`
--

DROP TABLE IF EXISTS `assessment_draft_raw_score`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_draft_raw_score` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `draft_id` bigint NOT NULL DEFAULT '0',
  `scale_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `raw_score` int NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_draft_raw_score` (`draft_id`,`scale_code`),
  KEY `idx_assessment_draft_raw_score_inst` (`inst_id`,`draft_id`,`scale_code`)
) ENGINE=InnoDB AUTO_INCREMENT=146 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估草稿原始分表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_execution_generation_task`
--

DROP TABLE IF EXISTS `assessment_iep_execution_generation_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_execution_generation_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `user_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `assessment_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `duration_months` int NOT NULL DEFAULT '3',
  `plan_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `target_month_index` int NOT NULL DEFAULT '0',
  `target_week_index` int NOT NULL DEFAULT '0',
  `rest_weekdays_json` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '[]',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `message` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `stream_text` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `usage_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `cost_amount_cny` decimal(12,6) NOT NULL DEFAULT '0.000000',
  `monthly_plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `weekly_plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `saved_execution_plans_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `error_message` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_execution_generation_task_task` (`task_id`,`del_flag`),
  KEY `idx_assessment_iep_execution_generation_task_active` (`inst_id`,`user_id`,`record_id`,`assessment_type`,`plan_type`,`target_month_index`,`target_week_index`,`status`,`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP执行生成任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_execution_plan`
--

DROP TABLE IF EXISTS `assessment_iep_execution_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_execution_plan` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `duration_months` int NOT NULL DEFAULT '3',
  `plan_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `target_month_index` int NOT NULL DEFAULT '0',
  `target_week_index` int NOT NULL DEFAULT '0',
  `plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_execution_plan_target` (`inst_id`,`record_id`,`duration_months`,`plan_type`,`target_month_index`,`target_week_index`,`del_flag`),
  KEY `idx_assessment_iep_execution_plan_record` (`inst_id`,`record_id`,`duration_months`,`plan_type`)
) ENGINE=InnoDB AUTO_INCREMENT=454 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP执行计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_generation_task`
--

DROP TABLE IF EXISTS `assessment_iep_generation_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_generation_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `user_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `assessment_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `duration_months` int NOT NULL DEFAULT '3',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `message` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `stream_text` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `usage_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `cost_amount_cny` decimal(12,6) NOT NULL DEFAULT '0.000000',
  `plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `saved_plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `error_message` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_generation_task_task` (`task_id`,`del_flag`),
  KEY `idx_assessment_iep_generation_task_active` (`inst_id`,`user_id`,`record_id`,`assessment_type`,`status`,`update_time`),
  KEY `idx_assessment_iep_generation_task_record` (`inst_id`,`record_id`,`assessment_type`,`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP计划生成任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_lesson_record`
--

DROP TABLE IF EXISTS `assessment_iep_lesson_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_lesson_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `duration_months` int NOT NULL DEFAULT '3',
  `target_month_index` int NOT NULL DEFAULT '0',
  `target_week_index` int NOT NULL DEFAULT '0',
  `lesson_date` date NOT NULL,
  `week_date_index` int NOT NULL DEFAULT '0',
  `weekly_row_index` int NOT NULL DEFAULT '0',
  `project` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `completion_code` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `teacher_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `course_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_lesson_record_target` (`inst_id`,`record_id`,`duration_months`,`target_month_index`,`target_week_index`,`lesson_date`,`weekly_row_index`,`del_flag`),
  KEY `idx_assessment_iep_lesson_record_week` (`inst_id`,`record_id`,`duration_months`,`target_month_index`,`target_week_index`,`lesson_date`),
  KEY `idx_assessment_iep_lesson_record_student_date` (`inst_id`,`student_id`,`lesson_date`)
) ENGINE=InnoDB AUTO_INCREMENT=2130 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP课次记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_lesson_session`
--

DROP TABLE IF EXISTS `assessment_iep_lesson_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_lesson_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `duration_months` int NOT NULL DEFAULT '3',
  `target_month_index` int NOT NULL DEFAULT '0',
  `target_week_index` int NOT NULL DEFAULT '0',
  `lesson_date` date NOT NULL,
  `week_date_index` int NOT NULL DEFAULT '0',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `elapsed_seconds` int NOT NULL DEFAULT '0',
  `started_at` datetime DEFAULT NULL,
  `last_resumed_at` datetime DEFAULT NULL,
  `last_heartbeat_at` datetime DEFAULT NULL,
  `paused_at` datetime DEFAULT NULL,
  `ended_at` datetime DEFAULT NULL,
  `operator_id` bigint NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_lesson_session_target` (`inst_id`,`record_id`,`duration_months`,`target_month_index`,`target_week_index`,`lesson_date`,`del_flag`),
  KEY `idx_assessment_iep_lesson_session_week` (`inst_id`,`record_id`,`duration_months`,`target_month_index`,`target_week_index`,`del_flag`),
  KEY `idx_assessment_iep_lesson_session_status` (`inst_id`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=89 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP课次会话表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_iep_plan`
--

DROP TABLE IF EXISTS `assessment_iep_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_iep_plan` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `duration_months` int NOT NULL DEFAULT '3',
  `status` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'draft',
  `plan_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_iep_plan_record_duration` (`inst_id`,`record_id`,`duration_months`,`del_flag`),
  KEY `idx_assessment_iep_plan_status` (`inst_id`,`status`,`update_time`),
  KEY `idx_assessment_iep_plan_record_status` (`inst_id`,`record_id`,`status`)
) ENGINE=InnoDB AUTO_INCREMENT=172 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估IEP计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_record`
--

DROP TABLE IF EXISTS `assessment_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assessment_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `birth_date` date DEFAULT NULL,
  `assessment_date` date DEFAULT NULL,
  `age_years` int NOT NULL DEFAULT '0',
  `age_months` int NOT NULL DEFAULT '0',
  `age_days` int NOT NULL DEFAULT '0',
  `norm_age_months` int NOT NULL DEFAULT '0',
  `examiner_id` bigint NOT NULL DEFAULT '0',
  `examiner_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `input_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `result_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `data_status` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `remark` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_assessment_record_inst_code_date` (`inst_id`,`assessment_code`,`assessment_date`,`id`),
  KEY `idx_assessment_record_inst_student` (`inst_id`,`student_id`,`assessment_date`,`id`),
  KEY `idx_assessment_record_created` (`inst_id`,`create_time`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_report_interpretation`
--

DROP TABLE IF EXISTS `assessment_report_interpretation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_report_interpretation` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `record_id` bigint NOT NULL DEFAULT '0',
  `assessment_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `source_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `model` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `generated_by` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_report_interpretation_record` (`inst_id`,`record_id`,`assessment_code`),
  KEY `idx_assessment_report_interpretation_update` (`inst_id`,`assessment_code`,`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估报告解读表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_scale_dataset`
--

DROP TABLE IF EXISTS `assessment_scale_dataset`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_scale_dataset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `data_status` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sources_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `metadata_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_scale_dataset` (`scale_code`,`scale_version`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估量表数据集表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_scale_domain`
--

DROP TABLE IF EXISTS `assessment_scale_domain`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_scale_domain` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `domain_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sort_no` int NOT NULL DEFAULT '0',
  `domain_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_scale_domain` (`scale_code`,`scale_version`,`domain_code`),
  KEY `idx_assessment_scale_domain_version` (`scale_code`,`scale_version`,`sort_no`)
) ENGINE=InnoDB AUTO_INCREMENT=89 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估量表领域表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_scale_item`
--

DROP TABLE IF EXISTS `assessment_scale_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_scale_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `item_no` int NOT NULL DEFAULT '0',
  `item_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_scale_item` (`scale_code`,`scale_version`,`item_no`),
  KEY `idx_assessment_scale_item_version` (`scale_code`,`scale_version`,`item_no`)
) ENGINE=InnoDB AUTO_INCREMENT=2056 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估量表题目表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_scale_item_record_field`
--

DROP TABLE IF EXISTS `assessment_scale_item_record_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_scale_item_record_field` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_code` varchar(64) NOT NULL DEFAULT '',
  `scale_version` varchar(64) NOT NULL DEFAULT '',
  `item_no` int NOT NULL DEFAULT '0',
  `field_key` varchar(100) NOT NULL DEFAULT '',
  `sort_no` int NOT NULL DEFAULT '0',
  `field_json` longtext NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_scale_item_record_field` (`scale_code`,`scale_version`,`item_no`,`field_key`),
  KEY `idx_assessment_scale_item_record_field_item` (`scale_code`,`scale_version`,`item_no`,`sort_no`)
) ENGINE=InnoDB AUTO_INCREMENT=509 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='评估量表题目记录字段表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessment_scale_norm_record`
--

DROP TABLE IF EXISTS `assessment_scale_norm_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `assessment_scale_norm_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `record_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sort_no` int NOT NULL DEFAULT '0',
  `norm_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_scale_norm_record` (`scale_code`,`scale_version`,`record_key`),
  KEY `idx_assessment_scale_norm_version` (`scale_code`,`scale_version`,`sort_no`)
) ENGINE=InnoDB AUTO_INCREMENT=26772 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评估量表常模记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `class_record_export_record`
--

DROP TABLE IF EXISTS `class_record_export_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `class_record_export_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `export_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `export_staff_id` bigint NOT NULL DEFAULT '0',
  `export_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_type` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_data` longblob NOT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `query_conditions_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_class_record_export_inst` (`inst_id`,`export_type`,`create_time`,`id`),
  KEY `idx_class_record_export_expire` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课堂记录导出记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `close_tuition_account_order`
--

DROP TABLE IF EXISTS `close_tuition_account_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `close_tuition_account_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `flow_source_id` bigint DEFAULT NULL,
  `tuition_account_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `course_id` bigint NOT NULL DEFAULT '0',
  `lesson_charging_mode` int NOT NULL DEFAULT '0',
  `quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `free_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT '1',
  `close_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `revert_valid_start_date` datetime DEFAULT NULL,
  `reverted_time` datetime DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_close_tuition_account_order_source` (`inst_id`,`flow_source_id`),
  KEY `idx_close_tuition_account_order_account` (`inst_id`,`tuition_account_id`,`del_flag`),
  KEY `idx_close_tuition_account_order_student_course` (`inst_id`,`student_id`,`course_id`,`del_flag`),
  KEY `idx_close_tuition_account_order_status` (`inst_id`,`status`,`close_time`)
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课消账户结清订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `enrolled_student_export_record`
--

DROP TABLE IF EXISTS `enrolled_student_export_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `enrolled_student_export_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `export_staff_id` bigint NOT NULL DEFAULT '0',
  `export_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_type` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_data` longblob NOT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `query_conditions_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_enrolled_student_export_inst` (`inst_id`,`create_time`,`id`),
  KEY `idx_enrolled_student_export_expire` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='在读学员导出记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `follow_record`
--

DROP TABLE IF EXISTS `follow_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `follow_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `student_id` bigint DEFAULT NULL COMMENT '学生ID',
  `follow_method` tinyint(1) DEFAULT NULL COMMENT '跟进方式',
  `content` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '沟通内容',
  `follow_images` varchar(1024) DEFAULT NULL COMMENT '跟进图片',
  `follow_up_time` datetime DEFAULT NULL COMMENT '跟进时间',
  `next_follow_up_time` datetime DEFAULT NULL COMMENT '下次跟进时间',
  `intended_course` varchar(255) DEFAULT NULL COMMENT '意向课程',
  `intention_level` int DEFAULT NULL COMMENT '意向等级',
  `follow_up_status` int DEFAULT NULL COMMENT '跟进状态',
  `visit_status` tinyint(1) DEFAULT NULL COMMENT '回访状态',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='跟进记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `government_user_profile`
--

DROP TABLE IF EXISTS `government_user_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `government_user_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `version` bigint NOT NULL DEFAULT '0',
  `user_id` bigint NOT NULL,
  `level` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `disabled` tinyint(1) NOT NULL DEFAULT '0',
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_government_user_profile_user_del` (`user_id`,`del_flag`),
  KEY `idx_government_user_profile_level_del` (`level`,`del_flag`),
  KEY `idx_government_user_profile_disabled_del` (`disabled`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='政府端用户扩展信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `government_user_scope`
--

DROP TABLE IF EXISTS `government_user_scope`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `government_user_scope` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `version` bigint NOT NULL DEFAULT '0',
  `user_id` bigint NOT NULL,
  `scope_level` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `province_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `province_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `city_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `city_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `district_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `district_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_government_user_scope_user_del` (`user_id`,`del_flag`),
  KEY `idx_government_user_scope_level_del` (`scope_level`,`del_flag`),
  KEY `idx_government_user_scope_code_del` (`province_code`,`city_code`,`district_code`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='政府端用户数据权限范围表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `homework_task`
--

DROP TABLE IF EXISTS `homework_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `homework_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `attachments_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `repeat_rule_json` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `publish_rule` int NOT NULL DEFAULT '1',
  `publish_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `publish_hour` int NOT NULL DEFAULT '0',
  `task_duration_hours` int NOT NULL DEFAULT '0',
  `end_hour` int NOT NULL DEFAULT '0',
  `is_visible_student` tinyint(1) NOT NULL DEFAULT '0',
  `source_type` int NOT NULL DEFAULT '1',
  `source_id` bigint NOT NULL DEFAULT '0',
  `source_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `selected_students_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `student_count` int NOT NULL DEFAULT '0',
  `unsubmitted_count` int NOT NULL DEFAULT '0',
  `rejected_count` int NOT NULL DEFAULT '0',
  `submitted_count` int NOT NULL DEFAULT '0',
  `re_submitted_count` int NOT NULL DEFAULT '0',
  `evaluated_count` int NOT NULL DEFAULT '0',
  `unevaluated_count` int NOT NULL DEFAULT '0',
  `read_count` int NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_homework_inst_publish` (`inst_id`,`publish_time`,`id`),
  KEY `idx_homework_inst_end` (`inst_id`,`end_time`,`id`),
  KEY `idx_homework_inst_source` (`inst_id`,`source_type`,`source_id`),
  KEY `idx_homework_inst_creator` (`inst_id`,`create_id`),
  KEY `idx_homework_inst_deleted` (`inst_id`,`del_flag`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭作业任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_approval_config`
--

DROP TABLE IF EXISTS `inst_approval_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_approval_config` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `name` varchar(255) DEFAULT NULL COMMENT '功能名称',
  `type` int DEFAULT NULL COMMENT '功能类型 1报名续费 2转课 3退课 4储值充值 5储值退费',
  `enable` tinyint(1) DEFAULT NULL COMMENT '是否开启',
  `rule_json` varchar(255) DEFAULT NULL COMMENT '审批条件',
  `config_version` int DEFAULT '0' COMMENT '配置版本',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构审批配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_approval_flow`
--

DROP TABLE IF EXISTS `inst_approval_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_approval_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `config_id` bigint DEFAULT NULL COMMENT '审批配置ID',
  `config_version` int DEFAULT NULL COMMENT '历史版本',
  `staff_id` varchar(100) DEFAULT NULL COMMENT '审批人',
  `staff_name` varchar(255) DEFAULT NULL COMMENT '审批人名称',
  `step` int DEFAULT NULL COMMENT '审批步骤',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=121 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构审批流程表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_channel`
--

DROP TABLE IF EXISTS `inst_channel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字段定义主键ID',
  `uuid` varchar(36) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '所属机构ID',
  `category_id` bigint DEFAULT NULL COMMENT '分类ID',
  `channel_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '渠道名称',
  `introduction` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '渠道简介',
  `is_default` tinyint(1) DEFAULT NULL COMMENT '是否系统默认',
  `is_disabled` tinyint(1) DEFAULT NULL COMMENT '是否停用',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='机构渠道表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_channel_category`
--

DROP TABLE IF EXISTS `inst_channel_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_channel_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字段定义主键ID',
  `uuid` varchar(36) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '所属机构ID',
  `category_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分类名称',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=50 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='机构渠道分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_classroom`
--

DROP TABLE IF EXISTS `inst_classroom`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_classroom` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_classroom_inst` (`inst_id`),
  KEY `idx_inst_classroom_name` (`name`),
  KEY `idx_inst_classroom_enabled` (`enabled`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构教室表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_compose_lesson`
--

DROP TABLE IF EXISTS `inst_compose_lesson`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_compose_lesson` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_icl_inst_del` (`inst_id`,`del_flag`),
  KEY `idx_icl_inst_ctime` (`inst_id`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构组合课表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_compose_lesson_product`
--

DROP TABLE IF EXISTS `inst_compose_lesson_product`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_compose_lesson_product` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `compose_lesson_id` bigint NOT NULL,
  `course_id` bigint NOT NULL,
  `sort_order` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_iclp_compose` (`compose_lesson_id`),
  KEY `idx_iclp_inst` (`inst_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构组合课产品表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_config`
--

DROP TABLE IF EXISTS `inst_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_config` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT NULL COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `school_home_banner` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '学校首页横幅设置',
  `auto_send_birthday_message` tinyint(1) DEFAULT NULL COMMENT '是否自动发送生日祝福消息',
  `deduct_when_leave` tinyint(1) DEFAULT NULL COMMENT '离开时是否扣除费用',
  `deduct_when_truancy` tinyint(1) DEFAULT NULL COMMENT '旷课时是否扣除费用',
  `default_class_time` int DEFAULT NULL COMMENT '默认课程时长（单位：小时）',
  `enabled_one2one` tinyint(1) DEFAULT NULL COMMENT '是否启用一对一课程',
  `enable_leave_apply_number_limit` tinyint(1) DEFAULT NULL COMMENT '是否启用请假申请次数限制',
  `leave_apply_type_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'course',
  `leave_apply_cycle_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'month',
  `leave_apply_number_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '2',
  `enable_leave_apply_time_limit` tinyint(1) DEFAULT NULL COMMENT '是否启用请假申请时间限制',
  `leave_apply_time_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '1.0',
  `enable_leave_deduct_money` tinyint(1) DEFAULT NULL COMMENT '请假是否扣除费用',
  `enable_truant_deduct_money` tinyint(1) DEFAULT NULL COMMENT '旷课是否扣除费用',
  `enabled_class_reminder` tinyint(1) DEFAULT NULL COMMENT '是否启用上课提醒',
  `enabled_class_consumption_reminder` tinyint(1) DEFAULT NULL COMMENT '是否启用上课消费提醒',
  `enabled_renew_reminder` tinyint(1) DEFAULT NULL COMMENT '是否启用续费提醒',
  `enable_renew_class_num` tinyint(1) DEFAULT NULL COMMENT '是否根据剩余课时数触发续费提醒',
  `enable_renew_validity_day` tinyint(1) DEFAULT NULL COMMENT '是否根据有效期天数触发续费提醒',
  `renew_class_num` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '5',
  `renew_validity_day` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '15',
  `enable_by_date_lesson` tinyint(1) DEFAULT NULL COMMENT '是否按日期查看课程',
  `enable_by_date_student_absent_record` tinyint(1) DEFAULT NULL COMMENT '是否按日期查看学生缺勤记录',
  `enable_by_date_student_teaching_record` tinyint(1) DEFAULT NULL COMMENT '是否按日期查看学生教学记录',
  `enable_general_lesson` tinyint(1) DEFAULT NULL COMMENT '是否启用普通课程',
  `enable_by_face_attendance` tinyint(1) DEFAULT NULL COMMENT '是否启用面部识别考勤',
  `enable_by_voice_tips` tinyint(1) DEFAULT NULL COMMENT '是否启用语音提示',
  `enable_face_attendance_relate_teaching` tinyint(1) DEFAULT NULL COMMENT '面部识别考勤是否关联教学记录',
  `enable_face_attendance_check_in_notice` tinyint(1) DEFAULT NULL COMMENT '是否启用签到通知',
  `enable_face_attendance_check_out_notice` tinyint(1) DEFAULT NULL COMMENT '是否启用签退通知',
  `face_attendance_interval` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '1',
  `face_attendance_split` int DEFAULT NULL COMMENT '面部识别考勤分割规则',
  `face_attendance_relate_rule` int DEFAULT NULL COMMENT '面部识别考勤关联规则',
  `enable_by_auto_teaching` tinyint(1) DEFAULT NULL COMMENT '是否启用自动教学',
  `enable_charge_by_price` tinyint(1) DEFAULT NULL COMMENT '是否启用按价格收费',
  `enable_charge_by_price_student_absent_record` tinyint(1) DEFAULT NULL COMMENT '是否启用按价格收费的学生缺勤记录',
  `charge_by_price_default_price` decimal(10,2) DEFAULT NULL COMMENT '按价格收费的默认价格',
  `enable_renew_price` tinyint(1) DEFAULT NULL COMMENT '是否启用续费价格',
  `renew_price` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '500',
  `enable_goods_management` tinyint(1) DEFAULT NULL COMMENT '是否启用商品管理',
  `enable_adjust_tuition_account_order` tinyint(1) DEFAULT NULL COMMENT '是否启用调整学费账户订单',
  `tuition_account_priority` int DEFAULT NULL COMMENT '学费账户优先级',
  `enable_recharge_account_change_message` tinyint(1) DEFAULT NULL COMMENT '是否启用充值账户变动消息',
  `enable_send_coupon_remind_sms` tinyint(1) DEFAULT NULL COMMENT '是否启用发送优惠券提醒短信',
  `enable_teaching_bill_remind_sms` tinyint(1) DEFAULT NULL COMMENT '是否启用教学账单提醒短信',
  `enable_compensation_send_message` tinyint(1) DEFAULT NULL COMMENT '是否启用补偿消息发送',
  `enable_arrearaged_send_message` tinyint(1) DEFAULT NULL COMMENT '是否启用欠费消息发送',
  `enable_tran_order_finished_send_message` tinyint(1) DEFAULT NULL COMMENT '是否启用转账订单完成消息发送',
  `enable_compose_lesson` tinyint(1) DEFAULT NULL COMMENT '是否启用组合课程',
  `enable_show_left_tuition` tinyint(1) DEFAULT NULL COMMENT '是否显示剩余学费',
  `enable_filter_holiday` tinyint(1) DEFAULT NULL COMMENT '是否过滤节假日',
  `book_lesson_locked_hours` int DEFAULT NULL COMMENT '预约课程锁定时间（单位：小时）',
  `book_lesson_opening_days` int DEFAULT NULL COMMENT '预约课程开放天数',
  `enabled_show_book_lesson_student_count` tinyint(1) DEFAULT NULL COMMENT '是否显示预约课程学生人数',
  `book_lesson_time` int DEFAULT NULL COMMENT '预约课程时长（单位：小时）',
  `enabled_book_lesson_excess` tinyint(1) DEFAULT NULL COMMENT '是否启用超额预约课程',
  `enable_book_lesson_times` tinyint(1) DEFAULT NULL COMMENT '是否启用预约课程次数限制',
  `book_lesson_times` int DEFAULT NULL COMMENT '预约课程次数限制',
  `book_lesson_times_cycle` int DEFAULT NULL COMMENT '预约课程次数限制周期',
  `enable_liquidation_remind_message` tinyint(1) DEFAULT NULL COMMENT '是否启用清算提醒消息',
  `no_saler_pool_day` int DEFAULT NULL COMMENT '销售人员池保留天数',
  `enabled_arrears_rollcall` tinyint(1) DEFAULT NULL COMMENT '是否启用欠费点名',
  `send_class_reminder_sms_hour` int DEFAULT NULL COMMENT '发送上课提醒短信的时间（单位：小时）',
  `send_class_reminder_msg_hour` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '19:00',
  `auto_assign_performance` tinyint(1) DEFAULT NULL COMMENT '是否自动分配业绩',
  `audition_record_automate` int DEFAULT NULL COMMENT '试听记录自动化设置',
  `add_intention_student_rule` int DEFAULT NULL COMMENT '添加意向学生的规则',
  `add_import_student_rule` int DEFAULT NULL COMMENT '导入学生规则',
  `default_student_class_time` int DEFAULT NULL COMMENT '默认学生课程时长（单位：小时）',
  `default_teacher_class_time` int DEFAULT NULL COMMENT '默认教师课程时长（单位：小时）',
  `student_absent_class_switch` tinyint(1) DEFAULT NULL COMMENT '学生缺勤课时开关',
  `student_absent_class_value` int DEFAULT NULL COMMENT '学生缺勤课时值',
  `arrive_class_switch` tinyint(1) DEFAULT NULL COMMENT '到达班级开关',
  `arrive_class_days_config` int DEFAULT NULL COMMENT '到达班级天数配置',
  `arrive_class_times_config` int DEFAULT NULL COMMENT '到达班级次数配置',
  `enable_show_recharge_account_balance` tinyint(1) DEFAULT NULL COMMENT '是否显示充值账户余额',
  `enable_point_change_remind_message` tinyint(1) DEFAULT NULL COMMENT '是否启用积分变动提醒消息',
  `should_student_source_type_config` int DEFAULT NULL COMMENT '应设置的学生来源类型',
  `actual_student_source_type_config` int DEFAULT NULL COMMENT '实际设置的学生来源类型',
  `enable_show_arrears_information` tinyint(1) DEFAULT NULL COMMENT '是否显示欠费信息',
  `time_table_changed_send_to_c_switch` tinyint(1) DEFAULT NULL COMMENT '时间表更改后发送给客服开关',
  `enable_grade_upgrade` tinyint(1) DEFAULT NULL COMMENT '是否启用年级升级',
  `enable_space_booking_notice` tinyint(1) DEFAULT NULL COMMENT '是否启用场地预订通知',
  `enable_show_school_on_order_receipt` tinyint(1) DEFAULT NULL COMMENT '是否在订单收据上显示学校名称',
  `enable_subject` tinyint(1) DEFAULT NULL COMMENT '是否启用科目',
  `enable_subject_online_sale_filter` tinyint(1) DEFAULT NULL COMMENT '是否启用在线销售科目的筛选',
  `enable_auto_deduct_stock` tinyint(1) DEFAULT NULL COMMENT '是否启用自动扣减库存',
  `enable_order_tag_required` tinyint(1) DEFAULT NULL COMMENT '是否启用订单标签必填项',
  `enable_left_class_time_remind` tinyint(1) DEFAULT NULL COMMENT '是否启用剩余课时提醒',
  `enable_audition_sms_remind` tinyint(1) DEFAULT NULL COMMENT '是否启用试听短信提醒',
  `enable_student_parent_transcript_chart` tinyint(1) DEFAULT NULL COMMENT '是否启用学生成绩单图表',
  `enable_class_comment_parent_feedback` tinyint(1) DEFAULT NULL COMMENT '是否启用班级评论家长反馈',
  `class_comment_parent_feedback_type` int DEFAULT NULL COMMENT '班级评论家长反馈类型',
  `discounts_mode` int DEFAULT NULL COMMENT '折扣模式',
  `enable_custom_sku` tinyint(1) DEFAULT NULL COMMENT '是否启用自定义SKU',
  `limit_same_weChat` tinyint(1) DEFAULT NULL COMMENT '是否限制相同的微信账号',
  `limit_import_same_weChat` tinyint(1) DEFAULT NULL COMMENT '是否限制导入相同的微信账号',
  `enable_send_face_attend_notice_to_parent` tinyint(1) DEFAULT NULL COMMENT '是否向家长发送面部识别考勤通知',
  `enable_send_face_attend_notice_to_admin` tinyint(1) DEFAULT NULL COMMENT '是否向管理员发送面部识别考勤通知',
  `enable_send_child_bind_notice_to_admin` tinyint(1) DEFAULT NULL COMMENT '是否向管理员发送孩子绑定通知',
  `allow_original_refund` tinyint(1) DEFAULT NULL COMMENT '是否允许原路退款',
  `enable_timetable_time_config` tinyint(1) DEFAULT NULL COMMENT '是否启用时间表时间配置',
  `enable_refund_zero` tinyint(1) DEFAULT NULL COMMENT '是否启用零金额退款',
  `enable_org_send_face_attend_notice_to_admin` tinyint(1) DEFAULT NULL COMMENT '是否组织向管理员发送面部识别考勤通知',
  `enable_org_send_child_bind_notice_to_admin` tinyint(1) DEFAULT NULL COMMENT '是否组织向管理员发送孩子绑定通知',
  `enable_peer_info_and_service_management` tinyint(1) DEFAULT NULL COMMENT '是否启用同行信息和服务管理',
  `maximum_class_size_policy` int DEFAULT NULL COMMENT '最大班级规模策略',
  `micro_school_order_expire_minutes` int DEFAULT NULL COMMENT '微校订单过期时间（单位：分钟）',
  `max_micro_school_online_user_count` bigint DEFAULT NULL COMMENT '微校在线用户最大数量',
  `enable_public_pool` tinyint(1) DEFAULT NULL COMMENT '是否开启公有池',
  `unfollowed_time` int DEFAULT NULL COMMENT '超过多少天未跟进自动转换',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `enable_quick_unified_period` tinyint(1) NOT NULL DEFAULT '0',
  `unified_time_period_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `group_class_roll_call_sheet_template` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enable_classroom_teaching` tinyint(1) NOT NULL DEFAULT '1',
  `enable_charge_by_hours` tinyint(1) NOT NULL DEFAULT '1',
  `enable_one_to_one_schedule_limit` tinyint(1) NOT NULL DEFAULT '0',
  `enable_schedule_conflict_continue` tinyint(1) NOT NULL DEFAULT '1',
  `schedule_teacher_selection_range` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'teacher-only',
  `default_class_time_record_mode` int NOT NULL DEFAULT '1',
  `enable_hour_truancy_normal_record` tinyint(1) NOT NULL DEFAULT '0',
  `enable_period_auto_finish_when_zero` tinyint(1) NOT NULL DEFAULT '0',
  `enable_price_leave_normal_record` tinyint(1) NOT NULL DEFAULT '0',
  `enable_limit_single_order_arrears_deduct` tinyint(1) NOT NULL DEFAULT '0',
  `enable_period_makeup` tinyint(1) NOT NULL DEFAULT '0',
  `enable_price_truancy_normal_record` tinyint(1) NOT NULL DEFAULT '0',
  `enable_price_makeup` tinyint(1) NOT NULL DEFAULT '0',
  `enable_hour_leave_normal_record` tinyint(1) NOT NULL DEFAULT '0',
  `enable_supervisor` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='机构配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course`
--

DROP TABLE IF EXISTS `inst_course`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `type` int DEFAULT NULL COMMENT '商品类型',
  `name` varchar(255) DEFAULT NULL COMMENT '课程名称',
  `course_category` int DEFAULT NULL COMMENT '课程类别',
  `course_attribute` int DEFAULT NULL COMMENT '课程属性',
  `sale_status` tinyint(1) DEFAULT NULL COMMENT '售卖状态',
  `teach_method` int DEFAULT NULL COMMENT '授课方式',
  `sale_volume` int DEFAULT NULL COMMENT '课程销量',
  `subject_ids` varchar(255) DEFAULT NULL COMMENT '科目',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  `roll_call_deduct_price` decimal(18,2) DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=223 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_category`
--

DROP TABLE IF EXISTS `inst_course_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_category` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `name` varchar(255) DEFAULT NULL COMMENT '类别名称',
  `sort` int DEFAULT NULL COMMENT '序号',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程分类表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_detail`
--

DROP TABLE IF EXISTS `inst_course_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `course_id` bigint DEFAULT NULL COMMENT '课程ID',
  `title` varchar(255) DEFAULT NULL COMMENT '商品名称',
  `images` longtext,
  `description` longtext,
  `is_show_mico_school` tinyint(1) DEFAULT NULL COMMENT '微校展示',
  `enable_buy_limit` tinyint(1) DEFAULT NULL COMMENT '是否限制购买',
  `is_allow_returning_student` tinyint(1) DEFAULT NULL COMMENT '是否允许老生购买',
  `allow_type` int DEFAULT NULL COMMENT '允许类型',
  `relate_product_ids` json DEFAULT NULL COMMENT '关联的商品ID列表',
  `student_statuses` varchar(255) DEFAULT NULL COMMENT '购买学生状态',
  `is_allow_freshman_student` tinyint(1) DEFAULT NULL COMMENT '是否允许新生购买',
  `limit_one_per` tinyint(1) DEFAULT NULL COMMENT '限购1单',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=221 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_property`
--

DROP TABLE IF EXISTS `inst_course_property`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_property` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `name` varchar(255) DEFAULT NULL COMMENT '属性名称',
  `enable` tinyint(1) DEFAULT NULL COMMENT '启用状态',
  `enable_online_filter` tinyint(1) DEFAULT NULL COMMENT '是否在线商城支持筛选',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程属性表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_property_option`
--

DROP TABLE IF EXISTS `inst_course_property_option`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_property_option` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `property_id` bigint DEFAULT NULL COMMENT '属性ID',
  `name` varchar(255) DEFAULT NULL COMMENT '属性值',
  `sort` int DEFAULT NULL COMMENT '排序',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程属性选项表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_property_result`
--

DROP TABLE IF EXISTS `inst_course_property_result`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_property_result` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `course_id` bigint DEFAULT NULL COMMENT '课程ID',
  `course_property_id` bigint DEFAULT NULL COMMENT '属性ID',
  `property_id_name` varchar(255) DEFAULT NULL COMMENT '属性ID名称',
  `course_property_value` bigint DEFAULT NULL COMMENT '属性结果ID',
  `property_value_name` varchar(255) DEFAULT NULL COMMENT '属性结果名称',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=293 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程属性结果表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_course_quotation`
--

DROP TABLE IF EXISTS `inst_course_quotation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_course_quotation` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `course_id` bigint DEFAULT NULL COMMENT '课程ID',
  `lesson_model` int DEFAULT NULL COMMENT '报价单模式(1按课时 2按时段 3按金额)',
  `name` varchar(255) DEFAULT NULL COMMENT '报价单名称',
  `unit` int DEFAULT NULL COMMENT '计价方式(1 课时 2按天 3按月 4按年 5按金额)',
  `quantity` int DEFAULT NULL COMMENT '数量',
  `price` decimal(20,4) DEFAULT NULL COMMENT '总价金额',
  `lesson_audition` tinyint(1) DEFAULT NULL COMMENT '体验价',
  `online_sale` tinyint(1) DEFAULT NULL COMMENT '微校售卖状态',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=562 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构课程报价表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_leave_action`
--

DROP TABLE IF EXISTS `inst_leave_action`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_leave_action` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `leave_request_id` bigint NOT NULL,
  `action_type` int NOT NULL DEFAULT '0',
  `action_staff_id` bigint NOT NULL DEFAULT '0',
  `action_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status_text` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `remark` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_leave_action_leave` (`inst_id`,`leave_request_id`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构请假操作记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_leave_request`
--

DROP TABLE IF EXISTS `inst_leave_request`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_leave_request` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_avatar_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `leave_type` int NOT NULL DEFAULT '0',
  `reason` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `proof_materials_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `remark` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `schedule_count` int NOT NULL DEFAULT '0',
  `initiate_staff_id` bigint NOT NULL DEFAULT '0',
  `initiate_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `is_agent` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否代办',
  `status` int NOT NULL DEFAULT '1',
  `approval_config_version` int NOT NULL DEFAULT '0',
  `current_step` int DEFAULT NULL,
  `current_approver_ids` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `current_approver_names` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_leave_request_inst_status` (`inst_id`,`status`,`create_time`),
  KEY `idx_inst_leave_request_student` (`inst_id`,`student_id`),
  KEY `idx_inst_leave_request_time` (`inst_id`,`start_time`,`end_time`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构请假申请表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_leave_schedule`
--

DROP TABLE IF EXISTS `inst_leave_schedule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_leave_schedule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `leave_request_id` bigint NOT NULL,
  `teaching_schedule_id` bigint NOT NULL,
  `class_type` int NOT NULL DEFAULT '0',
  `teaching_class_id` bigint NOT NULL DEFAULT '0',
  `teaching_class_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `lesson_id` bigint NOT NULL DEFAULT '0',
  `lesson_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `teacher_id` bigint NOT NULL DEFAULT '0',
  `teacher_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `roster_status_before` int NOT NULL DEFAULT '1',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_leave_schedule_unique` (`inst_id`,`leave_request_id`,`teaching_schedule_id`),
  KEY `idx_inst_leave_schedule_leave` (`inst_id`,`leave_request_id`),
  KEY `idx_inst_leave_schedule_schedule` (`inst_id`,`teaching_schedule_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构请假排课关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_ledger`
--

DROP TABLE IF EXISTS `inst_ledger`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_ledger` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `source_type` int NOT NULL DEFAULT '1',
  `system_type` int NOT NULL DEFAULT '0',
  `source_biz_type` int NOT NULL DEFAULT '0',
  `source_biz_id` bigint NOT NULL DEFAULT '0',
  `type` int NOT NULL DEFAULT '1',
  `ledger_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `ledger_category_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `ledger_category_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `ledger_sub_category_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `ledger_sub_category_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `ledger_category_icon` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `deal_staff_id` bigint NOT NULL DEFAULT '0',
  `deal_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `pay_time` datetime DEFAULT NULL,
  `pay_method` int DEFAULT NULL,
  `account_id` bigint NOT NULL DEFAULT '0',
  `account_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `reciprocal_account` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `bank_slip_no` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `order_id` bigint NOT NULL DEFAULT '0',
  `order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_phone_raw` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `payment_voucher_text` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `payment_voucher_images` json DEFAULT NULL,
  `ledger_confirm_status` int NOT NULL DEFAULT '0',
  `confirm_staff_id` bigint NOT NULL DEFAULT '0',
  `confirm_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `confirm_time` datetime DEFAULT NULL,
  `confirm_remark_text` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `confirm_remark_images` json DEFAULT NULL,
  `bill_flow_id` bigint NOT NULL DEFAULT '0',
  `bill_id` bigint NOT NULL DEFAULT '0',
  `error_message` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_ledger_source` (`inst_id`,`source_type`,`source_biz_type`,`source_biz_id`),
  UNIQUE KEY `uk_inst_ledger_number` (`inst_id`,`ledger_number`),
  KEY `idx_inst_ledger_list` (`inst_id`,`create_time`,`id`),
  KEY `idx_inst_ledger_order` (`inst_id`,`order_id`),
  KEY `idx_inst_ledger_student` (`inst_id`,`student_id`),
  KEY `idx_inst_ledger_confirm` (`inst_id`,`ledger_confirm_status`),
  KEY `idx_inst_ledger_sub_category` (`inst_id`,`ledger_sub_category_id`)
) ENGINE=InnoDB AUTO_INCREMENT=394 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构台账表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_order_tag`
--

DROP TABLE IF EXISTS `inst_order_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_order_tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '1',
  `org_order_tag_id` bigint NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_order_tag_inst` (`inst_id`),
  KEY `idx_inst_order_tag_enable` (`enable`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构订单标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_period_config_version`
--

DROP TABLE IF EXISTS `inst_period_config_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_period_config_version` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `effective_week_start` date NOT NULL,
  `payload_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_period_cfg_version_week` (`inst_id`,`effective_week_start`),
  KEY `idx_inst_period_cfg_version_inst_week` (`inst_id`,`effective_week_start`)
) ENGINE=InnoDB AUTO_INCREMENT=171 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构时段配置版本表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_period_group`
--

DROP TABLE IF EXISTS `inst_period_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_period_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `group_uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sort_order` int NOT NULL DEFAULT '0',
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_period_group_uuid` (`inst_id`,`group_uuid`),
  KEY `idx_inst_period_group_inst` (`inst_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=541 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构时段分组表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_period_group_teacher`
--

DROP TABLE IF EXISTS `inst_period_group_teacher`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_period_group_teacher` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL,
  `teacher_user_id` bigint NOT NULL,
  `teacher_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_period_group_teacher` (`group_id`,`teacher_user_id`),
  KEY `idx_inst_period_group_teacher_teacher` (`teacher_user_id`),
  CONSTRAINT `fk_inst_period_group_teacher_group` FOREIGN KEY (`group_id`) REFERENCES `inst_period_group` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=1770 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构时段分组教师表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_period_slot`
--

DROP TABLE IF EXISTS `inst_period_slot`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_period_slot` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL,
  `slot_index` int NOT NULL,
  `start_time` char(5) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `end_time` char(5) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_inst_period_slot_group` (`group_id`,`slot_index`),
  CONSTRAINT `fk_inst_period_slot_group` FOREIGN KEY (`group_id`) REFERENCES `inst_period_group` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6502 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构时段槽表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_school_holiday`
--

DROP TABLE IF EXISTS `inst_school_holiday`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_school_holiday` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `start_date` date NOT NULL,
  `end_date` date NOT NULL,
  `source` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'custom',
  `sort` int NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_school_holiday_inst` (`inst_id`,`del_flag`),
  KEY `idx_inst_school_holiday_dates` (`inst_id`,`start_date`,`end_date`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机构校历假期表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student`
--

DROP TABLE IF EXISTS `inst_student`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint NOT NULL COMMENT '机构ID',
  `stu_name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '学生姓名',
  `stu_sex` tinyint(1) DEFAULT NULL COMMENT '性别',
  `children_id` bigint DEFAULT NULL COMMENT '孩子ID',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `mobile` varchar(255) DEFAULT NULL COMMENT '手机号',
  `phone_relationship` tinyint(1) DEFAULT NULL COMMENT '手机关联人关系',
  `fee_id` bigint DEFAULT NULL COMMENT '余额账户ID',
  `avatar_url` varchar(1024) DEFAULT NULL COMMENT '头像地址',
  `channel_id` int DEFAULT NULL COMMENT '渠道ID',
  `intended_course` varchar(255) DEFAULT NULL COMMENT '意向课程',
  `sale_person` bigint DEFAULT NULL COMMENT '销售员',
  `sale_assigned_time` datetime DEFAULT NULL COMMENT '分配销售时间',
  `follow_up_status` int DEFAULT NULL COMMENT '跟进状态',
  `last_follow_up_time` datetime DEFAULT NULL COMMENT '最近跟进时间',
  `next_follow_up_time` datetime DEFAULT NULL COMMENT '下次跟进时间',
  `intent_level` int DEFAULT NULL COMMENT '意向度等级',
  `student_status` tinyint(1) DEFAULT '0' COMMENT '学员状态 0意向学员 1在读学员',
  `is_bind_child` tinyint(1) DEFAULT '0' COMMENT '是否关注家校通 0-未关注 1-已关注',
  `is_collect` tinyint(1) DEFAULT '0' COMMENT '是否人脸采集 0-未采集 1-已采集',
  `recommend_student_id` bigint DEFAULT NULL COMMENT '推荐人ID',
  `wechat_number` varchar(255) DEFAULT NULL COMMENT '微信号',
  `grade` varchar(255) DEFAULT NULL COMMENT '年级',
  `study_school` varchar(255) DEFAULT NULL COMMENT '就读学校',
  `interest` varchar(255) DEFAULT NULL COMMENT '兴趣爱好',
  `address` varchar(255) DEFAULT NULL COMMENT '家庭住址',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `supervisor_id` bigint DEFAULT NULL COMMENT '督导人员ID',
  `supervisor_assigned_time` datetime DEFAULT NULL COMMENT '分配督导时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE,
  KEY `children_id_index` (`children_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5102 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构学员表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_face_attendance_record`
--

DROP TABLE IF EXISTS `inst_student_face_attendance_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_face_attendance_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `face_image` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `record_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_inst_student_face_attendance_record_inst_time` (`inst_id`,`record_time`),
  KEY `idx_inst_student_face_attendance_record_student` (`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员人脸考勤记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_face_attendance_session`
--

DROP TABLE IF EXISTS `inst_student_face_attendance_session`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_face_attendance_session` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `attendance_date` date NOT NULL,
  `status` int NOT NULL DEFAULT '1',
  `sign_in_time` datetime DEFAULT NULL,
  `sign_in_image` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sign_out_time` datetime DEFAULT NULL,
  `sign_out_image` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_student_face_attendance_session_day` (`inst_id`,`student_id`,`attendance_date`),
  KEY `idx_inst_student_face_attendance_session_inst_day` (`inst_id`,`attendance_date`),
  KEY `idx_inst_student_face_attendance_session_student` (`inst_id`,`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员人脸考勤场次表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_face_profile`
--

DROP TABLE IF EXISTS `inst_student_face_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_face_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `face_descriptor` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `face_image` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_student_face_profile` (`inst_id`,`student_id`),
  KEY `idx_inst_student_face_profile_student` (`student_id`),
  KEY `idx_inst_student_face_profile_inst` (`inst_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员人脸档案表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_face_roll_call_task`
--

DROP TABLE IF EXISTS `inst_student_face_roll_call_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_face_roll_call_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `attendance_session_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `attendance_date` date NOT NULL,
  `teaching_schedule_id` bigint NOT NULL,
  `execute_at` datetime NOT NULL,
  `sign_in_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '1',
  `teaching_record_id` bigint NOT NULL DEFAULT '0',
  `last_error` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inst_student_face_roll_call_task` (`inst_id`,`attendance_session_id`,`teaching_schedule_id`),
  KEY `idx_inst_student_face_roll_call_task_execute` (`inst_id`,`status`,`execute_at`,`id`),
  KEY `idx_inst_student_face_roll_call_task_session` (`inst_id`,`attendance_session_id`),
  KEY `idx_inst_student_face_roll_call_task_student` (`inst_id`,`student_id`,`attendance_date`)
) ENGINE=InnoDB AUTO_INCREMENT=3405 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员人脸点名任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_field_key`
--

DROP TABLE IF EXISTS `inst_student_field_key`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_field_key` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字段定义主键ID',
  `uuid` varchar(36) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '所属机构ID',
  `field_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '字段名称',
  `field_type` int NOT NULL COMMENT '字段类型（text, number, date, select）',
  `required` tinyint(1) DEFAULT '0' COMMENT '是否必填',
  `searched` tinyint(1) DEFAULT '0' COMMENT '支持搜索',
  `options_json` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '字段选项（用于select）',
  `is_default` tinyint(1) DEFAULT NULL COMMENT '是否系统默认',
  `is_display` tinyint(1) DEFAULT NULL COMMENT '是否展示',
  `can_delete` tinyint(1) DEFAULT NULL COMMENT '是否可以移除',
  `can_edit` tinyint(1) DEFAULT NULL COMMENT '是否可以编辑',
  `sort` int DEFAULT '0' COMMENT '排序号',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=124 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='机构学员自定义字段定义表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_field_value`
--

DROP TABLE IF EXISTS `inst_student_field_value`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_field_value` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字段值ID',
  `uuid` varchar(36) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `student_id` bigint DEFAULT NULL COMMENT '学生ID',
  `field_id` bigint DEFAULT NULL COMMENT '字段定义ID',
  `field_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '字段编码（冗余字段加速查询）',
  `field_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '字段值（JSON字符串、数字、文本等）',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_student` (`student_id`) USING BTREE,
  KEY `idx_field` (`field_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2204 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='机构学员自定义字段值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_student_record`
--

DROP TABLE IF EXISTS `inst_student_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_student_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint NOT NULL COMMENT '机构ID',
  `stu_id` bigint NOT NULL COMMENT '学生姓名',
  `change_content` varchar(2550) DEFAULT NULL COMMENT '变更内容',
  `change_id` bigint DEFAULT NULL COMMENT '变更人',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=98 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构学员档案记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_user`
--

DROP TABLE IF EXISTS `inst_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'uuid',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT NULL COMMENT '删除标记',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `inst_id` bigint NOT NULL COMMENT '机构ID',
  `nick_name` varchar(40) NOT NULL COMMENT '昵称',
  `username` varchar(12) DEFAULT NULL COMMENT '用户名',
  `avatar` varchar(1024) DEFAULT NULL COMMENT '头像',
  `sex` tinyint(1) DEFAULT NULL COMMENT '性别',
  `mobile` varchar(18) NOT NULL COMMENT '手机号码',
  `is_admin` tinyint(1) DEFAULT '0' COMMENT '是否管理员',
  `is_manage` tinyint(1) DEFAULT '0' COMMENT '是否有全部数据权限',
  `disabled` tinyint(1) DEFAULT '0' COMMENT '是否禁用',
  `user_type` int DEFAULT NULL COMMENT '用户类型',
  `is_teacher` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否教师',
  `is_supervisor` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否督导',
  `activated_status` tinyint(1) DEFAULT NULL COMMENT '激活状态',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_inst_user_user_del_dis_inst` (`user_id`,`del_flag`,`disabled`,`inst_id`)
) ENGINE=InnoDB AUTO_INCREMENT=30000398 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `inst_user_dept`
--

DROP TABLE IF EXISTS `inst_user_dept`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inst_user_dept` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_user_id` bigint DEFAULT NULL COMMENT '机构学员ID',
  `dept_id` bigint DEFAULT NULL COMMENT '部门ID',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=261 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='机构用户部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `intent_student_export_record`
--

DROP TABLE IF EXISTS `intent_student_export_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `intent_student_export_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `export_staff_id` bigint NOT NULL DEFAULT '0',
  `export_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_type` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_data` longblob NOT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `query_conditions_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_intent_student_export_inst` (`inst_id`,`create_time`,`id`),
  KEY `idx_intent_student_export_expire` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意向学员导出记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `intention_student_import_task`
--

DROP TABLE IF EXISTS `intention_student_import_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `intention_student_import_task` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `upload_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `upload_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `execute_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `execute_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `executed_rows` int NOT NULL DEFAULT '0',
  `deleted_rows` int NOT NULL DEFAULT '0',
  `error_rows` int NOT NULL DEFAULT '0',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `confirm_time` datetime DEFAULT NULL,
  `complete_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `inst_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `columns_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_import_task_inst` (`inst_id`),
  KEY `idx_import_task_created` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意向学员导入任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `intention_student_import_task_record`
--

DROP TABLE IF EXISTS `intention_student_import_task_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `intention_student_import_task_record` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `row_no` int NOT NULL DEFAULT '0',
  `has_error` tinyint(1) NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '0',
  `result` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cells_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_import_task_record_task` (`task_id`),
  KEY `idx_import_task_record_row` (`row_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='意向学员导入明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `login_template`
--

DROP TABLE IF EXISTS `login_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `login_template` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `template_key` varchar(64) NOT NULL,
  `template_name` varchar(128) NOT NULL,
  `entry_type` varchar(32) NOT NULL DEFAULT 'all',
  `layout_type` varchar(32) NOT NULL DEFAULT 'split',
  `description` varchar(500) DEFAULT NULL,
  `preview_image` varchar(500) DEFAULT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `sort` int NOT NULL DEFAULT '0',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_login_template_key` (`template_key`),
  KEY `idx_login_template_entry` (`entry_type`,`enabled`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=4783 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='登录页模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `login_template_institution`
--

DROP TABLE IF EXISTS `login_template_institution`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `login_template_institution` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `template_id` bigint NOT NULL,
  `institution_id` bigint NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_login_template_institution` (`template_id`,`institution_id`),
  KEY `idx_login_template_institution_id` (`institution_id`,`del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='登录页模板机构关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `login_template_tenant`
--

DROP TABLE IF EXISTS `login_template_tenant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `login_template_tenant` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `template_id` bigint NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_login_template_tenant` (`template_id`,`tenant_id`),
  KEY `idx_login_template_tenant_id` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='登录页模板租户关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `message_event_log`
--

DROP TABLE IF EXISTS `message_event_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `message_event_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `topic` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `tag` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `payload` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息事件日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notice_record`
--

DROP TABLE IF EXISTS `notice_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notice_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `notice_template_id` bigint NOT NULL DEFAULT '0',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `summary` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `is_all_school` tinyint(1) NOT NULL DEFAULT '0',
  `is_delay_send` tinyint(1) NOT NULL DEFAULT '0',
  `is_confirm` tinyint(1) NOT NULL DEFAULT '0',
  `is_remind` tinyint(1) NOT NULL DEFAULT '0',
  `is_withdraw` tinyint(1) NOT NULL DEFAULT '0',
  `class_ids_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `classs_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `student_ids_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `target_student_ids_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `student_count` int NOT NULL DEFAULT '0',
  `read_student_count` int NOT NULL DEFAULT '0',
  `confirm_student_count` int NOT NULL DEFAULT '0',
  `operator_id` bigint NOT NULL DEFAULT '0',
  `operator_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `operation_date` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `publish_hour` int NOT NULL DEFAULT '0',
  `publish_time` datetime DEFAULT NULL,
  `reality_publish_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '4',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_notice_record_inst_deleted` (`inst_id`,`del_flag`,`id`),
  KEY `idx_notice_record_inst_status` (`inst_id`,`status`,`is_withdraw`,`id`),
  KEY `idx_notice_record_inst_publish` (`inst_id`,`publish_time`,`id`),
  KEY `idx_notice_record_inst_operation` (`inst_id`,`operation_date`,`id`),
  KEY `idx_notice_record_inst_operator` (`inst_id`,`operator_id`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知发送记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notice_template`
--

DROP TABLE IF EXISTS `notice_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notice_template` (
  `id` bigint NOT NULL,
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cover_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `tag` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `weight` int NOT NULL DEFAULT '0',
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `org_id` bigint NOT NULL DEFAULT '0',
  `school_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_notice_template_scope` (`school_id`,`org_id`,`weight`,`id`),
  KEY `idx_notice_template_deleted` (`del_flag`,`weight`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `order_import_task`
--

DROP TABLE IF EXISTS `order_import_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_import_task` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `upload_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `upload_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `execute_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `execute_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `executed_rows` int NOT NULL DEFAULT '0',
  `deleted_rows` int NOT NULL DEFAULT '0',
  `error_rows` int NOT NULL DEFAULT '0',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `confirm_time` datetime DEFAULT NULL,
  `complete_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `inst_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `columns_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_import_task_inst` (`inst_id`),
  KEY `idx_order_import_task_created` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单导入任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `order_import_task_record`
--

DROP TABLE IF EXISTS `order_import_task_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_import_task_record` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `row_no` int NOT NULL DEFAULT '0',
  `has_error` tinyint(1) NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '0',
  `result` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cells_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_order_import_task_record_task` (`task_id`),
  KEY `idx_order_import_task_record_row` (`row_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单导入明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `org_institution`
--

DROP TABLE IF EXISTS `org_institution`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_institution` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `gid` bigint DEFAULT NULL COMMENT '集团ID',
  `organ_name` varchar(64) NOT NULL COMMENT '机构名称',
  `organ_type` tinyint NOT NULL COMMENT '机构类型',
  `organ_label` varchar(127) DEFAULT NULL COMMENT '机构标签',
  `mobile` varchar(11) NOT NULL COMMENT '手机号',
  `organ_code` varchar(32) NOT NULL COMMENT '机构编号',
  `login_name` varchar(22) NOT NULL COMMENT '登录账户名',
  `open_type` tinyint NOT NULL DEFAULT '2' COMMENT '开通类型：1体验版 2正式版',
  `open_duration` varchar(16) NOT NULL DEFAULT '1y' COMMENT '开通时长编码',
  `expire_start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '账号有效期起始时间',
  `expire_end_time` datetime DEFAULT NULL COMMENT '账号有效期结束',
  `province_code` int DEFAULT NULL COMMENT '省Code',
  `province` varchar(32) DEFAULT NULL COMMENT '省',
  `city_code` int DEFAULT NULL COMMENT '市Code',
  `city` varchar(32) NOT NULL COMMENT '市',
  `region_code` int DEFAULT NULL COMMENT '区Code',
  `region` varchar(32) DEFAULT NULL COMMENT '区',
  `logo` varchar(2000) DEFAULT NULL COMMENT 'logo地址',
  `label` varchar(64) DEFAULT NULL COMMENT '机构标签',
  `principal` varchar(64) DEFAULT NULL COMMENT '负责人',
  `address` varchar(256) DEFAULT NULL COMMENT '详细地址',
  `video` varchar(2000) DEFAULT NULL COMMENT '介绍视频',
  `description` text COMMENT '企业描述',
  `lng` decimal(10,6) DEFAULT NULL COMMENT '经度',
  `lat` decimal(10,6) DEFAULT NULL COMMENT '纬度',
  `account_num` int DEFAULT '5' COMMENT '子账号数',
  `status` tinyint DEFAULT '1' COMMENT '状态',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否可用',
  `concat_phone` varchar(40) DEFAULT NULL COMMENT '联系电话',
  `fixed_phone` varchar(255) DEFAULT NULL COMMENT '固定电话',
  `business_time` varchar(255) DEFAULT NULL COMMENT '运营时间',
  `inst_images` json DEFAULT NULL COMMENT '机构图片集',
  `map_json` json DEFAULT NULL COMMENT '残联大屏地图json',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `guide_id` bigint DEFAULT NULL COMMENT '指导中心ID',
  `disable_id` bigint DEFAULT NULL COMMENT '残联机构ID',
  `fixed_point` tinyint(1) DEFAULT NULL COMMENT '是否定点机构',
  `dis_template` bigint DEFAULT '1' COMMENT '救助申请模板',
  `subsidy_price` decimal(20,6) DEFAULT NULL COMMENT '补贴单价',
  `wechat_code` varchar(255) DEFAULT NULL COMMENT '微信收款码',
  `alipay_code` varchar(255) DEFAULT NULL COMMENT '支付宝收款码',
  `union_pay_code` varchar(255) DEFAULT NULL COMMENT '银联收款码',
  `use_fee` tinyint(1) DEFAULT '0' COMMENT '是否使用现金收费销课',
  `use_train` tinyint(1) DEFAULT '1' COMMENT '是否使用交互训练',
  `show_mock` tinyint(1) DEFAULT '0' COMMENT '是否模拟数据',
  PRIMARY KEY (`id`,`city`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=10059 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='组织机构表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `org_institution_profile`
--

DROP TABLE IF EXISTS `org_institution_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_institution_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `institution_id` bigint NOT NULL,
  `organ_label` varchar(127) DEFAULT NULL,
  `description` text,
  `business_time` varchar(255) DEFAULT NULL,
  `video` varchar(2000) DEFAULT NULL,
  `gallery_images` json DEFAULT NULL,
  `login_slug` varchar(64) NOT NULL DEFAULT '',
  `login_brand_config` json DEFAULT NULL,
  `create_id` bigint DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint DEFAULT NULL,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_institution_profile_inst` (`institution_id`),
  KEY `idx_org_institution_profile_inst_del` (`institution_id`,`del_flag`),
  KEY `idx_org_institution_profile_login_slug` (`login_slug`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=49 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='组织机构扩展资料表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `org_institution_renewal_record`
--

DROP TABLE IF EXISTS `org_institution_renewal_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_institution_renewal_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `institution_id` bigint NOT NULL,
  `before_open_type` tinyint NOT NULL DEFAULT '2',
  `before_open_duration` varchar(16) NOT NULL DEFAULT '',
  `before_expire_end_time` datetime DEFAULT NULL,
  `after_open_type` tinyint NOT NULL DEFAULT '2',
  `renew_duration` varchar(16) NOT NULL DEFAULT '',
  `renew_start_time` datetime NOT NULL,
  `after_expire_end_time` datetime NOT NULL,
  `operator_id` bigint DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_org_institution_renewal_record_inst_del_time` (`institution_id`,`del_flag`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='组织机构续费记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `org_institution_version_change_record`
--

DROP TABLE IF EXISTS `org_institution_version_change_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_institution_version_change_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `institution_id` bigint NOT NULL,
  `before_open_type` tinyint NOT NULL DEFAULT '2',
  `before_module_id` bigint DEFAULT NULL,
  `before_version_name` varchar(64) NOT NULL DEFAULT '',
  `after_open_type` tinyint NOT NULL DEFAULT '2',
  `after_module_id` bigint DEFAULT NULL,
  `after_version_name` varchar(64) NOT NULL DEFAULT '',
  `operator_id` bigint DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_org_institution_version_change_record_inst_del_time` (`institution_id`,`del_flag`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='组织机构版本变更记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `org_module`
--

DROP TABLE IF EXISTS `org_module`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `org_module` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `org_id` bigint DEFAULT NULL COMMENT '机构id',
  `module_id` bigint DEFAULT NULL COMMENT '模块ID',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `status` tinyint DEFAULT NULL COMMENT '状态',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE,
  KEY `idx_org_module_org_del_id` (`org_id`,`del_flag`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb3 COMMENT='组织模块授权表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pending_renewal_student_export_record`
--

DROP TABLE IF EXISTS `pending_renewal_student_export_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pending_renewal_student_export_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `export_staff_id` bigint NOT NULL DEFAULT '0',
  `export_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_type` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_data` longblob NOT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `query_conditions_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pending_renewal_student_export_inst` (`inst_id`,`create_time`,`id`),
  KEY `idx_pending_renewal_student_export_expire` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='待续费学员导出记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_goal_material`
--

DROP TABLE IF EXISTS `pep3_iep_goal_material`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_goal_material` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `library_scope` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'institution',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `material_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'long_term',
  `parent_goal_material_id` bigint NOT NULL DEFAULT '0',
  `domain_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `domain` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `long_goal` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `short_goal` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `course_form` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `age_min_months` int NOT NULL DEFAULT '0',
  `age_max_months` int NOT NULL DEFAULT '0',
  `difficulty_level` int NOT NULL DEFAULT '0',
  `mastery_criteria` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `applicable_score_values` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `priority` int NOT NULL DEFAULT '0',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pep3_iep_goal_scope` (`library_scope`,`inst_id`,`status`,`del_flag`),
  KEY `idx_pep3_iep_goal_domain` (`library_scope`,`inst_id`,`domain_code`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=862 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP目标素材表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_item_option_rule`
--

DROP TABLE IF EXISTS `pep3_iep_item_option_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_item_option_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `library_scope` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'institution',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `item_no` int NOT NULL DEFAULT '0',
  `item_title` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `domain_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `domain` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `score_value` int NOT NULL DEFAULT '-1',
  `score_label` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `score_description` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `result_meaning` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `generate_policy` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `priority` int NOT NULL DEFAULT '0',
  `ai_instruction` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pep3_iep_item_rule_match` (`library_scope`,`inst_id`,`item_no`,`score_value`,`status`,`del_flag`),
  KEY `idx_pep3_iep_item_rule_domain` (`library_scope`,`inst_id`,`domain_code`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=426 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP题目选项规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_item_rule_goal_rel`
--

DROP TABLE IF EXISTS `pep3_iep_item_rule_goal_rel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_item_rule_goal_rel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL DEFAULT '0',
  `goal_material_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pep3_iep_rule_goal_rel` (`rule_id`,`goal_material_id`,`del_flag`),
  KEY `idx_pep3_iep_rule_goal_rule` (`rule_id`,`del_flag`),
  KEY `idx_pep3_iep_rule_goal_goal` (`goal_material_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=429 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP题目规则目标关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_material_import_task`
--

DROP TABLE IF EXISTS `pep3_iep_material_import_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_material_import_task` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `upload_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `upload_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `execute_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `execute_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `executed_rows` int NOT NULL DEFAULT '0',
  `deleted_rows` int NOT NULL DEFAULT '0',
  `error_rows` int NOT NULL DEFAULT '0',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `confirm_time` datetime DEFAULT NULL,
  `complete_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `inst_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `columns_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_pep3_iep_import_task_created` (`created_time`),
  KEY `idx_pep3_iep_import_task_status` (`status`,`del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP素材导入任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_material_import_task_record`
--

DROP TABLE IF EXISTS `pep3_iep_material_import_task_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_material_import_task_record` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `row_no` int NOT NULL DEFAULT '0',
  `has_error` tinyint(1) NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '0',
  `result` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cells_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pep3_iep_import_record_task` (`task_id`),
  KEY `idx_pep3_iep_import_record_row` (`row_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP素材导入明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pep3_iep_training_material`
--

DROP TABLE IF EXISTS `pep3_iep_training_material`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pep3_iep_training_material` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `library_scope` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'institution',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `goal_material_id` bigint NOT NULL DEFAULT '0',
  `training_project` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `training_content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `priority` int NOT NULL DEFAULT '0',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pep3_iep_training_scope` (`library_scope`,`inst_id`,`status`,`del_flag`),
  KEY `idx_pep3_iep_training_goal` (`goal_material_id`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=464 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PEP3 IEP训练素材表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_package`
--

DROP TABLE IF EXISTS `product_package`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `product_package` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `name` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `title` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `online_sale` tinyint(1) NOT NULL DEFAULT '1',
  `is_allow_edit_when_enroll` tinyint(1) NOT NULL DEFAULT '0',
  `images` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_show_mico_school` tinyint(1) NOT NULL DEFAULT '0',
  `is_online_sale_mico_school` tinyint(1) NOT NULL DEFAULT '0',
  `buy_rule_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `subject_ids_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `org_product_package_id` bigint NOT NULL DEFAULT '0',
  `editable` tinyint(1) NOT NULL DEFAULT '1',
  `is_sync_org_product_package` tinyint(1) NOT NULL DEFAULT '0',
  `sale_volume` int NOT NULL DEFAULT '0',
  `total_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `discount_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `final_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_product_package_inst` (`inst_id`,`del_flag`),
  KEY `idx_product_package_updated` (`inst_id`,`update_time`,`id`),
  KEY `idx_product_package_sale` (`inst_id`,`online_sale`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品套餐表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_package_item`
--

DROP TABLE IF EXISTS `product_package_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `product_package_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `product_package_id` bigint NOT NULL,
  `product_type` int NOT NULL DEFAULT '1',
  `product_id` bigint NOT NULL DEFAULT '0',
  `product_name` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sku_id` bigint NOT NULL DEFAULT '0',
  `sku_name` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sku_count` decimal(18,2) NOT NULL DEFAULT '0.00',
  `free_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `discount_type` int DEFAULT NULL,
  `discount_number` decimal(18,2) NOT NULL DEFAULT '0.00',
  `lesson_type` int NOT NULL DEFAULT '0',
  `lesson_mode` int NOT NULL DEFAULT '0',
  `lesson_audition` tinyint(1) NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_product_package_item_package` (`product_package_id`,`del_flag`),
  KEY `idx_product_package_item_product` (`product_id`,`sku_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品套餐项目表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `product_package_property_result`
--

DROP TABLE IF EXISTS `product_package_property_result`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `product_package_property_result` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `product_package_id` bigint NOT NULL,
  `property_id` bigint NOT NULL,
  `property_value` bigint NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_product_package_property_package` (`product_package_id`,`del_flag`),
  KEY `idx_product_package_property_filter` (`property_id`,`property_value`,`del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品套餐属性结果表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account`
--

DROP TABLE IF EXISTS `recharge_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `account_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `main_student_id` bigint NOT NULL DEFAULT '0',
  `phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `recharge_balance` decimal(18,2) NOT NULL DEFAULT '0.00',
  `residual_balance` decimal(18,2) NOT NULL DEFAULT '0.00',
  `giving_balance` decimal(18,2) NOT NULL DEFAULT '0.00',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_inst` (`inst_id`,`update_time`,`id`),
  KEY `idx_recharge_account_main_student` (`inst_id`,`main_student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1282 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_bill`
--

DROP TABLE IF EXISTS `recharge_account_bill`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_bill` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `recharge_account_order_id` bigint NOT NULL,
  `status` int NOT NULL DEFAULT '1',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_bill_order` (`inst_id`,`recharge_account_order_id`)
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户账单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_bill_flow`
--

DROP TABLE IF EXISTS `recharge_account_bill_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_bill_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `bill_id` bigint NOT NULL,
  `amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_bill_flow_bill` (`inst_id`,`bill_id`)
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户账单流水表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_flow`
--

DROP TABLE IF EXISTS `recharge_account_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `recharge_account_id` bigint NOT NULL,
  `student_id` bigint NOT NULL DEFAULT '0',
  `order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `flow_type` int NOT NULL DEFAULT '1',
  `amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `residual_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `giving_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_flow_account` (`inst_id`,`recharge_account_id`,`create_time`),
  KEY `idx_recharge_account_flow_student` (`inst_id`,`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户流水表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_import_task`
--

DROP TABLE IF EXISTS `recharge_account_import_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_import_task` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `upload_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `upload_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `execute_staff_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `execute_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `executed_rows` int NOT NULL DEFAULT '0',
  `deleted_rows` int NOT NULL DEFAULT '0',
  `error_rows` int NOT NULL DEFAULT '0',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `confirm_time` datetime DEFAULT NULL,
  `complete_time` datetime DEFAULT NULL,
  `status` int NOT NULL DEFAULT '0',
  `inst_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `columns_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_import_task_inst` (`inst_id`),
  KEY `idx_recharge_account_import_task_created` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户导入任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_import_task_record`
--

DROP TABLE IF EXISTS `recharge_account_import_task_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_import_task_record` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `row_no` int NOT NULL DEFAULT '0',
  `has_error` tinyint(1) NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '0',
  `result` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cells_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_import_task_record_task` (`task_id`),
  KEY `idx_recharge_account_import_task_record_row` (`row_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户导入明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_order`
--

DROP TABLE IF EXISTS `recharge_account_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `recharge_account_id` bigint NOT NULL,
  `sale_order_id` bigint NOT NULL DEFAULT '0',
  `order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT '1',
  `amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `giving_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `residual_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `deal_date` date DEFAULT NULL,
  `sale_person_id` bigint NOT NULL DEFAULT '0',
  `collector_staff_id` bigint NOT NULL DEFAULT '0',
  `phone_sell_staff_id` bigint NOT NULL DEFAULT '0',
  `foreground_staff_id` bigint NOT NULL DEFAULT '0',
  `vice_sell_staff_staff_id` bigint NOT NULL DEFAULT '0',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `external_remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_id` bigint NOT NULL DEFAULT '0',
  `bill_id` bigint NOT NULL DEFAULT '0',
  `approve_id` bigint DEFAULT NULL,
  `order_obsolete` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recharge_account_order_number` (`inst_id`,`order_number`),
  KEY `idx_recharge_account_order_inst` (`inst_id`,`create_time`,`id`),
  KEY `idx_recharge_account_order_student` (`inst_id`,`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=74 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_order_tag`
--

DROP TABLE IF EXISTS `recharge_account_order_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_order_tag` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `recharge_account_order_id` bigint NOT NULL,
  `tag_id` bigint NOT NULL,
  `tag_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_recharge_account_order_tag_order` (`inst_id`,`recharge_account_order_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户订单标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `recharge_account_student`
--

DROP TABLE IF EXISTS `recharge_account_student`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `recharge_account_student` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `recharge_account_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `is_main_student` tinyint(1) NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recharge_account_student` (`recharge_account_id`,`student_id`),
  KEY `idx_recharge_account_student_inst` (`inst_id`,`student_id`),
  KEY `idx_recharge_account_student_account` (`inst_id`,`recharge_account_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1282 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值账户学员关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `refund_tuition_account_order`
--

DROP TABLE IF EXISTS `refund_tuition_account_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `refund_tuition_account_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `tuition_account_id` bigint NOT NULL DEFAULT '0',
  `sale_order_id` bigint NOT NULL DEFAULT '0',
  `order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT '1',
  `total_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `real_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `charge_against_tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `refund_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `refund_free_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `handling_fee` decimal(18,2) NOT NULL DEFAULT '0.00',
  `is_recharge_account` tinyint(1) NOT NULL DEFAULT '0',
  `recharge_account_id` bigint NOT NULL DEFAULT '0',
  `deal_date` date DEFAULT NULL,
  `sale_person_id` bigint NOT NULL DEFAULT '0',
  `collector_staff_id` bigint NOT NULL DEFAULT '0',
  `phone_sell_staff_id` bigint NOT NULL DEFAULT '0',
  `foreground_staff_id` bigint NOT NULL DEFAULT '0',
  `vice_sell_staff_staff_id` bigint NOT NULL DEFAULT '0',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `external_remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `order_obsolete` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `student_id` bigint NOT NULL DEFAULT '0',
  `course_id` bigint NOT NULL DEFAULT '0',
  `auto_close_tuition` tinyint(1) NOT NULL DEFAULT '0',
  `completed_time` datetime DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_tuition_sale_order` (`inst_id`,`sale_order_id`),
  KEY `idx_refund_tuition_account_order_account` (`inst_id`,`tuition_account_id`,`del_flag`),
  KEY `idx_refund_tuition_account_order_student` (`inst_id`,`student_id`,`create_time`),
  KEY `idx_refund_tuition_account_order_status` (`inst_id`,`status`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课消账户退款订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `refund_tuition_account_order_item`
--

DROP TABLE IF EXISTS `refund_tuition_account_order_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `refund_tuition_account_order_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `refund_order_id` bigint NOT NULL,
  `sale_order_id` bigint NOT NULL DEFAULT '0',
  `tuition_account_id` bigint NOT NULL DEFAULT '0',
  `source_order_id` bigint NOT NULL DEFAULT '0',
  `source_order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `course_id` bigint NOT NULL DEFAULT '0',
  `lesson_type` int DEFAULT NULL,
  `lesson_charging_mode` int NOT NULL DEFAULT '0',
  `quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `free_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `refund_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `original_refund_amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `arrear_deduction` decimal(18,2) NOT NULL DEFAULT '0.00',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_refund_tuition_order_item_order` (`inst_id`,`refund_order_id`),
  KEY `idx_refund_tuition_order_item_account` (`inst_id`,`tuition_account_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课消账户退款订单明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sale_order`
--

DROP TABLE IF EXISTS `sale_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sale_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `student_id` bigint DEFAULT NULL COMMENT '学员ID',
  `order_number` varchar(255) DEFAULT NULL COMMENT '订单编号',
  `sale_person` bigint DEFAULT NULL COMMENT '销售人ID',
  `deal_date` date DEFAULT NULL COMMENT '经办时间',
  `order_discount_type` int DEFAULT NULL COMMENT '整单优惠类型',
  `order_discount_amount` decimal(20,2) DEFAULT NULL COMMENT '整单优惠金额',
  `order_discount_number` decimal(20,2) DEFAULT NULL COMMENT '整单优惠折扣',
  `order_real_amount` decimal(20,2) DEFAULT NULL COMMENT '订单实际金额',
  `order_tag_ids` varchar(255) DEFAULT NULL COMMENT '订单标签',
  `internal_remark` varchar(255) DEFAULT NULL COMMENT '对内备注',
  `external_remark` varchar(255) DEFAULT NULL COMMENT '对外备注',
  `order_type` int DEFAULT NULL COMMENT '订单类型',
  `order_status` int DEFAULT NULL COMMENT '订单状态',
  `order_source` int DEFAULT NULL COMMENT '订单来源',
  `is_bad_debt` tinyint(1) DEFAULT '0' COMMENT '是否坏账 0-否 1-是',
  `bad_debt_amount` decimal(20,2) DEFAULT '0.00' COMMENT '坏账金额',
  `bad_debt_remark` varchar(500) DEFAULT NULL COMMENT '坏账备注',
  `bad_debt_time` datetime DEFAULT NULL COMMENT '设为坏账时间',
  `bad_debt_operator_id` bigint DEFAULT NULL COMMENT '坏账操作人ID',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=424 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='销售订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sale_order_course_detail`
--

DROP TABLE IF EXISTS `sale_order_course_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sale_order_course_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '支付订单编号',
  `uuid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT 'uuid',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `order_id` bigint DEFAULT NULL COMMENT '订单ID',
  `handle_type` int DEFAULT NULL COMMENT '办理类型',
  `course_id` bigint DEFAULT NULL COMMENT '课程ID',
  `quote_id` bigint DEFAULT NULL COMMENT '报价单',
  `count` int DEFAULT NULL COMMENT '购买份数',
  `unit` int DEFAULT NULL,
  `free_quantity` decimal(20,2) DEFAULT NULL COMMENT '赠送数量',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '报价单金额',
  `discount_type` int DEFAULT NULL COMMENT '优惠类型',
  `discount_number` decimal(20,2) DEFAULT NULL COMMENT '单课优惠',
  `share_discount` decimal(20,2) DEFAULT NULL COMMENT '分摊整单优惠',
  `real_quantity` decimal(20,2) DEFAULT '0.00' COMMENT '总数量（购买+赠送，支持小数）',
  `has_valid_date` tinyint(1) DEFAULT NULL COMMENT '是否有有效期',
  `valid_date` date DEFAULT NULL COMMENT '有效时间',
  `end_date` date DEFAULT NULL COMMENT '截止时间',
  `create_id` bigint DEFAULT NULL COMMENT '创建者',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新者',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `order_index` (`order_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=377 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='销售订单课程明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sale_order_pay_detail`
--

DROP TABLE IF EXISTS `sale_order_pay_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sale_order_pay_detail` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `inst_id` bigint DEFAULT NULL COMMENT '机构ID',
  `order_id` bigint DEFAULT NULL COMMENT '订单ID',
  `amount_id` bigint DEFAULT NULL COMMENT '收款账户',
  `pay_method` int DEFAULT NULL COMMENT '支付方式',
  `pay_amount` decimal(20,2) DEFAULT NULL COMMENT '支付金额',
  `pay_time` date DEFAULT NULL COMMENT '支付时间',
  `payment_voucher` varchar(255) DEFAULT NULL COMMENT '支付凭证',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=405 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='销售订单支付明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sso_menu`
--

DROP TABLE IF EXISTS `sso_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sso_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `menu_name` varchar(255) DEFAULT NULL COMMENT '标题',
  `url_path` varchar(255) DEFAULT NULL COMMENT '路径',
  `menu_code` varchar(255) DEFAULT NULL COMMENT '标识',
  `menu_type` tinyint DEFAULT '0' COMMENT '类型（0：功能，1：数据）',
  `pid` bigint DEFAULT NULL COMMENT '父级ID',
  `sort` int DEFAULT NULL COMMENT '排序值',
  `icon` varchar(255) DEFAULT NULL COMMENT '图标',
  `is_system` tinyint(1) DEFAULT NULL COMMENT '是否系统菜单',
  `introduce` varchar(255) DEFAULT NULL COMMENT '菜单功能介绍',
  `access_denied_image` varchar(2000) DEFAULT NULL COMMENT '页面无权限展示图片',
  `own_type` tinyint DEFAULT NULL COMMENT '所属类型',
  `level` int DEFAULT NULL COMMENT '菜单级别',
  `weight` int DEFAULT NULL COMMENT '分组权重',
  `group_code` varchar(255) DEFAULT NULL COMMENT '分组标识',
  `create_id` varchar(255) DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` varchar(255) DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint DEFAULT NULL COMMENT '删除标记',
  `remark` varchar(1024) DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_sso_menu_own_code_del` (`own_type`,`menu_code`(191),`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=710 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='单点登录菜单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sso_role`
--

DROP TABLE IF EXISTS `sso_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sso_role` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `org_id` bigint DEFAULT NULL COMMENT '机构ID',
  `role_name` varchar(255) DEFAULT NULL COMMENT '角色名',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `sort` int DEFAULT NULL COMMENT '排序字段',
  `role_type` tinyint(1) DEFAULT NULL COMMENT '登录角色类型（0总控平台 1 集团 2机构 3家长）',
  `is_default` tinyint(1) DEFAULT NULL COMMENT '是否系统默认',
  `is_admin` tinyint(1) DEFAULT '0' COMMENT '是否管理员',
  `create_id` varchar(255) DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` varchar(255) DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint DEFAULT NULL COMMENT '删除标记',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_sso_role_org_type_del` (`org_id`,`role_type`,`del_flag`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=71 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='单点登录角色表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sso_role_menu`
--

DROP TABLE IF EXISTS `sso_role_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sso_role_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `role_id` bigint DEFAULT NULL COMMENT '角色ID',
  `menu_id` bigint DEFAULT NULL COMMENT '菜单ID',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `unique_index` (`role_id`,`menu_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=46153 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='单点登录角色菜单关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sso_user`
--

DROP TABLE IF EXISTS `sso_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sso_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` varchar(255) DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` varchar(255) DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint DEFAULT NULL COMMENT '删除标记',
  `username` varchar(255) DEFAULT NULL COMMENT '用户名',
  `password` varchar(255) DEFAULT NULL COMMENT '密码',
  `mobile` varchar(255) DEFAULT NULL COMMENT '手机号',
  `avatar` varchar(1024) DEFAULT NULL COMMENT '头像',
  `openid` varchar(255) DEFAULT NULL COMMENT '微信openid',
  `unionid` varchar(255) DEFAULT NULL COMMENT '微信unionid',
  `nick_name` varchar(255) DEFAULT NULL COMMENT '昵称',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `user_type` tinyint DEFAULT NULL COMMENT '用户类型',
  `dept_id` bigint DEFAULT NULL COMMENT '部门ID',
  `current_inst_id` bigint DEFAULT NULL COMMENT '当前登录机构ID',
  `is_admin` tinyint(1) DEFAULT '0' COMMENT '是否超级管理员（0 否 1是）',
  `disabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '总控账号是否停用',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2114 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='单点登录用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sso_user_role`
--

DROP TABLE IF EXISTS `sso_user_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sso_user_role` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `user_id` bigint DEFAULT NULL COMMENT '用户ID',
  `role_id` bigint DEFAULT NULL COMMENT '角色ID',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `user_id_idx` (`user_id`) USING BTREE,
  KEY `role_id_idx` (`role_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=379 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='单点登录用户角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `student_arrear_export_record`
--

DROP TABLE IF EXISTS `student_arrear_export_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `student_arrear_export_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `export_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `export_staff_id` bigint NOT NULL DEFAULT '0',
  `export_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `content_type` varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `file_data` longblob NOT NULL,
  `total_rows` int NOT NULL DEFAULT '0',
  `query_conditions_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_student_arrear_export_inst` (`inst_id`,`export_type`,`create_time`,`id`),
  KEY `idx_student_arrear_export_expire` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员欠费导出记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `student_rehab_record`
--

DROP TABLE IF EXISTS `student_rehab_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `student_rehab_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `student_teaching_record_id` bigint NOT NULL DEFAULT '0',
  `teaching_record_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `template_code` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `template_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `template_version` int NOT NULL DEFAULT '1',
  `template_scope` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'default',
  `template_assignment_id` bigint NOT NULL DEFAULT '0',
  `template_snapshot_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `draft_content_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `draft_saved_staff_id` bigint NOT NULL DEFAULT '0',
  `draft_saved_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `draft_saved_time` datetime DEFAULT NULL,
  `published_content_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `published_summary` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `published_staff_id` bigint NOT NULL DEFAULT '0',
  `published_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `published_time` datetime DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_student_rehab_record_inst_student` (`inst_id`,`student_teaching_record_id`),
  KEY `idx_student_rehab_record_teaching` (`inst_id`,`teaching_record_id`),
  KEY `idx_student_rehab_record_student` (`inst_id`,`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员康复记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `student_teaching_record`
--

DROP TABLE IF EXISTS `student_teaching_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `student_teaching_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `teaching_record_id` bigint NOT NULL DEFAULT '0',
  `teaching_schedule_id` bigint NOT NULL DEFAULT '0',
  `timetable_source_type` int NOT NULL DEFAULT '0',
  `timetable_source_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `avatar_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `source_type` int NOT NULL DEFAULT '0',
  `current_student_status` int NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '0',
  `is_late` tinyint(1) NOT NULL DEFAULT '0',
  `class_id` bigint NOT NULL DEFAULT '0',
  `class_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `one_to_one_id` bigint NOT NULL DEFAULT '0',
  `one_to_one_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `lesson_id` bigint NOT NULL DEFAULT '0',
  `lesson_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `subject_id` bigint NOT NULL DEFAULT '0',
  `subject_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `teaching_content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `teaching_content_images_json` json DEFAULT NULL,
  `classroom_id` bigint NOT NULL DEFAULT '0',
  `classroom_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `main_teacher_id` bigint NOT NULL DEFAULT '0',
  `main_teacher_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `teacher_employee_type` int NOT NULL DEFAULT '0',
  `assistant_teacher_ids_json` json DEFAULT NULL,
  `assistant_teacher_names_json` json DEFAULT NULL,
  `class_teacher_ids_json` json DEFAULT NULL,
  `class_teacher_names_json` json DEFAULT NULL,
  `roll_call_class_teacher_ids_json` json DEFAULT NULL,
  `roll_call_class_teacher_names_json` json DEFAULT NULL,
  `current_class_teacher_ids_json` json DEFAULT NULL,
  `current_class_teacher_names_json` json DEFAULT NULL,
  `one2one_teacher_ids_json` json DEFAULT NULL,
  `one2one_teacher_names_json` json DEFAULT NULL,
  `tuition_account_id` bigint NOT NULL DEFAULT '0',
  `tuition_account_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sku_mode` int NOT NULL DEFAULT '0',
  `quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `actual_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `amount` decimal(18,2) NOT NULL DEFAULT '0.00',
  `actual_deduct` decimal(18,2) NOT NULL DEFAULT '0.00',
  `actual_tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `arrear_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `teacher_class_time` decimal(18,2) NOT NULL DEFAULT '0.00',
  `remark` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `external_remark` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `is_auto_roll_call` tinyint(1) NOT NULL DEFAULT '0',
  `has_compensated` tinyint(1) NOT NULL DEFAULT '0',
  `advisor_staff_id` bigint NOT NULL DEFAULT '0',
  `advisor_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_manager_id` bigint NOT NULL DEFAULT '0',
  `student_manager_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `teaching_record_created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `record_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_staff_id` bigint NOT NULL DEFAULT '0',
  `updated_staff_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `updated_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_student_teaching_record_list` (`inst_id`,`start_time`,`updated_time`,`id`),
  KEY `idx_student_teaching_record_student` (`inst_id`,`student_id`,`start_time`),
  KEY `idx_student_teaching_record_teaching` (`inst_id`,`teaching_record_id`),
  KEY `idx_student_teaching_record_schedule` (`inst_id`,`teaching_schedule_id`)
) ENGINE=InnoDB AUTO_INCREMENT=216 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员教学记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `student_teaching_record_change_log`
--

DROP TABLE IF EXISTS `student_teaching_record_change_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `student_teaching_record_change_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_record_id` bigint NOT NULL DEFAULT '0',
  `student_teaching_record_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `action_type` int NOT NULL DEFAULT '0',
  `before_status` int NOT NULL DEFAULT '0',
  `before_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `after_status` int NOT NULL DEFAULT '0',
  `after_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `change_content` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `operator_id` bigint NOT NULL DEFAULT '0',
  `operator_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `operate_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_student_teaching_record_change_log_teaching` (`inst_id`,`teaching_record_id`,`operate_time`),
  KEY `idx_student_teaching_record_change_log_student` (`inst_id`,`student_id`,`operate_time`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学员教学记录变更日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `suspend_resume_tuition_account_order`
--

DROP TABLE IF EXISTS `suspend_resume_tuition_account_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `suspend_resume_tuition_account_order` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `tuition_account_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `course_id` bigint NOT NULL,
  `type` int NOT NULL DEFAULT '0',
  `expire_time` datetime DEFAULT NULL,
  `expire_type` int NOT NULL DEFAULT '0',
  `remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `suspend_date` datetime DEFAULT NULL,
  `resume_date` datetime DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_suspend_resume_tuition_account_order_inst` (`inst_id`,`tuition_account_id`,`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课消账户停复课订单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_depart`
--

DROP TABLE IF EXISTS `sys_depart`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_depart` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `depart_name` varchar(127) DEFAULT NULL COMMENT '部门名',
  `depart_code` varchar(40) DEFAULT NULL COMMENT '部门编号',
  `depart_man` varchar(40) DEFAULT NULL COMMENT '负责人名称',
  `depart_concat` varchar(40) DEFAULT NULL COMMENT '联系电话',
  `org_id` bigint DEFAULT NULL COMMENT '所属组织ID',
  `pid` bigint DEFAULT NULL COMMENT '父级部门',
  `is_enable` tinyint(1) DEFAULT '1' COMMENT '状态',
  `sort` int DEFAULT NULL COMMENT '排序',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=47 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='系统部门表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_dict`
--

DROP TABLE IF EXISTS `sys_dict`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `dict_name` varchar(127) DEFAULT NULL COMMENT '字典名称',
  `dict_code` varchar(127) DEFAULT NULL COMMENT '字典类型',
  `is_enable` tinyint(1) DEFAULT '1' COMMENT '状态',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=80 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='系统字典表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_dict_value`
--

DROP TABLE IF EXISTS `sys_dict_value`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_dict_value` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `dict_id` bigint DEFAULT NULL COMMENT '字典主项id',
  `dict_label` varchar(127) DEFAULT NULL COMMENT '字典名称',
  `dict_value` varchar(127) DEFAULT NULL COMMENT '字典类型',
  `sort` int DEFAULT NULL COMMENT '排序',
  `is_enable` tinyint(1) DEFAULT '1' COMMENT '状态',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=243 DEFAULT CHARSET=utf8mb3 ROW_FORMAT=DYNAMIC COMMENT='系统字典值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_login_log`
--

DROP TABLE IF EXISTS `sys_login_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_login_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT NULL COMMENT '删除标记',
  `uuid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'uuid',
  `user_id` bigint DEFAULT NULL COMMENT '用户ID',
  `nick_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '登录用户名',
  `user_type` tinyint DEFAULT NULL COMMENT '用户类型',
  `user_ip` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户IP地址',
  `user_agent` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '浏览器UA',
  `org_id` bigint DEFAULT NULL COMMENT '机构ID',
  `org_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '机构名称',
  `result` int DEFAULT NULL COMMENT '登录结果',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1345 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='系统登录日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_module`
--

DROP TABLE IF EXISTS `sys_module`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_module` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `tenant_id` varchar(64) NOT NULL DEFAULT 'platform' COMMENT '版本归属租户',
  `owner_type` varchar(32) NOT NULL DEFAULT 'platform_template' COMMENT '版本类型：平台模板/租户售卖版本',
  `source_module_id` bigint DEFAULT NULL COMMENT '来源平台模板ID',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `name` varchar(2400) DEFAULT NULL COMMENT '模块名称',
  `type` tinyint DEFAULT '1' COMMENT '模块类型（1：机构，2：集团）',
  `price` decimal(18,4) DEFAULT NULL COMMENT '模块价格',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE,
  KEY `idx_sys_module_type_name_del` (`type`,`name`(191),`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb3 COMMENT='系统模块表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_module_menu`
--

DROP TABLE IF EXISTS `sys_module_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_module_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(36) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `module_id` bigint DEFAULT NULL COMMENT '模块id',
  `menu_id` bigint DEFAULT NULL COMMENT '权限ID',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标记',
  `remark` varchar(255) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `uuid_index` (`uuid`) USING BTREE,
  KEY `idx_sys_module_menu_mod_del_menu` (`module_id`,`del_flag`,`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3375 DEFAULT CHARSET=utf8mb3 COMMENT='系统模块菜单关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_notice_info`
--

DROP TABLE IF EXISTS `sys_notice_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_notice_info` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
  `version` bigint DEFAULT '0' COMMENT '乐观锁',
  `create_id` bigint DEFAULT NULL COMMENT '创建人',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT NULL COMMENT '删除标记',
  `uuid` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'uuid',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '标题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '内容',
  `disable_id` bigint DEFAULT '-1' COMMENT '残联ID',
  `compel` tinyint(1) DEFAULT NULL COMMENT '是否强制提醒（0否 1是）',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='系统公告通知表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_scale`
--

DROP TABLE IF EXISTS `sys_scale`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_scale` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_name` varchar(128) NOT NULL,
  `scale_code` varchar(64) NOT NULL,
  `category` varchar(64) NOT NULL,
  `scenario` varchar(64) NOT NULL,
  `age_range` varchar(64) NOT NULL,
  `age_min_months` int NOT NULL DEFAULT '0',
  `age_max_months` int NOT NULL DEFAULT '0',
  `estimated_duration` varchar(64) NOT NULL DEFAULT '',
  `duration_min_minutes` int NOT NULL DEFAULT '0',
  `duration_max_minutes` int NOT NULL DEFAULT '0',
  `current_version` varchar(64) NOT NULL,
  `item_count` int NOT NULL DEFAULT '0',
  `domain_count` int NOT NULL DEFAULT '0',
  `institution_count` int NOT NULL DEFAULT '0',
  `month_usage` int NOT NULL DEFAULT '0',
  `data_status` varchar(512) DEFAULT NULL,
  `summary` varchar(1024) DEFAULT NULL,
  `poster_url` varchar(500) DEFAULT NULL,
  `execution_entry` varchar(256) DEFAULT NULL,
  `api_package` varchar(256) DEFAULT NULL,
  `sort` int NOT NULL DEFAULT '1',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_scale_code` (`scale_code`,`del_flag`),
  KEY `idx_sys_scale_sort` (`sort`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统量表表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_scale_acknowledgement`
--

DROP TABLE IF EXISTS `sys_scale_acknowledgement`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_scale_acknowledgement` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_id` bigint NOT NULL,
  `content` text NOT NULL,
  `sort` int NOT NULL DEFAULT '1',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_sys_scale_ack_scale` (`scale_id`,`del_flag`),
  KEY `idx_sys_scale_ack_sort` (`scale_id`,`sort`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统量表确认记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_scale_auth_institution`
--

DROP TABLE IF EXISTS `sys_scale_auth_institution`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_scale_auth_institution` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_id` bigint NOT NULL,
  `institution_name` varchar(128) NOT NULL,
  `contact` varchar(128) NOT NULL,
  `auth_state` varchar(32) NOT NULL,
  `expire_at` date DEFAULT NULL,
  `sort` int NOT NULL DEFAULT '1',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_sys_scale_auth_scale` (`scale_id`,`del_flag`),
  KEY `idx_sys_scale_auth_sort` (`scale_id`,`sort`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统量表授权机构表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_scale_reference`
--

DROP TABLE IF EXISTS `sys_scale_reference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_scale_reference` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_id` bigint NOT NULL,
  `content` text NOT NULL,
  `sort` int NOT NULL DEFAULT '1',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_sys_scale_reference_scale` (`scale_id`,`del_flag`),
  KEY `idx_sys_scale_reference_sort` (`scale_id`,`sort`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统量表参考资料表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_class`
--

DROP TABLE IF EXISTS `teaching_class`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_class` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `class_type` int NOT NULL DEFAULT '1',
  `course_id` bigint NOT NULL DEFAULT '0',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `advisor_id` bigint NOT NULL DEFAULT '0',
  `default_teacher_id` bigint NOT NULL DEFAULT '0',
  `status` int NOT NULL DEFAULT '1',
  `scheduled_lesson_count` int NOT NULL DEFAULT '0',
  `finished_lesson_count` int NOT NULL DEFAULT '0',
  `class_room_id` bigint NOT NULL DEFAULT '0',
  `class_room_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `classroom_enabled` tinyint(1) DEFAULT NULL,
  `remark` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `compose_lesson_id` bigint NOT NULL DEFAULT '0',
  `max_count` int NOT NULL DEFAULT '0',
  `closed_time` datetime DEFAULT NULL,
  `default_student_class_time` decimal(18,2) NOT NULL DEFAULT '1.00',
  `default_teacher_class_time` decimal(18,2) NOT NULL DEFAULT '0.00',
  `default_class_time_record_mode` int NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_teaching_class_inst_type` (`inst_id`,`class_type`,`del_flag`),
  KEY `idx_teaching_class_course` (`inst_id`,`course_id`),
  KEY `idx_teaching_class_advisor` (`inst_id`,`advisor_id`),
  KEY `idx_teaching_class_default_teacher` (`inst_id`,`default_teacher_id`),
  KEY `idx_teaching_class_created` (`inst_id`,`create_time`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=990 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学班级表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_class_entry_exit_record`
--

DROP TABLE IF EXISTS `teaching_class_entry_exit_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_class_entry_exit_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_class_id` bigint NOT NULL,
  `teaching_class_student_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `entry_exit_status` int NOT NULL DEFAULT '1',
  `entry_exit_time` datetime NOT NULL,
  `operator_id` bigint NOT NULL DEFAULT '0',
  `operate_time` datetime NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_teaching_class_entry_exit_class` (`inst_id`,`teaching_class_id`,`entry_exit_time`),
  KEY `idx_teaching_class_entry_exit_student` (`inst_id`,`student_id`,`entry_exit_time`),
  KEY `idx_teaching_class_entry_exit_member_status` (`inst_id`,`teaching_class_student_id`,`entry_exit_status`,`entry_exit_time`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学班级进出班记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_class_operation_log`
--

DROP TABLE IF EXISTS `teaching_class_operation_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_class_operation_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_class_id` bigint NOT NULL,
  `teaching_class_student_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `operation_type` int NOT NULL DEFAULT '0',
  `operation_content` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `operator_id` bigint NOT NULL DEFAULT '0',
  `operate_time` datetime NOT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_teaching_class_operation_log_class` (`inst_id`,`teaching_class_id`,`operate_time`),
  KEY `idx_teaching_class_operation_log_student` (`inst_id`,`student_id`,`operate_time`),
  KEY `idx_teaching_class_operation_log_member_type` (`inst_id`,`teaching_class_student_id`,`operation_type`,`operate_time`)
) ENGINE=InnoDB AUTO_INCREMENT=42 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学班级操作日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_class_student`
--

DROP TABLE IF EXISTS `teaching_class_student`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_class_student` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_class_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `order_id` bigint NOT NULL DEFAULT '0',
  `order_course_detail_id` bigint NOT NULL DEFAULT '0',
  `quote_id` bigint NOT NULL DEFAULT '0',
  `primary_tuition_account_id` bigint NOT NULL DEFAULT '0',
  `class_student_status` int NOT NULL DEFAULT '1',
  `class_time` decimal(18,2) NOT NULL DEFAULT '1.00',
  `student_class_time` decimal(18,2) NOT NULL DEFAULT '1.00',
  `teacher_class_time` decimal(18,2) NOT NULL DEFAULT '0.00',
  `class_time_record_mode` int NOT NULL DEFAULT '1',
  `last_finished_lesson_day` datetime DEFAULT NULL,
  `class_properties_json` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tcs_inst_class_ocd` (`inst_id`,`teaching_class_id`,`order_course_detail_id`),
  KEY `idx_teaching_class_student_class` (`inst_id`,`teaching_class_id`),
  KEY `idx_teaching_class_student_student` (`inst_id`,`student_id`),
  KEY `idx_teaching_class_student_tuition` (`inst_id`,`primary_tuition_account_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1023 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学班级学员关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_class_teacher`
--

DROP TABLE IF EXISTS `teaching_class_teacher`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_class_teacher` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_class_id` bigint NOT NULL,
  `teacher_id` bigint NOT NULL,
  `status` int NOT NULL DEFAULT '1',
  `is_default` tinyint(1) NOT NULL DEFAULT '0',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_teaching_class_teacher` (`inst_id`,`teaching_class_id`,`teacher_id`),
  KEY `idx_teaching_class_teacher_class` (`inst_id`,`teaching_class_id`),
  KEY `idx_teaching_class_teacher_teacher` (`inst_id`,`teacher_id`)
) ENGINE=InnoDB AUTO_INCREMENT=915 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学班级教师关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_record`
--

DROP TABLE IF EXISTS `teaching_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `class_id` bigint NOT NULL,
  `lesson_day` date NOT NULL,
  `start_minutes` int NOT NULL DEFAULT '0',
  `end_minutes` int NOT NULL DEFAULT '0',
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_teaching_record_slot` (`inst_id`,`class_id`,`lesson_day`,`start_minutes`,`end_minutes`),
  KEY `idx_teaching_record_class_day` (`inst_id`,`class_id`,`lesson_day`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='教学记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_schedule`
--

DROP TABLE IF EXISTS `teaching_schedule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_schedule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `class_type` int NOT NULL DEFAULT '0',
  `teaching_class_id` bigint NOT NULL DEFAULT '0',
  `teaching_class_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `lesson_id` bigint NOT NULL DEFAULT '0',
  `lesson_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `teacher_id` bigint NOT NULL DEFAULT '0',
  `teacher_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `assistant_ids_json` json DEFAULT NULL,
  `assistant_names_json` json DEFAULT NULL,
  `classroom_id` bigint NOT NULL DEFAULT '0',
  `classroom_name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `lesson_date` date NOT NULL,
  `lesson_start_at` datetime NOT NULL,
  `lesson_end_at` datetime NOT NULL,
  `batch_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `batch_size` int NOT NULL DEFAULT '1',
  `status` int NOT NULL DEFAULT '1',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_teaching_schedule_inst_date` (`inst_id`,`lesson_date`),
  KEY `idx_teaching_schedule_teacher` (`inst_id`,`teacher_id`,`lesson_date`),
  KEY `idx_teaching_schedule_classroom` (`inst_id`,`classroom_id`,`lesson_date`),
  KEY `idx_teaching_schedule_batch` (`inst_id`,`batch_no`)
) ENGINE=InnoDB AUTO_INCREMENT=18515 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学排课表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_schedule_batch_meta`
--

DROP TABLE IF EXISTS `teaching_schedule_batch_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_schedule_batch_meta` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `batch_key` varchar(96) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `batch_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `class_type` int NOT NULL DEFAULT '0',
  `teaching_class_id` bigint NOT NULL DEFAULT '0',
  `meta_json` json DEFAULT NULL,
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_teaching_schedule_batch_meta_key` (`inst_id`,`batch_key`),
  KEY `idx_teaching_schedule_batch_meta_batch_no` (`inst_id`,`batch_no`)
) ENGINE=InnoDB AUTO_INCREMENT=67 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学排课批次元数据表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `teaching_schedule_student`
--

DROP TABLE IF EXISTS `teaching_schedule_student`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `teaching_schedule_student` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `teaching_schedule_id` bigint NOT NULL DEFAULT '0',
  `teaching_class_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_type` int NOT NULL DEFAULT '1',
  `roster_status` int NOT NULL DEFAULT '1',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_teaching_schedule_student_unique` (`inst_id`,`teaching_schedule_id`,`student_id`),
  KEY `idx_teaching_schedule_student_schedule` (`inst_id`,`teaching_schedule_id`),
  KEY `idx_teaching_schedule_student_student` (`inst_id`,`student_id`)
) ENGINE=InnoDB AUTO_INCREMENT=103 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教学排课学员关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `template_message_record`
--

DROP TABLE IF EXISTS `template_message_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `template_message_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `business_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `channel` int NOT NULL DEFAULT '0',
  `template_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `template_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `notify_count` int NOT NULL DEFAULT '0',
  `success_count` int NOT NULL DEFAULT '0',
  `skipped_count` int NOT NULL DEFAULT '0',
  `failed_count` int NOT NULL DEFAULT '0',
  `operator_id` bigint NOT NULL DEFAULT '0',
  `operator_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_template_message_record_inst_business` (`inst_id`,`business_type`,`del_flag`,`id`),
  KEY `idx_template_message_record_inst_time` (`inst_id`,`business_type`,`create_time`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模板消息发送记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `template_message_record_item`
--

DROP TABLE IF EXISTS `template_message_record_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `template_message_record_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `record_id` bigint NOT NULL,
  `business_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `channel` int NOT NULL DEFAULT '0',
  `relation_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `student_id` bigint NOT NULL DEFAULT '0',
  `student_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sex` int DEFAULT NULL,
  `avatar` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `phone` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `biz_title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `biz_summary` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` int NOT NULL DEFAULT '0',
  `status_reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `recipient_count` int NOT NULL DEFAULT '0',
  `success_recipient_count` int NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_template_message_record_item_record` (`inst_id`,`record_id`,`del_flag`,`id`),
  KEY `idx_template_message_record_item_relation` (`inst_id`,`business_type`,`relation_id`,`id`),
  KEY `idx_template_message_record_item_student` (`inst_id`,`student_id`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模板消息发送明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_domain`
--

DROP TABLE IF EXISTS `tenant_domain`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_domain` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `domain` varchar(255) NOT NULL,
  `entry_type` varchar(32) NOT NULL DEFAULT 'institution-admin',
  `is_primary` tinyint(1) NOT NULL DEFAULT '0',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_domain_domain` (`domain`),
  KEY `idx_tenant_domain_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=772 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户域名表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_institution`
--

DROP TABLE IF EXISTS `tenant_institution`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_institution` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `institution_id` bigint NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_institution_inst` (`institution_id`),
  KEY `idx_tenant_institution_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=157 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户机构关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_menu`
--

DROP TABLE IF EXISTS `tenant_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `menu_id` bigint NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_menu` (`tenant_id`,`menu_id`),
  KEY `idx_tenant_menu_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=21647 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户菜单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_module`
--

DROP TABLE IF EXISTS `tenant_module`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_module` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `module_id` bigint NOT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_module` (`tenant_id`,`module_id`),
  KEY `idx_tenant_module_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户模块授权表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_profile`
--

DROP TABLE IF EXISTS `tenant_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `tenant_name` varchar(128) NOT NULL,
  `tenant_type` varchar(32) NOT NULL DEFAULT 'partner',
  `parent_tenant_id` varchar(64) DEFAULT NULL,
  `edition` varchar(64) NOT NULL DEFAULT 'enterprise',
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `isolation_mode` varchar(32) NOT NULL DEFAULT 'shared_db',
  `brand_config` json DEFAULT NULL,
  `remark` varchar(500) DEFAULT NULL,
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_profile_tenant` (`tenant_id`),
  KEY `idx_tenant_profile_parent` (`parent_tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户档案表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_storage_config`
--

DROP TABLE IF EXISTS `tenant_storage_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_storage_config` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `provider` varchar(32) NOT NULL DEFAULT 'qiniu',
  `access_key` varchar(255) NOT NULL DEFAULT '',
  `secret_key` varchar(255) NOT NULL DEFAULT '',
  `bucket` varchar(128) NOT NULL DEFAULT '',
  `bucket_host` varchar(255) NOT NULL DEFAULT '',
  `upload_prefix` varchar(255) NOT NULL DEFAULT '',
  `expires_seconds` bigint NOT NULL DEFAULT '72000',
  `image_max_size` bigint NOT NULL DEFAULT '10485760',
  `image_mime_types` varchar(255) NOT NULL DEFAULT 'image/*',
  `video_max_size` bigint NOT NULL DEFAULT '104857600',
  `video_mime_types` varchar(255) NOT NULL DEFAULT 'video/*',
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `remark` varchar(500) NOT NULL DEFAULT '',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_storage_provider` (`tenant_id`,`provider`),
  KEY `idx_tenant_storage_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户对象存储配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tenant_user`
--

DROP TABLE IF EXISTS `tenant_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tenant_user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `tenant_id` varchar(64) NOT NULL,
  `user_id` bigint NOT NULL,
  `user_role` varchar(32) NOT NULL DEFAULT 'tenant_admin',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user` (`tenant_id`,`user_id`),
  KEY `idx_tenant_user_tenant` (`tenant_id`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=1866 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户用户关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tuition_account`
--

DROP TABLE IF EXISTS `tuition_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tuition_account` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uuid` varchar(255) DEFAULT NULL COMMENT 'UUID',
  `version` bigint DEFAULT '0' COMMENT '版本号',
  `inst_id` bigint NOT NULL COMMENT '机构ID',
  `student_id` bigint NOT NULL COMMENT '学员ID',
  `order_id` bigint NOT NULL COMMENT '订单ID',
  `order_course_detail_id` bigint NOT NULL COMMENT '订单课程详情ID',
  `course_id` bigint NOT NULL COMMENT '课程ID',
  `quote_id` bigint DEFAULT NULL COMMENT '报价单ID',
  `total_quantity` decimal(20,2) DEFAULT '0.00' COMMENT '购买数量（实际购买，不含赠送）',
  `free_quantity` decimal(20,2) DEFAULT '0.00' COMMENT '赠送数量',
  `used_quantity` decimal(20,2) DEFAULT '0.00' COMMENT '已使用数量（支持小数）',
  `remaining_quantity` decimal(20,2) DEFAULT '0.00' COMMENT '剩余数量（支持小数）',
  `total_tuition` decimal(20,2) DEFAULT '0.00' COMMENT '总学费',
  `paid_tuition` decimal(20,2) DEFAULT '0.00' COMMENT '已支付学费',
  `used_tuition` decimal(20,2) DEFAULT '0.00' COMMENT '已用学费',
  `remaining_tuition` decimal(20,2) DEFAULT '0.00' COMMENT '剩余学费',
  `confirmed_tuition` decimal(20,2) DEFAULT '0.00' COMMENT '已确认学费',
  `status` tinyint DEFAULT '1' COMMENT '学费账户状态：1-正常 2-已停课 3-已结课',
  `handle_type` int DEFAULT NULL COMMENT '办理类型：0-试听 1-报读 2-续费 3-转课',
  `enable_expire_time` tinyint(1) DEFAULT '0' COMMENT '是否启用过期时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `valid_date` date DEFAULT NULL COMMENT '有效开始时间',
  `end_date` date DEFAULT NULL COMMENT '有效结束时间',
  `suspended_time` datetime DEFAULT NULL COMMENT '停课时间',
  `plan_suspend_time` datetime DEFAULT NULL COMMENT '计划停课时间',
  `plan_resume_time` datetime DEFAULT NULL COMMENT '计划恢复时间',
  `class_ending_time` datetime DEFAULT NULL COMMENT '结课时间',
  `status_change_time` datetime DEFAULT NULL COMMENT '状态变更时间',
  `assigned_class` tinyint(1) DEFAULT '0' COMMENT '是否已分配班级',
  `can_transfer` tinyint(1) DEFAULT '1' COMMENT '是否可以转学费账户',
  `has_grade_upgrade` tinyint(1) DEFAULT '0' COMMENT '是否有升班',
  `create_id` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_id` bigint DEFAULT NULL COMMENT '更新人ID',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `del_flag` tinyint(1) DEFAULT '0' COMMENT '删除标志：0-未删除 1-已删除',
  PRIMARY KEY (`id`),
  KEY `idx_student_id` (`student_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_course_id` (`course_id`),
  KEY `idx_inst_id` (`inst_id`),
  KEY `idx_status` (`status`),
  KEY `idx_order_course_detail_id` (`order_course_detail_id`)
) ENGINE=InnoDB AUTO_INCREMENT=521 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='课消账户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tuition_account_flow`
--

DROP TABLE IF EXISTS `tuition_account_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tuition_account_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `inst_id` bigint NOT NULL,
  `tuition_account_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `product_id` bigint NOT NULL,
  `lesson_type` int DEFAULT NULL,
  `lesson_charging_mode` int DEFAULT NULL,
  `source_type` int NOT NULL,
  `source_id` bigint NOT NULL DEFAULT '0',
  `teaching_record_id` bigint DEFAULT NULL,
  `order_number` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `created_time` datetime NOT NULL,
  `quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `balance_quantity` decimal(18,2) NOT NULL DEFAULT '0.00',
  `balance_tuition` decimal(18,2) NOT NULL DEFAULT '0.00',
  `create_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_id` bigint NOT NULL DEFAULT '0',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tuition_account_flow_backfill` (`inst_id`,`tuition_account_id`,`source_type`,`source_id`),
  KEY `idx_tuition_account_flow_list` (`inst_id`,`created_time`,`id`),
  KEY `idx_tuition_account_flow_product` (`inst_id`,`product_id`),
  KEY `idx_tuition_account_flow_student` (`inst_id`,`student_id`),
  KEY `idx_tuition_account_flow_source` (`inst_id`,`source_type`)
) ENGINE=InnoDB AUTO_INCREMENT=6286 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课消账户流水表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `vbmapp_material_item`
--

DROP TABLE IF EXISTS `vbmapp_material_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `vbmapp_material_item` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `library_scope` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'platform',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `profile_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `material_code` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `material_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `material_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `sort_no` int NOT NULL DEFAULT '0',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vbmapp_material_item` (`scale_version`,`library_scope`,`inst_id`,`profile_id`,`material_code`),
  KEY `idx_vbmapp_material_item_profile` (`scale_version`,`library_scope`,`inst_id`,`profile_id`,`status`,`del_flag`),
  KEY `idx_vbmapp_material_item_name` (`scale_version`,`material_name`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VB-MAPP素材条目表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `vbmapp_material_profile`
--

DROP TABLE IF EXISTS `vbmapp_material_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `vbmapp_material_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `library_scope` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'platform',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `profile_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `label` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `source_logic` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `suggested_types_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `preparation_checks_json` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `sort_no` int NOT NULL DEFAULT '0',
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vbmapp_material_profile` (`scale_version`,`library_scope`,`inst_id`,`profile_id`),
  KEY `idx_vbmapp_material_profile_scope` (`scale_version`,`library_scope`,`inst_id`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VB-MAPP素材档案表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `vbmapp_response_schema_override`
--

DROP TABLE IF EXISTS `vbmapp_response_schema_override`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `vbmapp_response_schema_override` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL DEFAULT '0',
  `scale_version` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `module_code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `item_code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `override_json` longtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vbmapp_response_schema_override` (`inst_id`,`scale_version`,`module_code`,`item_code`),
  KEY `idx_vbmapp_response_schema_override_scope` (`inst_id`,`scale_version`,`module_code`,`status`,`del_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VB-MAPP响应结构覆盖表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `vbmapp_response_schema_preset`
--

DROP TABLE IF EXISTS `vbmapp_response_schema_preset`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `vbmapp_response_schema_preset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `scale_version` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `module_code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `item_code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `schema_json` longtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `sort_no` int NOT NULL DEFAULT '0',
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `create_id` bigint NOT NULL DEFAULT '0',
  `update_id` bigint NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `del_flag` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_vbmapp_response_schema_preset` (`scale_version`,`module_code`,`item_code`),
  KEY `idx_vbmapp_response_schema_preset_scope` (`scale_version`,`module_code`,`status`,`del_flag`)
) ENGINE=InnoDB AUTO_INCREMENT=213 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='VB-MAPP响应结构预设表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `wechat_official_bind_ticket`
--

DROP TABLE IF EXISTS `wechat_official_bind_ticket`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wechat_official_bind_ticket` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `ticket` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `official_openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `event_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `scene_value` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `inst_id` bigint NOT NULL DEFAULT '0',
  `student_id` bigint NOT NULL DEFAULT '0',
  `status` tinyint NOT NULL DEFAULT '0',
  `expires_at` datetime DEFAULT NULL,
  `used_at` datetime DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wechat_official_bind_ticket_ticket` (`ticket`),
  KEY `idx_wechat_official_bind_ticket_openid` (`official_openid`,`create_time`),
  KEY `idx_wechat_official_bind_ticket_status` (`status`,`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信公众号绑定票据表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `wechat_official_student_binding`
--

DROP TABLE IF EXISTS `wechat_official_student_binding`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wechat_official_student_binding` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inst_id` bigint NOT NULL,
  `student_id` bigint NOT NULL,
  `official_openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `mini_openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `unionid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `subscribed` tinyint(1) NOT NULL DEFAULT '1',
  `last_bind_ticket` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `last_subscribe_time` datetime DEFAULT NULL,
  `last_unsubscribe_time` datetime DEFAULT NULL,
  `bind_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wechat_official_student_binding` (`inst_id`,`student_id`,`official_openid`),
  KEY `idx_wechat_official_student_binding_openid` (`official_openid`,`subscribed`),
  KEY `idx_wechat_official_student_binding_student` (`inst_id`,`student_id`,`subscribed`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信公众号学员绑定表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `wechat_official_user_link`
--

DROP TABLE IF EXISTS `wechat_official_user_link`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `wechat_official_user_link` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `official_openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mini_openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `unionid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `subscribed` tinyint(1) NOT NULL DEFAULT '0',
  `last_subscribe_time` datetime DEFAULT NULL,
  `last_unsubscribe_time` datetime DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wechat_official_user_link_official_openid` (`official_openid`),
  UNIQUE KEY `uk_wechat_official_user_link_mini_openid` (`mini_openid`),
  UNIQUE KEY `uk_wechat_official_user_link_unionid` (`unionid`),
  KEY `idx_wechat_official_user_link_phone` (`phone`,`subscribed`),
  KEY `idx_wechat_official_user_link_subscribed` (`subscribed`,`update_time`)
) ENGINE=InnoDB AUTO_INCREMENT=53 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信公众号用户关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping routines for database 'ybk_rebuild_edu'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-05-21 13:41:17
