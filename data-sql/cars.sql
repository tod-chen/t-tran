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
/*Table structure for table `cars` */

DROP TABLE IF EXISTS `cars`;

CREATE TABLE `cars` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tran_type` varchar(20) DEFAULT NULL,
  `seat_type` varchar(5) DEFAULT NULL,
  `seat_count` tinyint(3) DEFAULT NULL,
  `no_seat_count` tinyint(3) unsigned DEFAULT NULL,
  `remark` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8;

/*Data for the table `cars` */

insert  into `cars`(`id`,`tran_type`,`seat_type`,`seat_count`,`no_seat_count`,`remark`) values (1,'G','S',28,0,'高铁商务座'),(2,'G','FC',51,0,'高铁一等座'),(3,'G','SC',100,5,'高铁二等座'),(4,'G','SC',100,10,'高铁二等座高峰'),(5,'D','S',28,0,'动车商务座'),(6,'D','FC',51,0,'动车一等座'),(7,'D','SC',100,5,'动车二等座'),(8,'D','SC',100,10,'动车二等座高峰'),(9,'D','DS',36,0,'动车卧铺'),(10,'C','FC',51,0,'城际一等座'),(11,'C','SC',127,5,'城际二等座'),(12,'Z','SS',32,0,'直达软卧'),(13,'Z','HS',66,0,'直达硬卧'),(14,'Z','SST',78,0,'直达软座'),(15,'Z','HST',100,10,'直达硬座'),(16,'Z','HST',100,20,'直达硬座高峰'),(17,'T','SS',32,0,'特快软卧'),(18,'T','HS',66,0,'特快硬卧'),(19,'T','SST',78,0,'特快软座'),(20,'T','HST',100,10,'特快硬座'),(21,'T','HST',100,20,'特快硬座高峰'),(22,'T','MS',88,0,'特快动卧'),(23,'K','SS',32,0,'普快软卧'),(24,'K','HS',66,0,'普快硬卧'),(25,'K','MS',88,0,'普快动卧'),(26,'K','SST',78,0,'普快软座'),(27,'K','HST',100,10,'普快硬座'),(28,'K','HST',100,20,'普快硬座高峰'),(29,'O','SS',32,0,'其他软卧'),(30,'O','HS',66,0,'其他硬卧'),(31,'O','SST',78,0,'其他软座'),(32,'O','HST',100,10,'其他硬座'),(33,'O','HST',100,20,'其他硬座高峰');

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
