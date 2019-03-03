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
/*Table structure for table `orders` */

DROP TABLE IF EXISTS `orders`;

CREATE TABLE `orders` (
  `order_num` varchar(255) DEFAULT NULL,
  `user_id` int(11) DEFAULT NULL,
  `contact_id` int(11) DEFAULT NULL,
  `tran_dep_date` varchar(255) DEFAULT NULL,
  `tran_num` varchar(255) DEFAULT NULL,
  `car_num` tinyint(3) unsigned DEFAULT NULL,
  `seat_num` varchar(255) DEFAULT NULL,
  `seat_type` varchar(255) DEFAULT NULL,
  `departure_station` varchar(255) DEFAULT NULL,
  `check_ticket_gate` varchar(255) DEFAULT NULL,
  `departure_time` datetime DEFAULT NULL,
  `arrival_station` varchar(255) DEFAULT NULL,
  `arrival_time` datetime DEFAULT NULL,
  `price` double DEFAULT NULL,
  `book_time` datetime DEFAULT NULL,
  `pay_time` datetime DEFAULT NULL,
  `pay_type` int(11) DEFAULT NULL,
  `pay_account` varchar(255) DEFAULT NULL,
  `status` tinyint(4) DEFAULT NULL,
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `create_at` datetime DEFAULT NULL,
  `update_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*Data for the table `orders` */

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
