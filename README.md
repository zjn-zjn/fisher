# fisher

transfer with saga

## 表结构

conf/ddl.sql 为数据库表结构定义

## 初始化

conf.InitWith*() 初始化配置

## 转移方法

service.Transfer 转移物品

service.Rollback 回滚

service.Inspection 检查与推进

## 备注

推荐使用数据库事务隔离级别：`READ-COMMITTED`

务必保证不同转移之间的transfer_id和transfer_scene联合唯一
