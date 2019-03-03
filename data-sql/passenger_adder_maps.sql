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
/*Table structure for table `passenger_adder_maps` */

DROP TABLE IF EXISTS `passenger_adder_maps`;

CREATE TABLE `passenger_adder_maps` (
  `pid` bigint(20) unsigned DEFAULT NULL,
  `uid` bigint(20) unsigned DEFAULT NULL,
  KEY `query` (`pid`,`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

/*Data for the table `passenger_adder_maps` */

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
