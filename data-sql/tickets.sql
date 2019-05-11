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
/*Table structure for table `tickets` */

DROP TABLE IF EXISTS `tickets`;

CREATE TABLE `tickets` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint(20) unsigned NOT NULL,
  `passenger_id` bigint(20) unsigned NOT NULL,
  `is_student` bit(1) NOT NULL,
  `status` tinyint(4) unsigned NOT NULL,
  `price` double NOT NULL,
  `tran_dep_date` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `tran_num` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `car_num` tinyint(4) unsigned NOT NULL,
  `seat_idx` tinyint(4) unsigned NOT NULL,
  `seat_num` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `seat_type` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `check_ticket_gate` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `dep_station` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `dep_station_idx` tinyint(4) unsigned NOT NULL,
  `dep_time` datetime NOT NULL,
  `arr_station` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `arr_station_idx` tinyint(4) unsigned NOT NULL,
  `arr_time` datetime NOT NULL,
  `change_ticket_id` bigint(20) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `q_passenger` (`passenger_id`,`status`),
  KEY `q_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
