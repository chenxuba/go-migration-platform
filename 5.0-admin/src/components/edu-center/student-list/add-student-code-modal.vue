<script setup>
import { CloseOutlined, MenuOutlined, MinusCircleFilled, PlusCircleFilled, QuestionCircleOutlined } from '@ant-design/icons-vue'
import Sortable from 'sortablejs'
import { nextTick, onMounted, ref, watch } from 'vue'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  editMode: {
    type: Boolean,
    default: false,
  },
  editData: {
    type: Object,
    default: null,
  },
  isSystemDefault: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open', 'handleSubmitStuCode'])
const loading = ref(false)
const formRef = ref()
const optionsListRef = ref(null)

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 编辑模式标识
const isEditMode = computed(() => props.editMode)
// 是否为系统默认字段
const isSystemDefault = computed(() => props.isSystemDefault)

// 生成唯一ID的函数
function generateId() {
  return `_${Math.random().toString(36).substr(2, 9)}`
}

// 获取初始状态函数
function getInitialState() {
  return {
    fieldKey: '',
    fieldType: undefined,
    required: false,
    searched: false,
    optionsList: Array(3).fill().map(() => ({ id: generateId(), name: '' })),
    optionsJson: [],
  }
}

const formState = reactive(getInitialState())

let sortableInstance = null

// 初始化拖拽排序
function initSortable() {
  nextTick(() => {
    if (!optionsListRef.value)
      return

    // 如果已经存在实例，先销毁
    if (sortableInstance) {
      sortableInstance.destroy()
    }

    sortableInstance = Sortable.create(optionsListRef.value, {
      handle: '.drag-handle',
      animation: 150,
      ghostClass: 'sortable-ghost',
      onEnd: (evt) => {
        const { newIndex, oldIndex } = evt
        if (newIndex === undefined || oldIndex === undefined)
          return

        // 创建一个全新的数组来更新选项
        const newOptions = [...formState.optionsList]
        const movedItem = newOptions[oldIndex]
        newOptions.splice(oldIndex, 1)
        newOptions.splice(newIndex, 0, movedItem)
        formState.optionsList = newOptions
      },
    })
  })
}

// 处理选项数据
function parseOptionsJson(optionsJson) {
  if (!optionsJson)
    return []

  // 处理可能的格式: 字符串、已解析的数组或JSON字符串
  let options = []

  if (typeof optionsJson === 'string') {
    // 处理字符串
    options = optionsJson.split(',').filter(item => item)
  }
  else if (Array.isArray(optionsJson)) {
    // 如果已经是数组，直接使用
    options = optionsJson
  }

  // 将选项转换为所需格式
  return options.map(name => ({
    id: generateId(),
    name: name.trim(),
  }))
}

// 监听 type 变化
watch(() => formState.fieldType, (newVal) => {
  if (newVal === '4') {
    initSortable()
  }
})

// 监听openModal和editData
watch([() => openModal.value, () => props.editData], ([newOpenModal, newEditData]) => {
  if (newOpenModal) {
    if (isEditMode.value && newEditData) {
      // 编辑模式下填充数据
      const editData = { ...newEditData }

      // 转换字段类型为字符串
      if (typeof editData.fieldType === 'number') {
        editData.fieldType = String(editData.fieldType)
      }

      // 处理optionsList
      if ((editData.fieldType === '4' || editData.fieldType === 4) && editData.optionsJson) {
        editData.optionsList = parseOptionsJson(editData.optionsJson)
      }
      else {
        editData.optionsList = Array(3).fill().map(() => ({ id: generateId(), name: '' }))
      }

      console.log('Editing data:', editData)
      console.log('Options parsed:', editData.optionsList)

      // 将数据复制到formState
      Object.keys(formState).forEach((key) => {
        if (key in editData) {
          formState[key] = editData[key]
        }
      })

      // 确保在下一个tick初始化拖拽，以便DOM更新
      if (editData.fieldType === '4' || editData.fieldType === 4) {
        nextTick(() => initSortable())
      }
    }
    else {
      // 新增模式重置状态
      Object.assign(formState, getInitialState())
    }
    loading.value = false
  }
}, { immediate: true })

onMounted(() => {
  if (formState.fieldType === '4') {
    initSortable()
  }
})

function addOption() {
  if (formState.optionsList.length >= 20) {
    // 使用自定义的message组件
    messageService.warning('最多只能添加20个选项')
    return
  }
  // 创建新选项
  const newOption = { id: generateId(), name: '' }
  // 确保创建新数组并添加到末尾
  formState.optionsList = Array.from(formState.optionsList).concat(newOption)
}

function deleteOption(index) {
  formState.optionsList.splice(index, 1)
}

// 手动触发验证
async function handleSubmitStuCode() {
  loading.value = true
  try {
    await formRef.value.validate()
    const submitData = { ...formState }

    // Backend expects numeric fieldType; convert from select string value before submit.
    if (submitData.fieldType !== undefined && submitData.fieldType !== null) {
      submitData.fieldType = Number(submitData.fieldType)
    }

    // 处理选项类型的数据
    if (submitData.fieldType === 4) {
      const validOptions = submitData.optionsList.map(item => item.name.trim()).filter(Boolean)
      if (validOptions.length === 0) {
        messageService.warning('请至少填写一个选项')
        loading.value = false
        return
      }
      submitData.optionsJson = validOptions.join(',')
    }
    else {
      delete submitData.optionsList
      submitData.optionsJson = ''
    }

    // 保留编辑模式下的原始ID、uuid和version
    if (isEditMode.value && props.editData) {
      submitData.id = props.editData.id
      submitData.uuid = props.editData.uuid
      submitData.version = props.editData.version
    }

    console.log('验证通过，提交数据:', submitData)
    emit('handleSubmitStuCode', submitData)
  }
  catch (error) {
    console.log('验证失败:', error)
    loading.value = false
  }
}

function closeFun() {
  formRef.value.resetFields()
  Object.assign(formState, getInitialState())
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="580"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ isEditMode ? '编辑学员属性' : '新增学员属性' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState">
        <!-- 属性名称 必填 请填写属性名称 -->
        <a-form-item label="属性名称" name="fieldKey" class="w-60%" :rules="[{ required: true, message: '请输入属性名称' }]">
          <template v-if="isEditMode && isSystemDefault">
            <div>{{ formState.fieldKey }}<span class="text-#888 ml-4px">(属性名称不可编辑)</span></div>
            <a-input v-model:value="formState.fieldKey" class="hidden" :disabled="true" />
          </template>
          <a-input v-else v-model:value="formState.fieldKey" :maxlength="20" placeholder="请输入属性名称" />
        </a-form-item>
        <!-- 格式类型  文本 数字  日期 选项 -->
        <a-form-item label="格式类型" name="fieldType" class="w-60% flex" :rules="[{ required: true, message: '请选择格式类型' }]">
          <template v-if="isEditMode">
            <div :class="formState.fieldType == 4 ? 'mt-5px' : ''">
              {{ formState.fieldType === '1' ? '文本'
                : formState.fieldType === '2' ? '数字'
                  : formState.fieldType === '3' ? '日期'
                    : formState.fieldType === '4' ? '选项' : '' }}
              <span class="text-#888">(格式类型不可编辑)</span>
            </div>
            <a-select v-model:value="formState.fieldType" class="hidden" :disabled="true">
              <a-select-option value="1">
                文本
              </a-select-option>
              <a-select-option value="2">
                数字
              </a-select-option>
              <a-select-option value="3">
                日期
              </a-select-option>
              <a-select-option value="4">
                选项
              </a-select-option>
            </a-select>
          </template>
          <div v-else class="flex relative flex-center">
            <a-select v-model:value="formState.fieldType" placeholder="请选择格式类型" style="width: 300px;">
              <a-select-option value="1">
                文本
              </a-select-option>
              <a-select-option value="2">
                数字
              </a-select-option>
              <a-select-option value="3">
                日期
              </a-select-option>
              <a-select-option value="4">
                选项
              </a-select-option>
            </a-select>
            <span v-if="formState.fieldType === '1'" class="absolute right--31 text-#888 text-14px">文本内容限 100 字</span>
            <span v-if="formState.fieldType === '2'" class="absolute right--23 text-#888 text-14px">仅限输入数字</span>
            <span v-if="formState.fieldType === '3'" class="absolute right--34 text-#888 text-14px">选择格式：年-月-日</span>
          </div>
          <a-form-item-rest v-if="formState.fieldType === '4'">
            <div class="w-185% bg-#fafafa rounded-6px p-10px mt-10px">
              <div ref="optionsListRef" class="options-list">
                <div
                  v-for="(item, index) in formState.optionsList" :key="item.id"
                  class="flex flex-items-center mb-12px"
                >
                  <span class="cursor-pointer drag-handle">
                    <MenuOutlined class="text-#ccc" />
                  </span>
                  <a-input v-model:value="item.name" class="mx-10px w-60%" placeholder="请输入选项内容（20字以内）" />
                  <span @click="deleteOption(index)">
                    <MinusCircleFilled class="text-3.5 cursor-pointer text-#f33" />
                  </span>
                </div>
              </div>
              <div class="add text-14px text-#06f cursor-pointer ">
                <span>
                  <PlusCircleFilled class="text-#06f" />
                </span>
                <span class="ml-4px" @click="addOption">添加（{{ formState.optionsList.length }}/20）</span>
              </div>
            </div>
          </a-form-item-rest>
        </a-form-item>
        <!-- 是否必填 单选框 -->
        <a-form-item name="required" class="w-60%" :rules="[{ required: true, message: '请选择是否必填' }]">
          <template #label>
            <span>是否必填
              <a-popover title="字段说明">
                <template #content>
                  <div class="w-300px">开启必填后，报名创建/编辑学员、填写招生表单等相关学员属性填写业务时必填，否则无法保存</div>
                </template>
                <QuestionCircleOutlined />
              </a-popover>
            </span>
          </template>
          <a-radio-group v-model:value="formState.required" class="custom-radio">
            <a-radio :value="true">
              是
            </a-radio>
            <a-radio :value="false">
              否
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <!-- 支持搜索 -->
        <a-form-item name="searched" class="w-60%" :rules="[{ required: true, message: '请选择是否支持搜索' }]">
          <template #label>
            <span>支持搜索
              <a-popover title="字段说明">
                <template #content>
                  <div class="w-300px">开启支持筛选/搜索后，学员管理的"在读学员""意向学员"等相关页面将支持此属性的筛选/搜索功能</div>
                </template>
                <QuestionCircleOutlined />
              </a-popover>
            </span>
          </template>
          <a-radio-group v-model:value="formState.searched" class="custom-radio">
            <a-radio :value="true">
              是
            </a-radio>
            <a-radio :value="false">
              否
            </a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="loading" @click="handleSubmitStuCode">
        {{ isEditMode ? '保存' : '保存并选择'
        }}
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
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

.contenter {
  padding: 24px;
  max-height: calc(100vh - 300px);
  overflow-y: auto;
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

.drag-handle {
  cursor: move;

  &:hover {
    color: var(--pro-ant-color-primary);
  }
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
