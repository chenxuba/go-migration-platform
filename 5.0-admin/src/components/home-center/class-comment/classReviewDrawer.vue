<script setup>
import { CloseOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
// 去点评
import dayjs from 'dayjs'
import reviewDrawer from './reviewDrawer.vue'

const props = defineProps({
  type: {
    type: [String, Number],
    default: '1',
  },
})

const activeKey = ref('')

watch(
  () => props.type,
  (val) => {
    console.log('val', val)
    activeKey.value = val
  },
  {
    immediate: true,
  },
)

const open = defineModel({
  type: Boolean,
  default: false,
})

const openModal = ref(false)

const reviewDrawerOpen = ref(false)

const commentsDrawerOpen = ref(false)

// 上课内容
const content = ref('我是上课内容')
//
const modelValue = ref(content.value)

const columns = [
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    width: 550,
  },
  {
    title: '上课状态',
    dataIndex: 'status',
    key: 'status',
    width: 250,
    filters: [
      { text: '到课', value: '0' },
      { text: '请假', value: '1' },
      { text: '旷课', value: '2' },
      { text: '未记录', value: '3' },
    ],
    filterSearch: true,
    onFilter: (value, record) => record.status.startsWith(value),
  },
  {
    title: '操作',
    key: 'action',
  },
]

const data = [
  {
    key: '1',
    name: 'John Brown',
    sex: '0',
    status: '0',
  },
  {
    key: '2',
    name: 'Jim Green',
    sex: '1',
    status: '1',
  },
  {
    key: '3',
    name: 'Joe Black',
    sex: '0',
    status: '2',
  },
]

const selectedRowKeysState = reactive({
  selectedRowKeys: [], // Check here to configure the default column
})

const comments = ref([
  {
    name: '何老师',
    time: '2022-03-15 10:00',
    content: '这是评论',
  },
])

const commentContent = ref('')

function handleCommentSubmit() {
  if (!commentContent.value) {
    return
  }
  comments.value.push({
    name: '何老师',
    time: dayjs().format('YYYY-MM-DD HH:mm'),
    content: commentContent.value || '这是评论',
  })
  commentContent.value = ''
}
function handelDeleteComment(index) {
  comments.value.splice(index, 1)
}

function onSelectChange(selectedRowKeys) {
  console.log('selectedRowKeys changed: ', selectedRowKeys)
  selectedRowKeysState.selectedRowKeys = selectedRowKeys
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="open" :body-style="{ padding: '0', background: '#f7f7fd' }" :keyboard="false"
      :mask-closable="false" :closable="false" width="1165px"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            课堂点评详情
          </div>
          <a-button type="text" class="close-btn" @click="open = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" class=" rounded-100"
            src="https://pcsys.admin.ybc365.com//e64c7fd6-2edc-412f-9141-a9904be88b4f.png" alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                奥夫班
              </div>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">2025-04-14(周一)10:00 ~ 10:30</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">30分钟</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="3" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="上课老师">
            张晨
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            陈瑞生
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            -
          </a-descriptions-item>
          <a-descriptions-item label="所属课程">
            视只觉训练
          </a-descriptions-item>
          <a-descriptions-item label="点评统计">
            1/3
          </a-descriptions-item>
          <a-descriptions-item label="上课内容">
            <div class="flex items-start gap-8px">
              <span>{{ content }}</span>
              <span class="text-#06f cursor-pointer" @click="openModal = true">编辑</span>
            </div>
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tabs">
        <a-tabs
          v-model:active-key="activeKey" size="large" :tab-bar-style="{
            'border-radius': '0px', 'padding-left': '24px',
          }"
        >
          <a-tab-pane key="0" tab="已点评（1）">
            <div class="p-20px flex gap-20px">
              <div class="w-300px ">
                <div class="gap-8px flex items-center p-20px bg-white rounded-15px">
                  <div>
                    <img
                      width="46" height="46" class="mr-2" style="border-radius: 100%;"
                      src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120"
                      alt=""
                    >
                  </div>
                  <div class="flex flex-1 gap-10px">
                    <div class="flex  flex-col">
                      <span>lily</span>
                      <span class="text-#888">家长：未读</span>
                    </div>
                    <div>
                      <span class="bg-#e6f0ff text-#06f text-12px text-3 px-3 py-2px rounded-10 mr-10px">到课</span>
                      <span class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10">未记录</span>
                    </div>
                  </div>
                </div>
                <div class="gap-8px flex items-center p-20px cursor-pointer rounded-15px mt-15px">
                  <div>
                    <img
                      width="46" height="46" class="mr-2" style="border-radius: 100%;"
                      src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120"
                      alt=""
                    >
                  </div>
                  <div class="flex flex-1 gap-10px">
                    <div class="flex  flex-col">
                      <span>lily</span>
                      <span class="text-#888">家长：未读</span>
                    </div>
                    <div>
                      <span class="bg-#e6f0ff text-#06f text-12px text-3 px-3 py-2px rounded-10 mr-10px">到课</span>
                      <span class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10">未记录</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="flex-1">
                <a-card title="评价详情">
                  <template #extra>
                    <a-button>
                      编辑
                    </a-button>
                  </template>
                  <!-- 点评人 -->
                  <div class="flex gap-15px items-center">
                    <div
                      class="text-16px text-white flex items-center justify-center w-48px h-48px rounded-50% bg-#005ce6"
                    >
                      何
                    </div>
                    <div>
                      <div class="font-550 text-15px">
                        何老师
                      </div>
                      <div class="text-13px text-#888">
                        点评时间：2022-03-15 10:00
                      </div>
                    </div>
                  </div>
                  <!-- 点评分 -->
                  <div class="flex items-center gap-10px my-20px">
                    <div class="flex items-center">
                      <div class="pentagram active" />
                      <div class="pentagram active" />
                      <div class="pentagram active" />
                      <div class="pentagram active" />
                      <div class="pentagram" />
                    </div>
                    <div class="text-#ff9400 text-14px">
                      表现很棒，继续保持！
                    </div>
                  </div>
                  <!-- 点评内容 -->
                  <div class="text-14px">
                    今天在视知觉训练课上，查查表现得非常积极！
                    他一直在努力参与“发发发”的游戏，眼神专注，动作迅速，看得出他对这个活动非常感兴趣。
                    不过，查查在过程中有点着急，不停地喊“给我给我给我”，可能是太想拿到旺旺了。
                    老师已经提醒他，要耐心等待，按顺序来，这样大家都能玩得更开心。
                    查查也慢慢学会了控制自己的情绪，开始配合老师的指令。
                    希望下次他能继续保持这种热情，同时也能更好地遵守课堂规则，相信他会越来越棒的！
                  </div>
                  <!-- 评论 -->
                  <div class="bg-#fafafa rounded-15px p-15px">
                    <div class="text-#888 text-16px font-500 family">
                      共 {{ comments.length }} 条评论
                    </div>
                    <div v-for="(comment, index) in comments" :key="index" class="flex gap-15px items-center mt-15px">
                      <div
                        class="text-14px text-white flex items-center justify-center w-38px h-38px rounded-50% bg-#005ce6"
                      >
                        何
                      </div>
                      <div class="flex flex-col">
                        <div class="flex items-center gap-5px text-14px text-#888">
                          <div>
                            {{ comment.name }}
                          </div>
                          <div class="text-12px">
                            {{ comment.time }}
                          </div>
                          <div>
                            <a-popconfirm
                              title="确认删除次评论？" ok-text="删除" cancel-text="取消"
                              @confirm="handelDeleteComment(index)"
                            >
                              <a-button type="link" class=" text-#06f">
                                删除
                              </a-button>
                            </a-popconfirm>
                          </div>
                        </div>
                        <div>{{ comment.content }}</div>
                      </div>
                    </div>
                    <!-- 发布评论 -->
                    <div class="flex items-start justify-between gap-10px mt-20px">
                      <div
                        class="text-14px text-white flex items-center justify-center w-38px h-38px rounded-50% bg-#005ce6"
                      >
                        何
                      </div>
                      <a-textarea
                        v-model:value="commentContent" class="flex-1" :autosize="{ maxRows: 6 }"
                        :maxlength="300" :show-count="true" style="height: 98px;min-height: 98px;"
                        placeholder="输入评论内容..."
                      />
                      <a-button :disabled="!commentContent" type="primary" @click="handleCommentSubmit">
                        发布
                      </a-button>
                    </div>
                  </div>
                </a-card>
              </div>
            </div>
          </a-tab-pane>
          <a-tab-pane key="1" tab="待点评（2）">
            <div class="bg-white rounded-15px p-20px m-15px">
              <custom-title class="mb-12px" title="共3位待点评学员" font-size="15px" font-weight="400">
                <template #right>
                  <a-tooltip>
                    <template #title>
                      未勾选学员时，直接点击批量点评将会仅默认选中到课学员
                    </template>
                    <a-button type="primary" class="px-10px">
                      <InfoCircleOutlined />
                      批量点评
                    </a-button>
                  </a-tooltip>
                </template>
              </custom-title>
              <a-table
                :pagination="false"
                :row-selection="{ selectedRowKeys: selectedRowKeysState.selectedRowKeys, onChange: onSelectChange }"
                :columns="columns" :data-source="data"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'name'">
                    <div class="flex items-center gap-5px">
                      <student-avatar :auto-width="false" :name="record.name" :show-gender="false" :show-age="false" />
                      <span class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10">{{ '未关注' }}</span>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <span
                      v-if="record.status === '0'"
                      class="bg-#e6f0ff text-#06f text-12px text-3 px-3 py-2px rounded-10"
                    >到课</span>
                    <span
                      v-if="record.status === '1'"
                      class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10"
                    >请假</span>
                    <span
                      v-if="record.status === '2'"
                      class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10"
                    >旷课</span>
                    <span
                      v-if="record.status === '3'"
                      class="bg-#eee text-#888 text-12px text-3 px-3 py-2px rounded-10"
                    >未记录</span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <a-button type="link" class="text-14px text-#06f" @click="reviewDrawerOpen = true">
                      去点评
                    </a-button>
                  </template>
                </template>
              </a-table>
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-drawer>
    <a-modal
      v-model:open="openModal" :mask-closable="false" :keyboard="false" width="550px"
      @ok="content = modelValue; openModal = false" @cancel="modelValue = content"
    >
      <template #title>
        <span class="font-400 text-15px">编辑上课内容</span>
      </template>
      <div class="p-10px flex items-start">
        <span class="text-#888">上课内容：</span>
        <a-textarea
          v-model:value="modelValue" placeholder="选填（1000字以内）" style="flex:1;height: 120px;" show-count
          :maxlength="1000"
        />
      </div>
    </a-modal>
    <reviewDrawer v-model="reviewDrawerOpen" />
  </div>
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

.tabs {
  width: 100%;
  border-radius: 10px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0px !important;

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

.pentagram {
  width: 38px;
  height: 38px;
  background-color: #fff;
  background-image: url('https://pcsys.admin.ybc365.com/9930332d-53aa-4748-b4eb-ff3c6808832c.png');
  background-size: cover;
}

.pentagram.active {
  background-image: url('https://pcsys.admin.ybc365.com/2571c832-819d-47df-9d1a-c471ac9ade7f.png')
}

.family {
  font-family: 'PingFangSC-Medium, PingFang SC, sans-serif;'
}
</style>
