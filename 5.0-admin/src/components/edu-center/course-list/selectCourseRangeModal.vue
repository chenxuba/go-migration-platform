<script setup>
import { CloseOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { getCoursePageApi } from '~@/api/edu-center/course-list'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  // 用于回显的已选课程数据
  selectedCourses: {
    type: Array,
    default: () => [],
  },
  /** 为 true 时回显课程不可在右侧删除（默认）；创建组合课程等场景可设为 true 允许删除 */
  echoCoursesDeletable: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open', 'confirm'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 数据相关
const targetKeys = ref([])
const selectedKeys = ref([])
const courseList = ref([])
const loading = ref(false)
const searchValue = ref('')

// 分页相关
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  hasMore: true,
})

// 初始化数据方法
function initData() {
  targetKeys.value = []
  selectedKeys.value = []
  selectedOrder.value = []
  courseList.value = []
  pagination.value = {
    current: 1,
    pageSize: 20,
    total: 0,
    hasMore: true,
  }
  searchValue.value = ''
  leftSearchValue.value = ''
  rightSearchValue.value = ''
}

// 监听modal打开
watch(() => openModal.value, (newVal) => {
  if (newVal) {
    initData()
    getCourseList().then(() => {
      // 课程列表加载完成后设置回显数据
      setSelectedCourses()
    })
  }
})

// 设置回显的已选课程
function setSelectedCourses() {
  if (props.selectedCourses && props.selectedCourses.length > 0) {
    // 根据传入的课程数据设置targetKeys
    const selectedKeys = props.selectedCourses.map(course => course.id ? course.id.toString() : course.key)
    targetKeys.value = [...selectedKeys]

    const lockEcho = !props.echoCoursesDeletable

    // 如果传入的课程数据中有课程不在当前列表中，需要添加到courseList中
    props.selectedCourses.forEach(course => {
      const courseKey = course.id ? course.id.toString() : course.key
      const existsInList = courseList.value.some(item => item.key === courseKey)

      if (!existsInList) {
        const courseData = {
          ...course,
          id: course.id || course.key,
          key: courseKey,
          title: course.name || course.title,
        }
        courseData.disabled = lockEcho
        courseList.value.unshift(courseData)
      } else {
        const existingCourse = courseList.value.find(item => item.key === courseKey)
        if (existingCourse) {
          existingCourse.disabled = lockEcho
        }
      }
    })
  }
}

// 获取课程列表
async function getCourseList(isLoadMore = false) {
  if (loading.value) return Promise.resolve()

  loading.value = true
  
  try {
    const pageIndex = isLoadMore ? pagination.value.current + 1 : 1
    
    const res = await getCoursePageApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pageIndex,
      },
      sortModel: {
        byTotalSales: 0,
        byUpdateTime: 0,
      },
      queryModel: {
        searchKey: searchValue.value,
        delFlag: false, // 只获取在售课程
        saleStatus: 1, // 只获取售卖中的课程
        teachMethod: 1, // 只筛选班课
        courseType: 1, // 只筛选班课
      },
    })

    if (res.code === 200) {
      const newData = (res.result || []).map(item => ({
        id: item.id,
        key: item.id.toString(),
        title: item.name,
        ...item,
      }))

      if (isLoadMore) {
        // 加载更多时追加数据
        courseList.value = [...courseList.value, ...newData]
        pagination.value.current = pageIndex
      } else {
        // 首次加载或搜索时替换数据
        courseList.value = newData
        pagination.value.current = 1
      }

      pagination.value.total = res.total || 0
      pagination.value.hasMore = courseList.value.length < pagination.value.total
    }
  } catch (error) {
    console.error('获取课程列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 搜索功能
const leftSearchValue = ref('')
const rightSearchValue = ref('')

// 左侧搜索 - 调用API
const handleLeftSearch = debounce((value) => {
  leftSearchValue.value = value
  searchValue.value = value
  getCourseList(false)
}, 500)

// 右侧搜索 - 本地筛选
const handleRightSearch = (value) => {
  rightSearchValue.value = value
}

// 统一搜索处理函数
const handleSearch = (direction, value) => {
  if (direction === 'left') {
    handleLeftSearch(value)
  } else if (direction === 'right') {
    handleRightSearch(value)
  }
}

// 触底加载更多
function loadMoreCourses() {
  if (!loading.value && pagination.value.hasMore) {
    getCourseList(true)
  }
}

// 处理滚动到底部
function handleScroll(e) {
  const { target } = e
  if (target.scrollTop + target.offsetHeight >= target.scrollHeight - 10) {
    loadMoreCourses()
  }
}

const disabled = ref(false)

// 左侧搜索过滤（API搜索不需要本地过滤）
function filterOption(inputValue, option, direction) {
  if (direction === 'left') {
    // 左侧由API处理搜索，这里不做过滤
    return true
  } else if (direction === 'right') {
    // 右侧本地搜索过滤
    return option.title.toLowerCase().includes(inputValue.toLowerCase())
  }
  return option.title.includes(inputValue)
}

// 用于记录选中顺序
const selectedOrder = ref([])

function handleChange(keys) {
  if (keys.length > selectedKeys.value.length) {
    // 新增选中
    const newKey = keys.find(key => !selectedKeys.value.includes(key))
    if (newKey) {
      selectedOrder.value.push(newKey)
    }
  }
  else {
    // 取消选中
    const removedKey = selectedKeys.value.find(key => !keys.includes(key))
    if (removedKey) {
      selectedOrder.value = selectedOrder.value.filter(key => key !== removedKey)
    }
  }
  selectedKeys.value = [...keys]
}

// 排序函数
function sortBySelection(items) {
  return [...items].sort((a, b) => {
    const aSelected = selectedKeys.value.includes(a.key)
    const bSelected = selectedKeys.value.includes(b.key)

    if (aSelected && bSelected) {
      return selectedOrder.value.indexOf(a.key) - selectedOrder.value.indexOf(b.key)
    }
    if (aSelected)
      return -1
    if (bSelected)
      return 1
    return courseList.value.findIndex(item => item.key === a.key) - courseList.value.findIndex(item => item.key === b.key)
  })
}

// 右侧本地筛选的课程列表
const filteredCourseList = computed(() => {
  return courseList.value
})

// 筛选右侧已选课程（本地筛选）
const filteredTargetCourses = computed(() => {
  if (!rightSearchValue.value) {
    return courseList.value.filter(item => targetKeys.value.includes(item.key))
  }
  
  return courseList.value.filter(item => 
    targetKeys.value.includes(item.key) && 
    item.title.toLowerCase().includes(rightSearchValue.value.toLowerCase())
  )
})

// 计算属性用于排序后的列表
const sortedFilteredItems = computed(() => {
  return (direction, filteredItems) => {
    if (direction === 'left') {
      return sortBySelection(filteredItems)
    } else if (direction === 'right') {
      // 右侧使用本地筛选后的数据
      return filteredTargetCourses.value
    }
    return filteredItems
  }
})

function handleTransferChange(nextTargetKeys, direction, moveKeys) {
  targetKeys.value = nextTargetKeys
}

function handleDel(item) {
  if (item.disabled) {
    return
  }
  targetKeys.value = targetKeys.value.filter(key => key !== item.key)
}

// 手动触发验证
function handleSubmit() {
  // 获取选中的课程数据
  const selectedCourses = targetKeys.value.map(key => {
    return courseList.value.find(course => course.key === key)
  }).filter(Boolean)
  
  // 向父组件传递选中的课程数据
  emit('confirm', selectedCourses)
  
  console.log('选中的课程:', selectedCourses)
  
  closeFun()
}

function closeFun() {
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>选择课程</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-transfer
        v-model:target-keys="targetKeys" v-model:selected-keys="selectedKeys" :data-source="filteredCourseList"
        :one-way="true" :show-search="true" :filter-option="filterOption" :show-select-all="false"
        @change="handleTransferChange" @search="handleSearch"
      >
        <template #leftSelectAllLabel>
          <div v-if="selectedKeys.length === 0">
            {{ courseList.length }}个可选
            <span v-if="pagination.total > courseList.length">
              (共{{ pagination.total }}个)
            </span>
          </div>
          <div v-else>
            勾选：{{ selectedKeys.length }}
          </div>
        </template>
        <template #rightSelectAllLabel>
          <div>
            {{ rightSearchValue ? filteredTargetCourses.length : targetKeys.length }}个已选
            <span v-if="rightSearchValue && filteredTargetCourses.length !== targetKeys.length">
              (共{{ targetKeys.length }}个)
            </span>
          </div>
        </template>
        <template #children="{ direction, filteredItems }">
          <div v-if="direction === 'left'" class="course-list scrollbar" :class="{'flex justify-center items-center min-h-300px': loading}" @scroll="handleScroll">
            <a-spin :spinning="loading && pagination.current === 1" tip="加载中...">
              <div v-for="item in sortedFilteredItems(direction, filteredItems)" :key="item.key" class="course-item">
                <a-checkbox
                  :checked="selectedKeys.includes(item.key)"
                  @change="(e) => handleChange(e.target.checked ? [...selectedKeys, item.key] : selectedKeys.filter(key => key !== item.key))"
                >
                  <div class="course-info">
                    <div class="course-title">
                      {{ item.title }}
                    </div>
                    <a-space class="course-tags">
                      <span v-if="item.teachMethod" class="bg-#e6f0ff text-#06f text-3 px2 py2px rounded-10">
                        {{ item.teachMethod === 1 ? '班级授课' : item.teachMethod === 2 ? '1v1授课' : '其他授课' }}
                      </span>
                      <span v-if="item.chargeMethods" class="bg-#f0f9ff text-#1890ff text-3 px2 py2px rounded-10">
                        {{ item.chargeMethods }}
                      </span>
                      <span v-if="item.teachMethodName" class="bg-#e6f0ff text-#06f text-3 px2 py2px rounded-10">
                        {{ item.teachMethodName }}
                      </span>
                      <span v-if="item.hasExperiencePrice" class="bg-#fff5e6 text-#f90 text-3 px2 py2px rounded-10">
                        体验价
                      </span>
                    </a-space>
                  </div>
                </a-checkbox>
              </div>
              
              <!-- 加载更多提示 -->
              <div v-if="loading && pagination.current > 1" class="load-more-tip">
                <a-spin size="small" />
                <span class="ml-2">加载更多...</span>
              </div>
              
              <!-- 没有更多数据提示 -->
              <div v-else-if="!pagination.hasMore && courseList.length > 0" class="no-more-tip">
                没有更多数据了
              </div>
              
              <!-- 空状态 -->
              <div v-if="!loading && courseList.length === 0" class="empty-state">
                <a-empty description="暂无课程数据" />
              </div>
            </a-spin>
          </div>
          <div v-if="direction === 'right'" class="course-list">
            <div v-for="item in sortedFilteredItems(direction, filteredItems)" :key="item.key" class="course-item">
              <div class="course-info flex justify-between flex-center">
                <div>
                  <div class="course-title" :class="{ 'disabled-course': item.disabled }">
                    {{ item.title }}
                  </div>
                  <a-space class="course-tags">
                    <span v-if="item.teachMethod" class="bg-#e6f0ff text-#06f text-3 px2 py2px rounded-10">
                      {{ item.teachMethod === 1 ? '班级授课' : item.teachMethod === 2 ? '1v1授课' : '其他授课' }}
                    </span>
                    <span v-if="item.chargeMethods" class="bg-#f0f9ff text-#1890ff text-3 px2 py2px rounded-10">
                      {{ item.chargeMethods }}
                    </span>
                    <span v-if="item.teachMethodName" class="bg-#e6f0ff text-#06f text-3 px2 py2px rounded-10">
                      {{ item.teachMethodName }}
                    </span>
                    <span v-if="item.hasExperiencePrice" class="bg-#fff5e6 text-#f90 text-3 px2 py2px rounded-10">
                      体验价
                    </span>
                  </a-space>
                </div>
                <div>
                  <DeleteOutlined 
                    v-if="!item.disabled"
                    class="text-#06f cursor-pointer" 
                    @click="handleDel(item)" 
                  />
                  <DeleteOutlined 
                    v-else
                    class="text-#ccc cursor-not-allowed disabled-delete-icon" 
                    title="回显课程不可删除"
                  />
                </div>
              </div>
            </div>
          </div>
        </template>
      </a-transfer>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost @click="handleSubmit">
        确定
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
  display: flex;
  justify-content: center;

  :deep(.ant-transfer) {
    .ant-transfer-list {
      width: 336px;
      height: min(560px, calc(100vh - 200px));
      border: 1px solid #ddd;
      border-radius: 16px;
      overflow: hidden;
      background-color: #fafafa;
    }

    .ant-btn-icon-only {
      width: 40px;
      height: 40px;
      border-radius: 8px;
    }

    .ant-transfer-list-header {
      min-height: 54px;
      padding: 0 16px;
      line-height: 22px;
      border: none;
      font-size: 14px;
      font-weight: 500;
      color: #222;
      background-color: #fafafa;
    }

    .ant-transfer-list-body-search-wrapper {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 48px;
      padding: 0 16px;
      background-color: #fff;
      box-shadow: inset 0 -1px 0 0 #eee;
    }

    .course-list {
      background: #fff;
      max-height: min(460px, calc(100vh - 280px));
      overflow-y: auto;
      overflow-x: hidden;
    }

    .course-item {
      padding: 12px 16px;
      width: 100%;
      box-shadow: inset 0 -1px 0 0 #eee;

      .course-info {
        width: 300px;
      }

      .course-title {
        color: #666;
        margin-bottom: 8px;
        font-weight: 500;

        &.disabled-course {
          color: #999;
          position: relative;
        }
       
      }

      .disabled-delete-icon {
        opacity: 0.5;
        pointer-events: none;
      }
    }

    .load-more-tip {
      padding: 16px;
      text-align: center;
      color: #666;
      font-size: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .no-more-tip {
      padding: 16px;
      text-align: center;
      color: #999;
      font-size: 12px;
    }

    .empty-state {
      padding: 40px 16px;
      text-align: center;
    }
  }
}
</style>

<style>
.modal-content-box .ant-modal-footer {
  margin-top: 0 !important;
}

.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
