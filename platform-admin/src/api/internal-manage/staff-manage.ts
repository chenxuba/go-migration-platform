// 新增部门 /sysDepart/saveDepart
export function saveDepartApi(data) {
  return usePost('/sso/departs/create', data)
}

// 修改部门 /sysDepart/update
export function updateDepart(data) {
  return usePost('/sso/departs/update', data)
}

// 删除部门 /sysDepart/delete
export function deleteDepart(data) {
  return usePost('/sso/departs/delete', data)
}

// 树形部门列表  /sysDepart/listTree
export function getListTreeDepartApi() {
  return useGet('/sso/departs/tree')
}

// 获取机构用户列表 /instUser/getUserList
export function getUserListApi(data) {
  return usePost('/sso/manage-users/page', data)
}

// 新增机构用户 /instUser/saveInstUser
export function saveInstUser(data) {
  return usePost('/sso/manage-users/create', data, { silentError: true })
}

// 批量离职/复职 /instUser/batchDisabledApi
export function batchDisabledApi(data) {
  return usePost('/sso/manage-users/batch-disabled', data)
}

// 修改机构用户 /instUser/updateInstUser
export function updateInstUser(data) {
  return usePost('/sso/manage-users/update', data)
}

// 获取机构用户详情 /instUser/getInstUserDetail?id=1
export function getInstUserDetail(data) {
  return useGet('/sso/manage-users/detail', data)
}

// 修改机构用户  /instUser/updateInstUser
export function updateInstUserDetail(data) {
  return usePost('/sso/manage-users/update', data)
}

// 批量修改部门 /instUser/batchModifyDept
export function batchModifyDept(data) {
  return usePost('/sso/manage-users/batch-dept', data)
}

// 批量修改角色 /instUser/batchModifyRole
export function batchModifyRole(data) {
  return usePost('/sso/manage-users/batch-role', data)
}
// 检验手机号是否已使用 /instUser/checkPhoneUsed
export function checkPhoneUsedApi(data) {
  return useGet('/sso/manage-users/login-account-available', { username: data?.mobile || data?.username }, { silentError: true })
}
// 更换手机号 /instUser/changePhoneWithOther
export function changePhoneWithOtherApi(data) {
  return usePost('/sso/manage-users/change-phone', data)
}

// 校验登录账号是否全库唯一（新增员工默认使用手机号作为登录账号）
export function checkInstUserLoginAccountAvailableApi(data) {
  return useGet('/sso/manage-users/login-account-available', data, { silentError: true })
}
