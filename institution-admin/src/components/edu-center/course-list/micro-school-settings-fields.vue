<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import CustomTitle from '@/components/common/custom-title.vue'

const props = defineProps({
  settingFormState: {
    type: Object,
    required: true,
  },
  settingsTitle: {
    type: String,
    default: '微校课程设置',
  },
  showMicroSale: {
    type: Boolean,
    default: false,
  },
  microShowPopover: {
    type: String,
    default: '关闭后，学员将无法在微校看到此课程，但仍可通过分享购买',
  },
  microSalePopover: {
    type: String,
    default: '开启后，学员能够在微校中查看并购买此课程，或通过分享链接购买此课程',
  },
  purchaseTargetLabel: {
    type: String,
    default: '课程',
  },
  showError: {
    type: Boolean,
    default: false,
  },
  courseListOptions: {
    type: Array,
    default: () => [],
  },
  courseListLoading: {
    type: Boolean,
    default: false,
  },
  courseListPagination: {
    type: Object,
    default: () => ({ hasMore: false, total: 0 }),
  },
})

const emit = defineEmits([
  'search-course',
  'course-dropdown-visible-change',
  'course-popup-scroll',
  'course-selection-change',
  'load-more-courses',
])
</script>

<template>
  <CustomTitle :title="settingsTitle" font-size="16px" font-weight="500" class="mb-24px mt-24px" />
  <a-form-item label="微校展示：" name="isShow">
    <div class="flex items-center">
      <a-switch v-model:checked="settingFormState.isShow" />
      <a-popover :content="microShowPopover" title="微校展示">
        <ExclamationCircleOutlined class="ml-6px" />
      </a-popover>
    </div>
  </a-form-item>
  <a-form-item v-if="showMicroSale" label="微校售卖：" name="isOnlineSale">
    <div class="flex items-center">
      <a-switch v-model:checked="settingFormState.isOnlineSale" />
      <a-popover title="微校售卖">
        <template #content>
          {{ microSalePopover }}
        </template>
        <ExclamationCircleOutlined class="ml-6px" />
      </a-popover>
    </div>
  </a-form-item>
  <a-form-item label="购买限制：" name="buyLimit">
    <a-switch v-model:checked="settingFormState.buyLimit" />
  </a-form-item>
  <div v-if="settingFormState.buyLimit">
    <a-form-item label="允许老生购买：" name="oldBuy">
      <div class="flex items-center">
        <a-switch v-model:checked="settingFormState.oldBuy" />
        <a-popover title="允许老生购买">
          <template #content>
            <div class="w-440px">
              <div>老生：在读、历史学员。</div>
              <div>
                如果选择任意课程的在读学员，则机构所有在读学员均可购买此{{ purchaseTargetLabel }}；
                <div>
                  如果选择 A、B 课程的在读学员和历史学员，则机构只有购买过 A、B 课程的学员可购买此{{ purchaseTargetLabel }}。
                </div>
              </div>
            </div>
          </template>
          <ExclamationCircleOutlined class="ml-6px" />
        </a-popover>
      </div>
    </a-form-item>
    <a-form-item v-if="settingFormState.oldBuy" label="允许：" name="allowType">
      <div class="flex flex-col mt-5px">
        <a-form-item-rest>
          <a-radio-group v-model:value="settingFormState.allowType" class="custom-radio">
            <a-radio :value="1">
              任意课程
            </a-radio>
            <a-radio :value="2">
              部分课程
            </a-radio>
          </a-radio-group>
        </a-form-item-rest>
        <div class="flex items-center mt-5px">
          <a-form-item-rest>
            <a-tag v-if="settingFormState.allowType == 1"
              class="w-70px h-28px flex flex-items-center text-14px text-#888">
              任意课程
            </a-tag>
            <a-select v-else v-model:value="settingFormState.courseListIds"
              :status="settingFormState.courseListIds.length == 0 && showError && settingFormState.allowType == 2 ? 'error' : ''"
              mode="multiple" style="width: 380px;" class="mr-8px" placeholder="请选择课程（可多选）" :max-tag-count="2"
              :loading="courseListLoading" show-search :filter-option="false" @search="$emit('search-course', $event)"
              @dropdown-visible-change="(open) => $emit('course-dropdown-visible-change', open)"
              @popup-scroll="$emit('course-popup-scroll', $event)"
              @change="$emit('course-selection-change', $event)">
              <a-select-option v-for="course in courseListOptions" :key="course.id" :value="course.id"
                :label="course.name">
                {{ course.name }}
              </a-select-option>
              <template v-if="courseListLoading && courseListOptions.length === 0" #notFoundContent>
                <div class="text-center py-2">
                  <a-spin size="small" /> 加载中...
                </div>
              </template>
              <a-select-option
                v-if="courseListPagination.hasMore && !courseListLoading && courseListOptions.length > 0"
                :value="'load-more'" :disabled="true" class="text-center">
                <div class="py-1 text-gray-500 text-sm cursor-pointer hover:bg-gray-50"
                  @click.stop="$emit('load-more-courses')">
                  点击加载更多 ({{ courseListPagination.total - courseListOptions.length }} 条)
                </div>
              </a-select-option>
              <a-select-option v-if="courseListLoading && courseListOptions.length > 0"
                :value="'loading-more'" :disabled="true" class="text-center">
                <div class="py-1 text-gray-500 text-sm">
                  <a-spin size="small" /> 加载中...
                </div>
              </a-select-option>
              <a-select-option
                v-if="!courseListPagination.hasMore && !courseListLoading && courseListOptions.length > 0"
                :value="'no-more'" :disabled="true" class="text-center">
                <div class="py-1 text-gray-500 text-sm">
                  没有更多了
                </div>
              </a-select-option>
            </a-select>
            <span class="mr-6px">的</span>
            <a-select v-model:value="settingFormState.studentStatuses" mode="multiple"
              style="width: auto;min-width: 150px;" placeholder="请选择学员状态">
              <a-select-option :value="1">
                在读学员
              </a-select-option>
              <a-select-option :value="2">
                历史学员
              </a-select-option>
            </a-select>
            <span class="ml-8px whitespace-nowrap">购买</span>
          </a-form-item-rest>
        </div>
      </div>
    </a-form-item>
    <a-form-item label="允许新生购买：" name="newBuy">
      <div class="flex items-center">
        <a-switch v-model:checked="settingFormState.newBuy" />
        <a-popover title="允许新生购买">
          <template #content>
            新生：意向学员（未录入的学员）
          </template>
          <ExclamationCircleOutlined class="ml-6px" />
        </a-popover>
      </div>
    </a-form-item>
    <a-form-item label="限购 1 单：" name="buyOne">
      <div class="flex items-center">
        <a-switch v-model:checked="settingFormState.buyOne" />
        <a-popover title="限购1单">
          <template #content>
            <div class="w-300px">
              开启后，每个学员仅允许在微校购买此{{ purchaseTargetLabel }}一次。 如果某{{ purchaseTargetLabel }}有多个可售项目，学员购买完成后，不能再次重复购买。
            </div>
          </template>
          <ExclamationCircleOutlined class="ml-6px" />
        </a-popover>
      </div>
    </a-form-item>
  </div>
</template>
