# Fisher

[![Go Reference](https://pkg.go.dev/badge/github.com/zjn-zjn/fisher.svg)](https://pkg.go.dev/github.com/zjn-zjn/fisher)
[![Go Report Card](https://goreportcard.com/badge/github.com/zjn-zjn/fisher)](https://goreportcard.com/report/github.com/zjn-zjn/fisher)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## 简介

Fisher 是一个高可用的分布式资产转移系统，基于本地事务+状态表模式，实现N个账户扣除资产，M个账户接收的操作。系统保证资产转移的一致性和可追溯性，适用于各类需要跨账户资产转移的业务场景。

### 使用场景

**充值场景：** 官方账户扣减，用户账户增加

**买卖场景：** A账户扣减，B账户增加，C账户收取分成，官方D账户收取手续费

**合买场景：** A账户和B账户扣减，C账户收取款项，官方D账户收取手续费

**退款场景：** 用户账户扣减，官方账户增加

### 特色

- 预定义官方账户，官方账户与用户账户隔离，官方账户允许余额为负
- 不存在莫名其妙的增加和减少操作，所有资产转移都有迹可循，所有交易达到最终态时，账户余额总和为0
- 官方账户是个区间，交易过程中官方账户离散，避免热点官方账户
- 支持半成功，如A、B账户扣减完成时就算成功，C、D账户资产增加即可持续推进（除非手动调用回滚接口）
- 支持分库分表(跨库事务)
- 提供自动检查推进机制，确保系统最终一致性

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
1. **INIT(初始化)**：创建转移请求，记录转移元数据
2. **DEBITING(扣减中)**：执行源账户资产扣减操作
3. **HALF_SUCCESS(半成功)**：源账户扣减成功，等待目标账户增加
4. **SUCCESS(成功)**：所有源账户扣减和目标账户增加均完成
5. **FAILED(失败)**：转移失败，需执行补偿操作或回滚
6. **ROLLBACKING(回滚中)**：执行回滚操作，恢复账户状态
7. **ROLLBACKED(已回滚)**：回滚完成，所有资产变更已撤销

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
    DBs:               dbs,           // 数据库连接池
    StateSplitNum:     10,            // 转移状态分表数量
    RecordSplitNum:    10,            // 转移记录分表数量
    AccountSplitNum:   10,            // 账户分表数量
    OfficialAcctStep:  100000000,     // 官方账户类型步长
    OfficialAcctMin:   1,             // 官方账户最小值
    OfficialAcctMax:   100000000000,  // 官方账户最大值
})
```

### 使用示例

#### 转移示例

```go
// 简单转移示例：从账户A转移100单位资产到账户B
err := service.Transfer(ctx, &model.TransferReq{
    TransferId:     12345,            // 转移ID，确保唯一
    UseHalfSuccess: true,             // 启用半成功机制
    ItemType:       1,                // 物品类型（如金币、钻石等）
    FromAccounts: []*model.TransferItem{
        {
            AccountId:   100,         // 源账户ID
            Amount:      100,         // 扣减数量
            ChangeType:  1,           // 扣减类型
            Comment:     "购买商品",    // 备注
        },
    },
    ToAccounts: []*model.TransferItem{
        {
            AccountId:   200,         // 目标账户ID
            Amount:      98,          // 增加数量
            ChangeType:  2,           // 增加类型
            Comment:     "销售商品收入", // 备注
        },
        {
            AccountId:   300,         // 手续费账户ID（可以是官方账户）
            Amount:      2,           // 手续费数量
            ChangeType:  3,           // 增加类型
            Comment:     "交易手续费",  // 备注
        },
    },
    TransferScene: 1,                 // 转移场景
    Comment:       "商品交易",         // 转移备注
})

// 使用官方账户示例（官方账户ID在OfficialAcctMin和OfficialAcctMax之间）
err := service.Transfer(ctx, &model.TransferReq{
    TransferId:     12346,
    UseHalfSuccess: true,
    ItemType:       1,
    FromAccounts: []*model.TransferItem{
        {
            AccountId:   1,           // 官方账户ID
            Amount:      100,
            ChangeType:  10,
            Comment:     "系统赠送",
        },
    },
    ToAccounts: []*model.TransferItem{
        {
            AccountId:   500,         // 用户账户ID
            Amount:      100,
            ChangeType:  11,
            Comment:     "系统奖励",
        },
    },
    TransferScene: 2,
    Comment:       "系统赠送",
})
```

#### 回滚示例

```go
// 回滚之前的转移
err := service.Rollback(ctx, &model.RollbackReq{
    TransferId:    12345,             // 要回滚的转移ID
    TransferScene: 1,                 // 转移场景
})
```

#### 检查推进示例

```go
// 检查1小时前的未完成转移
now := time.Now().Unix()
oneHourAgo := now - 3600
errors := service.Inspection(ctx, oneHourAgo)
```

### 操作接口详情

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
