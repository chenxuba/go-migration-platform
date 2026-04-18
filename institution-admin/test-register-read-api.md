# 报读列表接口优化说明

## 已支持的筛选参数

根据你提供的参数示例，报读列表接口现已支持以下所有筛选条件：

### 1. 学员信息
- `studentId`: 学员ID（通过学员/手机号搜索筛选器）

### 2. 时间范围筛选
- `fromExpireTime` / `toExpireTime`: 有效期至范围
- `fromSuspendedTime` / `toSuspendedTime`: 停课时间范围
- `fromClosedTime` / `toClosedTime`: 结课时间范围
- `lastestTeachingRecordStartTime` / `lastestTeachingRecordEndTime`: 最近上课时间范围

### 3. 状态筛选
- `isSetExpireTime`: 是否设置有效期（true/false）
- `lessonType`: 授课方式（1-班级授课，2-1v1授课）
- `lessonChargingList`: 收费方式数组（1-按课时，2-按时段，3-按金额）
- `statusList`: 当前状态数组（1-正常，2-停课，3-结课）
- `assignedClass`: 分班状态（true-已分班，false-未分班）
- `isArrears`: 是否欠费（true/false）

### 4. 剩余数量筛选
- `fromRemainQuantity` / `toRemainQuantity`: 剩余数量范围
- `remainLessonChargingMode`: 剩余数量计费模式（1-按课时，2-按时段，3-按金额）

### 5. 人员筛选
- `classTeacherId`: 班主任ID
- `salespersonId`: 销售员ID

### 6. 课程/班级筛选
- `classIds`: 班级ID数组
- `productIds`: 报读课程ID数组

## 参数示例

```json
{
  "studentId": "696958030125764608",
  "fromExpireTime": "2026-02-09",
  "toExpireTime": "2026-02-14",
  "fromSuspendedTime": "2026-02-11",
  "toSuspendedTime": "2026-02-14",
  "fromClosedTime": "2026-02-10",
  "toClosedTime": "2026-02-12",
  "isSetExpireTime": true,
  "lessonType": 2,
  "lessonChargingList": [4, 1, 2],
  "statusList": [2, 1, 3],
  "fromRemainQuantity": 2,
  "toRemainQuantity": 23,
  "remainLessonChargingMode": 1,
  "classTeacherId": "677662659453186048",
  "salespersonId": "681664834470734848",
  "classIds": ["656895677254691840", "616735559406145536"],
  "productIds": ["617134292790761472", "617134640326594560"],
  "isArrears": true,
  "lastestTeachingRecordStartTime": "2026-02-10",
  "lastestTeachingRecordEndTime": "2026-02-20"
}
```

## 前端筛选器映射

| 筛选器名称 | 对应API参数 | 说明 |
|-----------|------------|------|
| stuPhoneSearchFilter | studentId | 学员/手机号搜索 |
| currentStatusFilter | statusList | 当前状态（多选） |
| orNotFenClassFilter | assignedClass | 分班状态 |
| billingModeFilter | lessonChargingList | 收费方式（多选） |
| isSetExpirationDateFilter | isSetExpireTime | 是否设置有效期 |
| classEndingTimeFilter | fromClosedTime, toClosedTime | 结课时间范围 |
| classStopTimeFilter | fromSuspendedTime, toSuspendedTime | 停课时间范围 |
| expiryDateFilter | fromExpireTime, toExpireTime | 有效期至范围 |
| salesPersonFilter | salespersonId | 销售员 |
| intentionCourseFilter | productIds | 报读课程 |
| teachingMethodFilter | lessonType | 授课方式 |
| classTeacherFilter | classTeacherId | 班主任 |
| classNameFilter | classIds | 班级名称 |
| enrolledCourseFilter | productIds | 报读课程 |
| isArrearsFilter | isArrears | 是否欠费 |
| lastClassTimeFilter | lastestTeachingRecordStartTime, lastestTeachingRecordEndTime | 最近上课时间 |
| remainingFilter | fromRemainQuantity, toRemainQuantity, remainLessonChargingMode | 剩余数量 |

## 修改的文件

1. **5.0-admin/src/pages/edu-center/register-read-list.vue**
   - 更新了 `queryState` 添加所有筛选参数
   - 更新了 `filterFieldMapping` 添加新的筛选器映射
   - 更新了 `filterUpdateHandlers` 处理所有筛选器的值转换
   - 添加了剩余数量、是否欠费、最近上课时间等筛选器的处理逻辑

2. **5.0-admin/src/components/common/all-filter.vue**
   - 添加了 `update:remainingFilter` 和 `update:billingModeFilter` emit事件
   - 更新了 `handleRemainingConfirm` 函数，添加emit逻辑
   - 将剩余数量的mode转换为对应的数字值（lesson=1, period=2, amount=3）

3. **5.0-admin/src/api/edu-center/register-read-list.ts**
   - API接口类型定义已包含所有参数，无需修改

## 使用说明

前端页面的筛选器会自动将用户选择的条件转换为API所需的参数格式：

1. **时间范围**：选择日期范围后自动拆分为 `from*` 和 `to*` 参数
2. **布尔值**：单选筛选器（如是否设置有效期）转换为 true/false
3. **数组值**：多选筛选器（如当前状态、收费方式）保持数组格式
4. **剩余数量**：包含模式选择（按课时/时段/金额）和数量范围

所有筛选条件都支持清空和重置功能。
