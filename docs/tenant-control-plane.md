# 多级租户控制台落地说明

## 当前分支实现范围

本分支先完成“最小可运行底座”：

- 自动创建 `tenant_profile`：租户主表。
- 自动创建 `tenant_institution`：机构归属租户。
- 自动创建 `tenant_menu`：租户可用菜单。
- 自动创建 `tenant_user`：租户后台管理员绑定。
- 自动创建 `tenant_domain`：租户独立域名配置。
- 启动时自动创建 `tenant-a`。
- 启动时把现有 `org_institution` 中未绑定租户的机构全部归到 `tenant-a`。
- 启动时把现有 `sso_menu` 中未绑定租户的菜单全部授权给 `tenant-a`。
- 启动时把现有 `admin` 用户绑定为 `tenant-a` 的 `tenant_admin`。
- 机构端登录时校验机构是否属于当前请求租户。

> 你现在库里如果正好有 4 个机构，那么这 4 个机构会全部绑定到 `tenant-a`。如果不止 4 个，当前策略会把所有历史机构作为迁移存量绑定到 `tenant-a`，避免老数据登录失败。

## 总控后台和子总控后台是不是一套代码

建议是：后端同一套代码，前端也尽量同一套代码，通过登录身份、租户、菜单和权限区分能力。

推荐入口：

- 公司总控后台：继续使用 `platform-admin`，登录类型仍是 `manage`。
- 客户子总控后台：第一阶段也复用 `platform-admin`，但账号绑定到 `tenant_user`，后续再把菜单裁剪成“租户管理、版本管理、菜单授权、机构管理”。
- 机构端后台：继续使用 `institution-admin`，登录类型仍是 `org`。

也就是说：

- 总控和子总控可以先共用 `platform-admin` 代码。
- 机构端所有客户也共用 `institution-admin` 代码。
- 不建议为每个客户复制一套前端代码。

## 租户账号怎么开通

当前最小方案：

1. 由公司总控创建或绑定 `sso_user` 用户。
2. 在 `tenant_user` 中绑定该用户和租户。
3. 用户登录 `platform-admin`。
4. 后续通过菜单权限只开放“子总控”菜单。

当前分支默认把已有 `admin` 用户绑定为 `tenant-a` 的租户管理员：

```sql
SELECT * FROM tenant_user WHERE tenant_id = 'tenant-a';
```

后续正式开发建议新增接口：

- `POST /api/v1/platform/tenants/create`：创建合作客户租户。
- `POST /api/v1/platform/tenants/admin/create`：创建租户管理员账号。
- `POST /api/v1/platform/tenants/modules/replace`：配置客户可售版本。
- `POST /api/v1/platform/tenants/menus/replace`：配置客户可用菜单上限。

## 租户在哪里登录

第一阶段：

- 公司总控：`platform-admin` 登录。
- 客户子总控：也登录 `platform-admin`，通过域名或请求头识别租户。
- 机构端：登录 `institution-admin`，通过域名或请求头识别租户。

机构端登录已经加了租户校验：如果访问的是 `tenant-b` 域名，但账号所属机构绑定在 `tenant-a`，会提示“该机构不属于当前租户”。

## 独立域名怎么配置

后端中间件支持两种方式识别租户：

### 方式一：请求头

适合本地开发：

```bash
curl -H 'X-Tenant-ID: tenant-a' http://127.0.0.1:8082/api/v1/tenant/customization-summary
```

### 方式二：域名映射

设置环境变量 `TENANT_DOMAIN_MAP`：

```bash
export TENANT_DOMAIN_MAP='tenant-a.example.com=tenant-a,a-school.example.com=tenant-a'
```

也支持 JSON：

```bash
export TENANT_DOMAIN_MAP='{"tenant-a.example.com":"tenant-a","a-school.example.com":"tenant-a"}'
```

然后 Nginx 把不同域名代理到同一套前端和同一套 API：

```nginx
server {
    listen 80;
    server_name tenant-a.example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
    }
}
```

生产建议：

- 域名配置落库到 `tenant_domain`。
- 网关或 BFF 根据 Host 查询租户并注入可信 `X-Tenant-ID`。
- 不要让公网用户随便伪造 `X-Tenant-ID`。

## 查看初始化结果

启动 `platform-service` 后，使用管理端 token 调用：

```bash
curl -H 'Authorization: Bearer <token>' \
  -H 'X-Tenant-ID: tenant-a' \
  http://127.0.0.1:8082/api/v1/platform/tenants/bootstrap-summary
```

返回内容包括：

- 当前租户。
- 已绑定机构数量。
- 已授权菜单数量。
- 租户管理员账号。
- 已配置域名。

## 下一步建议

当前分支只是租户控制面底座。下一步建议继续开发：

1. 租户列表、创建、停用、编辑。
2. 租户管理员账号开通和重置密码。
3. 租户菜单授权页面。
4. 租户版本包配置页面。
5. 客户子总控菜单裁剪。
6. 域名从数据库动态解析，不再依赖环境变量。
