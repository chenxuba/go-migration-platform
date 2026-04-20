<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { Empty, Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import createNoticeModel from './components/createNoticeModel.vue'
import {
  listNoticeTemplatesApi,
  pageNoticesApi,
  withdrawNoticeApi,
  type NoticeListItem,
  type NoticeTemplateItem,
} from '@/api/home-center/notice'
import messageService from '@/utils/messageService'

const TEMPLATE_CARD_WIDTH = 156
const TEMPLATE_CARD_GAP = 14

const displayArray = ref(['commentStatus', 'createUser', 'applyTime'])
const createNoticeOpen = ref(false)
const createTemplate = ref<NoticeTemplateItem | null>(null)
const templateLibraryOpen = ref(false)
const templateViewportRef = ref<HTMLElement | null>(null)
const visibleTemplateCount = ref(0)
const noticeTemplateLoading = ref(false)
const noticeTemplates = ref<NoticeTemplateItem[]>([])
const noticeListLoading = ref(false)
const noticeList = ref<NoticeListItem[]>([])
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const defaultTemplateCover = 'https://prod-cdn.schoolpal.cn/training/next-erp/h5/static/images/notice/cjbk2025.png'

const filters = reactive({
  status: undefined as number | undefined,
  isWithdraw: undefined as boolean | undefined,
  operatorId: undefined as string | undefined,
  dateRange: [] as string[],
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const tableColumns = [
  { title: '通知标题', dataIndex: 'title', key: 'title', width: 180 },
  { title: '摘要内容', dataIndex: 'summary', key: 'summary', width: 280 },
  { title: '通知范围', dataIndex: 'scope', key: 'scope', width: 220 },
  { title: '公告状态', dataIndex: 'status', key: 'status', width: 130 },
  { title: '家长端已读', dataIndex: 'read', key: 'read', width: 150 },
  { title: '创建人', dataIndex: 'operator', key: 'operator', width: 120 },
  { title: '发布时间', dataIndex: 'publishTime', key: 'publishTime', width: 170 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 120, fixed: 'right' as const },
]

const statusFilterOptions = [
  { id: '1', value: '待审核' },
  { id: '2', value: '审核未通过' },
  { id: '3', value: '待发布' },
  { id: '4', value: '已发布' },
  { id: 'withdrawn', value: '已撤回' },
]

let templateResizeObserver: ResizeObserver | null = null

const visibleNoticeTemplates = computed(() => {
  if (!noticeTemplates.value.length)
    return []
  const count = Math.max(1, visibleTemplateCount.value || 0)
  return noticeTemplates.value.slice(0, count)
})

const tablePagination = computed(() => ({
  current: pagination.current,
  pageSize: pagination.pageSize,
  total: pagination.total,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

function updateVisibleTemplateCount() {
  const containerWidth = Number(templateViewportRef.value?.clientWidth || 0)
  if (!noticeTemplates.value.length) {
    visibleTemplateCount.value = 0
    return
  }
  if (!containerWidth) {
    visibleTemplateCount.value = Math.min(6, noticeTemplates.value.length)
    return
  }
  const count = Math.floor((containerWidth + TEMPLATE_CARD_GAP) / (TEMPLATE_CARD_WIDTH + TEMPLATE_CARD_GAP))
  visibleTemplateCount.value = Math.max(1, Math.min(count, noticeTemplates.value.length))
}

function installTemplateResizeObserver() {
  templateResizeObserver?.disconnect()
  templateResizeObserver = null
  if (!(templateViewportRef.value instanceof HTMLElement)) {
    updateVisibleTemplateCount()
    return
  }
  if (typeof ResizeObserver === 'undefined') {
    updateVisibleTemplateCount()
    return
  }
  templateResizeObserver = new ResizeObserver(() => {
    updateVisibleTemplateCount()
  })
  templateResizeObserver.observe(templateViewportRef.value)
  updateVisibleTemplateCount()
}

function getNoticeTemplateStyle(item: NoticeTemplateItem) {
  const coverUrl = String(item.coverUrl || '').trim() || defaultTemplateCover
  return {
    backgroundImage: `url("${coverUrl}")`,
  }
}

function previewNoticeTemplate(item: NoticeTemplateItem) {
  const target = String(item.coverUrl || '').trim() || defaultTemplateCover
  window.open(target, '_blank', 'noopener,noreferrer')
}

function openTemplateLibrary() {
  templateLibraryOpen.value = true
}

function handleOpenCreateNotice(template?: NoticeTemplateItem | null) {
  createTemplate.value = template || null
  createNoticeOpen.value = true
  templateLibraryOpen.value = false
}

function buildPageQuery() {
  const beginPublishDate = filters.dateRange?.[0] || undefined
  const endPublishDate = filters.dateRange?.[1] || undefined
  return {
    pageRequestModel: {
      pageIndex: pagination.current,
      pageSize: pagination.pageSize,
    },
    queryModel: {
      statuses: filters.status ? [filters.status] : undefined,
      isWithdraw: filters.isWithdraw,
      beginPublishDate,
      endPublishDate,
      operatorId: filters.operatorId,
    },
  }
}

function normalizeSingleFilterValue(value: unknown) {
  if (Array.isArray(value))
    return value.length ? String(value[0] || '').trim() || undefined : undefined
  const text = String(value ?? '').trim()
  return text || undefined
}

function normalizeDateRange(value: unknown) {
  if (!Array.isArray(value) || value.length !== 2)
    return []
  const values = value.map(item => String(item || '').trim()).filter(Boolean)
  return values.length === 2 ? values : []
}

async function fetchNoticeTemplates() {
  noticeTemplateLoading.value = true
  try {
    const res = await listNoticeTemplatesApi()
    if (res.code !== 200) {
      messageService.error(res.message || '获取通知模板失败')
      return
    }
    noticeTemplates.value = Array.isArray(res.result) ? res.result : []
    await nextTick()
    updateVisibleTemplateCount()
  }
  catch (error: any) {
    console.error('fetch notice templates failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '获取通知模板失败')
  }
  finally {
    noticeTemplateLoading.value = false
  }
}

async function fetchNoticeList() {
  noticeListLoading.value = true
  try {
    const res = await pageNoticesApi(buildPageQuery())
    if (res.code !== 200) {
      messageService.error(res.message || '获取通知列表失败')
      return
    }
    noticeList.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    console.error('fetch notice list failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '获取通知列表失败')
  }
  finally {
    noticeListLoading.value = false
  }
}

function triggerFilterRefresh() {
  pagination.current = 1
  fetchNoticeList()
}

function handleStatusFilterChange(value: unknown) {
  const normalized = normalizeSingleFilterValue(value)
  if (normalized === 'withdrawn') {
    filters.status = undefined
    filters.isWithdraw = true
  }
  else {
    filters.status = normalized ? Number(normalized) : undefined
    filters.isWithdraw = normalized ? false : undefined
  }
  triggerFilterRefresh()
}

function handleCreateUserFilterChange(value: unknown) {
  filters.operatorId = normalizeSingleFilterValue(value)
  triggerFilterRefresh()
}

function handlePublishDateFilterChange(value: unknown) {
  filters.dateRange = normalizeDateRange(value)
  triggerFilterRefresh()
}

function handleTableChange(pageConfig: any) {
  const nextCurrent = Number(pageConfig?.current || 1)
  const nextPageSize = Number(pageConfig?.pageSize || pagination.pageSize)
  if (nextCurrent === pagination.current && nextPageSize === pagination.pageSize)
    return
  pagination.current = nextCurrent
  pagination.pageSize = nextPageSize
  fetchNoticeList()
}

function getScopeText(record: NoticeListItem) {
  if (record.isAllSchool)
    return '全校群发'
  const classNames = (Array.isArray(record.classs) ? record.classs : [])
    .map(item => String(item.className || '').trim())
    .filter(Boolean)
  if (classNames.length <= 2)
    return classNames.join('、')
  if (classNames.length > 2)
    return `${classNames.slice(0, 2).join('、')} 等 ${classNames.length} 个来源`
  if (record.studentCount > 0)
    return `指定学员（${record.studentCount} 人）`
  return '指定范围'
}

function asNoticeRecord(record: Record<string, any>) {
  return record as NoticeListItem
}

function getStatusMeta(record: NoticeListItem) {
  if (record.isWithdraw)
    return { text: '已撤回', color: 'default' }
  switch (Number(record.status || 0)) {
    case 1:
      return { text: '待审核', color: 'processing' }
    case 2:
      return { text: '审核驳回', color: 'error' }
    case 3:
      return { text: '待发布', color: 'warning' }
    case 4:
      return { text: '已发布', color: 'success' }
    default:
      return { text: '未知状态', color: 'default' }
  }
}

function getDisplayPublishTime(record: NoticeListItem) {
  return record.realityPublishTime || record.publishTime || record.operationDate || ''
}

function formatDateTime(value?: string | null) {
  if (!value || value.startsWith('0001-01-01'))
    return '--'
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

async function handleWithdraw(record: NoticeListItem) {
  Modal.confirm({
    title: '确认撤回该通知？',
    centered: true,
    content: '撤回后该通知会保留记录，但状态将更新为已撤回。',
    okText: '确认撤回',
    cancelText: '取消',
    async onOk() {
      try {
        const res = await withdrawNoticeApi({ noticeId: record.noticeId })
        if (res.code !== 200) {
          messageService.error(res.message || '撤回通知失败')
          return
        }
        messageService.success('通知已撤回')
        fetchNoticeList()
      }
      catch (error: any) {
        console.error('withdraw notice failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '撤回通知失败')
      }
    },
  })
}

function handleCreateSuccess() {
  pagination.current = 1
  fetchNoticeList()
}

onMounted(async () => {
  await Promise.all([fetchNoticeTemplates(), fetchNoticeList()])
  await nextTick()
  installTemplateResizeObserver()
})

onBeforeUnmount(() => {
  templateResizeObserver?.disconnect()
  templateResizeObserver = null
})
</script>

<template>
  <div class="noticePage">
    <div class="noticePage__hero">
      <div class="noticePage__heroHeader">
        <div>
          <div class="noticePage__heroTitle">
            通知公告模板
          </div>
          <div class="noticePage__heroDesc">
            多种模板，一键群发，已读未读及时跟进
          </div>
        </div>
        <div class="noticePage__more" @click="openTemplateLibrary">
          <img
            width="14"
            src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAgCAYAAAAFQMh/AAAAAXNSR0IArs4c6QAABIJJREFUWEe1ln1sE3UYx7/PvdDCnC8Mx0RM3DSRoBIQM9Z2aFG2dajbnBkxkQjEdyEmKo46iXbRZB3E+DoTjFETEqIh4rbo2g0iZLNtiJk6/yC+ZmIMOpYpDJm99nqPuaZX29rrVan31909z/P93O95nt/vHoLFxbsGq6HyMBgT6G3ZSCDODJnyco1IWEEaTgEor+ilYStN3U5WTrxz4AMQ2pN+zPdQb9v+zJgZLzcyYMDeXeSnrVaalmDuHKyDwJG0EGMS85Vl5NsYM96d8vIqAfg8Jbanwk+d5wVm3xEJ0dmjAFw5Qp3kb91jvJvu4nWkYSsxpjQBc5f20HP/GcxdgzdA41cA1OcR0UDcB2jd1NM+Yw5hcnrijzAJ1xOJ/vAQncj0zaox7zhYCUncDeBey/ozTkPA87B98TL5fFqmaEcHiydn428zka6jN8eJi0i+JhAgxfDLBj89MAJGQzGp+rvuvJ162/qM5+Zmtp3R1PdAaMvRcYaDcrpfssHegb0AHvwXYA2MFupt/ViPcbv5gphdHQBwS6YGA+OROqkWPkpnJhvs8wlQVm4G0wsAllh8wCigPUH+O8d1P0cTLyRSAwBqs+Pox7ggOj4bol9Na2wYeMdwGWRlF5i9eeAzAD9A/rYPDZurgZewqI4AuDbHf4aQcIWC9m9ydQoeIOztfwegLVlBRHdQT8tHxrv626I1iYR0mMDVOeJ/sob1kRE5nC9zhcHP9F+BBH0LwJ4KHiV/682GkLNRuQ6CoK/0shxxjYnuigSkfrNyWR+Z3gF9ez2VFCCuo562Y/pt/YbYGk0jvaaX/FOct4WD894o1CPWYL3eYvQlgCeMbVPXFF9PhH4CynLFGeiJBOUuq51hCc4VcHnUdgbrPwpbnpXuCwflzXpqSgp2emJbAHoLgJinSw9Hp6UN4+MUt4Imq1aMk+7j8sQeZ9CLeWMYX5JNuik0SGeL1SsK7GiKdxPh2fyifEJk2TE2TL8Y9oe/P1vJotDNxGOCBpGJnKwp3jevWnjG8LEAMzmaE68S83aTlfzGmuaKjNi+zrQ/OvlHVQJ0Mp0dgsoJZVFRYLebpZg9rh8gm0ygUQIaQkH503z2hybP/QCgJmWb2FtdtjLTL++K3W62x+zq+wBaTKAagTpCQelgPvu2n2cr1Lg4ldGEytz8BRX7quicaapdLVzOseQfZp1poxA/Fg7Me83Mfv93c0tFiV9P1pghMsgpyYn7+pZemB4cslZceytXSLIaBHCjORS7wwF5Z7Hda+aXBjtu58uhqiMELDd1Jt4fCsibijkgrD4sCXY0Ra8mkg4BfGWBgE9Ol0vNxw9QesK0Ei9kpzUeZYUEYZiBqgJp+Uolae2xAM2eDyyrq52euD4TryrQSD+RKjtCh5L7smSXPoLqA1idieLvImn1YwHb8ZIRU0LkaFSWkSgcBWNxjrgCAY3hIXm01FBdL9lca5uV5QkWjgCoTEE0Bt0dCUoH/g9oGqzfpMYYfaJYDOInCx0QpfiYrANk9WqWF1Ti4rEATZdCvJDGX55jnDAUsnPxAAAAAElFTkSuQmCC"
            alt=""
          >
          更多模板
        </div>
      </div>

      <div ref="templateViewportRef" class="templatePreview">
        <a-spin :spinning="noticeTemplateLoading">
          <div v-if="visibleNoticeTemplates.length" class="templatePreview__list">
            <div
              v-for="item in visibleNoticeTemplates"
              :key="item.id"
              class="templateItem templateItem--preview"
            >
              <div class="templateItem__cover" :style="getNoticeTemplateStyle(item)">
                <div class="mask">
                  <a-button type="primary" @click.stop="handleOpenCreateNotice(item)">
                    使用
                  </a-button>
                  <div class="eye" @click.stop="previewNoticeTemplate(item)">
                    <svg width="17px" height="12px" viewBox="0 0 17 12">
                      <title>预览</title>
                      <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g transform="translate(-1825.000000, -377.000000)">
                          <g transform="translate(1816.000000, 367.000000)">
                            <g transform="translate(9.000000, 10.000000)">
                              <path
                                d="M8.00044664,12 C3.58904385,12 0,9.30840603 0,6 C0,2.69159397 3.58922251,0 8.00044664,0 C12.4107775,0 16,2.69159397 16,6 C16,9.30840603 12.4107775,12 8.00044664,12 L8.00044664,12 Z M8.00044664,1.69822139 C4.55254196,1.69822139 1.64025146,3.66832095 1.64025146,6 C1.64025146,8.33167905 4.55254196,10.3017786 8.00044664,10.3017786 C11.447458,10.3017786 14.3595699,8.33167905 14.3595699,6 C14.3595699,3.66832095 11.4472794,1.69822139 8.00044664,1.69822139 Z"
                                fill="currentColor"
                              />
                              <path
                                d="M5.44993691,5.99981505 C5.44993691,6.94299358 5.93599319,7.81452648 6.7250131,8.28611575 C7.51403302,8.75770502 8.48614564,8.75770502 9.27516555,8.28611575 C10.0641855,7.81452648 10.5502417,6.94299358 10.5502417,5.99981505 C10.5502417,5.05663652 10.0641855,4.18510361 9.27516555,3.71351434 C8.48614564,3.24192507 7.51403302,3.24192507 6.7250131,3.71351434 C5.93599319,4.18510361 5.44993691,5.05663652 5.44993691,5.99981505 L5.44993691,5.99981505 Z"
                                fill="currentColor"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </div>
                </div>
              </div>
              <div class="templateItemTitle">
                {{ item.title }}
              </div>
            </div>
          </div>
          <div v-else class="noticePage__emptyWrap">
            <a-empty :image="simpleImage" description="暂无模板" />
          </div>
        </a-spin>
      </div>
    </div>

    <div class="noticePage__filters">
      <div class="noticePage__filtersRow">
        <all-filter
          class="noticePage__allFilter"
          :display-array="displayArray"
          comment-status-label="公告状态"
          :comment-status-options="statusFilterOptions"
          create-user-label="创建人"
          create-user-placeholder="请输入创建人"
          apply-time-label="发布时间"
          @update:comment-status-filter="handleStatusFilterChange"
          @update:create-user-filter="handleCreateUserFilterChange"
          @update:apply-time-filter="handlePublishDateFilterChange"
        />
      </div>
    </div>

    <div class="noticePage__tableCard">
      <div class="noticePage__tableHeader">
        <div class="noticePage__tableTitle">
          当前共计 {{ pagination.total }} 条通知公告
        </div>
        <a-button type="primary" @click="handleOpenCreateNotice()">
          创建通知
        </a-button>
      </div>

      <a-table
        row-key="noticeId"
        :loading="noticeListLoading"
        :columns="tableColumns"
        :data-source="noticeList"
        :pagination="tablePagination"
        :scroll="{ x: 1250 }"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'title'">
            <div class="noticeTable__title">
              {{ record.title || '--' }}
            </div>
          </template>

          <template v-else-if="column.key === 'summary'">
            <div class="noticeTable__summary">
              {{ record.summary || '--' }}
            </div>
          </template>

          <template v-else-if="column.key === 'scope'">
            <div class="noticeTable__scope">
              {{ getScopeText(asNoticeRecord(record)) }}
            </div>
          </template>

          <template v-else-if="column.key === 'status'">
            <a-tag :color="getStatusMeta(asNoticeRecord(record)).color">
              {{ getStatusMeta(asNoticeRecord(record)).text }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'read'">
            <div class="noticeTable__metrics">
              <div>已读 {{ record.readStudentCount }}/{{ record.studentCount }}</div>
              <div v-if="record.isConfirm">
                确认 {{ record.confirmStudentCount }}/{{ record.studentCount }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'operator'">
            <div>{{ record.operatorName || '--' }}</div>
          </template>

          <template v-else-if="column.key === 'publishTime'">
            <div>{{ formatDateTime(getDisplayPublishTime(asNoticeRecord(record))) }}</div>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-button
              v-if="!record.isWithdraw"
              type="link"
              class="noticeTable__action"
              @click="handleWithdraw(asNoticeRecord(record))"
            >
              撤回
            </a-button>
            <span v-else class="noticeTable__actionDisabled">已撤回</span>
          </template>
        </template>
      </a-table>
    </div>

    <createNoticeModel
      v-model="createNoticeOpen"
      :selected-template="createTemplate"
      @success="handleCreateSuccess"
    />

    <a-modal
      v-model:open="templateLibraryOpen"
      centered
      wrap-class-name="template-library-modal"
      :footer="false"
      :closable="false"
      :mask-closable="true"
      :keyboard="true"
      :width="760"
      :body-style="{ maxHeight: '68vh', overflowY: 'auto', padding: '12px 24px 24px' }"
      destroy-on-close
    >
      <template #title>
        <div class="templateLibraryModal__title">
          <span>模板库</span>
          <a-button type="text" class="templateLibraryModal__close" @click="templateLibraryOpen = false">
            <template #icon>
              <CloseOutlined />
            </template>
          </a-button>
        </div>
      </template>
      <div class="templateLibraryModal__body">
        <a-spin :spinning="noticeTemplateLoading">
          <div v-if="noticeTemplates.length" class="templateLibraryModal__grid">
            <div
              v-for="item in noticeTemplates"
              :key="`${item.id}-library`"
              class="templateItem templateItem--library"
            >
              <div class="templateItem__cover" :style="getNoticeTemplateStyle(item)">
                <div class="mask">
                  <a-button type="primary" @click.stop="handleOpenCreateNotice(item)">
                    使用
                  </a-button>
                  <div class="eye" @click.stop="previewNoticeTemplate(item)">
                    <svg width="17px" height="12px" viewBox="0 0 17 12">
                      <title>预览</title>
                      <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g transform="translate(-1825.000000, -377.000000)">
                          <g transform="translate(1816.000000, 367.000000)">
                            <g transform="translate(9.000000, 10.000000)">
                              <path
                                d="M8.00044664,12 C3.58904385,12 0,9.30840603 0,6 C0,2.69159397 3.58922251,0 8.00044664,0 C12.4107775,0 16,2.69159397 16,6 C16,9.30840603 12.4107775,12 8.00044664,12 L8.00044664,12 Z M8.00044664,1.69822139 C4.55254196,1.69822139 1.64025146,3.66832095 1.64025146,6 C1.64025146,8.33167905 4.55254196,10.3017786 8.00044664,10.3017786 C11.447458,10.3017786 14.3595699,8.33167905 14.3595699,6 C14.3595699,3.66832095 11.4472794,1.69822139 8.00044664,1.69822139 Z"
                                fill="currentColor"
                              />
                              <path
                                d="M5.44993691,5.99981505 C5.44993691,6.94299358 5.93599319,7.81452648 6.7250131,8.28611575 C7.51403302,8.75770502 8.48614564,8.75770502 9.27516555,8.28611575 C10.0641855,7.81452648 10.5502417,6.94299358 10.5502417,5.99981505 C10.5502417,5.05663652 10.0641855,4.18510361 9.27516555,3.71351434 C8.48614564,3.24192507 7.51403302,3.24192507 6.7250131,3.71351434 C5.93599319,4.18510361 5.44993691,5.05663652 5.44993691,5.99981505 L5.44993691,5.99981505 Z"
                                fill="currentColor"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </div>
                </div>
              </div>
              <div class="templateItemTitle">
                {{ item.title }}
              </div>
            </div>
          </div>
          <div v-else class="templateLibraryModal__empty">
            <a-empty :image="simpleImage" description="暂无模板" />
          </div>
        </a-spin>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.noticePage {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.noticePage__hero,
.noticePage__filters,
.noticePage__tableCard {
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 35, 41, 0.04);
}

.noticePage__hero {
  padding: 20px 20px 16px;
}

.noticePage__heroHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.noticePage__heroTitle {
  font-size: 20px;
  font-weight: 700;
  line-height: 30px;
  color: #1f2329;
}

.noticePage__heroDesc {
  margin-top: 4px;
  font-size: 13px;
  line-height: 20px;
  color: #8c94a4;
}

.noticePage__more {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 12px;
  background: #eef4ff;
  color: #2468f2;
  cursor: pointer;
  padding: 10px 12px;
  font-size: 13px;
  line-height: 20px;
  transition: all 0.2s ease;

  &:hover {
    background: #e5efff;
  }

  &::after {
    position: absolute;
    top: 7px;
    right: 5px;
    width: 8px;
    height: 8px;
    border: 1px solid #fff;
    border-radius: 50%;
    background: #ee1625;
    content: "";
  }
}

.noticePage__filters {
  padding: 16px 20px;
}

.noticePage__filtersRow {
  display: flex;
  align-items: flex-start;
}

.noticePage__allFilter {
  width: 100%;
}

.noticePage__tableCard {
  padding: 18px 20px 20px;
}

.noticePage__tableHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.noticePage__tableTitle {
  position: relative;
  padding-left: 12px;
  font-size: 14px;
  line-height: 22px;
  color: #1f2329;

  &::before {
    position: absolute;
    top: 5px;
    left: 0;
    width: 4px;
    height: 12px;
    border-radius: 2px;
    background: #2468f2;
    content: "";
  }
}

.noticePage__emptyWrap {
  padding: 48px 0 36px;
}

.noticeTable__title {
  font-weight: 600;
  color: #1f2329;
  line-height: 22px;
}

.noticeTable__summary {
  display: -webkit-box;
  overflow: hidden;
  color: #5f6b7c;
  line-height: 22px;
  word-break: break-word;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.noticeTable__scope {
  color: #1f2329;
  line-height: 22px;
}

.noticeTable__metrics {
  color: #5f6b7c;
  line-height: 22px;
}

.noticeTable__action {
  padding-left: 0;
}

.noticeTable__actionDisabled {
  color: #b7bfcc;
}

.templatePreview {
  margin-top: 12px;
}

.templatePreview__list {
  display: flex;
  gap: 14px;
  overflow: hidden;
}

.templateItem {
  .templateItemTitle {
    margin-top: 6px;
    font-size: 14px;
    line-height: 22px;
    color: #1f2329;
    font-weight: 500;
  }

  &--preview {
    width: 156px;
    min-width: 156px;
  }

  &--library {
    width: 100%;
  }

  .templateItem__cover {
    width: 100%;
    height: 156px;
    border-radius: 8px;
    background: url("https://prod-cdn.schoolpal.cn/training/next-erp/h5/static/images/notice/cjbk2025.png");
    background-size: 100% 100%;

    .mask {
      display: flex;
      width: 100%;
      height: 100%;
      justify-content: space-between;
      border-radius: 8px;
      background: rgba(0, 0, 0, 0.1);
      opacity: 0;
      padding: 12px;

      .eye {
        display: flex;
        width: 34px;
        height: 32px;
        align-items: center;
        justify-content: center;
        border: 1px solid #fff;
        border-radius: 8px;
        background: #fff;
        box-sizing: border-box;
        color: #666;
        cursor: pointer;
        transition: all 0.3s ease;

        &:hover {
          border-color: #06f;
          color: #06f;
        }
      }
    }

    &:hover {
      cursor: pointer;

      .mask {
        opacity: 1;
      }
    }
  }
}

.templateLibraryModal__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 16px;
  font-weight: 600;
  color: #222;
}

.templateLibraryModal__close {
  color: #999;

  &:hover {
    color: #333;
    background: transparent;
  }
}

.templateLibraryModal__body {
  padding-top: 8px;
}

.templateLibraryModal__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(156px, 1fr));
  gap: 22px 14px;
}

.templateLibraryModal__empty {
  padding: 48px 0 36px;
}

@media (max-width: 960px) {
  .noticePage__heroHeader,
  .noticePage__tableHeader,
  .noticePage__filtersRow {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>

<style lang="less">
.template-library-modal {
  .ant-modal-content {
    overflow: hidden;
    border-radius: 16px;
  }

  .ant-modal-header {
    border-radius: 16px 16px 0 0;
  }

  .ant-modal-body {
    scrollbar-gutter: stable;
    scrollbar-width: thin;
    scrollbar-color: #cfd6e4 transparent;
  }

  .ant-modal-body::-webkit-scrollbar {
    width: 8px;
  }

  .ant-modal-body::-webkit-scrollbar-track {
    background: transparent;
  }

  .ant-modal-body::-webkit-scrollbar-thumb {
    border: 2px solid transparent;
    border-radius: 999px;
    background: #cfd6e4;
    background-clip: padding-box;
  }

  .ant-modal-body::-webkit-scrollbar-thumb:hover {
    background: #b9c3d4;
    background-clip: padding-box;
  }

  .ant-modal-body::-webkit-scrollbar-corner {
    background: transparent;
  }
}
</style>
