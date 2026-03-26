<script setup>
import { CheckOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import { createStudentFollowUpApi, getFollowUpRecordPagedApi, updateFollowRecordApi, updateVisitStatusApi } from '~@/api/enroll-center/intention-student'
import { FollowMethodLabel } from '@/enums'
import { messageService } from '@/utils/messageService'
import emitter, { EVENTS } from '~@/utils/eventBus'

const props = defineProps({
  studentId: {
    type: [Number, String],
    default: '',
  },
  studentDetail: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['refresh-student-detail'])
const spinning = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)

const list = ref([])
const openAddFollowUpModal = ref(false)
const isEdit = ref(false)
const editRecord = ref({})

// 哨兵元素引用，用于 IntersectionObserver
const sentinelRef = ref(null)
// IntersectionObserver 实例
let intersectionObserver = null

function handleAddFollowUp() {
  isEdit.value = false
  editRecord.value = {}
  openAddFollowUpModal.value = true
}

function handleEditFollowUp(item) {
  isEdit.value = true
  const editData = { ...item }
  // 后端分页返回的是 intendedCourse（id 数组）；旧版 intentionLessonList 若有则优先
  if (editData.intentionLessonList?.length) {
    editData.intentCourseIds = editData.intentionLessonList.map(lesson => String(lesson.lessonId))
  }
  else if (editData.intendedCourse?.length) {
    editData.intentCourseIds = editData.intendedCourse.map(id => String(id))
  }
  else {
    editData.intentCourseIds = []
  }
  editRecord.value = editData
  openAddFollowUpModal.value = true
}

/** 时间轴展示意向课程：优先接口返回的 intendedCourseName */
function formatFollowUpIntendedCourses(item) {
  if (item.intendedCourseName?.length)
    return item.intendedCourseName.join('、')
  if (item.intentionLessonList?.length)
    return item.intentionLessonList.map(l => l.lessonName).join('、')
  const ids = item.intendedCourse
  if (!ids?.length)
    return ''
  const lessons = props.studentDetail?.lessons || []
  return ids.map((id) => {
    const row = lessons.find(l => String(l.id) === String(id))
    return row?.name || String(id)
  }).join('、')
}

const handleMarkAsVisited = debounce(async (item) => {
  try {
    spinning.value = true
    const res = await updateVisitStatusApi({
      'id': item.id,
      'uuid': item.uuid,
      'version': item.version,
      'visitStatus': true,
    })
    if (res.code === 200) {
      getFollowUpRecord()
    }
  }
  catch (error) {
    console.log(error)
  }
}, 300, {
  leading: true,
  trailing: false,
})

const handleMarkAsUnvisited = debounce(async (item) => {
  try {
    spinning.value = true
    const res = await updateVisitStatusApi({
      'id': item.id,
      'uuid': item.uuid,
      'version': item.version,
      'visitStatus': false,
    })
    if (res.code === 200) {
      getFollowUpRecord()
    }
  }
  catch (error) {
    console.log(error)
  }
}, 300, {
  leading: true,
  trailing: false,
})

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

// 设置 IntersectionObserver 监听哨兵元素
function setupIntersectionObserver() {
  // 清理之前的观察器
  if (intersectionObserver) {
    intersectionObserver.disconnect()
  }

  // 重试机制，等待哨兵元素渲染完成
  const trySetupObserver = (retryCount = 0) => {
    if (!sentinelRef.value) {
      if (retryCount < 10) { // 最多重试10次
        console.log(`哨兵元素不存在，第${retryCount + 1}次重试...`)
        setTimeout(() => trySetupObserver(retryCount + 1), 100)
        return
      }
      else {
        console.log('哨兵元素不存在，停止重试')
        return
      }
    }

    console.log('设置 IntersectionObserver 监听哨兵元素')

    // 创建新的观察器
    intersectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          console.log('哨兵元素视口状态:', entry.isIntersecting, 'hasMore:', hasMore.value, 'loadingMore:', loadingMore.value, 'listLength:', list.value.length)
          // 当哨兵元素进入视口且有更多数据时，触发加载
          if (entry.isIntersecting && hasMore.value && !loadingMore.value && list.value.length > 0) {
            console.log('触发加载更多')
            loadMore()
          }
        })
      },
      {
        // 设置根边距，提前200px触发
        rootMargin: '200px',
        threshold: 0,
      },
    )

    // 开始观察哨兵元素
    intersectionObserver.observe(sentinelRef.value)
    console.log('开始观察哨兵元素')
  }

  // 开始尝试设置观察器
  trySetupObserver()
}

// 加载更多数据
async function loadMore() {
  if (!hasMore.value || loadingMore.value)
    return

  try {
    loadingMore.value = true
    pagination.value.current += 1

    const res = await getFollowUpRecordPagedApi({
      'queryModel': {
        'studentId': props.studentId,
      },
      'pageRequestModel': {
        'pageIndex': pagination.value.current,
        'pageSize': pagination.value.pageSize,
      },
    })

    if (res.code === 200) {
      // 处理图片数据
      res.result.forEach((item) => {
        if (item.followImages) {
          item.followImages = JSON.parse(item.followImages)
        }
      })

      // 追加数据而不是替换
      list.value.push(...res.result)
      pagination.value.total = res.total

      // 检查是否还有更多数据
      const totalPages = Math.ceil(res.total / pagination.value.pageSize)
      hasMore.value = pagination.value.current < totalPages
    }
  }
  catch (error) {
    console.log(error)
    // 加载失败时回退页码
    pagination.value.current -= 1
  }
  finally {
    loadingMore.value = false
  }
}

async function getFollowUpRecord(isRefresh = false) {
  try {
    spinning.value = true
    // list.value = []

    // 如果是刷新，重置分页和状态
    if (isRefresh) {
      pagination.value.current = 1
      hasMore.value = true
      pagination.value.total = 0
      loadingMore.value = false
    }

    const res = await getFollowUpRecordPagedApi({
      'queryModel': {
        'studentId': props.studentId,
      },
      'pageRequestModel': {
        'pageIndex': pagination.value.current,
        'pageSize': pagination.value.pageSize,
      },
    })
    if (res.code === 200) {
      // 把res.result里的followImages转成数组 格式为"[{\"type\":1,\"url\":\"https://prod-tbu-next"]" followImages会为null 排除null和空字符串
      res.result.forEach((item) => {
        if (item.followImages) {
          item.followImages = JSON.parse(item.followImages)
        }
      })

      if (isRefresh) {
        list.value = res.result
      }
      else {
        list.value = res.result
      }

      pagination.value.total = res.total

      // 检查是否还有更多数据
      const totalPages = Math.ceil(res.total / pagination.value.pageSize)
      hasMore.value = pagination.value.current < totalPages
    }
  }
  catch (error) {
    console.log(error)
  }
  finally {
    spinning.value = false
  }
}

// 提交跟进记录
async function handleFollowUpSubmit(data) {
  try {
    let res
    // 将 intentCourseIds 数组转为逗号分隔的字符串
    const submitData = {
      ...data,
      intentCourseIds: Array.isArray(data.intentCourseIds) ? data.intentCourseIds.join(',') : data.intentCourseIds,
    }
    
    if (isEdit.value) {
      // 编辑模式，使用更新接口
      res = await updateFollowRecordApi({
        ...submitData,
        id: editRecord.value.id,
        uuid: editRecord.value.uuid,
        version: editRecord.value.version,
      })
      if (res.code === 200) {
        messageService.success('编辑跟进记录成功')
        openAddFollowUpModal.value = false
        getFollowUpRecord(true) // 刷新数据
        // 通知父组件刷新学员详情
        emit('refresh-student-detail')
      }
    }
    else {
      // 新增模式，使用创建接口
      res = await createStudentFollowUpApi(submitData)
      if (res.code === 200) {
        messageService.success('添加跟进记录成功')
        openAddFollowUpModal.value = false
        getFollowUpRecord(true) // 刷新数据
        // 通知父组件刷新学员详情
        emit('refresh-student-detail')
      }
    }
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
    // emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
  }
  catch (error) {
    console.log(error)
    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
}

// 组件挂载时设置 IntersectionObserver
onMounted(() => {
  // if (props.studentId) {
  //   getFollowUpRecord(true)
  // }

  // 使用 nextTick 确保 DOM 完全渲染后再设置观察器
  nextTick(() => {
    setupIntersectionObserver()
  })
})

// 当组件被激活时重新设置观察器（用于 keep-alive 或 tab 切换场景）
onActivated(() => {
  nextTick(() => {
    setupIntersectionObserver()
  })
})

// 组件卸载时清理观察器
onUnmounted(() => {
  if (intersectionObserver) {
    intersectionObserver.disconnect()
    intersectionObserver = null
  }
})

// 当组件被停用时清理观察器（用于 keep-alive 或 tab 切换场景）
onDeactivated(() => {
  if (intersectionObserver) {
    intersectionObserver.disconnect()
  }
})

// 监听 studentId 变化，当切换不同学生数据时重新设置观察器
watch(() => props.studentId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    // 当 studentId 变化时，重新设置观察器
    nextTick(() => {
      setupIntersectionObserver()
    })
  }
}, { immediate: false })

// 监听列表数据变化，当有数据时设置观察器
watch(() => list.value.length, (newLength) => {
  if (newLength > 0 && hasMore.value) {
    // 当有数据且还有更多数据时，设置观察器
    nextTick(() => {
      setupIntersectionObserver()
    })
  }
}, { immediate: false })

defineExpose({
  getFollowUpRecord,
})
</script>

<template>
  <div class="follow-up-record p3 pr0">
    <a-spin :spinning="spinning">
      <div class="record bg-white rounded-3 p3">
        <div class="t flex flex-center justify-between border border-b-solid border-color-#eee pb2">
          <span class="text-4 font-500">跟进信息</span>
          <a-button type="primary" @click="handleAddFollowUp">
            添加跟进记录
          </a-button>
        </div>
        <div class="timeLine mt-8 pl-3">
          <a-timeline v-if="list.length > 0">
            <a-timeline-item v-for="(item, index) in list" :key="index">
              <div>
                <div class="time h-8 bg-#06f rounded-10 px-4 flex flex-center w-40 whitespace-nowrap text-#fff">
                  {{
                    dayjs(item.createTime).format('YYYY-MM-DD HH:mm') }}
                </div>
                <div class="bg-#fafafa p4 rounded-2 mt-3">
                  <div class="name flex justify-between flex-center">
                    <div class="flex-center">
                      <span class="w-10 h-10 rounded-10 bg-#06f text-#fff flex-center text-4">{{
                        item.createName.slice(0,
                                              1) }}</span>
                      <span class="text-#666 ml-3 text-3.5">{{ item.createName }}</span>
                    </div>
                    <div>
                      <a type="text" class="text-#06f" @click="handleEditFollowUp(item)">编辑</a>
                    </div>
                  </div>
                  <div class="mb-4">
                    <div class="content whitespace-pre-wrap text-#222 mt-4 mb-2">
                      {{ item.content }}
                    </div>
                    <div v-if="item.followImages && item.followImages.length > 0" class="content-img">
                      <a-space>
                        <a-image
                          v-for="(img, index) in item.followImages" :key="index" :width="80" :height="80"
                          class="rounded-1" :src="img.url" alt=""
                        />
                      </a-space>
                    </div>
                  </div>
                  <!-- 使用枚举  -->
                  <div v-if="item.followMethod !== 0" class="type text-#888 text-3">
                    跟进方式：{{
                      FollowMethodLabel[item.followMethod] }}
                  </div>
                  <div
                    v-for="icText in [formatFollowUpIntendedCourses(item)]"
                    v-show="icText"
                    :key="`${index}-${icText}`"
                    class="type text-#888 text-3"
                  >
                    意向课程：{{ icText }}
                  </div>
                  <div v-if="item.nextFollowUpTime" class="type text-#888 text-3 flex-center justify-start mt-2">
                    下次跟进时间：
                    <div class="bg-#eee rounded-1 px3 py1">
                      {{ dayjs(item.nextFollowUpTime).format('YYYY-MM-DD HH:mm') }}
                      <span
                        v-if="item.visitStatus === 0"
                        class="text-#ff3333 bg-#ffe6e6 px1.5 py0.5 mx2 rounded-10"
                      >未回访</span>
                      <span
                        v-if="item.visitStatus === 1"
                        class="text-#0c3 bg-#e6ffec px1.5 py0.5 mx2 rounded-10"
                      >已回访</span>
                      <a-popconfirm title="标记为已回访？" @confirm="handleMarkAsVisited(item)">
                        <a v-if="item.visitStatus === 0" class="font500">
                          <CheckOutlined /> 标记为已回访
                        </a>
                      </a-popconfirm>
                      <a-popconfirm title="取消标记已回访？" @confirm="handleMarkAsUnvisited(item)">
                        <a v-if="item.visitStatus === 1" class="font500">
                          取消标记已回访
                        </a>
                      </a-popconfirm>
                    </div>
                  </div>
                </div>
              </div>
            </a-timeline-item>
            <!-- 加载更多指示器 -->
            <a-timeline-item v-if="loadingMore">
              <div class="text-center text-#666 py-4">
                <a-spin :spinning="true">
                  <span class="ml-2">正在加载更多...</span>
                </a-spin>
              </div>
            </a-timeline-item>
            <!-- 没有更多数据 -->
            <a-timeline-item v-if="list.length > 0 && (!hasMore || list.length === pagination.pageSize) && !loadingMore">
              <span style="position: relative;top: 6px;">到底了～</span>
            </a-timeline-item>
          </a-timeline>

          <!-- 滚动监听哨兵元素 - 移到时间轴外面，只在有数据且还有更多数据时显示 -->
          <div
            v-if="list.length > 0 && hasMore"
            ref="sentinelRef"
            class="scroll-sentinel"
            style="height: 20px; background: transparent;"
          />

          <div v-else-if="list.length === 0" class="empty-container">
            <a-empty
              image="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12283/static/no-data.88c62015.png"
            >
              <template #description>
                <span class="text-#888">没有跟进信息</span>
              </template>
            </a-empty>
          </div>
        </div>
      </div>
    </a-spin>
    <add-follow-up-modal
      v-model:open="openAddFollowUpModal" :record="studentDetail" :edit-record="editRecord"
      :is-edit="isEdit" @handle-follow-up-submit="handleFollowUpSubmit"
    />
  </div>
</template>

<style lang="less" scoped>
.timeLine {
  :deep(.ant-timeline-item-content) {
    inset-block-start: -12px;
  }
}

.empty-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 400px);
}
</style>
