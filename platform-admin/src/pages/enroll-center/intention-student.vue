<script setup>
import { ref, watch } from 'vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'

const activeKey = ref('1')
const activeKey2 = ref('1')
const userStore = useUserStore()
const instConfig = ref(userStore.instConfig)
const publicDataIsShow = ref(false)
const publicPoolRef = ref(null)
// 子组件引用
const allIntentionStudentRef = ref(null)
const dptIntentionStudentRef = ref(null)

function diaplayPublicDataFun(e) {
  publicDataIsShow.value = e
}

// 监听主tab切换
watch(activeKey, (newKey, oldKey) => {
  // 从其他tab切换到意向学员tab时，刷新数据
  if (newKey === '1' && oldKey !== '1') {
    refreshIntentionStudentData()
  }
  // 刷新公有池
  if (newKey === '2' && oldKey !== '2') {
    publicPoolRef.value?.getIntentStudentList()  // 刷新公有池数据
  }
})

// 监听意向学员内部tab切换
watch(activeKey2, (newKey, oldKey) => {
  // 当在意向学员tab内部切换时，刷新当前激活的子组件数据
  if (activeKey.value === '1' && newKey !== oldKey) {
    refreshIntentionStudentData()
  }
})

// 刷新意向学员数据
function refreshIntentionStudentData() {
  if (activeKey2.value === '1' && allIntentionStudentRef.value) {
    // 刷新全部意向学员数据
    allIntentionStudentRef.value?.getIntentStudentList()
  } else if (activeKey2.value === '2' && dptIntentionStudentRef.value) {
    // 刷新部门意向学员数据
    dptIntentionStudentRef.value?.getIntentStudentList()
  } 
}

onMounted(async () => {
  await userStore.getInstConfig()
  instConfig.value = userStore.instConfig
  if (instConfig.value) {
    publicDataIsShow.value = instConfig.value.enablePublicPool
  }
})
</script>

<template>
  <div class="home">
    <div class="tabs">
      <a-tabs 
        v-model:active-key="activeKey" :animated="publicDataIsShow" :tab-bar-style="{
          'border-bottom-left-radius': '0px',
          'border-bottom-right-radius': '0px',
        }"
      >
        <!-- force-render 强制渲染 -->
        <a-tab-pane  key="1" tab="意向学员">
          <a-tabs
            v-model:active-key="activeKey2" animated :tab-bar-style="{
              'height': '46px',
              'border-top-left-radius': '0px',
              'border-top-right-radius': '0px',
            }" class="twoTab"
          >
            <a-tab-pane key="1">
              <template #tab>
                <span class="custom-tab">
                  全部意向学员
                  <a-tooltip
                    :overlay-style="{
                      maxWidth: '300px', // 最大宽度
                      whiteSpace: 'normal', // 允许换行
                    }"
                  >
                    <template #title>查看机构内所有的意向学员</template>

                    <QuestionCircleOutlined />
                  </a-tooltip>
                </span>
              </template>
              <div class="tab-content">
                <all-intention-student ref="allIntentionStudentRef" :public-data-is-show="publicDataIsShow" />
              </div>
            </a-tab-pane>
            <a-tab-pane key="2">
              <template #tab>
                <span class="custom-tab">
                  部门意向学员
                  <a-tooltip
                    :overlay-style="{
                      maxWidth: '300px', // 最大宽度
                      whiteSpace: 'normal', // 允许换行
                    }"
                  >
                    <template #title>查看所在部门及下级部门销售员的意向学员</template>

                    <QuestionCircleOutlined />
                  </a-tooltip>
                </span>
              </template>
              <div class="tab-content">
                <dpt-intention-student ref="dptIntentionStudentRef" :public-data-is-show="publicDataIsShow"/>
              </div>
            </a-tab-pane>
          </a-tabs>
        </a-tab-pane>
        <a-tab-pane v-if="publicDataIsShow" key="2" tab="公有池">
          <public-pool ref="publicPoolRef" />
        </a-tab-pane>
        <a-tab-pane key="3" tab="渠道管理">
          <channel-management />
        </a-tab-pane>
        <a-tab-pane key="4" tab="设置">
          <setting @diaplay-public-data="diaplayPublicDataFun" />
        </a-tab-pane>
      </a-tabs>
    </div>
  </div>
</template>

<style scoped lang="less">
  .home {
    color: #666;

    .tabs {
      width: 100%;
      border-radius: 10px;
      line-height: 40px;

      :deep(.ant-tabs-nav) {
        background: #fff;
        border-radius: 16px;
        margin: 0;
      }

      :deep(.ant-tabs-nav-wrap) {
        padding-left: 36px;
      }

      :deep(.ant-tabs-ink-bar) {
        text-align: center;
        height: 9px !important;
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

      .twoTab {
        padding: 0;

        :deep(.ant-tabs-nav-wrap) {
          padding-left: 10px;
          margin: 6px 0;
        }

        :deep(.ant-tabs-nav) {
          &::before {
            display: none;
          }
        }

        :deep(.ant-tabs-tab) {
          padding: 6px 14px !important;
        }

        :deep(.ant-tabs-tab-active) {
          background: #e6f0ff;
          border-radius: 8px;
        }

        :deep(.ant-tabs-ink-bar) {
          display: none;
        }

      }
    }
  }
</style>
