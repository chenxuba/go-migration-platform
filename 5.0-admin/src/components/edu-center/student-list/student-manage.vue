<script setup>
import { onMounted } from 'vue'
import { getStudentOverviewStatisticsApi } from '@/api/edu-center/student-list'

const router = useRouter()
const currentType = ref(1)
const totalStudents = ref(0)
const loadingOverview = ref(false)

// 子组件引用
const studyingOrHistoryRef = ref(null)
const pendingFeesRef = ref(null)
const owePriceRef = ref(null)
const waitClassRef = ref(null)
const birthdayStudentRef = ref(null)
const waitFocusRef = ref(null)
const missSchoolRef = ref(null)

const itemsList = ref([
  { type: 1, selected: true, name: '在读学员', num: 2, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/reading-icon.a7ae1ed0.svg' },
  { type: 2, selected: false, name: '历史学员', num: 0, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/history-icon.f39cedb5.svg' },
  { type: 3, selected: false, name: '意向学员', num: 2, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/intention-icon.d6515e60.svg' },
  { type: 4, selected: false, name: '待续费学员', num: 2, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/renew-icon.50f2cafa.svg' },
  { type: 5, selected: false, name: '欠费学员', num: 1, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/arrears-icon.6f028695.svg' },
  { type: 6, selected: false, name: '待分班学员', num: 2, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/streaming-icon.703400f2.svg' },
  { type: 7, selected: false, name: '生日学员', num: 0, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/birthday-icon.7ec2c5c7.svg' },
  { type: 8, selected: false, name: '待关注学员', num: 1, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/follow-icon.4393714e.svg' },
  { type: 9, selected: false, name: '缺课学员', num: 0, icon: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/absentee-icon.5c1240ed.svg' },
].map(item => ({
  ...item,
  selected: item.type === 1, // 使用 map 自动设置选中状态
})))

function applyOverviewStatistics(data = {}) {
  totalStudents.value = Number(data.totalStudents || 0)
  const countMap = {
    1: Number(data.readingStudents || 0),
    2: Number(data.historyStudents || 0),
    3: Number(data.intentStudents || 0),
    4: Number(data.pendingRenewalStudents || 0),
    5: Number(data.arrearStudents || 0),
    6: Number(data.pendingClassStudents || 0),
    7: Number(data.birthdayStudents || 0),
    8: Number(data.pendingAttentionStudents || 0),
    9: Number(data.absentStudents || 0),
  }
  itemsList.value = itemsList.value.map(item => ({
    ...item,
    num: countMap[item.type] ?? 0,
  }))
}

async function getOverviewStatistics() {
  loadingOverview.value = true
  try {
    const res = await getStudentOverviewStatisticsApi()
    if (res.code === 200) {
      applyOverviewStatistics(res.data)
    }
  }
  catch (error) {
    console.error('get student overview statistics failed', error)
  }
  finally {
    loadingOverview.value = false
  }
}

function handleSelect(selectedItem) {
  if (selectedItem.type == 3) {
    router.replace({
      path: '/enroll-center/intention-student',
    })
    return
  }
  currentType.value = selectedItem.type
  itemsList.value = itemsList.value.map(item => ({
    ...item,
    selected: item.type === selectedItem.type,
  }))
}

// 注意：不需要在这里监听 currentType 变化并刷新数据
// studying-or-history 组件内部已经通过 watch props.currentType 来处理数据刷新
// 避免重复调用接口

// 刷新当前类型的数据（仅用于手动刷新，不自动触发）
function refreshCurrentTypeData() {
  switch (currentType.value) {
    case 1:
    case 2:
      studyingOrHistoryRef.value?.getEnrolledStudentList()
      break
    case 4:
      pendingFeesRef.value?.getList()
      break
    case 5:
      owePriceRef.value?.getList()
      break
    case 6:
      waitClassRef.value?.getList()
      break
    case 7:
      birthdayStudentRef.value?.getList()
      break
    case 8:
      waitFocusRef.value?.getList()
      break
    case 9:
      missSchoolRef.value?.getList()
      break
  }
}

onMounted(() => {
  getOverviewStatistics()
})

</script>

<template>
  <div class="home">
    <!-- 快捷入口 -->
    <div class="kuaiRun">
      <custom-title title="快捷入口" font-size="14px" font-weight="500" />
      <div class="items-list mt-4">
        <a-space :size="12" wrap>
          <div class="items one ">
            <div class="items-t">
              <span class="text-3.5 text-#222 font-600">学员总数 {{ totalStudents }}</span>
              <span class="text-3 text-#0066ff"> <i class="bg-#e6f0ff font-500">近1个月新增涨幅 -</i> <img
                class="ml--3"
                src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/go-up.5c9cb6eb.svg"
                alt=""
              ></span>
            </div>
            <div class="items-b text-3.25 text-#888 mt-2">
              当前机构（在读+意向+历史）学员数量
            </div>
          </div>
          <div
            v-for="(item, index) in itemsList" :key="index" class="items"
            :class="{ 'selected-item': item.selected }" @click="handleSelect(item)"
          >
            <div class="items-l">
              <img :src="item.icon" alt="">
            </div>
            <div class="items-r">
              <div class="text-#222">
                {{ item.name }}
              </div>
              <div class="text-#000 font-600 text-4 mt-1.5">
                {{ loadingOverview ? '-' : item.num }}
              </div>
            </div>
          </div>
        </a-space>
      </div>
    </div>
    <!-- 学员列表 -->
    <studying-or-history 
      v-if="currentType == 1 || currentType == 2" 
      ref="studyingOrHistoryRef"
      :current-type="currentType"
    />
    <pending-fees v-if="currentType == 4" ref="pendingFeesRef" />
    <owe-price v-if="currentType == 5" ref="owePriceRef" />
    <wait-class v-if="currentType == 6" ref="waitClassRef" />
    <birthday-student v-if="currentType == 7" ref="birthdayStudentRef" />
    <wait-focus v-if="currentType == 8" ref="waitFocusRef" />
    <miss-school v-if="currentType == 9" ref="missSchoolRef" />
  </div>
</template>

<style lang="less" scoped>
.home {
  .kuaiRun {
    padding: 12px 24px;
    background: #fff;
    border-radius: 0 0 16px 16px;

    .items-list {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      margin-bottom: 16px;
      .items {
        height: 80px;
        border-radius: 14px;
        background: #fbfcff;
        cursor: pointer;
        width: 160px;
        padding: 16px 8px;
        display: flex;

        &:hover {
          background: #e6f0ff;
        }

        i {
          font-style: normal;
          font-weight: bold;
          background: #e6f0ff;
          padding: 3px 8px;
          border-radius: 12px;
        }

        img {
          width: 30px;
          margin-right: 5px;
        }
      }

      /* 添加选中状态样式 */
      .selected-item {
        background: #e6f0ff;
      }

      .items.one {
        width: 332px;
        padding: 16px 24px;
        flex-direction: column;

        .items-t {
          display: flex;
          justify-content: space-between;
          align-items: center;

          img {
            width: 14px;
          }
        }
      }
    }
  }

  // .student-list{

  // }
}
</style>
