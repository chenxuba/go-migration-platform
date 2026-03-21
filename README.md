# Go Migration Platform

这是一个面向 SaaS 定制化场景的 Go 微服务迁移骨架。

目标不是把服务拆很多，而是控制在 3 个核心服务：

- `iam-service`：登录、用户、租户、角色权限、SSO 迁移承载点
- `platform-service`：租户配置、功能开关、字典、菜单、品牌、规则绑定
- `education-service`：机构端与核心业务域，如学员、订单、审批、报表

## 为什么只拆 3 个服务

- 避免把当前 Java 项目过早拆成十几个服务，增加迁移复杂度
- 保留后续继续拆分的空间，但初期按业务大域划分即可
- SaaS 定制差异优先通过租户配置、规则包、工作流方案、外部集成来承载

## SaaS 设计原则

- 服务按业务域拆，不按客户拆
- 所有请求默认带 `X-Tenant-ID`
- 客户差异通过配置和规则表达，不写死 `if tenant == A`
- 支持不同客户启用不同功能、字段、工作流和集成

## 目录

```text
configs/                  # 示例租户配置
docs/                     # 架构说明
pkg/config                # 服务配置加载
pkg/customization         # 租户配置、功能开关、规则包
pkg/httpx                 # 统一响应
pkg/logx                  # 结构化日志
pkg/tenant                # 租户上下文与中间件
services/
  iam/
  platform/
  education/
```

## 启动

```bash
cd /Users/chenrui/Desktop/go-migration-platform
go run ./services/iam/cmd/api
go run ./services/platform/cmd/api
go run ./services/education/cmd/api
```

或者：

```bash
./scripts/dev-up.sh
```

默认端口：

- `iam-service`: `8081`
- `platform-service`: `8082`
- `education-service`: `8083`

## 当前已经完成的真实迁移切片

### iam-service

已接真实 MySQL `ybk_rebuild_edu`：

- 兼容老 SSO 登录接口：`POST /sso/doLogin`
- 当前登录信息：`GET /sso/info`
- 登录状态：`GET /sso/isLogin`
- 菜单列表：`GET /sso/menuList`
- 角色列表：`GET /sso/roleList`
- 管理后台用户分页：`GET /api/v1/users`

已验证可用账号：

- `admin / 123456`：管理后台
- `chenrui / 123456`：机构后台

### platform-service

已接真实平台数据：

- 字典分页：`GET /sysDict/page`
- 字典值查询：`GET /sysDictValue/listByCode`
- 系统模块分页：`GET /sysModule/page`
- 租户功能集：`GET /api/v1/tenant/features`
- 租户定制摘要：`GET /api/v1/tenant/customization-summary`

### education-service

已接真实机构业务数据：

- 机构渠道分类：`GET /instChannelCategory/getChannelCategoryList`
- 机构渠道树：`GET /instChannel/getChannelTree`
- 系统默认渠道：`GET /instChannel/getDefaultChannelList`
- 机构自有渠道：`GET /instChannel/getChannelList`
- 课程分类分页：`POST /instCourseCategory/getCourseCategoryPage`
- 课程分页：`POST /instCourse/getCoursePage`
- 课程下拉：`POST /instCourse/getCourseIdAndName`
- 意向学员分页：`POST /instStudent/getIntendedStudentPage`
- 意向学员详情：`GET /instStudent/getIntentStudentDetail`
- 在读学员分页：`POST /instStudent/getCurrentStudentPage`
- 财务订单列表：`POST /saleOrder/getOrderList`
- 财务订单详情：`POST /saleOrder/getOrderDetail`
- ES 基础状态：`GET /esSync/status`
- 意向学生 ES 同步入口：`POST /esSync/syncIntentStudent`

## SaaS 试用方式

请求时带上租户头：

```bash
curl -H 'X-Tenant-ID: tenant-a' http://127.0.0.1:8082/api/v1/tenant/features
curl -H 'X-Tenant-ID: tenant-b' http://127.0.0.1:8083/api/v1/students
```
