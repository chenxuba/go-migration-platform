// 新建角色 /role/saveRole
export function saveRoleApi(data) {
  return usePost('/sso/roles/create', data)
}

// 更新角色 /role/updateRoleApi
export function updateRoleApi(data) {
  return usePost('/sso/roles/update', data)
}

// 获取角色列表 /role/getInstRolePageApi
export function getInstRolePageApi(data) {
  return usePost('/sso/roles/page', data)
}

// 根据机构ID获取权限ID集合 /role/instMenuList
export function instMenuList(data) {
  return useGet('/sso/roles/inst-menu', data)
}

// 根据角色ID获取权限ID集合 /role/menuList
export function roleList(data) {
  return useGet('/sso/roles/menu-ids', data)
}

// 获取权限列表 /menu/instList
export function getMenuListApi(data) {
  return useGet('/sso/menus/inst-tree', data)
}

// 获取全量机构权限树 /menu/list
export function getFullMenuListApi(data) {
  return useGet('/sso/menus/tree', data)
}

// 机构后台获取默认角色 /role/getDefaultRole
export function getDefaultRole(data) {
  return useGet('/sso/roles/templates', data)
}

// 角色权限对比 /role/roleMenuCompare
export function roleMenuCompare(data) {
  return usePost('/sso/roles/compare', data)
}

// 获取默认角色详情 /role/getDefaultRoleDetail?roleId=2
export function getDefaultRoleDetailApi(data) {
  return useGet('/sso/roles/default-detail', data)
}

// 获取角色下的员工列表 /role/getStaffListByRoleId?roleId=26
export function getStaffListByRoleId(data) {
  return useGet('/sso/roles/staff', data)
}
