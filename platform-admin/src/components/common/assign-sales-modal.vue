<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import StaffSelect from './staff-select.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '分配销售',
  },
  type: {
    type: [String, Number],
    default: 1,
  },
  selectedStudents: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:open', 'submit'])

const formRef = ref()
const loading = ref(false)

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  studentIds: [],
  salespersonId: undefined,
})

// 监听学员数据变化，更新studentIds和销售员ID
watch(() => props.selectedStudents, (newStudents) => {
  formState.studentIds = newStudents.map(student => student.id)
  // 如果只有一个学员，回显当前销售员
  if (newStudents.length === 1) {
    const student = newStudents[0]
    // 确保 salePerson 存在且有效（不为空、不为0、不为null、不为undefined）
    if (student.salePerson && student.salePerson !== '0' && student.salePerson !== 0) {
      // salePerson 是字符串，转换为数字
      formState.salespersonId = Number(student.salePerson)
    } else {
      formState.salespersonId = undefined
    }
  } else {
    formState.salespersonId = undefined
  }
}, { immediate: true })

// 获取完整的学员姓名列表用于tooltip
const fullStudentNames = computed(() => {
  return props.selectedStudents.map(s => s.stuName).join('、')
})

// 手动触发验证
async function handleSubmit() {
  try {
    loading.value = true
    
    // 如果是批量转入公有池（type === 3），不需要验证表单
    if (props.type !== 3) {
      await formRef.value.validate()
    }
    
    // 准备提交数据
    const submitData = {
      studentIds: formState.studentIds,
      salespersonId: formState.salespersonId,
    }
    
    // 触发提交事件
    emit('submit', submitData)
  }
  catch (error) {
    console.log('验证失败:', error)
    loading.value = false
  }
}

function closeFun() {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  loading.value = false
  openModal.value = false
}

// 监听modal打开/关闭，重置表单或回显数据
watch(openModal, (newVal) => {
  if (!newVal) {
    // 关闭时重置
    formState.salespersonId = undefined
    if (formRef.value) {
      formRef.value.resetFields()
    }
    loading.value = false
  } else {
    // 打开时回显当前销售员（如果只有一个学员）
    if (props.selectedStudents.length === 1) {
      const student = props.selectedStudents[0]
      // 确保 salePerson 存在且有效
      if (student.salePerson && student.salePerson !== '0' && student.salePerson !== 0) {
        formState.salespersonId = Number(student.salePerson)
      } else {
        formState.salespersonId = undefined
      }
    }
  }
})

// 从父组件调用的关闭loading方法
defineExpose({
  closeLoading: () => {
    loading.value = false
  }
})
</script>

<template>
  <a-modal v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ title }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="rounded-12px bg-#fafafa mx-24px">
      <div v-if="type === 1 || type === 3" class="contenter flex student-display" style="margin-bottom: 0;">
        <span class="student-label">学员：</span>
        <span class="student-count">共 <span class="text-#06f mx-2">{{ selectedStudents.length }}</span> 人，</span>
        <a-tooltip :title="fullStudentNames" placement="topLeft">
          <span class="student-names">{{ fullStudentNames }}</span>
        </a-tooltip>
      </div>
      <div v-if="type === 1" class="w-752px">
        <a-divider class="my-0" />
      </div>
      <div v-if="type !== 3" class="contenter scrollbar" :class="type == 1 ? 'mt0' : ''">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 3 }" :wrapper-col="{ span: 21 }">
          <!-- 销售员 必选 -->
          <a-form-item label="销售员" name="salespersonId" :rules="[{ required: true, message: '请选择销售员' }]">
            <StaffSelect v-model="formState.salespersonId" placeholder="请选择销售员" width="300px" :status="0" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
      </a-button>
      <a-button v-if="type !== 3" type="primary" :loading="loading" @click="handleSubmit">
        确定
      </a-button>
      <a-button v-if="type === 3" type="primary" :loading="loading" @click="handleSubmit">
        确定转入公有池
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
  background: #fafafa;
  margin: 24px;

  .multiple-select {
    :deep(.ant-select-selection-item) {
      background-color: #e6f0ff;
      border: 1px solid #99c2ff;
    }
  }
}

.student-display {
  align-items: center;
  
  .student-label,
  .student-count {
    flex-shrink: 0;
  }
  
  .student-names {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
    
    &:hover {
      color: #1890ff;
    }
  }
}

.mt0 {
  margin-top: 0 !important;
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
