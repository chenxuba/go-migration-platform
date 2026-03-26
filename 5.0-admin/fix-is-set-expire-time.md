# 修复"是否设置有效期"参数传递问题

## 问题描述
当选择"未设置有效期"时，后端没有收到`isSetExpireTime`参数。

实际情况：
- 选择"已设置有效期"：传递 `{"isSetExpireTime": true, "statusList": [1,2]}`  ✅ 正确
- 选择"未设置有效期"：传递 `{"statusList": [1,2]}`  ❌ 缺少`isSetExpireTime`参数

期望情况：
- 选择"已设置有效期"：传递 `{"isSetExpireTime": true, "statusList": [1,2]}`
- 选择"未设置有效期"：传递 `{"isSetExpireTime": false, "statusList": [1,2]}`

## 问题原因

### 1. all-filter组件的选项定义
在 `all-filter.vue` 中：
```javascript
const isSetExpirationDateOptions = ref([
  { id: 1, value: '已设置有效期' },
  { id: 2, value: '未设置有效期' },
])
```

### 2. 报读列表页面的处理逻辑（修复前）
在 `register-read-list.vue` 中：
```javascript
// 是否设置有效期：0-未设置，1-已设置
if (fieldName === 'isSetExpireTime') {
  if (val === 0) {  // ❌ 错误：应该是 2
    handleFilterUpdate({ isSetExpireTime: false }, isClearAll, id, type)
  } else if (val === 1) {  // ✅ 正确
    handleFilterUpdate({ isSetExpireTime: true }, isClearAll, id, type)
  } else {
    handleFilterUpdate({ isSetExpireTime: undefined }, isClearAll, id, type)
  }
  return
}
```

### 3. 问题分析
- all-filter组件传递的值是 `1` 或 `2`
- 但报读列表页面判断的是 `0` 或 `1`
- 导致选择"未设置有效期"（值为2）时，走到了`else`分支，设置为`undefined`
- `undefined`的值在API请求时会被过滤掉，所以后端收不到参数

## 解决方案

修改 `register-read-list.vue` 中的判断逻辑：

```javascript
// 是否设置有效期：1-已设置，2-未设置
if (fieldName === 'isSetExpireTime') {
  if (val === 2) {  // ✅ 修复：判断值为2
    handleFilterUpdate({ isSetExpireTime: false }, isClearAll, id, type)
  } else if (val === 1) {
    handleFilterUpdate({ isSetExpireTime: true }, isClearAll, id, type)
  } else {
    handleFilterUpdate({ isSetExpireTime: undefined }, isClearAll, id, type)
  }
  return
}
```

## 修改的文件
- `5.0-admin/src/pages/edu-center/register-read-list.vue`
  - 修改了`filterUpdateHandlers`中对`isSetExpireTime`的判断逻辑
  - 将判断条件从`val === 0`改为`val === 2`

## 验证方法

1. 重启前端服务
2. 在报读列表页面选择"已设置有效期"
   - 打开浏览器开发者工具 → Network
   - 查看请求参数，应该包含 `"isSetExpireTime": true`
3. 选择"未设置有效期"
   - 查看请求参数，应该包含 `"isSetExpireTime": false`
4. 不选择（清空筛选）
   - 查看请求参数，不应该包含`isSetExpireTime`字段

## 相关说明

### 为什么all-filter使用1和2而不是0和1？
这是为了避免与falsy值混淆。在JavaScript中：
- `0` 是falsy值，在某些判断中可能被当作false
- `1` 和 `2` 都是truthy值，更容易区分

### 其他类似的筛选器
- **是否欠费**：使用 `0` 和 `1`（因为使用了`IsCommonYesNo`枚举）
  - `0` → "否"（不欠费）
  - `1` → "是"（欠费）
- **分班状态**：使用 `0` 和 `1`
  - `0` → "未分班"
  - `1` → "已分班"

这些筛选器的值映射需要与all-filter组件的定义保持一致。
