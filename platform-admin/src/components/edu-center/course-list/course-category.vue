<script setup>
import {
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { Modal } from 'ant-design-vue'
import { createVNode } from 'vue'
import { addCourseCategoryApi, deleteCourseCategoryApi, getCourseCategoryPageApi, updateCourseCategoryApi } from '~@/api/edu-center/course-list'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '~@/utils/messageService'

const displayArray = ref(['courseCategory'])
const checked = ref(true)
const dataSource = ref([])
const allColumns = ref([
  {
    title: '序号',
    dataIndex: 'order',
    key: 'order',
    width: 100,
    // 排序
    sorter: true,
  },
  {
    title: '课程类别',
    dataIndex: 'courseCategory',
    key: 'courseCategory',
  },
  // {
  //   title: "是否启用",
  //   dataIndex: "status",
  //   key: "status",
  //   width: 140,
  // },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'course-category', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
const allFilterRef = ref(null)
const createCourseCategoryModalOpen = ref(false)
const formRef = ref(null)
const formState = reactive({
  name: undefined,
  sort: undefined,
})
const loading = ref(false)
const sortModel = ref({
  orderBySortNumber: 2,
})

// 存储所有查询条件的响应式对象
const queryState = ref({
  searchKey: undefined,
  courseCategoryId: undefined,
})

// 重置所有查询条件
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    queryState.value[key] = undefined
  })
}

// 使用防抖处理所有筛选条件的更新
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    // 如果是清空所有，重置所有查询条件
    resetQueryState()
  }
  else {
    // 处理更新
    Object.entries(updates).forEach(([key, value]) => {
      if (Array.isArray(value) && value.length === 0) {
        // 如果是空数组，清除字段
        queryState.value[key] = undefined
      }
      else {
        queryState.value[key] = value
      }
    })
  }

  pagination.value.current = 1
  getCourseCategoryPage(id, type)
}, 300, { leading: true, trailing: false })

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})

// 过滤器字段映射
const filterFieldMapping = {
  courseCategoryFilter: 'courseCategoryId',
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })

  return handlers
})

// 添加一个单独的搜索关键字，只用于课程类别选项筛选
const searchCategoryKey = ref(undefined)

// 处理表格排序变化
function handleTableChange(paginationInfo, filters, sorter) {
  console.log('排序变化:', sorter)
  console.log(paginationInfo)
  // 重置所有排序字段为0
  Object.keys(sortModel.value).forEach((key) => {
    sortModel.value[key] = 0
  })

  // 处理排序逻辑（支持单列排序和多列排序）
  const sortFieldMap = {
    'order': 'orderBySortNumber',
  }

  // 如果是数组，说明是多列排序，我们只取第一个（因为后端只支持单列排序）
  const currentSorter = Array.isArray(sorter) ? sorter[0] : sorter
  if (currentSorter && currentSorter.order) {
    const sortField = sortFieldMap[currentSorter.field]
    if (sortField) {
      // ascend: 升序(1), descend: 降序(2)
      sortModel.value[sortField] = currentSorter.order === 'ascend' ? 1 : 2
    }
  }

  // 更新分页信息
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize

  // 重新获取数据
  getCourseCategoryPage()
}

function closeCreateCourseCategoryModal() {
  formRef.value.resetFields()
  createCourseCategoryModalOpen.value = false
}
const confirmLoading = ref(false)
async function handleCreateCourseCategory() {
  // 先校验非空
  try {
    await formRef.value.validate() // 关键3：通过引用调用验证方法
    confirmLoading.value = true
    const res = formState.id ? await updateCourseCategoryApi(formState) : await addCourseCategoryApi(formState)
    if (res.code === 200) {
      createCourseCategoryModalOpen.value = false
      getCourseCategoryPage()
      messageService.success('创建成功')
      formRef.value.resetFields()
      confirmLoading.value = false
    }
    else {
      messageService.error('创建失败')
    }
  }
  catch (error) {
    console.log(error)
  }
}
async function handleEditCourseCategory(record) {
  createCourseCategoryModalOpen.value = true
  formState.name = record.name
  formState.sort = record.sort
  formState.id = record.id
  formState.version = record.version
  formState.uuid = record.uuid
}
async function handleDeleteCourseCategory(record) {
  Modal.confirm({
    title: '确定删除该课程类别？',
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    content: '删除后无法恢复，请谨慎操作。',
    onOk: async () => {
      try {
        const res = await deleteCourseCategoryApi(record)
        if (res.code === 200) {
          getCourseCategoryPage()
          messageService.success('删除成功')
        }
        else {
          messageService.error('删除失败')
        }
      }
      catch (error) {
        console.log(error)
      }
    },
    onCancel() { },
  })
}

async function getCourseCategoryPage(id, type) {
  try {
    loading.value = true

    // 过滤掉undefined的值，只传递有效的查询条件
    const validQueryParams = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => value !== undefined),
    )

    const res = await getCourseCategoryPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
        'skipCount': 1,
      },
      'queryModel': {
        ...validQueryParams, // 展开所有有效的查询条件
      },
      'sortModel': {
        'orderBySortNumber': sortModel.value.orderBySortNumber,
      },
    })
    if (res.code === 200) {
      dataSource.value = res.result
      pagination.value.total = res.total
      loading.value = false
      // 清除快捷筛选标记
      if (allFilterRef.value && id && type) {
        allFilterRef.value.clearQuickFilter(id, type)
      }
    }
  }
  catch (error) {
    console.log(error)
  }
}

onMounted(() => {
  getCourseCategoryPage()
})
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-tl-0 rounded-tr-0">
      <all-filter ref="allFilterRef" :display-array="displayArray" v-on="filterUpdateHandlers" />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计{{ dataSource.length }}条类别
          </div>
          <div class="edit flex">
            <a-button type="primary" class="mr-2" @click="createCourseCategoryModalOpen = true">
              创建课程类别
            </a-button>
            <!-- 自定义字段 -->
            <!-- <customize-code
                v-model:checkedValues="selectedValues"
                :options="columnOptions"
                :total="allColumns.length - 1"
                :num="selectedValues.length"
              /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :loading="loading" :pagination="pagination" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'courseCategory'">
                <span class="flex justify-start">{{ column.title }}</span>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'order'">
                {{ record.sort }}
              </template>
              <template v-if="column.key === 'courseCategory'">
                <div>{{ record.name }}</div>
              </template>
              <!-- <template v-if="column.key === 'status'">
                <a-switch v-model:checked="checked" />
              </template> -->
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <a class="mr-3" @click="handleEditCourseCategory(record)">编辑</a>
                  <a class="text-#f03" @click="handleDeleteCourseCategory(record)">删除</a>
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <!-- 创建课程类别modal -->
    <a-modal
      v-model:open="createCourseCategoryModalOpen" width="400px" :confirm-loading="confirmLoading"
      @ok="handleCreateCourseCategory" @cancel="closeCreateCourseCategoryModal"
    >
      <template #title>
        <span>创建课程类别</span>
      </template>
      <div>
        <a-form ref="formRef" :model="formState" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
          <a-form-item label="课程类别：" name="name" :rules="[{ required: true, message: '请输入课程类别名称（20字以内）' }]">
            <a-input v-model:value="formState.name" placeholder="请输入名称（20字以内）" />
          </a-form-item>
          <a-form-item label="排序：" name="sort" :rules="[{ required: true, message: '请输入排序' }]">
            <a-input-number v-model:value="formState.sort" placeholder="请排序" />
          </a-form-item>
        </a-form>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.studentStatus {
  display: flex;
  align-items: center;

  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
  }
}
</style>
