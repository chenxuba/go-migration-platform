<script setup>
import { FormOutlined } from '@ant-design/icons-vue'
import { reactive, ref } from 'vue'
import { message } from 'ant-design-vue'
import { getCoursePropertyListApi, updateCoursePropertyEnableApi } from '~@/api/edu-center/course-list'

const state = reactive({
  checked: false,
})
const propertyList = ref([])
const propertyLastList = ref([])
const loading = ref(true)
const switchLoading = ref(false)
const editPropertyModalOpen = ref(false)
async function getCoursePropertyList(oldItem) {
  try {
    const res = await getCoursePropertyListApi()
    if (res.code === 200) {
      if (oldItem) {
        // 根据oldItem的id找出元素 并更新 version
        const index = res.result.findIndex(item => item.id === oldItem.id)
        if (index !== -1) {
          currentEditProperty.value.version = res.result[index].version
        }
      }
      // propertyLastList 取最后一个
      propertyLastList.value = [res.result[res.result.length - 1]]
      // 过滤掉最后一个
      res.result.pop()
      propertyList.value = res.result
      loading.value = false
      switchLoading.value = false
    }
  }
  catch (err) {
    loading.value = false
  }
}
async function updateCoursePropertyEnable(item, oldItem) {
  try {
    switchLoading.value = true
    const res = await updateCoursePropertyEnableApi(item)
    getCoursePropertyList(oldItem)
  }
  catch (err) {
    console.log(err)
  }
}
const currentEditProperty = ref({})
function editProperty(item) {
  currentEditProperty.value = item
  editPropertyModalOpen.value = true
}
async function handleRefreshList(updatedProperty, originalProperty) {
  try {
    // 调用API更新属性
    const res = await updateCoursePropertyEnableApi(updatedProperty)

    if (res.code === 200) {
      // 使用API返回的最新数据更新本地状态（包含新的version）
      const latestProperty = updatedProperty
      currentEditProperty.value = latestProperty

      // 刷新列表
      getCoursePropertyList(originalProperty)
    }
    else {
      // API返回错误状态码，恢复原值
      currentEditProperty.value = {
        ...originalProperty,
        enableOnlineFilter: originalProperty.enableOnlineFilter,
      }
      message.error(res.message || '更新失败，请重试')
    }
  }
  catch (error) {
    console.error('更新在线筛选设置失败:', error)
    // 网络错误或其他异常，恢复原值
    currentEditProperty.value = {
      ...originalProperty,
      enableOnlineFilter: originalProperty.enableOnlineFilter,
    }
    message.error('更新失败，请重试')
  }
}

onMounted(() => {
  getCoursePropertyList()
})
</script>

<template>
  <div class="tab-content">
    <div class="setting" :class="loading ? '' : 'pb3'">
      <custom-title title="课程属性" font-size="18px" font-weight="800" />
      <div class="table-wrap mt-3 mb-3">
        <div class="property-grid">
          <div
            v-for="item in propertyList"
            :key="item.id"
            class="property-card"
          >
            <div class="property-card-head">
              <span class="property-card-title">{{ item.name }}
                <FormOutlined class="property-card-edit" @click="editProperty(item)" />
              </span>
              <a-switch
                v-model:checked="item.enable" :loading="switchLoading"
                class="shrink-0"
                @change="updateCoursePropertyEnable(item)"
              />
            </div>
            <div v-if="(item.options || []).length" class="property-card-body">
              <a-tag
                v-for="opt in item.options"
                :key="opt.id"
                class="option-tag"
                :title="opt.name"
              >
                {{ opt.name }}
              </a-tag>
            </div>
            <div v-else class="property-card-body property-card-body--empty">
              暂无选项，点击图标编辑添加
            </div>
          </div>
          <template v-if="loading">
            <div v-for="i in 5" :key="`sk-${i}`" class="property-card property-card--skeleton skeleton-item" />
          </template>
        </div>
      </div>
    </div>
  </div>
  <!-- <div class="bg-white rounded-3 p2 px5 mt3">
    <div class="table-wrap mt-3 mb-3 flex-items-center flex justify-between whitespace-nowrap">
      <div class="flex flex-items-center overflow-auto scrollbar" v-if="propertyLastList.length > 0">
        <custom-title class="whitespace-nowrap" v-for="item in propertyLastList" :key="item.id" :title="item.name" font-size="18px"
          font-weight="800"></custom-title>
        <a-switch class="ml3 mr2" :loading="switchLoading" @change="updateCoursePropertyEnable(propertyLastList[0])"
          v-model:checked="propertyLastList[0].enable" />
        <span>开启后，可对课程打标签，并在排课、记上课、数据报表中区分，一般适用于综合类课程。</span>
      </div>
      <a-button type="primary" @click="editProperty(propertyLastList[0])">科目设置</a-button>
    </div>
  </div> -->
  <edit-property-modal
    v-model:open="editPropertyModalOpen" :current-edit-property="currentEditProperty"
    @refresh-list="handleRefreshList"
  />
</template>

<style lang="less" scoped>
.tab-content {
  background: #fff;
  border-radius: 12px;
  border-top-right-radius: 0;
  border-top-left-radius: 0;
  padding: 12px 20px;
}

.property-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
  align-items: stretch;
}

.property-card {
  display: flex;
  flex-direction: column;
  min-height: 120px;
  height: 100%;
  padding: 12px 14px;
  background: #f6f7f8;
  border: 1px solid #ebebeb;
  border-radius: 8px;
  box-sizing: border-box;
}

.property-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-shrink: 0;
}

.property-card-title {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  font-size: 14px;
  font-weight: 500;
  color: #222;
}

.property-card-edit {
  margin-left: 6px;
  color: #06f;
  font-size: 16px;
  cursor: pointer;
  flex-shrink: 0;
}

.property-card-body {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 8px 8px;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e8e8e8;
  min-height: 0;
}

.property-card-body--empty {
  font-size: 12px;
  color: #999;
  line-height: 1.5;
  align-content: flex-start;
}

.property-card--skeleton {
  min-height: 120px;
}

:deep(.option-tag) {
  max-width: min(100%, 200px);
  margin: 0 !important;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}
</style>
