<script setup>
const emit = defineEmits(['change'])
const selectItems = ref([
  { title: '课程商品' },
  { title: '学杂费' },
  { title: '教材商品' },
])
const courseList = ref([
  { id: 1, name: '初级言语课' },
  { id: 2, name: '初级认知课' },
  { id: 3, name: '初级感统课' },
])
const openSelect = defineModel({
  default: false,
})
const activeIndex = ref(0)
const activeCourse = ref([])
const displayArray = ref(['subject', 'courseCategory'])

function changeSelectItems(index) {
  activeIndex.value = index
}

function changeSelectCourse(item) {
  emit('change', item)
  openSelect.value = false
}

function handleOk() {
  emit('change', item)
  openSelect.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openSelect" centered wrap-class-name="modal" width="1000px" :body-style="{ padding: 0 }"
    title="选择课程/学杂费/教材商品" :keyboard="false" :mask-closable="false" @ok="handleOk"
  >
    <div class="modal-wrap">
      <div class="modal-left">
        <a-list>
          <a-list-item
            v-for="(item, index) in selectItems" :key="index" :class="{ active: activeIndex === index }"
            @click="changeSelectItems(index)"
          >
            <custom-title
              v-if="activeIndex === index" :title="item.title" font-size="16px"
              :font-weight="activeIndex === index ? '500' : '300'"
            />
            <span v-else class="pl-2.5">{{ item.title }}</span>
            <span class="num">1</span>
          </a-list-item>
        </a-list>
      </div>
      <div class="modal-right">
        <div class="m-r-t">
          <all-filter
            :display-array="displayArray" :is-quick-show="false" search-label="课程名称"
            search-placeholder="请输入课程名称" :is-show-search-input="true" :is-show-search-stu-phone="false"
          />
        </div>
        <div class="m-r-b">
          <a-list>
            <a-list-item
              v-for="(item, index) in courseList" :key="index"
              :class="activeCourse.includes(index) ? 'activeCourse' : ''"
              class="flex flex-items-center justify-between r-item" @click="changeSelectCourse(item)"
            >
              <div class="m-r-b-l pt-1 pb-1">
                <div class="text-4 text-#222 font-500 mb-1">
                  {{ item.name }}
                </div>
                <div>
                  <a-tag v-if="index === 0 || index === 2" color="#0066ff" style="color:#fff;border-radius: 10px;">
                    通用课
                  </a-tag>
                  <a-tag color="#e6f0ff" style="color:#0066ff;border-radius: 10px;">
                    全部课程
                  </a-tag>
                  <a-tag color="#e6f0ff" style="color:#0066ff;border-radius: 10px;">
                    课时
                  </a-tag>
                </div>
              </div>
              <div class="m-r-b-r pt-1 pb-1 select">
                <a v-if="activeCourse.includes(index)" class="active-a">取消选择</a>
                <a v-if="!activeCourse.includes(index)">点击选择</a>
              </div>
            </a-list-item>
          </a-list>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
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

    .m-r-b {
      height: calc(100% - 70px);
      overflow: auto;
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
