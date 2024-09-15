# coin-trade

coin trade with saga

## 表结构

conf/ddl.sql 为数据库表结构定义

## 初始化

conf.InitWith*() 初始化配置

## 交易方法

service.CoinTrade 虚拟币交易

service.RollbackTrade 交易回滚

service.CoinTradeInspection 交易检查推进

## 备注

推荐使用数据库事务隔离级别：`READ-COMMITTED`