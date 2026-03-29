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

推荐先用一键重启脚本，它会按顺序执行：

- `scripts/ensure-dev-infra.sh`：按需拉起 Elasticsearch、RocketMQ、Canal
- `scripts/preflight-dev-deps.sh`：检查中间件是否就绪
- `scripts/dev-down.sh` / `scripts/dev-up.sh`：重启 3 个 Go 服务

```bash
cd /Users/chenrui/Desktop/go-migration-platform
./scripts/restart.sh
```

如果你只想启动 Go 服务，不检查也不代起中间件：

```bash
./scripts/dev-up.sh
```

如果你想手动跳过中间件拉起，再执行完整重启：

```bash
export SKIP_ENSURE_INFRA=1
./scripts/restart.sh
```

### 本地依赖

默认依赖：

- Elasticsearch：`https://127.0.0.1:9200`
- RocketMQ NameServer：`127.0.0.1:9876`
- RocketMQ Broker：`127.0.0.1:10911`
- Canal：本地 `canal.deployer` 进程

### Elasticsearch 启动方式

可以任选下面两种方式，脚本都支持。

#### 方式一：Homebrew 安装

```bash
brew tap elastic/tap
brew install elastic/tap/elasticsearch-full
brew services start elastic/tap/elasticsearch-full
```

#### 方式二：手工安装目录

如果本机已经有类似 `/usr/local/elasticsearch-8.5.3` 这样的目录，`scripts/ensure-dev-infra.sh` 会自动探测并尝试执行：

```bash
/path/to/elasticsearch/bin/elasticsearch -d
```

也可以显式指定：

```bash
export ES_HOME=/usr/local/elasticsearch-8.5.3
./scripts/restart.sh
```

### 常用环境变量

- `ES_HOME`：手工安装版 Elasticsearch 根目录
- `ES_URI`：默认 `https://127.0.0.1:9200`
- `ES_USERNAME`：默认 `elastic`
- `ES_PASSWORD`：默认见 `pkg/config/config.go`
- `ROCKETMQ_HOME`：默认 `~/rocketmq`
- `CANAL_HOME`：默认 `/usr/local/canal.deployer-1.1.8`；若不存在会尝试探测 `~/canal.deployer-*` 等路径。未安装时脚本会报错并提示设置环境变量；不需要 Canal 时可 `SKIP_ENSURE_INFRA=1` 或 `PREFLIGHT_SKIP_CANAL=1`。
- `ENSURE_INFRA_TIMEOUT`：中间件等待超时，默认 `120`
- `SKIP_ENSURE_INFRA=1`：跳过自动拉起中间件
- `SKIP_PREFLIGHT=1`：跳过依赖预检

### 手动启动服务

```bash
cd /Users/chenrui/Desktop/go-migration-platform
go run ./services/iam/cmd/api
go run ./services/platform/cmd/api
go run ./services/education/cmd/api
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
