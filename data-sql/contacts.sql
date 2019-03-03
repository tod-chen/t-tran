/*
SQLyog Ultimate v12.08 (64 bit)
MySQL - 8.0.13 : Database - t-tran
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
/*Table structure for table `contacts` */

DROP TABLE IF EXISTS `contacts`;

CREATE TABLE `contacts` (
  `uid` bigint(20) unsigned NOT NULL,
  `pid` bigint(20) unsigned NOT NULL,
  `is_male` tinyint(1) NOT NULL,
  `area` varchar(5) COLLATE utf8mb4_general_ci NOT NULL,
  `paperwork_type` tinyint(3) unsigned NOT NULL,
  `paperwork_num` varchar(50) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `status` tinyint(3) unsigned NOT NULL,
  `passenger_type` tinyint(3) unsigned NOT NULL,
  `phone_num` varchar(20) COLLATE utf8mb4_general_ci NOT NULL,
  `tel_num` varchar(20) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `email` varchar(50) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `addr` varchar(200) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `zip_code` varchar(10) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `add_date` date NOT NULL,
  KEY `query` (`uid`,`pid`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

/*Data for the table `contacts` */

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
