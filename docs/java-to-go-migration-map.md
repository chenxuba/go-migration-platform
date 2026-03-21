# Java To Go Migration Map

## 迁移目标

不是把当前 Java 工程按模块原样复制到 Go，而是收敛成 3 个核心服务。

## 服务映射

### iam-service

来源模块：

- `sso`

优先迁移内容：

- 登录
- 用户
- 角色
- 菜单授权
- 机构用户
- 登录日志
- 错误日志写入入口

暂缓内容：

- 完整 SSO 对接细节
- 微信相关复杂登录联动

### platform-service

来源模块：

- `common`
- `system/operatelog`
- `manage` 中平台能力部分

优先迁移内容：

- 租户配置
- 字典
- 系统菜单
- 平台公告
- 设备信息
- 七牛上传 token
- 功能开关
- 工作流方案绑定
- 规则包绑定

暂缓内容：

- 重度 Office 文档能力
- 与多个外部平台的深度集成

### education-service

来源模块：

- `institution`
- `manage` 中教育业务部分

优先迁移内容：

- 学员
- 课程
- 渠道
- 订单
- 审批
- 跟进记录
- 机构配置

暂缓内容：

- ES 同步
- RocketMQ 消费
- XXL-Job 复杂任务
- Word/Excel/PDF 模板生成

## 第一批建议落地顺序

1. `sso` 中用户、角色、登录基础接口迁入 `iam-service`
2. `manage` 中字典、菜单、公告、设备等平台接口迁入 `platform-service`
3. `institution` 中渠道、渠道分类、课程分类这类边界清晰模块迁入 `education-service`
4. 学员、订单、审批最后进入主迁移阶段

## 关键注意点

- Java 的公共基类 Controller 不要直接翻译，Go 应改为显式 handler + service
- MyBatis XML 查询先兼容数据库，必要时可先保留原 SQL 语义
- 事务逻辑必须逐个服务重新设计，不能简单照搬 `@Transactional`
- 模板文档导出应最后迁移，优先保证业务主链路
