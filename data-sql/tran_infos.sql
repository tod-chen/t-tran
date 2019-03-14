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
/*Table structure for table `tran_infos` */

DROP TABLE IF EXISTS `tran_infos`;

CREATE TABLE `tran_infos` (
  `id` int(4) unsigned NOT NULL AUTO_INCREMENT,
  `tran_num` varchar(10) DEFAULT NULL,
  `route_dep_corss_days` int(4) unsigned DEFAULT NULL,
  `schedule_days` int(4) unsigned DEFAULT '1',
  `is_sale_ticket` tinyint(1) DEFAULT '1',
  `sale_ticket_time` datetime DEFAULT NULL,
  `non_sale_remark` varchar(100) DEFAULT NULL,
  `enable_start_date` datetime DEFAULT '0000-00-00 00:00:00',
  `enable_end_date` datetime DEFAULT '9999-12-31 00:00:00',
  `car_ids` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `main` (`tran_num`)
) ENGINE=InnoDB AUTO_INCREMENT=10427 DEFAULT CHARSET=utf8;

/*Data for the table `tran_infos` */


/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;