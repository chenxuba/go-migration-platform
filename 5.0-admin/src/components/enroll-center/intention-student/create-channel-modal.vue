<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { debounce } from 'lodash-es'

const props = defineProps({
  open: { type: Boolean, default: false },
  type: { type: String, default: 'create' },
  data: { type: Object, default: () => ({}) },
  categoryList: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:open', 'onCreateChannel'])
const formRef = ref()
const loading = ref(false)
const showCategorySelect = ref(false)

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 表单数据
const formState = reactive({
  channelName: '',
  isClassification: '1',
  categoryId: undefined,
  remark: '',
})

// 初始化表单数据
function initFormData() {
  // 重置表单验证状态
  formRef.value?.resetFields()
  // 先隐藏下拉框，防止闪烁
  showCategorySelect.value = false

  if (props.type === 'create') {
    // 创建模式：重置为默认值
    resetFormToDefaults()
  }
  else {
    // 编辑模式：加载已有数据
    loadEditData()
  }
}

// 重置表单为默认值
function resetFormToDefaults() {
  Object.assign(formState, {
    channelName: '',
    isClassification: '1',
    categoryId: undefined,
    remark: '',
  })
}

// 加载编辑数据
function loadEditData() {
  const editData = { ...props.data }

  // 处理分类ID为0的情况，暂时不设置分类ID
  if (editData.categoryId === 0 || editData.categoryId === null) {
    const { categoryId, ...restData } = editData
    Object.assign(formState, restData)
  }
  else {
    Object.assign(formState, editData)
  }

  // 在下一个tick中显示下拉框，避免显示默认值时的闪烁
  nextTick(() => {
    showCategorySelect.value = true
  })
}

// 监听弹窗打开，初始化数据
watch(() => openModal.value, (newVal) => {
  if (!newVal)
    return
  initFormData()
})

// 表单提交核心逻辑
async function submitHandler() {
  if (loading.value)
    return

  try {
    loading.value = true
    await formRef.value.validate()

    // 根据不同模式处理提交数据
    const submitData = prepareSubmitData()

    emit('onCreateChannel', submitData)
  }
  catch (error) {
    console.log('验证失败:', error)
  }
  finally {
    loading.value = false
  }
}

// 使用 lodash-es 防抖处理的表单提交函数
const handleSubmit = debounce(submitHandler, 300, {
  leading: true,
  trailing: false,
})

// 根据模式准备提交数据
function prepareSubmitData() {
  const baseData = { ...formState }

  if (props.type === 'create') {
    // 创建模式：只保留必要字段，过滤掉编辑模式特有字段
    const createFields = ['channelName', 'isClassification', 'categoryId', 'remark']
    const submitData = {}

    createFields.forEach((field) => {
      if (baseData[field] !== undefined && baseData[field] !== '') {
        submitData[field] = baseData[field]
      }
    })

    // 如果选择不分类，移除categoryId
    if (submitData.isClassification === '1') {
      delete submitData.categoryId
    }

    return submitData
  }
  else {
    // 编辑模式：如果未选择分类，设置为0
    if (baseData.categoryId === undefined) {
      baseData.categoryId = 0
    }

    return baseData
  }
}

// 关闭弹窗
function closeFun() {
  formRef.value?.resetFields()
  showCategorySelect.value = false
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="518"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ props.type === 'create' ? '创建渠道' : '编辑渠道' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState" :label-col="{ span: 5 }" :wrapper-col="{ span: 18 }">
        <!-- 渠道名称  必选 -->
        <a-form-item label="渠道名称" name="channelName" :rules="[{ required: true, message: '请输入渠道名称' }]">
          <a-input v-model:value="formState.channelName" placeholder="请输入渠道名称" />
        </a-form-item>
        <!-- 是否分类 -->
        <a-form-item
          v-if="props.type === 'create'" label="是否分类" name="isClassification"
          :rules="[{ required: true, message: '请选择是否分类' }]"
        >
          <a-radio-group v-model:value="formState.isClassification" class="custom-radio">
            <a-radio value="1">
              不分类
            </a-radio>
            <a-radio value="2">
              是，添加至已有分类
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="props.type === 'edit'" label="归属分类">
          <a-select
            v-if="showCategorySelect" v-model:value="formState.categoryId" placeholder="请选择分类"
            :allow-clear="true"
          >
            <a-select-option v-for="category in props.categoryList" :key="category.id" :value="category.id">
              {{ category.categoryName }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <!-- 选择分类 必选 -->
        <a-form-item
          v-if="formState.isClassification === '2'" label="选择分类" name="categoryId"
          :rules="[{ required: true, message: '请选择分类' }]"
        >
          <a-select v-model:value="formState.categoryId" placeholder="请选择分类">
            <a-select-option v-for="category in props.categoryList" :key="category.id" :value="category.id">
              {{ category.categoryName }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <!-- 备注 -->
        <a-form-item label="备注" name="remark">
          <a-textarea v-model:value="formState.remark" placeholder="请输入备注" :auto-size="{ minRows: 1, maxRows: 1 }" />
        </a-form-item>
      </a-form>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="loading" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
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
  background: #fafafa;
  margin: 24px;
  border-radius: 12px;
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
