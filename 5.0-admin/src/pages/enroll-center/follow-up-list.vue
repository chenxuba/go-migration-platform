<script setup>
import { ref, watch } from 'vue'
import { Empty } from 'ant-design-vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { AccessEnum } from '@/constants/access'
import { useAccess } from '@/composables/access'

const activeKey2 = ref('1')
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const { hasAccess } = useAccess()

const canViewAllFollow = computed(() => hasAccess(AccessEnum.enroll_follow_view_all))
const canViewMyFollow = computed(() => hasAccess(AccessEnum.enroll_follow_view_my))
const canViewDeptFollow = computed(() => hasAccess(AccessEnum.enroll_follow_view_dept))
const primaryFollowTabKey = computed(() => {
  if (canViewAllFollow.value || canViewMyFollow.value)
    return '1'
  if (canViewDeptFollow.value)
    return '2'
  return ''
})
const primaryFollowTabLabel = computed(() =>
  canViewAllFollow.value ? '全部跟进记录' : '我的跟进记录',
)
const primaryFollowTabTip = computed(() =>
  canViewAllFollow.value
    ? '查看机构内所有的跟进记录'
    : '仅查看采单员/前台/电话销售/副销售员/销售员/班主任为自己的跟进记录',
)

watch(
  [primaryFollowTabKey, canViewDeptFollow],
  ([primaryKey, deptVisible]) => {
    const availableKeys = [primaryKey, deptVisible ? '2' : ''].filter(Boolean)
    if (availableKeys.length === 0)
      return
    if (!availableKeys.includes(activeKey2.value))
      activeKey2.value = availableKeys[0]
  },
  { immediate: true },
)
</script>

<template>
  <div class="tabs">
    <a-tabs
      v-if="primaryFollowTabKey || canViewDeptFollow"
      v-model:active-key="activeKey2" animated :tab-bar-style="{
        'height': '46px',
        'border-bottom-left-radius': '0px',
        'border-bottom-right-radius': '0px',
      }" class="twoTab"
    >
      <a-tab-pane v-if="primaryFollowTabKey === '1'" key="1">
        <template #tab>
          <span class="custom-tab">
            <a-tooltip
              :overlay-style="{
                maxWidth: '300px',
                whiteSpace: 'normal',
              }"
            >
              <template #title>{{ primaryFollowTabTip }}</template>
              {{ primaryFollowTabLabel }}
              <QuestionCircleOutlined />
            </a-tooltip>
          </span>
        </template>
        <div class="tab-content">
          <all-follow-up-list />
        </div>
      </a-tab-pane>
      <a-tab-pane v-if="canViewDeptFollow" key="2">
        <template #tab>
          <a-tooltip
            :overlay-style="{
              maxWidth: '300px',
              whiteSpace: 'normal',
            }"
          >
            <template #title>
              查看所在部门及下级部门人员的跟进记录
            </template>
            部门跟进记录
            <QuestionCircleOutlined />
          </a-tooltip>
        </template>
        <div class="tab-content">
          <dpt-follow-up-list />
        </div>
      </a-tab-pane>
    </a-tabs>
    <div v-if="!primaryFollowTabKey && !canViewDeptFollow" class="tab-empty bg-white rounded-b-4 h-260px flex items-center justify-center">
      <a-empty :image="simpleImage" description="暂无可查看的跟进记录权限" />
    </div>
  </div>
</template>

<style lang="less" scoped>
.tabs {
  width: 100%;
  border-radius: 10px;
  line-height: 40px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    border-radius: 16px;
    margin: 0;
  }

  .twoTab {
    :deep(.ant-tabs-nav-wrap) {
      padding-left: 10px;
      margin: 6px 0;
    }

    :deep(.ant-tabs-tab) {
      padding: 6px 14px !important;
    }

    :deep(.ant-tabs-ink-bar) {
      text-align: center;
      height: 3px !important;
      background: transparent;
      bottom: 1px !important;

      &::after {
        position: absolute;
        top: 0;
        left: calc(50% - 12px);
        width: 24px !important;
        height: 4px !important;
        border-radius: 2px;
        background-color: var(--pro-ant-color-primary);
        content: "";
      }
    }

  }
}
</style>
