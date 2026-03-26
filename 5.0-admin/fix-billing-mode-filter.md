# 修复收费方式筛选器参数传递问题

## 问题描述
选择收费方式筛选条件后，没有传递`lessonChargingList`参数给后端接口。

后端期望的参数：
```java
@ApiModelProperty("收费方式列表（1=按课时，2=按时段，3=按金额）")
private List<Integer> lessonChargingList;
```

## 问题原因

### 1. all-filter组件emit的事件名
在 `all-filter.vue` 中：
```javascript
function handleBillingModeChange(e) {
  nextTick(() => {
    console.log('收费方式:', selectBillingModeVals.value)
    debouncedEmit('chargingMethodFilter', selectBillingModeVals.value)  // ← emit的事件名
  })
}
```

emit的事件名是：`chargingMethodFilter`

### 2. 报读列表页面监听的事件名（修复前）
在 `register-read-list.vue` 中：
```javascript
const filterFieldMapping = {
  stuPhoneSearchFilter: 'studentId',
  currentStatusFilter: 'statusList',
  orNotFenClassFilter: 'assignedClass',
  billingModeFilter: 'lessonChargingList',  // ← 监听的事件名
  // ...
}
```

监听的事件名是：`billingModeFilter`

### 3. 问题分析
- all-filter组件emit的事件是：`update:chargingMethodFilter`
- 报读列表页面监听的事件是：`update:billingModeFilter`
- 两个事件名不匹配，导致事件无法触发
- 所以选择收费方式后，`lessonChargingList`参数没有更新

## 解决方案

在 `register-read-list.vue` 的 `filterFieldMapping` 中添加对 `chargingMethodFilter` 的映射：

```javascript
const filterFieldMapping = {
  stuPhoneSearchFilter: 'studentId',
  currentStatusFilter: 'statusList',
  orNotFenClassFilter: 'assignedClass',
  billingModeFilter: 'lessonChargingList',
  chargingMethodFilter: 'lessonChargingList', // ← 添加这一行
  isSetExpirationDateFilter: 'isSetExpireTime',
  // ...
}
```

这样，无论all-filter组件emit的是`billingModeFilter`还是`chargingMethodFilter`，都能正确处理。

## 为什么有两个不同的事件名？

这是历史遗留问题：
- `billingMode`（计费模式）是早期的命名
- `chargingMethod`（收费方式）是后来统一的命名
- all-filter组件使用了新的命名，但某些页面还在使用旧的命名

## 修改的文件
- `5.0-admin/src/pages/edu-center/register-read-list.vue`
  - 在`filterFieldMapping`中添加了`chargingMethodFilter`的映射

## 验证方法

1. 重启前端服务
2. 在报读列表页面选择收费方式（如"按课时"）
   - 打开浏览器开发者工具 → Network
   - 查看请求参数，应该包含 `"lessonChargingList": [1]`
3. 选择多个收费方式（如"按课时"和"按时段"）
   - 查看请求参数，应该包含 `"lessonChargingList": [1, 2]`
4. 清空收费方式筛选
   - 查看请求参数，不应该包含`lessonChargingList`字段

## 收费方式的值映射

| 选项 | 值 | 说明 |
|------|---|------|
| 按课时 | 1 | 按上课课时数计费 |
| 按时段 | 2 | 按时间段计费 |
| 按金额 | 3 | 按固定金额计费 |

## 相关筛选器的事件名对照

| 筛选器 | all-filter emit的事件名 | 报读列表监听的事件名 | 是否匹配 |
|--------|------------------------|-------------------|---------|
| 收费方式 | chargingMethodFilter | billingModeFilter | ❌ 不匹配（已修复） |
| 授课方式 | teachingMethodFilter | teachingMethodFilter | ✅ 匹配 |
| 当前状态 | currentStatusFilter | currentStatusFilter | ✅ 匹配 |
| 分班状态 | orNotFenClassFilter | orNotFenClassFilter | ✅ 匹配 |

## 建议

为了避免类似问题，建议：
1. 统一事件命名规范
2. 在all-filter组件中添加事件名映射表
3. 或者让all-filter组件同时emit两个事件名（兼容旧代码）
