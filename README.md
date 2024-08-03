# coin-trade

基于saga语义的转账交易实现，如A转账给B和C

conf/ddl.sql 为数据库表结构定义

conf.InitWith*() 初始化配置

service.CoinTrade 虚拟币交易

service.RollbackTrade 交易回滚

service.CoinTradeInspection 交易检查推进
