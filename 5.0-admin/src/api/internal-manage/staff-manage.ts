// 新增部门 /sysDepart/saveDepart
export function saveDepartApi(data) {
  return usePost('/sso/sysDepart/saveDepart', data)
}

// 修改部门 /sysDepart/update
export function updateDepart(data) {
  return usePost('/sso/sysDepart/update', data)
}

// 删除部门 /sysDepart/delete
export function deleteDepart(data) {
  return usePost('/sso/sysDepart/delete', data)
}

// 树形部门列表  /sysDepart/listTree
export function getListTreeDepartApi() {
  return useGet('/sso/sysDepart/listTree')
}

// 获取机构用户列表 /instUser/getUserList
export function getUserListApi(data) {
  return usePost('/api/v1/inst-users/page', data)
}

// 新增机构用户 /instUser/saveInstUser
export function saveInstUser(data) {
  return usePost('/api/v1/inst-users/create', data)
}

// 批量离职/复职 /instUser/batchDisabledApi
export function batchDisabledApi(data) {
  return usePost('/api/v1/inst-users/batch-disabled', data)
}

// 修改机构用户 /instUser/updateInstUser
export function updateInstUser(data) {
  return usePost('/api/v1/inst-users/update', data)
}

// 获取机构用户详情 /instUser/getInstUserDetail?id=1
export function getInstUserDetail(data) {
  return useGet('/api/v1/inst-users/detail', data)
}

// 修改机构用户  /instUser/updateInstUser
export function updateInstUserDetail(data) {
  return usePost('/api/v1/inst-users/update', data)
}

// 批量修改部门 /instUser/batchModifyDept
export function batchModifyDept(data) {
  return usePost('/api/v1/inst-users/batch-dept', data)
}

// 批量修改角色 /instUser/batchModifyRole
export function batchModifyRole(data) {
  return usePost('/api/v1/inst-users/batch-role', data)
}
// 检验手机号是否已使用 /instUser/checkPhoneUsed
export function checkPhoneUsedApi(data) {
  return usePost('/api/v1/inst-users/check-phone', data)
}
// 更换手机号 /instUser/changePhoneWithOther
export function changePhoneWithOtherApi(data) {
  return usePost('/api/v1/inst-users/change-phone', data)
}
