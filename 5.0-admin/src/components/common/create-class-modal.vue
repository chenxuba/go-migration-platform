<script setup>
import { CloseOutlined, QuestionCircleOutlined } from "@ant-design/icons-vue";
import CreateCombinedCourseModal from "./create-combined-course-modal.vue";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update:open"]);
const formRef = ref();
const combinedCourseModalOpen = ref(false);
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
  teacher: undefined,
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
          <a-select v-model:value="formState.course" placeholder="请选择课程">
            <a-select-option value="1"> 课程1 </a-select-option>
            <a-select-option value="2"> 课程2 </a-select-option>
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
              placeholder="请选择组合课程"
            >
              <a-select-option value="1"> 组合课程1 </a-select-option>
              <a-select-option value="2"> 组合课程2 </a-select-option>
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
        <!-- 班主任 -->
        <a-form-item label="班主任" name="teacher">
          <a-select
            v-model:value="formState.teacher"
            placeholder="请选择班主任"
          >
            <a-select-option value="1"> 班主任1 </a-select-option>
            <a-select-option value="2"> 班主任2 </a-select-option>
          </a-select>
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

  <CreateCombinedCourseModal v-model:open="combinedCourseModalOpen" />
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
