# Fisher

[![Go Reference](https://pkg.go.dev/badge/github.com/zjn-zjn/fisher.svg)](https://pkg.go.dev/github.com/zjn-zjn/fisher)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## 简介

Fisher 是一个高可用的分布式资产转移系统，基于本地事务+状态表模式，实现N个账户扣除资产，M个账户接收的操作。系统保证资产转移的一致性和可追溯性，适用于各类需要跨账户资产转移的业务场景。

### 使用场景

**充值场景：** 官方账户扣减，用户账户增加

**买卖场景：** A账户扣减，B账户增加，C账户收取分成，官方D账户收取手续费

**合买场景：** A账户和B账户扣减，C账户收取款项，官方D账户收取手续费

**退款场景：** 用户账户扣减，官方账户增加

### 特色

- **官方账户隔离设计**：预定义官方账户区间，与用户账户隔离管理，允许官方账户余额为负，满足特殊业务场景需求
- **完美的零和对账系统**：任何时刻系统中所有账户余额之和严格为0，这一数学特性提供了强大的自检机制，简化对账流程，使资产异常无所遁形
- **热点账户避免机制**：官方账户采用区间设计，交易时动态分散，有效避免热点账户问题，提升系统并发处理能力
- **半成功状态支持**：源账户扣减完成即视为半成功，目标账户增加操作可持续推进，显著提高系统可用性
- **分库分表支持**：内置分库分表能力，支持跨库事务处理，满足大规模业务数据存储需求
- **自动恢复机制**：提供自检和自动推进功能，确保系统最终一致性，减少人工干预成本

## 架构原理

Fisher 采用基于SAGA模式的分布式事务实现，通过"本地事务+补偿事务"的组合确保跨账户资产转移的一致性：

1. **资产扣减阶段**：首先执行源账户资产扣减的本地事务
2. **资产增加阶段**：执行目标账户资产增加的本地事务
3. **状态管理阶段**：根据执行结果更新转移状态为成功、半成功或失败
4. **补偿机制**：
   - 如果任一阶段失败，则触发对应的补偿事务，恢复系统一致性
   - 对于半成功状态，系统会持续推进资产增加操作，不执行补偿
   - 提供自动检查机制，确保长时间未完成的转移被适当处理

这种基于SAGA的设计使系统能够在分布式环境中保持数据一致性，同时通过半成功机制提升了系统的可用性。相比传统的两阶段提交(2PC)协议，该方案具有更高的性能和更低的资源锁定成本。

### 系统架构

系统采用三层架构：
- **Model层**：定义数据模型和请求结构
- **DAO层**：负责数据访问和事务处理
- **Service层**：实现业务逻辑和交易流程

### 数据模型

系统使用三张核心表：
- **state表**：记录转移状态和过程
- **record表**：记录具体的转移记录和补偿操作
- **account表**：记录账户资产信息和余额变更

### 状态流转

转移操作经历以下状态流转：
1. **StateStatusDoing (1)**：转移进行中，执行源账户资产扣减和目标账户增加操作
2. **StateStatusRollbackDoing (2)**：回滚进行中，执行补偿操作恢复账户状态
3. **StateStatusHalfSuccess (3)**：半成功状态，源账户扣减成功，等待目标账户增加
4. **StateStatusSuccess (4)**：转移成功，所有源账户扣减和目标账户增加均完成
5. **StateStatusRollbackDone (5)**：回滚完成，所有资产变更已撤销

记录状态定义：
1. **RecordStatusNormal (1)**：正常记录状态
2. **RecordStatusRollback (2)**：回滚记录状态
3. **RecordStatusEmptyRollback (3)**：空回滚记录状态（未执行原操作的回滚）

## 安装与依赖

### 依赖

- Go 1.22+
- MySQL 5.7+

### 安装

```bash
go get github.com/zjn-zjn/fisher
```

## 快速使用

### 表结构

conf/ddl.sql 为数据库表结构定义，包含了state、record和account三张核心表。

### 初始化

两种初始化方式：

```go
// 方式一：使用默认配置初始化（单表）
err := basic.InitWithDefault(dbs []*gorm.DB)

// 方式二：使用自定义配置初始化
err := basic.InitWithConf(&basic.TransferConf{
    DBs:                 dbs,          // 数据库列表
    StateSplitNum:       3,            // 转移状态分表数量
    RecordSplitNum:      3,            // 转移记录分表数量
    AccountSplitNum:     3,            // 账户分表数量
    OfficialAccountStep: 100000000,    // 官方账户步长
    OfficialAccountMin:  1,            // 官方账户最小值
    OfficialAccountMax:  100000000000  // 官方账户最大值
})
```

### 使用示例

以下是一个完整的资产转移示例，展示了用户购买商品时的资金流向，包括卖家收款、版权分成以及平台手续费（官方账户）：

```go
// 定义常量
const (
    ItemTypeGold basic.ItemType = 1    // 资产类型：金币
    OfficialAccountTypeFee basic.OfficialAccountType = 100000000  // 官方账户：手续费账户
    
    TransferSceneBuyGoods basic.TransferScene = 1    // 转账场景：购买商品
    ChangeTypeSpend basic.ChangeType = 1             // 变更类型：消费支出
    ChangeTypeSellGoodsIncome basic.ChangeType = 2   // 变更类型：商品销售收入
    ChangeTypeSellGoodsCopyright basic.ChangeType = 3 // 变更类型：版权分成收入
    ChangeTypePlatformFee basic.ChangeType = 4       // 变更类型：平台手续费
)

// 执行资产转移
ctx := context.Background()
buyerAccountId := int64(100000000001)     // 买家账户ID
sellerAccountId := int64(100000000002)    // 卖家账户ID
copyrightAccountId := int64(100000000003) // 版权方账户ID

err := service.Transfer(ctx, &model.TransferReq{
    TransferId:     12345,                // 转移ID，和转移场景确保联合唯一
    UseHalfSuccess: true,                 // 启用半成功机制
    ItemType:       ItemTypeGold,         // 物品类型：金币
    TransferScene:  TransferSceneBuyGoods, // 转账场景：购买商品
    Comment:        "购买数字商品",         // 转移备注
    
    // 资金来源账户
    FromAccounts: []*model.TransferItem{
        {
            AccountId:  buyerAccountId,   // 买家账户
            Amount:     100,              // 总金额
            ChangeType: ChangeTypeSpend,  // 变更类型：消费支出
            Comment:    "购买数字商品支出",
        },
    },
    
    // 资金目标账户
    ToAccounts: []*model.TransferItem{
        {
            AccountId:  sellerAccountId,  // 卖家账户
            Amount:     85,               // 卖家获得85%
            ChangeType: ChangeTypeSellGoodsIncome,
            Comment:    "商品销售收入",
        },
        {
            AccountId:  copyrightAccountId, // 版权方账户
            Amount:     10,                 // 版权方获得10%
            ChangeType: ChangeTypeSellGoodsCopyright,
            Comment:    "版权分成收入",
        },
        {
            AccountId:  int64(OfficialAccountTypeFee), // 官方手续费账户
            Amount:     5,                            // 平台收取5%手续费
            ChangeType: ChangeTypePlatformFee,
            Comment:    "平台手续费",
        },
    },
})

if err != nil {
    // 处理错误
    log.Println("转账失败:", err)
    return
}

// 成功处理
log.Println("转账成功")
```

更多复杂场景示例请参考测试代码：[demo_test.go](service/demo_test.go) 和 [extreme_test.go](service/extreme_test.go)

### 结果查询与验证

1. 查询state表了解整体转移状态
2. 查询record表了解具体的转移记录和状态
3. 检查account表确认账户余额变更是否符合预期
4. 使用Inspection接口推进半成功状态或回滚错误转移

## 操作接口详情

#### Transfer 转移接口

- 入参
    - req
        - TransferId 转移ID
        - UseHalfSuccess 是否使用半成功
        - ItemType 转移物品类型
        - FromAccounts 转移发起者
            - AccountId 发起账户ID
            - Amount 扣减数量
            - ChangeType 扣减类型
            - Comment 扣减备注
        - ToAccounts 转移接收者
            - AccountId 接收账户ID
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

## 最佳实践

- **唯一性保证**：务必保证不同转移之间的transfer_id和transfer_scene联合唯一
- **事务隔离级别**：推荐使用数据库事务隔离级别：`READ-COMMITTED`
- **定期检查**：建议设置定时任务执行Inspection接口，处理半成功状态的转移
- **官方账户管理**：合理设置官方账户区间，避免耗尽可用账户
- **异常监控**：对系统错误和半成功状态进行监控，便于及时发现问题
- **分表策略**：根据业务量合理配置分表数量，避免单表数据过大
- **补偿监控**：关注补偿事务执行情况，确保资产一致性不被破坏
- **并发控制**：对同一账户的多笔并发转移建议进行业务侧排队处理，降低死锁风险

## 性能与扩展

- 系统支持水平扩展，通过分表配置可扩展到多个分片
- 官方账户区间设计有效避免了热点账户问题
- 半成功机制提高了系统整体成功率和可用性
- 通过合理的索引设计和查询优化提升系统处理能力
- SAGA模式降低了资源锁定时间，提高了并发处理能力

## 故障排除

### 常见错误

1. **InsufficientAmountErr**：账户余额不足，请检查源账户余额是否充足
2. **AlreadyRolledBackErr**：转移已被回滚，无法执行新操作
3. **StateMutationErr**：状态变更错误，可能是并发操作导致

### 问题排查步骤

1. 使用转移ID和场景查询state表，检查转移状态
2. 查询record表了解具体的转移记录和状态
3. 检查account表确认账户余额变更是否符合预期
4. 使用Inspection接口推进半成功状态或回滚错误转移

## 许可证

本项目采用Apache License 2.0许可证 - 详见LICENSE文件
