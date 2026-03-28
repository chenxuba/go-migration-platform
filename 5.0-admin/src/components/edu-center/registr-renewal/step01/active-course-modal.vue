<script setup>
import { computed, ref, watch, onUnmounted } from 'vue'
import { InfoCircleOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { getProcessContentPageApi } from '~@/api/edu-center/registr-renewal'
import { Empty } from 'ant-design-vue'
// Props
const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  // 已选中的课程列表（用于重新打开弹窗时初始化选中状态）
  selectedCourses: {
    type: Array,
    default: () => [],
  },
})
// Emits
const emit = defineEmits(['update:open', 'confirm'])

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 数据
const displayArray = ref(['courseCategory'])
const activeIndex = ref(0)
// 改进：为每个商品类型维护独立的选中状态（存储完整的课程信息）
const activeCourseByType = ref({
  1: [], // 课程商品
  2: [], // 学杂费
  3: [], // 教材商品
})
// 计算当前类型的选中课程索引（用于界面显示）
const activeCourse = computed(() => {
  const selectedCourses = activeCourseByType.value[productType.value] || []
  const selectedIds = selectedCourses.map(course => course.id)
  return courseList.value
    .map((item, index) => selectedIds.includes(item.id) ? index : -1)
    .filter(index => index !== -1)
})
const productType = ref(1)
const selectItems = ref([
  { title: '课程商品', productType: 1 },
  { title: '学杂费', productType: 2 },
  { title: '教材用品', productType: 3 },
])
const searchKey = ref(null)
// 原始课程列表数据
const allCourseList = ref([])

// 分页相关状态
const pagination = ref({
  pageIndex: 1,
  pageSize: 10,
  total: 0,
  hasMore: true,
})

// 加载状态
const loading = ref(false)
const loadingMore = ref(false)

// 列表容器引用
const listContainerRef = ref()

// 课程列表（直接使用 allCourseList，搜索由服务器端处理）
const courseList = computed(() => {
  return allCourseList.value
})

// 计算每个商品类型的选中数量
function getSelectedCountByType(type) {
  return activeCourseByType.value[type]?.length || 0
}

// 获取课程标签列表
function getCourseTagList(item) {
  const tags = []

  // 通用课标签（蓝色背景）
  if (item.courseType === 2 || item.courseType === 3) {
    tags.push({
      text: '通用课',
      color: '#0066ff',
      textColor: '#fff',
      type: 'primary',
    })
  }

  // 授课类型标签
  if (item.teachMethod === 1) {
    tags.push({
      text: '班级授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  if (item.teachMethod === 2) {
    tags.push({
      text: '1v1授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  // 课程范围标签
  if (item.courseType === 2) {
    tags.push({
      text: '全部课程',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  if (item.courseType === 3) {
    tags.push({
      text: '部分课程',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'tooltip',
      tooltipTitle: '初级言语课、高级认知课',
    })
  }



  tags.push({
    text: item.chargeMethods,
    color: '#e6f0ff',
    textColor: '#0066ff',
    type: 'normal',
  })

  // 课程分类标签
  if (item.productCategoryName) {
    tags.push({
      text: item.productCategoryName,
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  // 课程属性标签
  if (item.properties && item.properties.length > 0) {
    item.properties.forEach((property) => {
      tags.push({
        text: property.lessonPropertyOptionName,
        color: '#e6f0ff',
        textColor: '#0066ff',
        type: 'normal',
        key: property.lessonPropertyOptionId,
      })
    })
  }
  // 是否有体验价
  if (item.hasExperiencePrice) {
    tags.push({
      text: '体验价',
      color: '#fff5e6',
      textColor: '#ff9900',
      type: 'normal',
    })
  }

  return tags
}

// 通用标签样式
function getTagStyle(type = 'normal') {
  const baseStyle = {
    borderRadius: '20px',
    marginRight: '0',
    height: '20px',
  }

  if (type === 'primary') {
    return {
      ...baseStyle,
      color: '#fff',
    }
  }

  return {
    ...baseStyle,
    color: '#0066ff',
  }
}

// 重置分页状态
function resetPagination() {
  pagination.value = {
    pageIndex: 1,
    pageSize: 10,
    total: 0,
    hasMore: true,
  }
  allCourseList.value = []
}

// 监听搜索关键词变化
watch(searchKey, () => {
  clearTimeout(searchTimeout.value)
  searchTimeout.value = setTimeout(() => {
    resetPagination()
    getProcessContentPage(true)
  }, 500)
})

// 搜索防抖定时器
const searchTimeout = ref(null)

// 初始化已选中的课程
function initializeSelectedCourses() {
  if (props.selectedCourses && props.selectedCourses.length > 0) {
    // 按商品类型分组已选中的课程
    const coursesByType = {
      1: [],
      2: [],
      3: [],
    }
    
    props.selectedCourses.forEach(course => {
      const type = course.productType || 1
      if (coursesByType[type]) {
        coursesByType[type].push(course)
      }
    })
    
    // 更新选中状态
    activeCourseByType.value = coursesByType
    console.log('初始化选中课程:', coursesByType)
  } else {
    // 清空选中状态
    activeCourseByType.value = {
      1: [],
      2: [],
      3: [],
    }
  }

}

watch(openModal, (newValue, oldValue) => {
  if (newValue) {
    // 重置状态
    activeIndex.value = 0
    productType.value = 1 // 重置为课程商品
    searchKey.value = null

    // 初始化已选中的课程
    initializeSelectedCourses()

    // 获取当前类型的课程列表
    getCourseList(productType.value)
  }
})

function getCourseList(type) {
  // 重置分页状态
  resetPagination()
  getProcessContentPage(true)
}

// 获取课程列表
async function getProcessContentPage(isRefresh = false) {
  if (loading.value || loadingMore.value) return
  
  // 如果没有更多数据且不是刷新操作，直接返回
  if (!pagination.value.hasMore && !isRefresh) return
  
  try {
    if (isRefresh) {
      loading.value = true
      resetPagination()
    } else {
      loadingMore.value = true
    }

    const params = {
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.pageIndex,
        skipCount: (pagination.value.pageIndex - 1) * pagination.value.pageSize
      },
      queryModel: {
        delFlag: false,
        productType: productType.value,
        ...(searchKey.value && { searchKey: searchKey.value }),
        saleStatus: true, // 只获取可销售的商品
      },
      sortModel: {}
    }

    const res = await getProcessContentPageApi(params)

    if (res.code === 200) {
      const { result = [], total = 0 } = res || {}
      
      if (isRefresh) {
        // 刷新时替换数据
        allCourseList.value = Array.isArray(result) ? result : []
      } else {
        // 加载更多时追加数据
        allCourseList.value.push(...(Array.isArray(result) ? result : []))
      }
      
      // 更新分页信息
      pagination.value.total = total
      pagination.value.hasMore = allCourseList.value.length < total
      
      // 如果还有更多数据，增加页码
      if (pagination.value.hasMore) {
        pagination.value.pageIndex++
      }
      
    }
  } catch (error) {
    console.error('获取课程列表失败:', error)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// 滚动事件处理
function handleScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  
  // 当滚动到底部附近时加载更多数据
  if (scrollHeight - scrollTop - clientHeight < 50) {
    loadMore()
  }
}

// 加载更多数据
function loadMore() {
  if (!pagination.value.hasMore || loading.value || loadingMore.value) return
  
  getProcessContentPage(false)
}

// 切换课程选中状态
function changeSelectCourse(index) {
  const course = courseList.value[index]
  if (!course) return

  const currentSelection = activeCourseByType.value[productType.value]
  const existingIndex = currentSelection.findIndex(item => item.id === course.id)

  if (existingIndex > -1) {
    // 如果已选中，则取消选择
    currentSelection.splice(existingIndex, 1)
  }
  else {
    // 如果未选中，则添加到选中列表（存储完整的课程信息）
    currentSelection.push({
      ...course,
      productType: productType.value,
    })
  }

  // 触发响应式更新
  activeCourseByType.value = { ...activeCourseByType.value }
}

function handleOk() {
  // 收集所有类型的选中课程
  const allSelectedCourses = []

  // 为每个商品类型收集选中的课程
  Object.keys(activeCourseByType.value).forEach((type) => {
    const typeNum = Number.parseInt(type)
    const selectedCourses = activeCourseByType.value[typeNum]
    if (selectedCourses && selectedCourses.length > 0) {
      // 直接使用缓存的完整课程信息
      allSelectedCourses.push(...selectedCourses)
    }
  })

  emit('confirm', allSelectedCourses)
  emit('update:open', false)
}

function changeSelectItems(item, index) {
  productType.value = item.productType
  activeIndex.value = index
  searchKey.value = null // 切换类型时重置搜索
  getCourseList(productType.value)
}

function handleCancel() {
  emit('update:open', false)
}

// 取消选中课程（由父组件调用）
function cancelCourseSelection(course) {
  if (!course)
    return

  // 获取课程类型
  const type = course.productType || 1

  // 直接从选中列表中移除课程
  const selectedCourses = activeCourseByType.value[type] || []
  const existingIndex = selectedCourses.findIndex(item => item.id === course.id)
  
  if (existingIndex > -1) {
    selectedCourses.splice(existingIndex, 1)
    // 触发响应式更新
    activeCourseByType.value = { ...activeCourseByType.value }
    console.log(`已取消选中课程: ${course.name || course.id}`)
  } else {
    console.warn(`课程 ${course.name || course.id} 未在选中列表中找到`)
  }

}

// 组件销毁时清理定时器
onUnmounted(() => {
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }
})

// 暴露方法给父组件调用
defineExpose({
  cancelCourseSelection,
})
</script>

<template>
  <!-- 选择课程/学杂费/教材用品 -->
  <a-modal :open="openModal" centered wrap-class-name="modal" width="1000px" :body-style="{ padding: 0 }"
    title="选择课程/学杂费/教材用品" @ok="handleOk" @cancel="handleCancel" @update:open="$emit('update:open', $event)">
    <div class="modal-wrap">
      <div class="modal-left">
        <a-list>
          <a-list-item v-for="(item, index) in selectItems" :key="index" :class="{ active: activeIndex === index }"
            @click="changeSelectItems(item, index)">
            <custom-title v-if="activeIndex === index" :title="item.title" font-size="16px"
              :font-weight="activeIndex === index ? '500' : '300'" />
            <span v-else class="pl-2.5">{{ item.title }}</span>
            <a-badge :count="getSelectedCountByType(item.productType)">
              <span class="w-0 h-0" />
            </a-badge>
          </a-list-item>
        </a-list>
      </div>
      <div class="modal-right">
        <div v-if="productType === 1" class="m-r-t m-r-t-2">
          <a-input allow-clear v-model:value="searchKey" placeholder="请输入课程名称">
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>
        <div v-if="productType === 2" class="m-r-t m-r-t-2">
          <a-input v-model:value="searchKey" placeholder="搜索学杂费">
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>
        <div v-if="productType === 3" class="m-r-t m-r-t-2">
          <a-input v-model:value="searchKey" placeholder="搜索教材用品">
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>
        <div ref="listContainerRef" class="m-r-b" @scroll="handleScroll">
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-container">
            <a-spin tip="加载中..." />
          </div>
          
          <!-- 课程列表 -->
          <a-list v-else>
            <a-list-item v-for="(item, index) in courseList" :key="item.id || index"
              :class="activeCourse.includes(index) ? 'activeCourse' : ''"
              class="flex flex-items-center justify-between r-item" @click="changeSelectCourse(index)">
              <div class="m-r-b-l pt-1 pb-1">
                <div class="text-4 text-#222 font-500 mb-1">
                  {{ item.name }}
                </div>
                <a-space :size="5" class="w-100% flex flex-wrap">
                  <template v-for="tag in getCourseTagList(item)" :key="tag.key || tag.text">
                    <a-tag v-if="tag.type === 'tooltip'" :style="getTagStyle(tag.type)" :color="tag.color">
                      {{ tag.text }}
                      <a-tooltip>
                        <template #title>
                          {{ tag.tooltipTitle }}
                        </template>
                        <InfoCircleOutlined class="ml-1" />
                      </a-tooltip>
                    </a-tag>
                    <a-tag v-else :style="getTagStyle(tag.type)" :color="tag.color">
                      <span class="text-#ff9900" v-if="tag.text === '体验价'">{{ tag.text }}</span>
                      <span v-else>{{ tag.text }}</span>
                    </a-tag>
                  </template>
                </a-space>
              </div>
              <div class="m-r-b-r pt-1 pb-1 select ml-70px whitespace-nowrap">
                <a v-if="activeCourse.includes(index)" class="active-a">取消选择</a>
                <a v-if="!activeCourse.includes(index)">点击选择</a>
              </div>
            </a-list-item>
            
            <!-- 加载更多状态 -->
            <div v-if="loadingMore" class="load-more-container">
              <a-spin size="small" />
              <span class="ml-2">加载更多...</span>
            </div>
            
            <!-- 没有更多数据提示 -->
            <div v-else-if="!pagination.hasMore && allCourseList.length > 0" class="no-more-container">
              <span>已加载全部数据</span>
            </div>
            
            <!-- 空数据提示 -->
            <div v-else-if="!loading && allCourseList.length === 0" class="empty-container">
              <a-empty description="暂无数据" :image="Empty.PRESENTED_IMAGE_SIMPLE" />
            </div>
          </a-list>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.r-item {
  border-bottom: 1px solid #eee !important;
  cursor: pointer;
}

.modal-wrap {
  display: flex;
  height: 70vh;

  .modal-left {
    width: 160px;

    .ant-list-item {
      border: none;
      font-size: 16px;
      color: #666;
      padding-left: 14px;
      cursor: pointer;
      line-height: 2.5;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .num {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        background: #f33;
        border-radius: 100px;
        color: #fff;
        font-size: 12px;
      }
    }

    .active.ant-list-item {
      background: #f2f8ff;

      .title {
        color: var(--pro-ant-color-primary);
      }
    }
  }

  .modal-right {
    flex: 1;
    border-left: 1px solid #eee;

    .m-r-t {
      display: flex;
      align-items: center;
      padding: 0 12px;
      border-bottom: 1px solid #eee;
    }

    .m-r-t-2 {
      padding: 12px;
    }

    .m-r-b {
      height: calc(100% - 70px);
      overflow: auto;
      
      .loading-container {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 200px;
        flex-direction: column;
      }
      
      .load-more-container {
        display: flex;
        justify-content: center;
        align-items: center;
        padding: 16px;
        color: #666;
        font-size: 14px;
      }
      
      .no-more-container {
        display: flex;
        justify-content: center;
        align-items: center;
        padding: 16px;
        color: #999;
        font-size: 12px;
      }
      
      .empty-container {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 200px;
      }
    }
  }

  .activeCourse {
    background: #f2f8ff;
  }

  .select {
    a {
      color: var(--pro-ant-color-primary);
    }
  }

  .active-a {
    position: relative;
    color: #666 !important;

    &::before {
      display: inline-block;
      position: absolute;
      content: "✓";
      font-size: 16px;
      line-height: 20px;
      top: -2px;
      left: -30px;
      color: var(--pro-ant-color-primary);
      font-family: "Franklin Gothic Medium", "Arial Narrow", Arial, sans-serif;
    }
  }
}
</style>

<style lang="less">
.modal {
  .ant-modal-header {
    margin-bottom: 0;
  }

  .ant-modal-footer {
    margin-top: 0;
  }
}
</style>
