# StaffSelect 员工选择组件

一个带有搜索、分页加载和接口调用功能的员工选择下拉框组件，支持单选和多选模式。

## 基本使用

```vue
<template>
  <div>
    <!-- 基本使用 -->
    <StaffSelect v-model="selectedStaffId" />
    
    <!-- 多选模式 -->
    <StaffSelect 
      v-model="selectedStaffIds" 
      :multiple="true"
      placeholder="请选择多个员工"
    />
    
    <!-- 自定义配置 -->
    <StaffSelect
      v-model="selectedStaffId"
      placeholder="请选择销售员"
      width="300px"
      :status="0"
      @change="handleStaffChange"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import StaffSelect from './staff-select.vue'

const selectedStaffId = ref(undefined)
const selectedStaffIds = ref([])

function handleStaffChange(value, staffInfo) {
  console.log('选择的员工ID:', value)
  console.log('员工信息:', staffInfo)
}
</script>
```

## Props

| 参数 | 说明 | 类型 | 默认值 |
|------|------|------|--------|
| modelValue | 绑定值(员工ID)，多选时为数组 | string \| number \| Array | undefined |
| placeholder | 占位符文本 | string | '请选择员工' |
| width | 组件宽度 | string | '240px' |
| allowClear | 是否允许清空 | boolean | true |
| status | 员工状态筛选，0-在职，1-离职，undefined-全部 | number | 0 |
| pageSize | 每页加载数量 | number | 10 |
| multiple | 是否多选模式 | boolean | false |

## Events

| 事件名 | 说明 | 回调参数 |
|--------|------|----------|
| update:modelValue | 绑定值变化时触发 | (value: string \| number \| Array) |
| change | 选择变化时触发 | (value: string \| number \| Array, staffInfo: object \| Array) |

## Methods

通过 ref 可以调用以下方法：

| 方法名 | 说明 | 参数 | 返回值 |
|--------|------|------|-------|
| refresh | 刷新员工列表 | - | - |
| getSelectedStaff | 获取当前选中的员工信息 | - | object \| Array |

### 使用示例

```vue
<template>
  <div>
    <!-- 单选模式 -->
    <StaffSelect ref="staffSelectRef" v-model="selectedStaffId" />
    
    <!-- 多选模式 -->
    <StaffSelect 
      ref="multiStaffSelectRef" 
      v-model="selectedStaffIds" 
      :multiple="true"
      placeholder="请选择多个员工"
    />
    
    <a-button @click="refreshList">刷新列表</a-button>
    <a-button @click="getSelected">获取选中员工</a-button>
    <a-button @click="getMultiSelected">获取多选员工</a-button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import StaffSelect from './staff-select.vue'

const selectedStaffId = ref(undefined)
const selectedStaffIds = ref([])
const staffSelectRef = ref()
const multiStaffSelectRef = ref()

function refreshList() {
  staffSelectRef.value.refresh()
}

function getSelected() {
  const selectedStaff = staffSelectRef.value.getSelectedStaff()
  console.log('当前选中的员工:', selectedStaff)
}

function getMultiSelected() {
  const selectedStaffs = multiStaffSelectRef.value.getSelectedStaff()
  console.log('当前选中的多个员工:', selectedStaffs)
}
</script>
```

## 特性

1. **单选/多选支持**: 通过 `multiple` 属性控制单选或多选模式
2. **搜索功能**: 支持按员工姓名和手机号搜索，带有防抖处理
3. **分页加载**: 支持滚动加载更多数据，避免一次性加载大量数据
4. **状态筛选**: 可根据员工状态（在职/离职）进行筛选
5. **智能显示**: 自动获取并显示当前选中员工的信息，即使未在当前列表中
6. **Loading效果**: 完善的加载状态提示
   - 初始加载loading效果
   - 滚动加载更多的loading效果
   - 搜索时的loading效果
   - 下拉框loading状态
7. **自动管理**: 自动管理加载状态、分页状态等
8. **错误处理**: 包含完整的错误处理机制
9. **接口封装**: 内置员工列表接口调用逻辑
10. **状态重置**: 关闭下拉框时自动重置状态，重新打开时重新加载

## 显示逻辑说明

### 单选模式
- 当绑定值存在但在当前列表中找不到时，组件会自动调用接口获取员工信息
- 如果接口获取失败，会使用当前登录用户的信息作为默认显示
- 选中的员工会优先显示在下拉列表顶部

### 多选模式
- 支持选择多个员工，返回员工ID数组
- `getSelectedStaff` 方法在多选模式下返回员工对象数组
- `change` 事件的第二个参数在多选模式下为员工对象数组

## Loading状态说明

组件包含以下几种loading状态：

1. **初始加载**: 首次打开下拉框或搜索时显示
2. **滚动加载**: 滚动到底部加载更多数据时显示
3. **下拉框Loading**: 在Select组件上显示的loading图标
4. **暂无数据**: 当没有匹配的员工时显示
5. **没有更多**: 当所有数据加载完成时显示

各种状态会根据实际情况自动切换，无需手动管理。

## 注意事项

1. 组件依赖 `getUserListApi` 接口，请确保接口路径正确
2. 组件使用了 `lodash-es` 的 `debounce` 函数，请确保已安装依赖
3. 搜索时会自动清空当前列表并重新加载
4. 关闭下拉框时会重置所有状态，重新打开时会重新加载数据
5. Loading状态会自动管理，无需手动干预
6. 多选模式下，modelValue 应为数组类型
7. 组件会自动获取当前登录用户信息用于回退显示
8. 选中的员工信息会被缓存，避免重复加载 