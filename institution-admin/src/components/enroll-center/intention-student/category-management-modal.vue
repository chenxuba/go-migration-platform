<script setup>
import { CloseOutlined, PlusCircleFilled } from '@ant-design/icons-vue'
import { computed, h, ref, watch } from 'vue'
import { deleteCategoryApi, saveCategoryApi, updateCategoryApi } from '~@/api/enroll-center/intention-student'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  channelCategoryList: {
    type: Array,
    default: () => [],
  },
})
const emit = defineEmits(['update:open', 'update:channelCategoryList'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 确认loading
const confirmLoading = ref(false)
// 模式
const modalType = ref('create')

const list = ref([])

// Watch for changes to channelCategoryList and update list accordingly
watch(() => props.channelCategoryList, (newVal) => {
  list.value = newVal || []
}, { immediate: true })

const createCategoryModal = ref(false)
function createCategory() {
  // 创建模式
  modalType.value = 'create'
  // 重置表单数据，清除所有属性只保留categoryName
  Object.keys(formState).forEach((key) => {
    if (key !== 'categoryName') {
      delete formState[key]
    }
  })
  formState.categoryName = ''
  createCategoryModal.value = true
}
const formState = reactive({
  categoryName: '',
})
// 删除分类
async function deleteCategory(item) {
  const res = await deleteCategoryApi(item)
  if (res.code === 200) {
    messageService.success('删除成功')
    emit('update:channelCategoryList')
  }
}
async function deleteChannel(item) {
  // 无法删除，此分类下有关联渠道
  messageService.error('无法删除，此分类下有关联渠道')
}
// 编辑分类
async function editCategory(item) {
  // 打开编辑弹窗
  createCategoryModal.value = true
  // 设置表单数据
  Object.assign(formState, item)
  // 编辑模式
  modalType.value = 'edit'
}
// 手动触发验证
async function createCategoryOk() {
  confirmLoading.value = true
  if (modalType.value === 'create') {
    try {
      await formRef.value.validate() // 关键3：通过引用调用验证方法
      console.log('验证通过，提交数据:', formState)
      const res = await saveCategoryApi(formState)
      console.log(res)
      if (res.code === 200) {
        messageService.success('创建成功')
        createCategoryModal.value = false
        emit('update:channelCategoryList')
      }
    }
    catch (error) {
      console.log('验证失败:', error)
    }
    finally {
      confirmLoading.value = false
    }
  }
  else {
    // 编辑模式
    console.log('编辑模式，提交数据:', formState)
    try {
      await formRef.value.validate() // 关键3：通过引用调用验证方法
      console.log('验证通过，提交数据:', formState)
      const res = await updateCategoryApi(formState)
      if (res.code === 200) {
        messageService.success('编辑成功')
        createCategoryModal.value = false
        emit('update:channelCategoryList')
      }
    }
    catch (error) {
      console.log('验证失败:', error)
    }
    finally {
      confirmLoading.value = false
    }
  }
}
function createCategoryCancel() {
  formRef.value.resetFields()
  createCategoryModal.value = false
}
function closeFun() {
  openModal.value = false
}
// 监听openModal，如果打开则重新请求分类列表
watch(openModal, (newVal) => {
  if (newVal) {
    emit('update:channelCategoryList')
  }
})
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="700" :footer="false"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>分类管理</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="flex justify-between items-center px-24px py-12px">
        <span class="text-#666">共 {{ list.length }} 个分类(仅支持删除不含渠道的分类)</span>
        <a-button type="primary" :icon="h(PlusCircleFilled)" @click="createCategory">
          创建分类
        </a-button>
      </div>
      <div class="line" />
      <div v-for="item in list" :key="item" class="flex justify-between items-center px-24px pb-12px ul text-#222">
        <div
          class="flex justify-between flex-1 bg-#fafafa px-16px h-48px rounded-8px items-center cursor-pointer hover"
        >
          <span>{{ item.categoryName }}</span>
          <span
            class="text-3 px3 py0.5 rounded-10 ml2"
            :class="item.channelCount > 0 ? 'bg-#e6f0ff text-#06f' : 'bg-#eee text-#888'"
          >含 {{ item.channelCount || 0 }}
            个渠道</span>
        </div>
        <div class="flex mx-8px">
          <div
            class="mr-8px w-48px h-48px rounded-8px bg-#fafafa text-#06f flex-center cursor-pointer hover"
            @click="editCategory(item)"
          >
            编辑
          </div>
          <a-popconfirm
            title="确定删除该分类吗？" ok-text="确定" cancel-text="取消"
            @confirm="deleteCategory(item)"
          >
            <div
              v-if="item.channelCount == 0" class="w-48px h-48px rounded-8px bg-#fafafa  flex-center cursor-pointer hover"
              :class="item.channelCount == 0 ? 'text-#ff3333' : 'text-#ccc'"
            >
              删除
            </div>
          </a-popconfirm>
          <div
            v-if="item.channelCount > 0" class="w-48px h-48px rounded-8px bg-#fafafa  flex-center cursor-pointer hover" :class="item.channelCount == 0 ? 'text-#ff3333' : 'text-#ccc'"
            @click="deleteChannel(item)"
          >
            删除
          </div>
        </div>
      </div>
    </div>
    <a-modal
      v-model:open="createCategoryModal" centered :confirm-loading="confirmLoading"
      :title="modalType === 'create' ? '创建分类' : '编辑分类'" @ok="createCategoryOk" @cancel="createCategoryCancel"
    >
      <!-- form -->
      <a-form ref="formRef" :model="formState">
        <!-- 必填 -->
        <a-form-item name="categoryName" :rules="[{ required: true, message: '请输入分类名称' }]">
          <a-textarea
            v-model:value="formState.categoryName" class="bg-#f6f7f8 border-none" placeholder="请输入分类名称，最多10字"
            :auto-size="{ minRows: 4, maxRows: 6 }"
          />
        </a-form-item>
      </a-form>
    </a-modal>
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
  height: calc(100vh - 300px);
  overflow-y: auto;

  .line {
    height: 13px;
    background: linear-gradient(0deg, hsla(0, 0%, 100%, 0), rgba(0, 0, 0, .3));
    opacity: .1;
  }

  .ul {
    &:hover {
      .hover {
        background: #e6f0ff;
      }
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
