# Fisher

## 简介

一个简易的账号间数量转移系统，基于本地事务+状态表，实现N个账号扣除数量，M个账号接收的操作。

### 使用场景

**充值场景：** 官方账号扣减，用户账号增加

**买卖场景：** A账号扣减，B账号增加，C账号收取分成，官方D账号收取手续费

**合买场景：** A账号和B账号扣减，C账号收取款项，官方D账号收取手续费

### 特色

- 预定义官方背包，官方背包与用户背包隔离，官方背包允许账户为负
- 不存在莫名其妙的增加和减少操作，所有数量转移都有迹可循，所有交易达到最终态时，背包账户余额总和为0
- 官方背包是个区间，交易过程中官方背包离散，避免热点官方背包
- 支持半成功，如A、B背包扣减完成时就算成功，C、D加背包即可持续推进增加（除非手动调用回滚接口）
- 支持分库分表(跨库事务)

## 快速使用

### 表结构

conf/ddl.sql 为数据库表结构定义

### 初始化

basic.InitWithDefault(dbs []*gorm.DB) 使用默认值初始化配置，默认单表

basic.InitWithConf(conf *TransferConf) 更丰富的配置初始化

### 操作接口

#### Transfer 转移接口

- 入参
    - req
        - TransferId 转移ID
        - UseHalfSuccess 是否使用半成功
        - ItemType 转移物品类型
        - FromBags 转移发起者
            - BagId 发起背包ID
            - Amount 扣减数量
            - ChangeType 扣减类型
            - Comment 扣减备注
        - ToBags 转移接收者
            - BagId 接收背包ID
            - Amount 增加数量
            - ChangeType 增加类型
            - Comment 增加备注
        - TransferScene 转移场景
        - Comment 转移备注
- 返回
    - error 转移错误原因

#### Rollback 回滚转移

- 入参
    - req
        - TransferId 回滚的转移ID
        - TransferScene 回滚的转移场景
- 返回
    - error 回滚错误原因

#### Inspection 检查推进

- 入参
    - lastTime 这个时间之前的所有历史交易状态检查与推进，需注意**如果转移还在进行中，会尝试回滚该转移**
- 返回
    - []error 推进产生错误的列表

## 备注

**务必保证不同转移之间的transfer_id和transfer_scene联合唯一**

**推荐使用数据库事务隔离级别：`READ-COMMITTED`**




