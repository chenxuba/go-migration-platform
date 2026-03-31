<script setup>
import { CloseOutlined, QuestionCircleOutlined } from "@ant-design/icons-vue";
import { debounce } from "lodash-es";
import { getCoursePageApi } from "~/api/edu-center/course-list";
import { pageComposeLessonsForPcApi } from "~/api/edu-center/compose-lesson";
import CreateCombinedCourseModal from "./create-combined-course-modal.vue";
import StaffSelect from "./staff-select.vue";
import messageService from "~/utils/messageService";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update:open"]);
const formRef = ref();
const combinedCourseModalOpen = ref(false);
const composeLessonOptions = ref([]);
const composeListLoading = ref(false);
const composeMoreLoading = ref(false);
const composeSearchKeyword = ref("");
const composePagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});
const composeFinished = ref(false);

function mapComposeListItem(it) {
  return {
    label:
      it.productCount != null
        ? `${it.name}（${it.productCount}门课）`
        : it.name,
    value: String(it.id),
  };
}

async function getComposeLessonListPage() {
  if (composeListLoading.value || composeMoreLoading.value)
    return;

  const pageIdx = composePagination.value.current;
  const isFirst = pageIdx === 1;

  if (isFirst)
    composeListLoading.value = true;
  else
    composeMoreLoading.value = true;

  try {
    const res = await pageComposeLessonsForPcApi({
      queryModel: { searchKey: composeSearchKeyword.value || "" },
      pageRequestModel: {
        needTotal: true,
        skipCount: 0,
        pageSize: composePagination.value.pageSize,
        pageIndex: pageIdx,
      },
    });
    if (res.code === 200 && Array.isArray(res.result?.list)) {
      const list = res.result.list;
      const mapped = list.map(mapComposeListItem);
      if (isFirst) {
        composeLessonOptions.value = mapped;
      }
      else {
        const existing = new Set(composeLessonOptions.value.map((o) => o.value));
        const extra = mapped.filter((o) => !existing.has(o.value));
        composeLessonOptions.value = [...composeLessonOptions.value, ...extra];
      }
      composePagination.value.total = Number(res.result.total ?? 0);
      const reachedEndByTotal =
        composePagination.value.total > 0
        && composeLessonOptions.value.length >= composePagination.value.total;
      const reachedEndByPage = list.length < composePagination.value.pageSize;
      composeFinished.value =
        composePagination.value.total === 0 || reachedEndByTotal || reachedEndByPage;
    }
    else {
      composeLessonOptions.value = [];
      composeFinished.value = true;
      if (res.code !== 200 && res.message)
        messageService.error(res.message);
    }
  }
  catch (e) {
    if (pageIdx > 1)
      composePagination.value.current -= 1;
    else
      composeLessonOptions.value = [];
    messageService.error(e?.message || "加载组合课程失败");
  }
  finally {
    composeListLoading.value = false;
    composeMoreLoading.value = false;
  }
}

const debouncedFetchComposeList = debounce(() => {
  composePagination.value.current = 1;
  composeFinished.value = false;
  if (composeSearchKeyword.value?.trim()) {
    composeLessonOptions.value = [];
  }
  getComposeLessonListPage();
}, 400);

function onComposeDropdownVisible(open) {
  if (open) {
    composePagination.value.current = 1;
    composeFinished.value = false;
    getComposeLessonListPage();
  }
}

function onComposeSearch(value) {
  composeSearchKeyword.value = value;
  debouncedFetchComposeList();
}

function onComposePopupScroll(event) {
  const { target } = event;
  const { scrollTop, scrollHeight, clientHeight } = target;
  if (scrollHeight - scrollTop - clientHeight >= 12)
    return;
  if (composeListLoading.value || composeMoreLoading.value || composeFinished.value)
    return;
  composePagination.value.current += 1;
  getComposeLessonListPage();
}

function onComposeLessonCreated() {
  composePagination.value.current = 1;
  composeFinished.value = false;
  getComposeLessonListPage();
}

/** 班级授课（班课）课程下拉，与选择课程范围弹窗 queryModel 一致；分页 + 滚动加载 */
const singleCourseOptions = ref([]);
const singleCourseLoading = ref(false);
const singleCourseMoreLoading = ref(false);
const singleCourseSearchKeyword = ref("");
const singleCoursePagination = ref({
  current: 1,
  pageSize: 30,
  total: 0,
});
const singleCourseFinished = ref(false);

async function getSingleCourseListPage() {
  if (singleCourseLoading.value || singleCourseMoreLoading.value)
    return;

  const pageIdx = singleCoursePagination.value.current;
  const isFirst = pageIdx === 1;

  if (isFirst)
    singleCourseLoading.value = true;
  else
    singleCourseMoreLoading.value = true;

  try {
    const res = await getCoursePageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: singleCoursePagination.value.pageSize,
        pageIndex: pageIdx,
      },
      sortModel: {
        byTotalSales: 0,
        byUpdateTime: 0,
      },
      queryModel: {
        searchKey: singleCourseSearchKeyword.value || "",
        delFlag: false,
        saleStatus: 1,
        teachMethod: 1,
        courseType: 1,
      },
    });
    if (res.code === 200) {
      const list = res.result || [];
      const mapped = list.map((item) => ({
        label: item.name || item.title || String(item.id),
        value: String(item.id),
      }));
      if (isFirst) {
        singleCourseOptions.value = mapped;
      }
      else {
        const existing = new Set(singleCourseOptions.value.map((o) => o.value));
        const extra = mapped.filter((o) => !existing.has(o.value));
        singleCourseOptions.value = [...singleCourseOptions.value, ...extra];
      }
      singleCoursePagination.value.total = Number(res.total || 0);
      const reachedEndByTotal =
        singleCoursePagination.value.total > 0
        && singleCourseOptions.value.length >= singleCoursePagination.value.total;
      const reachedEndByPage = list.length < singleCoursePagination.value.pageSize;
      singleCourseFinished.value =
        singleCoursePagination.value.total === 0 || reachedEndByTotal || reachedEndByPage;
    }
    else {
      singleCourseOptions.value = [];
      singleCourseFinished.value = true;
      if (res.message)
        messageService.error(res.message);
    }
  }
  catch (e) {
    if (pageIdx > 1)
      singleCoursePagination.value.current -= 1;
    else
      singleCourseOptions.value = [];
    messageService.error(e?.message || "加载课程失败");
  }
  finally {
    singleCourseLoading.value = false;
    singleCourseMoreLoading.value = false;
  }
}

const debouncedFetchSingleCourseList = debounce(() => {
  singleCoursePagination.value.current = 1;
  singleCourseFinished.value = false;
  if (singleCourseSearchKeyword.value?.trim()) {
    singleCourseOptions.value = [];
  }
  getSingleCourseListPage();
}, 400);

function onSingleCourseDropdownVisible(open) {
  if (open) {
    singleCoursePagination.value.current = 1;
    singleCourseFinished.value = false;
    getSingleCourseListPage();
  }
}

function onSingleCourseSearch(value) {
  singleCourseSearchKeyword.value = value;
  debouncedFetchSingleCourseList();
}

function onSingleCoursePopupScroll(event) {
  const { target } = event;
  const { scrollTop, scrollHeight, clientHeight } = target;
  if (scrollHeight - scrollTop - clientHeight >= 12)
    return;
  if (singleCourseLoading.value || singleCourseMoreLoading.value || singleCourseFinished.value)
    return;
  singleCoursePagination.value.current += 1;
  getSingleCourseListPage();
}

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: (value) => emit("update:open", value),
});
const formState = reactive({
  mode: "1",
  course: undefined,
  totalCourse: undefined,
  className: undefined,
  defaultClassTimeRecordMode: 1,
  defaultStudentClassTime: 1,
  defaultTeacherClassTime: 0,
  maxNum: undefined,
  teacher: [],
  classRoom: undefined,
  remark: "",
});

const classTimeUnitLabel = computed(() =>
  Number(formState.defaultClassTimeRecordMode) === 2 ? "课时/小时" : "课时",
);

const classTimeHint = computed(() =>
  Number(formState.defaultClassTimeRecordMode) === 2
    ? "每次点名，学员和上课教师记录的课时会根据日程时长自动计算课时（点名时支持调整）"
    : "每次点名，学员和上课教师记录的课时数默认为此数值（点名时支持调整）",
);
// 手动触发验证
async function handleSubmit() {
  try {
    await formRef.value.validate(); // 关键3：通过引用调用验证方法
    console.log("验证通过，提交数据:", formState);
  } catch (error) {
    console.log("验证失败:", error);
  }
}
function closeFun() {
  formRef.value.resetFields();
  openModal.value = false;
}
</script>

<template>
  <div class="create-class-modals-root">
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>创建班级</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form
        ref="formRef"
        :model="formState"
        :label-col="{ span: 4 }"
        :wrapper-col="{ span: 10 }"
      >
        <!-- 设置模式  单选框  课程 组合课程 -->
        <a-form-item
          label="设置模式"
          name="mode"
          :rules="[{ required: true, message: '请选择设置模式' }]"
        >
          <a-radio-group v-model:value="formState.mode" class="custom-radio">
            <a-space :size="100">
              <a-radio value="1">
                <a-popover title="课程">
                  <template #content>
                    <div class="w-220px">
                      设置后，该课程下的学员可在同一班级上课
                    </div>
                  </template>
                  课程
                  <QuestionCircleOutlined />
                </a-popover>
              </a-radio>
              <a-radio value="2">
                <a-popover title="组合课程">
                  <template #content>
                    <div class="w-220px">
                      设置后，该组合课程范围内，多个课程的对应学员可在同一班级上课
                    </div>
                  </template>
                  组合课程
                  <QuestionCircleOutlined />
                </a-popover>
              </a-radio>
            </a-space>
          </a-radio-group>
        </a-form-item>

        <!-- 选择课程 -->
        <a-form-item
          v-if="formState.mode === '1'"
          label="选择课程"
          name="course"
          :rules="[{ required: true, message: '请选择课程' }]"
        >
          <a-select
            v-model:value="formState.course"
            show-search
            :filter-option="false"
            :loading="singleCourseLoading"
            placeholder="请选择课程"
            option-filter-prop="label"
            @dropdown-visible-change="onSingleCourseDropdownVisible"
            @search="onSingleCourseSearch"
            @popup-scroll="onSingleCoursePopupScroll"
          >
            <a-select-option
              v-for="opt in singleCourseOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </a-select-option>
            <a-select-option
              v-if="singleCourseMoreLoading"
              key="__single_course_loading_more__"
              disabled
            >
              <div class="text-center text-#999 text-12px py-1">
                加载中…
              </div>
            </a-select-option>
            <a-select-option
              v-else-if="singleCourseFinished && singleCourseOptions.length > 0"
              key="__single_course_no_more__"
              disabled
            >
              <div class="text-center text-#999 text-12px py-1">
                没有更多了
              </div>
            </a-select-option>
          </a-select>
        </a-form-item>
        <!-- 选择组合课程 -->
        <a-form-item
          v-if="formState.mode === '2'"
          label="选择组合课程"
          name="totalCourse"
          :rules="[{ required: true, message: '请选择组合课程' }]"
        >
          <div class="flex-items-center flex w-420px">
            <a-select
              v-model:value="formState.totalCourse"
              class="flex-1 min-w-0"
              show-search
              :filter-option="false"
              :loading="composeListLoading"
              placeholder="请选择组合课程"
              option-filter-prop="label"
              @dropdown-visible-change="onComposeDropdownVisible"
              @search="onComposeSearch"
              @popup-scroll="onComposePopupScroll"
            >
              <a-select-option
                v-for="opt in composeLessonOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </a-select-option>
              <a-select-option
                v-if="composeMoreLoading"
                key="__compose_loading_more__"
                disabled
              >
                <div class="text-center text-#999 text-12px py-1">
                  加载中…
                </div>
              </a-select-option>
              <a-select-option
                v-else-if="composeFinished && composeLessonOptions.length > 0"
                key="__compose_no_more__"
                disabled
              >
                <div class="text-center text-#999 text-12px py-1">
                  没有更多了
                </div>
              </a-select-option>
            </a-select>
            <span class="whitespace-nowrap">
              <a-button
                type="link"
                class="text-3.5"
                @click="combinedCourseModalOpen = true"
              >
                设置组合课
              </a-button>
            </span>
          </div>
        </a-form-item>
        <a-form-item
          name="className"
          :rules="[{ required: true, message: '请输入班级名称' }]"
          class="custom-form-item"
        >
          <template #label>
            <div>
              <img
                class="w-48px h-48px"
                src="https://pcsys.admin.ybc365.com/c6221215-3203-4563-bbe9-5f4d8ffcd2a1.png"
                alt=""
              />
            </div>
          </template>
          <a-input
            v-model:value="formState.className"
            placeholder="请输入班级名称"
          />
        </a-form-item>
        <!-- 满班人数 数字选择器 -->
        <a-form-item label="满班人数" name="maxNum">
          <a-input-number
            v-model:value="formState.maxNum"
            placeholder="不限"
            class="w-160px"
          />
        </a-form-item>
        <a-form-item label="班主任" name="teacher">
          <StaffSelect
            v-model="formState.teacher"
            placeholder="请选择班主任"
            :width="'100%'"
            :multiple="true"
            :status="0"
            :allow-clear="true"
          />
        </a-form-item>
        <!-- 上课教室 -->
        <a-form-item label="上课教室" name="classRoom">
          <a-select
            v-model:value="formState.classRoom"
            placeholder="请选择上课教室"
          >
            <a-select-option value="1"> 教室1 </a-select-option>
            <a-select-option value="2"> 教室2 </a-select-option>
          </a-select>
        </a-form-item>
        <!-- 与一对一编辑弹窗一致的课时记录方式、默认记录课时 -->
        <a-form-item
          label="课时记录方式"
          name="defaultClassTimeRecordMode"
          :rules="[{ required: true, message: '请选择课时记录方式' }]"
        >
          <a-radio-group
            v-model:value="formState.defaultClassTimeRecordMode"
            class="custom-radio"
          >
            <a-radio :value="1"> 按固定课时记录 </a-radio>
            <a-radio :value="2"> 按上课时长记录 </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item
          label="默认记录课时"
          required
          :wrapper-col="{ span: 20 }"
        >
          <div class="one-to-one-class-time-inputs">
            <span class="one-to-one-ct-group">
              <span>学员</span>
              <a-form-item
                name="defaultStudentClassTime"
                class="mb-0 create-class-nested-fi"
                :rules="[{ required: true, message: '请输入学员课时' }]"
              >
                <a-input-number
                  v-model:value="formState.defaultStudentClassTime"
                  :min="0"
                  :precision="2"
                  style="width: 100px"
                />
              </a-form-item>
              <span class="one-to-one-ct-unit">{{ classTimeUnitLabel }}</span>
            </span>
            <span class="one-to-one-ct-group">
              <span>上课教师课时</span>
              <a-form-item
                name="defaultTeacherClassTime"
                class="mb-0 create-class-nested-fi"
                :rules="[{ required: true, message: '请输入教师课时' }]"
              >
                <a-input-number
                  v-model:value="formState.defaultTeacherClassTime"
                  :min="0"
                  :precision="2"
                  style="width: 100px"
                />
              </a-form-item>
              <span class="one-to-one-ct-unit">{{ classTimeUnitLabel }}</span>
            </span>
          </div>
          <div class="create-class-class-time-hint">
            {{ classTimeHint }}
          </div>
        </a-form-item>
        <a-form-item label="备注" name="remark">
          <a-input
            v-model:value="formState.remark"
            placeholder="请输入"
            class="w-450px"
          />
        </a-form-item>
      </a-form>
    </div>
    <template #footer>
      <a-button @click="closeFun"> 关闭 </a-button>
      <!-- 警告 -->
      <a-button @click="closeFun"> 保存并下一个 </a-button>
      <a-button type="primary" @click="handleSubmit"> 确定 </a-button>
    </template>
  </a-modal>

  <CreateCombinedCourseModal
    v-model:open="combinedCourseModalOpen"
    @created="onComposeLessonCreated"
  />
  </div>
</template>

<style lang="less" scoped>
.create-class-modals-root {
  display: contents;
}

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

.contenter {
  padding: 24px;
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

:deep(
    :where(.css-dev-only-do-not-override-1mphclt).ant-form-item.custom-form-item
      .ant-form-item-label
      > label
  ) {
  height: 48px !important;
}

.custom-form-item :deep(.ant-form-item-row) {
  // display: flex;
  // align-items: center;
  .ant-input {
    margin-top: 7px;
  }
}

.custom-form-item :deep(.ant-form-item-required) {
  &::after {
    opacity: 0;
  }
}

/* 默认记录课时：单行横排，与加宽的 wrapper-col 一起撑开 */
.one-to-one-class-time-inputs {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  column-gap: 16px;
  width: 100%;
  min-width: 0;
}

.one-to-one-ct-group {
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  flex-shrink: 0;
  gap: 8px;
}

.one-to-one-ct-unit {
  flex-shrink: 0;
  white-space: nowrap;
}

.create-class-nested-fi :deep(.ant-form-item-row) {
  display: inline-flex;
  margin-bottom: 0;
}

.create-class-class-time-hint {
  color: #888;
  font-size: 13px;
  margin-top: 6px;
  line-height: 1.5;
  white-space: normal;
  word-break: break-word;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
