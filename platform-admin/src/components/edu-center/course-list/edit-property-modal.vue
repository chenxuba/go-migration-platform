<script setup>
import { CloseOutlined, ExclamationCircleOutlined, FormOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { Modal } from 'ant-design-vue'
import { createVNode } from 'vue'
import messageService from '~@/utils/messageService'
import { addCoursePropertyOptionApi, deleteOptionApi, getCoursePropertyOptionsApi, updateCoursePropertyOptionApi } from '~@/api/edu-center/course-list'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  currentEditProperty: {
    type: Object,
    default: () => ({}),
  },
})
const emit = defineEmits(['update:open', 'refresh-list'])
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

async function getCoursePropertyOptions() {
  try {
    const res = await getCoursePropertyOptionsApi({
      propertyId: props.currentEditProperty.id,
    })
    if (res.code === 200) {
      data.value = res.result
      resetEditingState()
    }
  }
  catch (err) {
    console.log(err)
  }
}
watch(() => props.open, (newVal) => {
  if (newVal) {
    // 请求获取下拉属性选项
    getCoursePropertyOptions()
  }
}, {
  immediate: true,
})
// 手动触发验证
async function handleSubmit() {
  try {
    console.log('验证通过，提交数据:', props.currentEditProperty)
  }
  catch (error) {
    console.log('验证失败:', error)
  }
}
function closeFun() {
  resetEditingState()
  openModal.value = false
}
const columns = ref([
  {
    title: '序号',
    dataIndex: 'order',
    key: 'order',
    width: 60,
  },
  {
    title: '选项名称',
    dataIndex: 'name',
    key: 'name',
  },
  // {
  //   title: '启用状态',
  //   dataIndex: 'enable',
  //   key: 'enable',
  //   width: 100,
  // },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 150,
  },
])
const data = ref([])

// 添加当前编辑行的跟踪
const currentEditingRowSort = ref(null)

// 原始的处理函数
function handleEnableOnlineFilterChangeCore(value) {
  const e = value.target.value
  // 创建更新后的属性对象
  const updatedProperty = {
    ...props.currentEditProperty,
    enableOnlineFilter: e,
  }
  emit('refresh-list', updatedProperty, props.currentEditProperty)
}

// 使用防抖包装的函数
const handleEnableOnlineFilterChange = debounce(handleEnableOnlineFilterChangeCore, 300, {
  leading: true,
  trailing: false,
})
function handleAdd() {
  if (data.value.length >= 50) {
    messageService.error('最多添加50个选项')
    return
  }

  // 检查是否有行正在编辑
  if (currentEditingRowSort.value !== null) {
    // console.log('已有行正在编辑，阻止添加')
    messageService.warning('请先保存或取消当前编辑的行')
    return
  }

  const newSort = data.value.length + 1
  // console.log('添加新行，sort:', newSort)

  data.value.push({
    sort: newSort,
    name: '',
    readonly: true,
  })
  currentEditingRowSort.value = newSort
  // console.log('设置当前编辑行为:', currentEditingRowSort.value)
}

function handleEdit(record) {
  // 检查是否有其他行正在编辑
  if (currentEditingRowSort.value !== null && currentEditingRowSort.value !== record.sort) {
    messageService.warning('请先保存或取消当前编辑的行')
    return
  }
  // 保存原始数据
  record.originalName = record.name
  record.readonly = true
  currentEditingRowSort.value = record.sort
}

function handleCancel(record) {
  if (record.originalName !== undefined) {
    // 如果是编辑已存在的行，恢复原始数据
    record.name = record.originalName
    delete record.originalName
    record.readonly = false
  }
  else {
    // 如果是新增行，直接删除
    data.value = data.value.filter(item => item.sort !== record.sort)
  }
  currentEditingRowSort.value = null
}

async function handleDelete(record) {
  // // 删除指定sort的元素
  // data.value = data.value.filter(item => item.sort !== record.sort)
  // // 如果删除的是当前编辑行，清除编辑状态
  // if (currentEditingRowSort.value === record.sort) {
  //   currentEditingRowSort.value = null
  // }
  Modal.confirm({
    title: `确定删除${record.name}吗？`,
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    content: '删除后，该选项将从所有关联的课程中清空。此操作不可撤销，且可能影响已发布和正在售卖的课程。',
    onOk: async () => {
      try {
        const res = await deleteOptionApi({
          id: record.id,
          uuid: record.uuid,
          version: record.version,
        })
        if (res.code === 200) {
          messageService.success('删除成功')
          await getCoursePropertyOptions()
        }
      }
      catch (error) {
        console.log(error)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onCancel() { },
  })
}

async function handleSave(record) {
  // 基本验证
  if (!record.name || record.name.trim() === '') {
    messageService.error('请输入选项名称')
    return
  }

  // 选项名称 不能重复
  if (data.value.some(item => item.name === record.name.trim() && item.sort !== record.sort)) {
    messageService.error('选项名称不能重复')
    return
  }
  // 有record.id 是编辑
  if (record.id) {
    try {
      const res = await updateCoursePropertyOptionApi({
        id: record.id,
        name: record.name.trim(),
        uuid: record.uuid,
        version: record.version,
      })
      if (res.code === 200) {
        messageService.success('保存成功')
        // 保存成功后更新行状态
        record.readonly = false
        delete record.originalName
        currentEditingRowSort.value = null
        // 刷新数据
        await getCoursePropertyOptions()
      }
    }
    catch (err) {
      console.log(err)
    }
  }
  else {
    try {
      const res = await addCoursePropertyOptionApi({
        propertyId: props.currentEditProperty.id,
        name: record.name.trim(),
        sort: record.sort,
      })
      if (res.code === 200) {
        messageService.success('保存成功')
        // 保存成功后更新行状态
        record.readonly = false
        delete record.originalName
        currentEditingRowSort.value = null
        // 刷新数据
        await getCoursePropertyOptions()
      }
    }
    catch (err) {
      console.log(err)
    }
  }
}

// 使用防抖包装保存函数，防止重复点击
const debouncedHandleSave = debounce(handleSave, 300, {
  leading: true,
  trailing: false,
})

// 重置编辑状态
function resetEditingState() {
  currentEditingRowSort.value = null
  data.value.forEach((item) => {
    if (item.readonly) {
      item.readonly = false
      delete item.originalName
    }
  })
}
const editPropertyModalOpen = ref(false)
function handleEditProperty() {
  editPropertyModalOpen.value = true
}
const formState = ref({
  name: '',
})
watch(() => props.currentEditProperty, (newVal) => {
  formState.value.name = newVal.name
}, {
  immediate: true,
})
const editPropertyFormRef = ref(null)
function handleEditPropertyOkCore() {
  editPropertyFormRef.value.validate().then(async () => {
    // 创建更新后的属性对象
    const updatedProperty = {
      ...props.currentEditProperty,
      name: formState.value.name,
    }
    await emit('refresh-list', updatedProperty, props.currentEditProperty)
    editPropertyModalOpen.value = false
    messageService.success('编辑成功')
  })
}

// 使用防抖包装编辑属性确认函数，防止重复点击
const handleEditPropertyOk = debounce(handleEditPropertyOkCore, 300, {
  leading: true,
  trailing: false,
})

function handleEditPropertyCancel() {
  // 清空表单验证
  editPropertyFormRef.value.resetFields()
  editPropertyModalOpen.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="540"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>编辑{{ currentEditProperty.name }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form layout="inline" :model="currentEditProperty">
        <div class="flex justify-between w-full mb-16px">
          <!-- 属性名称 -->
          <a-form-item v-if="currentEditProperty.name != '科目'" label="属性名称" name="name" :rules="[{ required: true, message: '请输入属性名称' }]">
            <div class="flex flex-items-center">
              {{ currentEditProperty.name }}
              <FormOutlined class="text-#06f cursor-pointer text-16px ml-6px " @click="handleEditProperty" />
            </div>
          </a-form-item>
          <!-- 在线商城支持筛选 单选 -->
          <a-form-item label="在线商城支持筛选" name="enableOnlineFilter" :rules="[{ required: true, message: '请选择在线商城支持筛选' }]">
            <a-radio-group
              class="custom-radio" :value="currentEditProperty.enableOnlineFilter"
              @change="handleEnableOnlineFilterChange"
            >
              <a-radio :value="true">
                是
              </a-radio>
              <a-radio :value="false">
                否
              </a-radio>
            </a-radio-group>
          </a-form-item>
        </div>
        <!-- table 展示顺序 选项名称 操作（编辑、删除） -->
        <a-table :columns="columns" :data-source="data" class="w-full" :pagination="false">
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'order'">
              <div class="flex items-center justify-between relative top-2px">
                <span class="mt--4px">{{ index + 1 }}</span>
                <!-- <div class="flex flex-items-center cursor-pointer ">
                  <a-tooltip placement="top">
                    <template #title>
                      长按拖拽调整顺序
                    </template>
                    <SwapLeftOutlined :class="[
                      'rotate--90 text-18px',
                      index === data.length - 1 ? 'text-#ccc' : 'text-#06f'
                    ]" />
                    <SwapLeftOutlined :class="[
                      'rotate-90 text-18px ml--8px',
                      index === 0 ? 'text-#ccc' : 'text-#06f'
                    ]" />
                  </a-tooltip>
                </div> -->
              </div>
            </template>
            <template v-if="column.key === 'name'">
              <a-input v-if="record.readonly" v-model:value="record.name" :maxlength="8" placeholder="请输入（8字以内）" />
              <span v-else>{{ record.name }}</span>
            </template>
            <!-- <template v-if="column.key === 'enable'">
              <a-switch  v-model:checked="record.enable" @change="handleEnableOnlineFilterChange" />
            </template> -->
            <template v-if="column.key === 'action'">
              <a-space v-if="record.readonly" :size="14">
                <a @click="debouncedHandleSave(record)">保存</a>
                <a class="text-#ff3333" @click="handleCancel(record)">取消</a>
              </a-space>
              <a-space v-else :size="14">
                <a @click="handleEdit(record)">编辑</a>
                <a class="text-#ff3333" @click="handleDelete(record)">删除</a>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-form>
    </div>
    <template #footer>
      <a-button type="dashed" color="#06f" class="w-full" @click="handleAdd">
        添加（{{ data.length }}/50）
      </a-button>
    </template>
    <a-modal
      v-model:open="editPropertyModalOpen" centered :keyboard="false" :closable="false" :mask-closable="false"
      :title="`编辑${currentEditProperty.name}`" :width="540" @ok="handleEditPropertyOk"
      @cancel="handleEditPropertyCancel"
    >
      <a-form ref="editPropertyFormRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 14 }">
        <a-form-item label="属性名称" name="name" :rules="[{ required: true, message: '请输入属性名称' }]">
          <a-input v-model:value="formState.name" :maxlength="6" placeholder="请输入（6字以内）" />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-modal>
</template>

<style lang="less" scoped>
.contenter {
  padding: 24px;
}

/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
