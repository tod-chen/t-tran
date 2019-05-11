
DROP TABLE IF EXISTS `orders`;

CREATE TABLE `orders` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `order_num` varchar(15) NOT NULL,
  `user_id` bigint(20) unsigned NOT NULL,
  `price` double NOT NULL,
  `book_time` datetime NOT NULL,
  `pay_time` datetime NOT NULL,
  `pay_type` tinyint(4) unsigned NOT NULL COMMENT '支付类型 1.支付宝 2.微信',
  `pay_account` varchar(30) NOT NULL,
  `status` tinyint(4) unsigned NOT NULL COMMENT '订单状态 0.未支付 1.已取消 2.订单超时 3.已支付 4.已退票',
  PRIMARY KEY (`id`),
  KEY `idx_udi_pt_s` (`user_id`, `pay_time`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=ascii;