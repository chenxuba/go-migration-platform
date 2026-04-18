<!-- 编辑部门 -->
<script setup>
import { message } from 'ant-design-vue'
import { updateDepart } from '@/api/internal-manage/staff-manage'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  params: {
    type: Object,
    required: true,
    default: () => ({}),
  },
  treeData: {
    type: Array,
    required: true,
    default: () => [],
  },
})

const emit = defineEmits(['update:open', 'success'])

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const form = ref()

const formState = reactive({
  departName: '',
  pid: undefined,
  id: undefined,
})

// 表单验证规则
const rules = {
  departName: [
    { required: true, message: '请输入部门名称', trigger: 'blur' },
    { max: 15, message: '部门名称不能超过15个字符', trigger: 'blur' },
  ],
  pid: [
    { required: true, message: '请选择上级部门', trigger: 'change' },
  ],
}

async function handleOk() {
  try {
    await form.value?.validate()

    const { uuid, version } = props.params
    loading.value = true

    const res = await updateDepart({
      ...formState,
      uuid,
      version,
    })
    if (res.code === 200) {
      messageService.success('部门更新成功')
      emit('success')
    }
    else {
      messageService.error(res.message)
    }
    handleCancel()
  }
  catch (error) {
    console.error('更新部门失败:', error)
    if (error?.message) {
      message.error(error.message)
    }
  }
  finally {
    loading.value = false
  }
}

function handleCancel() {
  openModal.value = false
  form.value?.resetFields()
}

// 初始化表单数据
function initFormData(params) {
  if (params) {
    formState.departName = params.departName || ''
    formState.pid = params.data?.pid
    formState.id = params.id
  }
}

// 监听参数变化，初始化表单
watch(
  () => props.params,
  (newParams) => {
    if (newParams) {
      initFormData(newParams)
    }
  },
  { immediate: true },
)

// 处理树数据，禁用当前编辑部门及其所有子孙部门
const processedTreeData = computed(() => {
  if (!props.treeData || !formState.id) {
    return props.treeData
  }
  
  // 获取当前编辑部门的所有子孙部门ID
  function getDescendantIds(targetId, treeData) {
    const descendantIds = new Set([targetId]) // 包含自身
    
    // 递归查找指定节点的所有子孙节点
    function findDescendants(nodes) {
      nodes.forEach(node => {
        if (node.id === targetId && node.children) {
          // 找到目标节点，收集其所有子孙节点
          collectAllChildren(node.children)
        } else if (node.children) {
          // 继续在子树中查找
          findDescendants(node.children)
        }
      })
    }
    
    // 收集所有子孙节点ID
    function collectAllChildren(children) {
      children.forEach(child => {
        descendantIds.add(child.id)
        if (child.children) {
          collectAllChildren(child.children)
        }
      })
    }
    
    findDescendants(treeData)
    return descendantIds
  }
  
  const disabledIds = getDescendantIds(formState.id, props.treeData)
  
  // 递归处理树数据，禁用当前编辑部门及其所有子孙部门
  function processNode(node) {
    const processed = { ...node }
    
    // 如果当前节点ID在禁用列表中，则禁用
    if (disabledIds.has(processed.id)) {
      processed.disabled = true
    }
    
    // 递归处理子节点
    if (processed.children && processed.children.length > 0) {
      processed.children = processed.children.map(child => processNode(child))
    }
    
    return processed
  }
  
  return props.treeData.map(node => processNode(node))
})
</script>

<template>
  <a-modal
    v-model:open="openModal" title="编辑部门" :mask-closable="false" :keyboard="false" destroy-on-close
    :centered="true" width="400px" @ok="handleOk" @cancel="handleCancel"
  >
    <a-form ref="form" :model="formState" :rules="rules" autocomplete="off" layout="vertical">
      <a-form-item label="部门名称" name="departName">
        <a-input v-model:value="formState.departName" placeholder="请输入部门名称（15字以内）" :maxlength="15" show-count />
      </a-form-item>

      <a-form-item label="上级部门" name="pid">
        <a-tree-select
          v-model:value="formState.pid" placeholder="请选择上级部门" allow-clear tree-default-expand-all
          :tree-data="processedTreeData" :field-names="{
            children: 'children',
            label: 'departName',
            value: 'id',
          }"
        />
      </a-form-item>
    </a-form>

    <template #footer>
      <a-space>
        <a-button @click="handleCancel">
          取消
        </a-button>
        <a-button :loading="loading" type="primary" @click="handleOk">
          确定
        </a-button>
      </a-space>
    </template>
  </a-modal>
</template>

<style scoped lang="less">
:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}
</style>
