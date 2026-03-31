<script setup>
import { MenuOutlined } from "@ant-design/icons-vue";
import Sortable from "sortablejs";
import SelectCourseRangeModal from "~/components/edu-center/course-list/selectCourseRangeModal.vue";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:open"]);

const openModel = computed({
  get: () => props.open,
  set: (v) => emit("update:open", v),
});

const formRef = ref();
const selectCourseModalOpen = ref(false);
/** 已选课程（顺序即组合内顺序），与选择弹窗 confirm 结构一致 */
const selectedCourseRows = ref([]);
const sortableTbodyRef = ref(null);
let sortableInstance = null;

const formState = reactive({
  name: "",
  courseIds: [],
});

function syncCourseIds() {
  formState.courseIds = selectedCourseRows.value
    .map((r) => r.id)
    .filter((id) => id != null);
}

function destroySortable() {
  if (sortableInstance) {
    sortableInstance.destroy();
    sortableInstance = null;
  }
}

function initSortable() {
  destroySortable();
  nextTick(() => {
    const el = sortableTbodyRef.value;
    if (!el)
      return;
    sortableInstance = Sortable.create(el, {
      handle: ".drag-handle",
      animation: 150,
      ghostClass: "sortable-ghost-row",
      onEnd: (evt) => {
        const { newIndex, oldIndex } = evt;
        if (
          newIndex === undefined
          || oldIndex === undefined
          || newIndex === oldIndex
        ) {
          return;
        }
        const list = [...selectedCourseRows.value];
        const [moved] = list.splice(oldIndex, 1);
        list.splice(newIndex, 0, moved);
        selectedCourseRows.value = list;
        syncCourseIds();
      },
    });
  });
}

watch(
  () => [props.open, selectedCourseRows.value.length],
  () => {
    if (!props.open) {
      destroySortable();
      return;
    }
    if (selectedCourseRows.value.length > 0)
      initSortable();
    else
      destroySortable();
  },
  { flush: "post" },
);

onBeforeUnmount(() => destroySortable());

watch(
  () => props.open,
  (visible) => {
    if (visible) {
      formState.name = "";
      selectedCourseRows.value = [];
      formState.courseIds = [];
      nextTick(() => formRef.value?.resetFields());
    }
  },
);

function openSelectCourseModal() {
  selectCourseModalOpen.value = true;
}

function onSelectCourseConfirm(courses) {
  selectedCourseRows.value = (courses || []).map((c) => ({
    ...c,
    key: c.key != null ? String(c.key) : String(c.id),
    title: c.title || c.name,
  }));
  syncCourseIds();
  nextTick(() => {
    formRef.value?.validateFields(["courseIds"]).catch(() => {});
  });
}

function removeCourse(index) {
  selectedCourseRows.value.splice(index, 1);
  syncCourseIds();
  formRef.value?.validateFields(["courseIds"]).catch(() => {});
}

function close() {
  openModel.value = false;
}

async function handleOk() {
  try {
    await formRef.value?.validate();
    close();
  }
  catch {
    /* 校验失败保留在弹窗 */
  }
}

function handleCancel() {
  close();
}
</script>

<template>
  <div class="create-combined-course-modals">
    <a-modal
      v-model:open="openModel"
      centered
      title="创建组合课程"
      class="modal-content-box create-combined-course-modal"
      :width="800"
      destroy-on-close
      ok-text="确定"
      cancel-text="取消"
      wrap-class-name="create-combined-course-modal-wrap"
      @ok="handleOk"
      @cancel="handleCancel"
    >
      <div class="combined-course-modal-body contenter">
        <a-form
          ref="formRef"
          :model="formState"
          :label-col="{ span: 4 }"
          :wrapper-col="{ span: 10 }"
          class="combined-course-form"
        >
          <a-form-item
            label="组合课程名称"
            name="name"
            :rules="[{ required: true, message: '请输入组合课程名称' }]"
          >
            <a-input
              v-model:value="formState.name"
              placeholder="请输入"
              allow-clear
              :maxlength="100"
            />
          </a-form-item>
          <a-form-item
            label="添加课程"
            name="courseIds"
            :wrapper-col="{ span: 19 }"
            :rules="[
              {
                required: true,
                type: 'array',
                min: 1,
                message: '请选择课程',
                trigger: ['change', 'blur'],
              },
            ]"
          >
            <div class="combined-course-add-block">
              <a-button type="primary" ghost @click="openSelectCourseModal">
                {{ selectedCourseRows.length ? "编辑已选" : "选择课程" }}
              </a-button>

              <div
                v-if="selectedCourseRows.length"
                class="combined-course-table-wrap"
              >
                <div class="combined-course-table-scroll">
                  <table class="combined-course-table">
                    <thead>
                      <tr>
                        <th class="col-drag" />
                        <th>课程顺序</th>
                        <th>课程名称</th>
                        <th class="col-action">
                          操作
                        </th>
                      </tr>
                    </thead>
                    <tbody ref="sortableTbodyRef">
                      <tr
                        v-for="(row, index) in selectedCourseRows"
                        :key="row.key || row.id"
                      >
                        <td class="col-drag">
                          <MenuOutlined class="drag-handle cursor-move text-#999" />
                        </td>
                        <td>{{ index + 1 }}</td>
                        <td>{{ row.title || row.name }}</td>
                        <td class="col-action">
                          <a-button
                            type="link"
                            danger
                            size="small"
                            @click="removeCourse(index)"
                          >
                            删除
                          </a-button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p class="combined-course-drag-tip">
                  鼠标拖拽可以变动以上课程排列顺序
                </p>
              </div>
            </div>
          </a-form-item>
        </a-form>
      </div>
    </a-modal>

    <SelectCourseRangeModal
      v-model:open="selectCourseModalOpen"
      :selected-courses="selectedCourseRows"
      echo-courses-deletable
      @confirm="onSelectCourseConfirm"
    />
  </div>
</template>

<style lang="less" scoped>
.create-combined-course-modals {
  display: contents;
}

.combined-course-modal-body.contenter {
  padding: 24px;
  max-height: min(700px, calc(100vh - 160px));
  overflow-y: auto;
}

.combined-course-add-block {
  max-width: 100%;
}

.combined-course-table-wrap {
  margin-top: 12px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  background: #fafafa;
}

.combined-course-table-scroll {
  max-height: 360px;
  overflow-y: auto;
  overflow-x: hidden;
  background: #fff;
}

.combined-course-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  font-size: 14px;

  thead th {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 10px 12px;
    text-align: left;
    font-weight: 500;
    color: #333;
    background: #fafafa;
    border-bottom: 1px solid #f0f0f0;
    box-shadow: 0 1px 0 #f0f0f0;
  }

  tbody td {
    padding: 10px 12px;
    border-bottom: 1px solid #f0f0f0;
    vertical-align: middle;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }

  .col-drag {
    width: 40px;
    text-align: center;
  }

  .col-action {
    width: 88px;
    text-align: center;
  }
}

:deep(.sortable-ghost-row) {
  opacity: 0.55;
  background: #e6f4ff;
}

.combined-course-drag-tip {
  margin: 0;
  padding: 8px 12px;
  font-size: 12px;
  color: #888;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
}
</style>

<style>
/* 与创建班级弹窗 modal-content-box 一致 */
.create-combined-course-modal.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.create-combined-course-modal.modal-content-box .ant-modal-body {
  padding: 0 !important;
}

.create-combined-course-modal-wrap .ant-modal-header {
  margin-bottom: 0;
}
</style>
