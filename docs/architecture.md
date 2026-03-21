# Architecture

## 服务边界

### 1. iam-service

负责：

- 用户、角色、权限
- 登录、会话、令牌
- SSO 迁移
- 租户基础身份信息

对照当前 Java：

- `sso`

### 2. platform-service

负责：

- 租户配置
- 功能开关
- 品牌主题
- 菜单、字典、系统配置
- 规则包、工作流方案绑定
- 公共后台能力

对照当前 Java：

- `common`
- `system/operatelog`
- `manage` 中偏平台能力部分

### 3. education-service

负责：

- 学员
- 课程
- 订单
- 审批
- 机构业务
- 业务报表

对照当前 Java：

- `institution`
- `manage` 中偏教育业务部分

## SaaS 预留点

每个租户至少有：

- `tenant_id`
- `edition`
- `feature_flags`
- `branding`
- `custom_fields`
- `workflow_scheme`
- `rule_pack`
- `integrations`

## 后续迁移策略

第一批优先迁：

- 登录和租户上下文基础设施
- 租户配置中心
- 平台型 CRUD
- 轻量查询接口

第二批迁：

- 学员和机构主数据
- 订单和审批
- 异步任务与 ES 同步

第三批迁：

- 文档模板
- Excel/Word/PDF 重度生成
- 复杂对外集成
