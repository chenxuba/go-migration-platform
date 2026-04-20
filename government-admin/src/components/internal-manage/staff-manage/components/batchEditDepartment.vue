<!-- 批量修改所属部门 -->
<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { TreeSelect, message } from 'ant-design-vue'
import { batchModifyDept } from '@/api/internal-manage/staff-manage'
import messageService from '~@/utils/messageService'

const props = defineProps({
  batchUserList: {
    type: Array,
    default: () => [],
  },
  departmentList: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['success'])
const loading = ref(false)

const open = defineModel({
  type: Boolean,
  default: false,
})

const formState = reactive({
  deptIds: [],
})

const form = ref(null)

// 表单验证规则
const rules = {
  deptIds: [
    { required: true, message: '请选择所属部门', trigger: 'change' },
  ],
}

function cancel() {
  open.value = false
  form.value?.resetFields()
  formState.deptIds = []
}

async function handleOk() {
  try {
    await form.value?.validate()

    if (props.batchUserList.length === 0) {
      message.error('没有可修改的员工')
      return
    }

    loading.value = true

    const params = {
      ...formState,
      userIds: props.batchUserList.map(item => item.id)
    }

    const res = await batchModifyDept(params)

    if (res.code === 200) {
      messageService.success('批量修改部门成功')
      emit('success')
      cancel()
    } else {
      messageService.error(res.message || '修改失败')
    }
  } catch (error) {
    console.error('批量修改部门失败:', error)
  } finally {
    loading.value = false
  }
}

// 计算员工名称显示文本
const employeeNames = computed(() => {
  if (props.batchUserList.length === 0) return ''

  const names = props.batchUserList.map(item => item.nickName)
  if (names.length <= 3) {
    return names.join('、')
  }
  return `${names.slice(0, 3).join('、')}等`
})
</script>

<template>
  <a-modal v-model:open="open" :mask-closable="false" :keyboard="false" width="800px" :destroy-on-close="true"
    :closable="false" :centered="true" :confirm-loading="loading" @cancel="cancel">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>批量修改所属部门</span>
        <a-button type="text" class="close-btn" @click="cancel">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <template #footer>
      <a-button @click="cancel">取消</a-button>
      <a-button type="primary" :loading="loading" @click="handleOk">确定</a-button>
    </template>

    <div>
      <div class="bg-#fafafa rd-10px">
        <div class="p-20px">
          <div class="flex items-center gap-2">
            <span>员工：共</span>
            <span class="text-#06f font-medium">{{ batchUserList.length }}</span>
            <span>人</span>,
            <div class="text-#888" v-if="employeeNames">
              {{ employeeNames }}
            </div>
          </div>

        </div>

        <div class="w-100% h-.5px bg-#ddd" />

        <div class="p-20px">
          <a-form ref="form" :model="formState" :rules="rules">
            <a-form-item label="所属部门" name="deptIds" extra="选择员工的新部门，可以选择多个">
              <a-tree-select v-model:value="formState.deptIds" style="width: 100%"
                :dropdown-style="{ maxHeight: '400px', overflow: 'auto' }" placeholder="请选择部门（可多选）" allow-clear multiple
                tree-default-expand-all :tree-data="departmentList" tree-node-filter-prop="departName"
                :field-names="{ children: 'children', label: 'departName', value: 'id' }"
                :show-checked-strategy="TreeSelect.SHOW_ALL" show-search />
            </a-form-item>
          </a-form>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
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

:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}

:deep(.ant-modal-body) {
  padding: 12px 24px;
}

:deep(.modal-wrap .ant-modal) {
  top: 30px;
}

:deep(.ant-tree-select-dropdown) {
  .ant-select-tree-treenode {
    padding: 4px 8px;

    &:hover {
      background-color: #f5f5f5;
    }
  }

  .ant-select-tree-node-content-wrapper {
    flex: 1;

    &:hover {
      background-color: transparent;
    }
  }
}

:deep(.ant-form-item-extra) {
  color: #999;
  font-size: 12px;
  margin-top: 4px;
}
</style>
