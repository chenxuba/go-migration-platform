<script setup>
import { CloseOutlined, SoundFilled, EditOutlined, CopyOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
const formRef = ref()

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 工资项目列表
const salaryProjects = ref([
  { id: 1, name: '基本工资' },
  { id: 2, name: '课时费' },
  { id: 3, name: '销售业绩' },
])

// 新增项目输入框
const newProjectName = computed({
  get: () => formState.newProjectName,
  set: (value) => formState.newProjectName = value
})
const maxLength = 20

// 显示添加输入框
const showAddInput = ref(false)

// 检查是否有正在进行的操作
const hasActiveOperation = computed(() => {
  return editingId.value !== null || showAddInput.value
})

// 编辑状态管理
const editingId = ref(null)
const editingName = computed({
  get: () => formState.editingName,
  set: (value) => formState.editingName = value
})

const formState = reactive({
  newProjectName: '',
  editingName: ''
})

// 表单验证规则
const rules = {
  newProjectName: [
    { required: true, message: '请输入工资项目名称', trigger: 'blur' },
    { min: 1, max: 20, message: '长度在 1 到 20 个字符', trigger: 'blur' }
  ],
  editingName: [
    { required: true, message: '请输入工资项目名称', trigger: 'blur' },
    { min: 1, max: 20, message: '长度在 1 到 20 个字符', trigger: 'blur' }
  ]
}

// 手动触发验证
async function handleSubmit() {
  try {
    await formRef.value.validate()
    console.log('验证通过，提交数据:', formState)
    console.log('工资项目列表:', salaryProjects.value)
  }
  catch (error) {
    console.log('验证失败:', error)
  }
}

function closeFun() {
  formRef.value?.resetFields()
  openModal.value = false
  showAddInput.value = false
  formState.newProjectName = ''
  // 清除编辑状态
  editingId.value = null
  formState.editingName = ''
}

// 编辑项目
function editProject(project) {
  // 如果正在新增项目，先取消新增
  if (showAddInput.value) {
    cancelAdd()
  }
  
  // 如果正在编辑其他项目，先取消编辑
  if (editingId.value && editingId.value !== project.id) {
    cancelEdit()
  }
  
  editingId.value = project.id
  formState.editingName = project.name
  nextTick(() => {
    // 聚焦编辑输入框
    const input = document.querySelector('.edit-input input')
    if (input) input.focus()
  })
}

// 保存编辑
async function saveEdit() {
  try {
    await formRef.value.validateFields(['editingName'])
    const index = salaryProjects.value.findIndex(p => p.id === editingId.value)
    if (index !== -1) {
      salaryProjects.value[index].name = formState.editingName.trim()
    }
    cancelEdit()
  } catch (error) {
    console.log('编辑验证失败:', error)
  }
}

// 取消编辑
function cancelEdit() {
  editingId.value = null
  formState.editingName = ''
  // 清除验证错误
  formRef.value?.clearValidate(['editingName'])
}

// 删除项目
function deleteProject(index) {
  const project = salaryProjects.value[index]
  
  // 如果正在编辑被删除的项目，先取消编辑
  if (editingId.value === project.id) {
    cancelEdit()
  }
  
  salaryProjects.value.splice(index, 1)
}

// 显示添加输入框
function showAddInputBox() {
  // 如果正在编辑项目，先取消编辑
  if (editingId.value) {
    cancelEdit()
  }
  
  showAddInput.value = true
  nextTick(() => {
    // 聚焦输入框
    const input = document.querySelector('.add-input input')
    if (input) input.focus()
  })
}

// 保存新项目
async function saveNewProject() {
  try {
    await formRef.value.validateFields(['newProjectName'])
    salaryProjects.value.push({
      id: Date.now(),
      name: formState.newProjectName.trim()
    })
    formState.newProjectName = ''
    showAddInput.value = false
  } catch (error) {
    console.log('添加验证失败:', error)
  }
}

// 取消添加
function cancelAdd() {
  formState.newProjectName = ''
  showAddInput.value = false
  // 清除验证错误
  formRef.value?.clearValidate(['newProjectName'])
}

// 拖拽相关
function onDragStart(event, index) {
  // 如果有活动操作，阻止拖拽
  if (hasActiveOperation.value) {
    event.preventDefault()
    return
  }
  
  event.dataTransfer.setData('text/plain', index)
}

function onDragOver(event) {
  event.preventDefault()
}

function onDrop(event, dropIndex) {
  event.preventDefault()
  const dragIndex = parseInt(event.dataTransfer.getData('text/plain'))
  if (dragIndex !== dropIndex) {
    const dragItem = salaryProjects.value[dragIndex]
    salaryProjects.value.splice(dragIndex, 1)
    salaryProjects.value.splice(dropIndex, 0, dragItem)
  }
}
</script>

<template>
  <a-modal v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="520" :footer="false">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>工资项目管理</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules">
        <div class="tip-text">
          <SoundFilled class="sound-icon" />
          按住拖拽图标可进行排序，此处的顺序将影响工资构成模板中的工资项目排序
        </div>

        <div class="project-list">
                  <div v-for="(project, index) in salaryProjects" :key="project.id" class="project-item"
          :class="{ 'editing': editingId === project.id, 'disabled': hasActiveOperation && editingId !== project.id }" 
          :draggable="!hasActiveOperation"
          @dragstart="onDragStart($event, index)" @dragover="onDragOver" @drop="onDrop($event, index)">
            <div class="project-content">
              <div class="drag-handle">
                <div class="drag-dots">
                  <div class="dot"></div>
                  <div class="dot"></div>
                  <div class="dot"></div>
                  <div class="dot"></div>
                  <div class="dot"></div>
                  <div class="dot"></div>
                </div>
              </div>

              <!-- 编辑状态显示输入框 -->
              <div v-if="editingId === project.id" class="edit-input-container">
                <a-form-item name="editingName" class="edit-form-item">
                  <a-input v-model:value="formState.editingName" placeholder="请输入" class="edit-input"
                    :maxlength="maxLength" @pressEnter="saveEdit" />
                </a-form-item>
                <span class="char-count">{{ formState.editingName.length }} / {{ maxLength }}</span>
                <div class="edit-actions">
                  <a-button type="link" size="small" @click="saveEdit">保存</a-button>
                  <a-button type="link" size="small" @click="cancelEdit">取消</a-button>
                </div>
              </div>

              <!-- 正常状态显示项目名称 -->
              <span v-else class="project-name">{{ project.name }}</span>
            </div>
                      <div v-if="editingId !== project.id" class="project-actions">
            <a-button type="text" size="small" :disabled="hasActiveOperation" @click="editProject(project)">
              <EditOutlined />
            </a-button>
            <a-button type="text" size="small" :disabled="hasActiveOperation" @click="deleteProject(index)">
              <DeleteOutlined />
            </a-button>
          </div>
          </div>
        </div>

        <div v-if="!showAddInput" class="add-project-btn" 
             :class="{ 'disabled': hasActiveOperation }"
             @click="!hasActiveOperation && showAddInputBox()">
          <PlusOutlined />
          添加工资项目
        </div>
        <div v-if="showAddInput" class="add-input-container">
          <div class="input-row">
            <a-form-item name="newProjectName" class="add-form-item">
              <a-input v-model:value="formState.newProjectName" class="add-input" :maxlength="maxLength"
                placeholder="请输入" @pressEnter="saveNewProject" />
            </a-form-item>
            <span class="char-count">{{ formState.newProjectName.length }} / {{ maxLength }}</span>
            <div class="input-actions">
              <a-button type="link" size="small" @click="saveNewProject">保存</a-button>
              <a-button type="link" size="small" @click="cancelAdd">取消</a-button>
            </div>
          </div>
        </div>
      </a-form>
    </div>

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

.contenter {
  padding: 14px 24px;
  padding-bottom: 40px;
  :deep(.ant-form-item-explain-error) {
    position: absolute;
    font-size: 12px;
  }
}

.tip-text {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  padding: 12px 16px;
  background: #f6f6f6;
  border-radius: 6px;
  font-size: 12px;
  color: #666;
  line-height: 1.4;

  .sound-icon {
    margin-right: 8px;
    color: #1890ff;
    font-size: 14px;
  }
}

.project-list {
  margin-bottom: 20px;
}

.project-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  margin-bottom: 8px;
  background: #fff;
  cursor: move;
  transition: all 0.2s;

  &:hover {
    border-color: #1890ff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  &:last-child {
    margin-bottom: 0;
  }

  &.editing {
    border-color: #1890ff;
    background: #f0f9ff;
    cursor: default;
  }

  &.disabled {
    opacity: 0.6;
    pointer-events: none;
  }
}

.project-content {
  display: flex;
  align-items: center;
  flex: 1;
}

.drag-handle {
  margin-right: 12px;
  cursor: grab;

  &:active {
    cursor: grabbing;
  }
}

.drag-dots {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px;
  width: 12px;
  height: 12px;

  .dot {
    width: 3px;
    height: 3px;
    background: #bbb;
    border-radius: 50%;
  }
}

.project-name {
  font-size: 14px;
  color: #333;
}

.edit-input-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;

  .edit-form-item {
    flex: 1;
    margin-bottom: 0;

    .edit-input {
      width: 100%;
    }
  }

  .char-count {
    font-size: 12px;
    color: #999;
    white-space: nowrap;
  }

  .edit-actions {
    display: flex;
  }
}

.project-actions {
  display: flex;
  gap: 4px;

  .ant-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border: none;
    color: #666;

    &:hover {
      color: #1890ff;
      background: #f0f9ff;
    }
  }
}

.add-project-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 12px;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 12px;

  &:hover:not(.disabled) {
    border-color: #1890ff;
    color: #1890ff;
    background: #f0f9ff;
  }

  &.disabled {
    opacity: 0.5;
    cursor: not-allowed;
    color: #bbb;
    border-color: #f0f0f0;
  }

  .anticon {
    margin-right: 8px;
  }
}

.add-input-container {
  margin-bottom: 12px;

  .input-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .add-form-item {
    flex: 1;
    margin-bottom: 0;

    .add-input {
      width: 100%;
    }
  }

  .char-count {
    font-size: 12px;
    color: #999;
    white-space: nowrap;
  }

  .input-actions {
    display: flex;
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
