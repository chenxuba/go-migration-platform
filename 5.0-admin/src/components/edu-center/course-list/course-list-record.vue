<script setup>
import { computed, createVNode, onMounted, onUnmounted, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { addCourseApi, batchDelOrResCourseApi, batchOpenMicroSchoolShowApi, batchSaleStatusApi, getCoursePageApi, updateCourseApi } from '~@/api/edu-center/course-list'
import { useTableColumns } from '@/composables/useTableColumns'
import emitter, { EVENTS } from '~@/utils/eventBus'
import messageService from '~@/utils/messageService'

const displayArray = ref(['courseCategory', 'teachingMethod', 'billingMode', 'saleStatus', 'hasTrialPrice', 'isMicroSchoolSale', 'isMicroSchoolDisplay', 'courseAttribute'])

const assignSalesVisible = ref(false)
const modalType = ref('create')
const selectedRows = ref([])
const selectedRowKeys = ref([])
const dataSource = ref([])
const allColumns = ref([
  {
    title: '课程名称',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 180,
    required: true, // 新增必选标识
  },
  {
    title: '课程类别',
    key: 'categoryName',
    dataIndex: 'categoryName',
    width: 120,
  },
  {
    title: '是否是通用课',
    key: 'courseType',
    dataIndex: 'courseType',
    width: 120,
  },
  {
    title: '授课方式',
    key: 'teachingMethod',
    dataIndex: 'teachingMethod',
    width: 110,
  },
  {
    title: '收费方式',
    dataIndex: 'chargingMethod',
    key: 'chargingMethod',
    width: 200,
  },
  {
    title: '售卖状态',
    dataIndex: 'saleStatus',
    key: 'saleStatus',
    width: 120,
  },
  {
    title: '是否有体验价',
    dataIndex: 'hasExperiencePrice',
    key: 'hasExperiencePrice',
    width: 120,
  },
  {
    title: '是否开启微校售卖',
    dataIndex: 'onlineSale',
    key: 'onlineSale',
    width: 150,
  },
  // 是否开启微校展示
  {
    title: '是否开启微校展示',
    dataIndex: 'isShowMicoSchool',
    key: 'isShowMicoSchool',
    width: 150,
  },
  {
    title: '报价单数量',
    dataIndex: 'quoteCount',
    key: 'quoteCount',
    width: 120,
  },
  {
    title: '总销量',
    dataIndex: 'saleVolume',
    key: 'saleVolume',
    width: 120,
    sorter: true,
  },
  {
    title: '更新时间',
    key: 'updateTime',
    dataIndex: 'updateTime',
    width: 180,
  },

  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 180,
  },

])

const systemDefaultIsDisplayCodeList = ref([
  {
    title: '科目',
    key: 'subject',
    dataIndex: 'subject',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '学季',
    dataIndex: 'term',
    key: 'term',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '学年',
    dataIndex: 'schoolYear',
    key: 'schoolYear',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '年级',
    key: 'singleGrade',
    dataIndex: 'singleGrade',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '班型',
    key: 'classType',
    dataIndex: 'classType',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '课程属性',
    key: 'courseAttribute',
    dataIndex: 'courseAttribute',
    width: 120,
    show: false,
    isDynamic: true,
  },
])
const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()
// Add watch effect for enabledCourseProperties
// 控制显示自定义字段和列的逻辑
watch(enabledCourseProperties, (newList) => {
  // Update show field based on enabledCourseProperties
  systemDefaultIsDisplayCodeList.value.forEach((item) => {
    const matchingField = newList.find(field => field.name === item.title)

    if (matchingField) {
      item.show = matchingField.enable
    }
  })

  updateDynamicColumns()
}, { deep: true })

// 新增一个函数来统一处理动态列的更新
function updateDynamicColumns() {
  // 保存操作列和倒计时列
  const actionColumn = allColumns.value.find(item => item.key === 'action')

  // 获取所有非动态、非操作、非倒计时的基础列
  const baseColumns = allColumns.value.filter(col =>
    !col.isDynamic
    && col.key !== 'action',
  )

  // 合并所有需要显示的动态列
  const visibleSystemColumns = systemDefaultIsDisplayCodeList.value.filter(item => item.show)
  const allDynamicColumns = [...visibleSystemColumns]

  // 重新组装列顺序：基础列 -> 动态列 -> 倒计时列 -> 操作列
  allColumns.value = [
    ...baseColumns,
    ...allDynamicColumns,
  ]

  // 添加操作列到最后
  if (actionColumn) {
    allColumns.value.push(actionColumn)
  }
}

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'course-list-record', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
    defaultSelectedKeys: ['name', 'categoryName', 'courseType', 'teachingMethod', 'chargingMethod', 'saleStatus', 'isShowMicoSchool', 'hasExperiencePrice', 'onlineSale', 'quoteCount', 'saleVolume', 'updateTime'],
  })

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (selectedRowKeysParam, rows) => {
    // console.log(`selectedRowKeys: ${selectedRowKeysParam}`, 'selectedRows: ', rows);
    selectedRows.value = rows
    selectedRowKeys.value = selectedRowKeysParam
    // console.log(selectedRows.value);
  },
}))

const loading = ref(false)

// 存储所有查询条件的响应式对象
const queryState = ref({
  'delFlag': false,
  'courseCategory': undefined,
  'commonCourse': undefined,
  'teachMethod': undefined,
  'chargeTypes': undefined,
  'saleStatus': undefined,
  'lessonAudition': undefined,
  'isOpenMicroSchoolBuy': undefined,
  'isShowMicroSchool': undefined,
  'isOnlineClassSelection': undefined,
  'courseAttribute': undefined,
  'term': undefined,
  'schoolYear': undefined,
  'courseIds': undefined,
  'searchKey': undefined,
})

// 重置所有查询条件
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    if (key !== 'delFlag') {
      queryState.value[key] = undefined
    }
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
      queryState.value[key] = value
    })
  }

  pagination.value.current = 1
  getCoursePage(queryState.value, id, type)
}, 300, { leading: true, trailing: false })

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['5', '10', '20', '50'],
  hideOnSinglePage: true,
  showQuickJumper: true,
})

// 排序状态管理
const sortModel = ref({
  byUpdateTime: 0,
  byTotalSales: 0,

})
// 获取课程列表
async function getCoursePage(newQueryParams = {}, id, type) {
  console.log('newQueryParams: ', id, type)
  loading.value = true
  try {
    const res = await getCoursePageApi({
      'pageRequestModel': {
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
      },
      'sortModel': sortModel.value,
      'queryModel': {
        ...queryState.value, // 展开所有有效的查询条件
      },
    })
    dataSource.value = res.result || []
    pagination.value.total = res.total

    // 添加调试信息，帮助确认数据结构
    // if (dataSource.value.length > 0) {
    //   console.log('课程数据示例:', dataSource.value[0]);
    //   console.log('课程属性字段:', {
    //     courseProperties: dataSource.value[0].courseProperties,
    //     customInfo: dataSource.value[0].customInfo,
    //     properties: dataSource.value[0].properties
    //   });
    // }

    allFilterRef.value.clearQuickFilter(id, type)
  }
  catch (error) {
    console.log('error: ', error)
  }
  finally {
    loading.value = false
  }
}

// 处理表格排序变化
function handleTableChange(paginationInfo, filters, sorter) {
  // console.log('排序变化:', sorter);

  // 重置所有排序字段为0
  Object.keys(sortModel.value).forEach((key) => {
    sortModel.value[key] = 0
  })

  // 处理排序逻辑（支持单列排序和多列排序）
  const sortFieldMap = {
    'saleVolume': 'byTotalSales',
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
  getCoursePage()
}

onUnmounted(() => {
  // 组件卸载时移除事件监听
  emitter.off(EVENTS.REFRESH_STUDENT_LIST, getCoursePage)
})

onMounted(() => {
  getCoursePage()
  getEnabledCourseProperties()
  // 监听刷新列表事件
  emitter.on(EVENTS.REFRESH_STUDENT_LIST, getCoursePage)
})
const allFilterRef = ref(null)

// 过滤器字段映射
const filterFieldMapping = {
  courseCategoryFilter: 'courseCategory',
  teachingMethodFilter: 'teachMethod',
  chargingMethodFilter: 'chargeTypes',
  hasTrialPriceFilter: 'lessonAudition',
  isMicroSchoolSaleFilter: 'isOpenMicroSchoolBuy',
  isMicroSchoolDisplayFilter: 'isShowMicroSchool',
}

const courseAttributeFieldMapping = {
  学季: 'term',
  学年: 'schoolYear',
  课程属性: 'courseAttribute',
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })
  handlers['update:courseAttributeFilter'] = (payload, isClearAll, id, type) => {
    const fieldName = courseAttributeFieldMapping[payload?.itemName] || courseAttributeFieldMapping[enabledCourseProperties.value.find(item => item.id === Number(payload?.itemId))?.name]
    if (!fieldName) return
    const value = typeof payload?.value === 'object' && payload?.value !== null ? payload.value.id : payload?.value
    handleFilterUpdate({ [fieldName]: value }, isClearAll, id, type)
  }
  return handlers
})
const openDrawer = ref(false)

// 创建课程
function handleCreateCourse() {
  console.log('创建课程')
  openDrawer.value = true
  modalType.value = 'create'
}
async function handleSubmit(formState) {
  console.log('formState: ', formState)
  if (formState.courseCategory == undefined) {
    formState.courseCategory = ''
  }
  try {
    let res
    if (modalType.value === 'edit') {
      // 编辑课程
      res = await updateCourseApi({
        ...formState,
        id: editCourseId.value,
      })
      console.log('edit res: ', res)
      if (res.code === 200) {
        messageService.success('修改课程成功')
        openDrawer.value = false
        getCoursePage()
      }
    }
    else {
      // 创建课程
      res = await addCourseApi(formState)
      console.log('create res: ', res)
      if (res.code === 200) {
        messageService.success('创建课程成功')
        openDrawer.value = false
        getCoursePage()
      }
    }
  }
  catch (error) {
    console.log('error: ', error)
  }
}
function handleDelCourse(record) {
  Modal.confirm({
    title: '确定删除？',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: '删除课程将放入回收站，同时学员将无法报读',
    onOk: async () => {
      try {
        const res = await batchDelOrResCourseApi({
          courseIds: [record.id],
          delFlag: true,
        })
        if (res.code === 200) {
          messageService.success('删除成功')
          getCoursePage()
        }
      }
      catch (err) {
        console.log(err)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onCancel() { },
  })
}

// 批量删除课程
function handleBatchDelCourse(selectedCourses) {
  const courseIds = selectedCourses.map(course => course.id)

  Modal.confirm({
    title: '确定批量删除？',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: `删除课程将放入回收站，同时学员将无法报读`,
    onOk: async () => {
      try {
        const res = await batchDelOrResCourseApi({
          courseIds,
          delFlag: true,
        })
        if (res.code === 200) {
          messageService.success(`成功删除 ${selectedCourses.length} 个课程`)
          selectedRows.value = [] // 清空选中状态
          selectedRowKeys.value = [] // 清空选中的key
          getCoursePage()
        }
      }
      catch (err) {
        console.log(err)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onCancel() { },
  })
}
// 批量售卖/停售
async function handleBatchSaleStatus(selectedCourses, isSale) {
  const courseIds = selectedCourses.map(course => course.id)
  const courseNames = selectedCourses.map(course => course.name).join('、')

  Modal.confirm({
    title: isSale ? '确定批量售卖？' : '确定批量停售？',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: isSale
      ? `课程开启售卖后，学员将能正常报读`
      : `课程停售后，学员将无法报读`,
    onOk: async () => {
      try {
        const res = await batchSaleStatusApi({
          courseIds,
          saleStatus: isSale,
        })
        if (res.code === 200) {
          messageService.success(`成功${isSale ? '售卖' : '停售'} ${selectedCourses.length} 个课程`)
          selectedRows.value = [] // 清空选中状态
          selectedRowKeys.value = [] // 清空选中的key
          getCoursePage()
        }
      }
      catch (err) {
        console.log(err)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onCancel() { },
  })
}
// 批量开启微校展示/关闭微校展示
async function handleBatchOpenMicroSchoolShow(selectedCourses, isShow) {
  const courseIds = selectedCourses.map(course => course.id)

  // 检查所选课程的当前状态
  if (isShow) {
    // 要开启微校展示，检查是否都已经开启
    const allAlreadyShow = selectedCourses.every(course => course.isShowMicoSchool)
    if (allAlreadyShow) {
      messageService.warning('当前所选课程均已开启微校展示，请重新选择')
      return
    }
  }
  else {
    // 要关闭微校展示，检查是否都已经关闭
    const allAlreadyHidden = selectedCourses.every(course => !course.isShowMicoSchool)
    if (allAlreadyHidden) {
      messageService.warning('当前所选课程均已关闭微校展示，请重新选择')
      return
    }
  }

  Modal.confirm({
    title: isShow ? '确定批量开启微校展示？' : '确定批量关闭微校展示？',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: isShow
      ? '开启后，课程将在微校中展示给学员'
      : '关闭后，课程将不在微校中展示',
    onOk: async () => {
      try {
        const res = await batchOpenMicroSchoolShowApi({
          courseIds,
          isShowMicoSchool: isShow,
        })
        if (res.code === 200) {
          messageService.success(`成功${isShow ? '开启' : '关闭'}微校展示 ${selectedCourses.length} 个课程`)
          selectedRows.value = [] // 清空选中状态
          selectedRowKeys.value = [] // 清空选中的key
          getCoursePage()
        }
      }
      catch (err) {
        console.log(err)
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    onCancel() { },
  })
}
function onClick({ key }) {
  switch (key) {
    case 1:
      // 批量删除
      if (selectedRows.value.length === 0) {
        messageService.warning('请选择要删除的课程')
        return
      }
      handleBatchDelCourse(selectedRows.value)
      break
    case 2:
      // 批量售卖
      if (selectedRows.value.length === 0) {
        messageService.warning('请选择要批量售卖的课程')
        return
      }
      handleBatchSaleStatus(selectedRows.value, true)
      break
    case 3:
      // 批量停售
      if (selectedRows.value.length === 0) {
        messageService.warning('请选择要批量停售的课程')
        return
      }
      handleBatchSaleStatus(selectedRows.value, false)
      break
    case 4:
      // 批量开启微校展示
      if (selectedRows.value.length === 0) {
        messageService.warning('请选择要批量开启微校展示的课程')
        return
      }
      handleBatchOpenMicroSchoolShow(selectedRows.value, true)
      break
    case 5:
      // 批量关闭微校展示
      if (selectedRows.value.length === 0) {
        messageService.warning('请选择要批量关闭微校展示的课程')
        return
      }
      handleBatchOpenMicroSchoolShow(selectedRows.value, false)
      break
    default:
      break
  }
}
const editCourseId = ref(null)
// 编辑按钮触发
function handleEditCourse(record) {
  console.log('record: ', record)
  openDrawer.value = true
  modalType.value = 'edit'
  editCourseId.value = record.id
}
// 复制按钮触发
function handleCopyCourse(record) {
  console.log('record: ', record)
  openDrawer.value = true
  modalType.value = 'copy'
  editCourseId.value = record.id
}
// 搜索课程列表（带防抖）
const handleSearchInput = debounce((value, id, type) => {
  console.log('搜索关键词: ', value)
  queryState.value.searchKey = value
  pagination.value.current = 1
  getCoursePage(queryState.value, id, type)
}, 10)
</script>

<template>
  <div class="tab-content">
    <all-filter ref="allFilterRef" :display-array="displayArray" :is-show-search-input="true" search-label="课程名称" search-placeholder="请输入课程名称"
      :render-class-list-options="false"  :course-attribute-list="enabledCourseProperties"
      v-on="filterUpdateHandlers" @searchInputFun="handleSearchInput" />
    <div class="tab-table">
      <div class="table-title flex justify-between items-center">
        <div class="total whitespace-nowrap">
          总计 {{ dataSource.length }} 条课程，5 条在售卖
        </div>
        <!-- 隐藏滚动条 -->
        <div class="edit ml10px flex overflow-x-auto">
          <a-button class="mr-2">
            已删除课程
          </a-button>
          <a-dropdown class="mr-2">
            <template #overlay>
              <a-menu>
                <a-menu-item key="1">
                  导入课程商品
                </a-menu-item>
                <a-menu-item key="2">
                  导出课程商品
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              导入/导出课程商品
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-dropdown class="mr-2">
            <template #overlay>
              <a-menu @click="onClick">
                <a-menu-item :key="1">
                  批量删除
                </a-menu-item>
                <a-menu-item :key="2">
                  批量售卖
                </a-menu-item>
                <a-menu-item :key="3">
                  批量停售
                </a-menu-item>
                <a-menu-item :key="4">
                  批量开启微校展示
                </a-menu-item>
                <a-menu-item :key="5">
                  批量关闭微校展示
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              批量操作
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-button type="primary" class="mr-2" @click="handleCreateCourse">
            创建课程
          </a-button>
          <!-- 自定义字段 -->
          <customize-code v-model:checked-values="selectedValues" :options="columnOptions"
            :total="allColumns.length - 1" :num="selectedValues.length" />
        </div>
      </div>
      <div class="table-content mt-2">
        <a-table :data-source="dataSource" row-key="id" :loading="loading" :pagination="pagination"
          :columns="filteredColumns" :row-selection="rowSelection" :scroll="{ x: totalWidth }"
          @change="handleTableChange">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <clamped-text :text="record.name" :lines="1" />
            </template>
            <template v-if="column.key === 'categoryName'">
              {{ record.categoryName || '-' }}
            </template>
            <template v-if="column.key === 'courseType'">
              <span v-if="record.courseType === 1">不通用</span>
              <span v-if="record.courseType === 2">全部通用</span>
              <span v-if="record.courseType === 3">部分通用</span>
              <span v-if="record.courseType === 4">混合课程通用</span>
            </template>
            <template v-if="column.key === 'teachingMethod'">
              <span v-if="record.teachMethod === 1">班级授课</span>
              <span v-if="record.teachMethod === 2">1v1授课</span>
            </template>
            <template v-if="column.key === 'chargingMethod'">
              {{ record.chargeMethods }}
            </template>
            <template v-if="column.key === 'saleStatus'">
              <span v-if="record.saleStatus" class="bg-#e6ffec text-#0c3 text-3 px3 py0.8 rounded-10">在售</span>
              <span v-if="!record.saleStatus" class="bg-#eee text-#888 text-3 px3 py0.8 rounded-10">停售</span>
            </template>
            <template v-if="column.key === 'hasExperiencePrice'">
              <div v-if="!record.hasExperiencePrice" class="studentStatus">
                <a-badge status="default" />
                <span class="text-#999">否</span>
              </div>
              <div v-if="record.hasExperiencePrice" class="studentStatus">
                <a-badge status="default" color="#0c3" />
                <span class="text-#666">是</span>
              </div>
            </template>
            <template v-if="column.key === 'onlineSale'">
              <div v-if="!record.onlineSale" class="studentStatus">
                <a-badge status="default" />
                <span class="text-#999">否</span>
              </div>
              <div v-if="record.onlineSale" class="studentStatus">
                <a-badge status="default" color="#0c3" />
                <span class="text-#666">是</span>
              </div>
            </template>
            <template v-if="column.key === 'isShowMicoSchool'">
              <div v-if="!record.isShowMicoSchool" class="studentStatus">
                <a-badge status="default" />
                <span class="text-#999">否</span>
              </div>
              <div v-if="record.isShowMicoSchool" class="studentStatus">
                <a-badge status="default" color="#0c3" />
                <span class="text-#666">是</span>
              </div>
            </template>
            <template v-if="column.key === 'quoteCount'">
              {{ record.quoteCount }}个
            </template>
            <template v-if="column.key === 'saleVolume'">
              {{ record.saleVolume }}
            </template>
            <template v-if="column.key === 'updateTime'">
              {{ dayjs(record.updateTime).format('YYYY-MM-DD HH:mm') || '-' }}
            </template>
            <!-- <template v-if="column.key === 'subject'"> - </template> -->
            <template v-if="enabledCourseProperties.some(item => item.name === column.title)">
              <clamped-text :text="(() => {
                // 根据列标题匹配课程属性数据
                const coursePropertiesArray = record.courseProductProperties;

                if (Array.isArray(coursePropertiesArray)) {
                  const propertyData = coursePropertiesArray.find(item =>
                    item.coursePropertyName === column.title,
                  );
                  return propertyData?.coursePropertyOptionName || '-';
                }

                return '-';
              })()" />
            </template>
            <template v-else-if="column.key === 'action'">
              <span class="flex action">
                <a class="mr-3" @click="handleEditCourse(record)">编辑</a>
                <a class="mr-3" @click="handleCopyCourse(record)">复制</a>
                <a @click="handleDelCourse(record)">删除</a>
              </span>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
  <create-course-drawer v-model:open="openDrawer" :modal-type="modalType" :edit-course-id="editCourseId"
    @handle-submit="handleSubmit" />
</template>

<style lang="less" scoped>
.tab-content {
  margin: 8px 0;

  .tab-table {
    background: #fff;
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;

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

    .intention {
      display: flex;
      align-items: center;

      .statusTag {
        width: 54px;
        height: 22px;
        border-radius: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        font-family: PingFangSC-Regular, PingFang SC, sans-serif;
        font-weight: 600;
      }
    }

    .intentionTag {
      display: inline-block;
      width: 6px;
      height: 6px;
      border-radius: 100px;
      margin-right: 3px;
    }

    .action {
      a {
        color: var(--pro-ant-color-primary);
        ;
      }
    }
  }

}

.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
