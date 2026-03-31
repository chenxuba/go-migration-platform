<script setup>
import { CloseOutlined, QuestionCircleOutlined } from "@ant-design/icons-vue";
import { debounce } from "lodash-es";
import { getCoursePageApi } from "~/api/edu-center/course-list";
import { pageComposeLessonsForPcApi } from "~/api/edu-center/compose-lesson";
import {
  checkGroupClassNameApi,
  createGroupClassApi,
  getGroupClassDetailApi,
  updateGroupClassApi,
} from "~/api/edu-center/group-class";
import CreateCombinedCourseModal from "./create-combined-course-modal.vue";
import StaffSelect from "./staff-select.vue";
import messageService from "~/utils/messageService";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  /** 列表行数据，有值时表示编辑模式 */
  editRecord: {
    type: Object,
    default: null,
  },
});
const emit = defineEmits(["update:open", "created", "updated"]);
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

/** 组合课展示、回显不带「（N门课）」后缀 */
function stripComposeLessonCountSuffix(text) {
  if (text == null)
    return "";
  return String(text)
    .replace(/[（(]\s*\d+\s*门课\s*[）)]\s*$/u, "")
    .trim();
}

function mapComposeListItem(it) {
  const rawName = it.name != null ? String(it.name) : "";
  const short
    = stripComposeLessonCountSuffix(rawName) || rawName || String(it.id);
  const n = it.productCount;
  const listLabel
    = n != null && n !== "" && Number.isFinite(Number(n))
      ? `${short}（${n}门课）`
      : short;
  return {
    label: short,
    listLabel,
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
function getCreateFormDefaults() {
  return {
    mode: "1",
    course: undefined,
    totalCourse: undefined,
    className: undefined,
    defaultClassTimeRecordMode: 1,
    defaultStudentClassTime: 1,
    defaultTeacherClassTime: 0,
    maxNum: undefined,
    teacher: [],
    defaultTeacher: undefined,
    classRoom: undefined,
    remark: "",
  };
}

const formState = reactive(getCreateFormDefaults());

/** 满班人数：0 或未填表示不限，用空值回显以展示 placeholder「不限」 */
function syncMaxNumInput(v) {
  if (v == null || v === "" || Number(v) <= 0)
    formState.maxNum = undefined;
  else
    formState.maxNum = Number(v);
}

const classTimeUnitLabel = computed(() =>
  Number(formState.defaultClassTimeRecordMode) === 2 ? "课时/小时" : "课时",
);

const classTimeHint = computed(() =>
  Number(formState.defaultClassTimeRecordMode) === 2
    ? "每次点名，学员和上课教师记录的课时会根据日程时长自动计算课时（点名时支持调整）"
    : "每次点名，学员和上课教师记录的课时数默认为此数值（点名时支持调整）",
);
const submitting = ref(false);

const skipAutoDefaultTeacher = ref(false);
const syncingDefaultFromTeacher = ref(false);

const isEdit = computed(() => !!props.editRecord?.id);

/** 切换创建/编辑或不同班级时重挂 StaffSelect，避免内部 options/加载态残留（如默认老师一直「加载中」） */
const staffSelectInstanceKey = computed(() =>
  props.editRecord?.id ? `e-${props.editRecord.id}` : "c",
);

const editDetailLoading = ref(false);
let editDetailReqSeq = 0;

/** 传给班主任 StaffSelect，用班级接口里的 teachers 直接显示姓名 */
const teacherPresetForSelect = ref([]);
/** 默认上课老师：用 defaultTeacherName / teachers 匹配，避免单选卡在「加载中」 */
const defaultTeacherPresetForSelect = ref([]);

/** 编辑时当前 value 可能不在分页首屏，补一条 option 才能显示课程名而非纯 ID */
function mergeLessonOptionFromRecord(rec) {
  const lid = rec?.lessonId != null && String(rec.lessonId).trim() !== ""
    ? String(rec.lessonId).trim()
    : "";
  if (!lid)
    return;
  const name = String(rec.lessonName || "").trim();
  if (rec.isMultiProduct) {
    const short = stripComposeLessonCountSuffix(name) || name || lid;
    const m = name.match(/[（(]\s*(\d+)\s*门课\s*[）)]/u);
    const listLabel = m ? `${short}（${m[1]}门课）` : short;
    if (!composeLessonOptions.value.some((o) => String(o.value) === lid)) {
      composeLessonOptions.value = [
        { label: short, listLabel, value: lid },
        ...composeLessonOptions.value,
      ];
    }
  }
  else {
    const label = name || lid;
    if (!singleCourseOptions.value.some((o) => String(o.value) === lid)) {
      singleCourseOptions.value = [
        { label, value: lid },
        ...singleCourseOptions.value,
      ];
    }
  }
}

/** 创建模式 / 关弹窗时恢复空白表单（避免编辑残留进「创建班级」） */
function resetFormToCreateDefaults() {
  skipAutoDefaultTeacher.value = false;
  syncingDefaultFromTeacher.value = false;
  const d = getCreateFormDefaults();
  Object.keys(d).forEach((k) => {
    if (k === "teacher")
      formState.teacher = [];
    else
      formState[k] = d[k];
  });
  teacherPresetForSelect.value = [];
  defaultTeacherPresetForSelect.value = [];
  nextTick(() => {
    formRef.value?.clearValidate?.();
  });
}

function applyEditRecord(rec) {
  syncingDefaultFromTeacher.value = true;
  skipAutoDefaultTeacher.value = true;
  formState.mode = rec.isMultiProduct ? "2" : "1";
  const lessonIdStr
    = rec.lessonId != null && String(rec.lessonId).trim() !== ""
      ? String(rec.lessonId).trim()
      : undefined;
  formState.course = rec.isMultiProduct ? undefined : lessonIdStr;
  formState.totalCourse = rec.isMultiProduct ? lessonIdStr : undefined;
  formState.className = rec.name;
  {
    const mc = rec.maxCount;
    formState.maxNum
      = mc != null && Number(mc) > 0 ? Number(mc) : undefined;
  }
  formState.teacher = (rec.teachers || []).map((t) => t.id);
  const dtRaw
    = rec.defaultTeacherId && rec.defaultTeacherId !== "0"
      ? rec.defaultTeacherId
      : undefined;
  const dt = dtRaw != null ? String(dtRaw) : undefined;
  formState.defaultTeacher = dt;
  skipAutoDefaultTeacher.value = !dt;
  defaultTeacherPresetForSelect.value = [];
  if (dt) {
    const nm = String(rec.defaultTeacherName || "").trim();
    const fromT = (rec.teachers || []).find(t => String(t.id) === String(dt));
    const nick = nm || (fromT?.name ? String(fromT.name) : "");
    defaultTeacherPresetForSelect.value = [{
      id: dt,
      name: nick || dt,
      nickName: nick || dt,
      mobile: fromT?.mobile ?? "",
    }];
  }
  formState.defaultStudentClassTime = rec.defaultStudentClassTime ?? 1;
  formState.defaultTeacherClassTime = rec.defaultTeacherClassTime ?? 0;
  formState.defaultClassTimeRecordMode = rec.defaultClassTimeRecordMode ?? 1;
  formState.remark = rec.remark || "";
  const roomRaw = rec.classroomId ?? rec.classRoomId;
  formState.classRoom
    = roomRaw != null && String(roomRaw) !== "" && String(roomRaw) !== "0"
      ? String(roomRaw)
      : undefined;
  teacherPresetForSelect.value = (rec.teachers || []).map((t) => ({
    id: t.id,
    name: t.name,
    nickName: t.nickName,
    mobile: t.mobile,
    status: t.status,
  }));
  mergeLessonOptionFromRecord(rec);
  nextTick(() => {
    syncingDefaultFromTeacher.value = false;
  });
}

watch(
  () => [props.open, props.editRecord],
  () => {
    if (!props.open) {
      editDetailReqSeq += 1;
      editDetailLoading.value = false;
      resetFormToCreateDefaults();
      return;
    }
    if (!props.editRecord?.id) {
      editDetailReqSeq += 1;
      editDetailLoading.value = false;
      resetFormToCreateDefaults();
      return;
    }
    applyEditRecord(props.editRecord);
    const classId = String(props.editRecord.id);
    const seq = ++editDetailReqSeq;
    editDetailLoading.value = true;
    getGroupClassDetailApi({ id: classId })
      .then((res) => {
        if (seq !== editDetailReqSeq)
          return;
        const row = res.result ?? res.data;
        if (res.code === 200 && row && typeof row === "object")
          applyEditRecord(row);
        else if (res.message)
          messageService.error(res.message || "加载班级详情失败");
      })
      .catch((e) => {
        if (seq !== editDetailReqSeq)
          return;
        messageService.error(
          e?.response?.data?.message || e?.message || "加载班级详情失败",
        );
      })
      .finally(() => {
        if (seq === editDetailReqSeq)
          editDetailLoading.value = false;
      });
  },
  { flush: "sync", immediate: true },
);

watch(
  () => formState.teacher,
  (ids) => {
    const list = Array.isArray(ids) ? [...ids] : [];
    if (list.length === 0) {
      formState.defaultTeacher = undefined;
      skipAutoDefaultTeacher.value = false;
      return;
    }
    const idSet = new Set(list.map((id) => String(id)));
    const cur = formState.defaultTeacher;
    if (cur != null && cur !== "" && !idSet.has(String(cur))) {
      formState.defaultTeacher = undefined;
    }
    if (skipAutoDefaultTeacher.value)
      return;
    if (formState.defaultTeacher == null || formState.defaultTeacher === "") {
      syncingDefaultFromTeacher.value = true;
      formState.defaultTeacher = list[0];
      nextTick(() => {
        syncingDefaultFromTeacher.value = false;
      });
    }
  },
  { deep: true },
);

watch(
  () => formState.defaultTeacher,
  (v, ov) => {
    if (syncingDefaultFromTeacher.value)
      return;
    const had = ov != null && ov !== "";
    const empty = v == null || v === "";
    if (had && empty)
      skipAutoDefaultTeacher.value = true;
  },
);

watch(
  () => props.open,
  (open) => {
    if (open && !props.editRecord?.id)
      skipAutoDefaultTeacher.value = false;
  },
);

async function handleSubmit() {
  try {
    await formRef.value.validate();
  }
  catch {
    return;
  }
  const lessonId
    = formState.mode === "1" ? formState.course : formState.totalCourse;
  if (!lessonId) {
    messageService.error(
      formState.mode === "1" ? "请选择课程" : "请选择组合课程",
    );
    return;
  }
  const teacherIds = (formState.teacher || []).map((id) => String(id));
  if (teacherIds.length === 0) {
    messageService.error("请选择班主任");
    return;
  }
  const defaultTeacherId =
    formState.defaultTeacher != null && formState.defaultTeacher !== ""
      ? String(formState.defaultTeacher)
      : "";
  if (
    defaultTeacherId
    && !teacherIds.includes(defaultTeacherId)
  ) {
    messageService.error("默认上课教师须在所选班主任中");
    return;
  }
  submitting.value = true;
  try {
    const className = String(formState.className || "").trim();
    const checkNamePayload = {
      name: className,
      isOne2One: false,
    };
    if (props.editRecord?.id)
      checkNamePayload.exceptId = String(props.editRecord.id);

    const checkRes = await checkGroupClassNameApi(checkNamePayload);
    if (checkRes.code !== 200) {
      messageService.error(checkRes.message || "校验班级名称失败");
      return;
    }
    if (checkRes.result) {
      messageService.error("班级名称已存在");
      return;
    }

    const commonBody = {
      name: className,
      lessonId: String(lessonId),
      maxCount: formState.maxNum != null ? Number(formState.maxNum) : 0,
      teacherIds,
      defaultTeacherId: defaultTeacherId || "0",
      defaultStudentClassTime: Number(formState.defaultStudentClassTime) || 1,
      defaultTeacherClassTime: Number(formState.defaultTeacherClassTime) || 0,
      defaultClassTimeRecordMode: Number(formState.defaultClassTimeRecordMode) || 1,
      isCopyStudent: false,
      copiedStudents: [],
      isCopyTimetable: false,
      classProperties: [],
      remark: String(formState.remark || "").trim(),
    };

    if (props.editRecord?.id) {
      const classId = String(props.editRecord.id);
      const res = await updateGroupClassApi({
        ...commonBody,
        id: classId,
        copyFromClassId: classId,
      });
      const updated = res.result ?? res.data;
      if (res.code === 200 && updated?.id) {
        messageService.success("保存成功");
        emit("updated", updated);
        closeFun();
        return;
      }
      if (res.code !== 500)
        messageService.error(res.message || "保存失败");
      return;
    }

    const res = await createGroupClassApi(commonBody);
    const created = res.result ?? res.data;
    if (res.code === 200 && created?.id) {
      messageService.success("创建班级成功");
      emit("created", created);
      closeFun();
      return;
    }
    if (res.code !== 500)
      messageService.error(res.message || "创建班级失败");
  }
  catch (e) {
    const msg =
      e?.response?.data?.message
      || e?.response?.data?.Message
      || e?.message
      || "创建班级失败";
    messageService.error(msg);
  }
  finally {
    submitting.value = false;
  }
}
function closeFun() {
  resetFormToCreateDefaults();
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
        <span>{{ isEdit ? "编辑班级" : "创建班级" }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-spin :spinning="editDetailLoading && isEdit" tip="加载班级信息…">
      <a-form
        ref="formRef"
        :model="formState"
        :label-col="{ span: 4 }"
        :wrapper-col="{ span: 13 }"
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
              option-label-prop="label"
              @dropdown-visible-change="onComposeDropdownVisible"
              @search="onComposeSearch"
              @popup-scroll="onComposePopupScroll"
            >
              <a-select-option
                v-for="opt in composeLessonOptions"
                :key="opt.value"
                :value="opt.value"
                :label="opt.label"
              >
                {{ opt.listLabel ?? opt.label }}
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
            :value="formState.maxNum"
            placeholder="不限"
            :min="0"
            :precision="0"
            class="w-160px"
            @update:value="syncMaxNumInput"
          />
        </a-form-item>
        <a-form-item
          label="班主任"
          name="teacher"
          :rules="[
            {
              required: true,
              type: 'array',
              min: 1,
              message: '请选择班主任',
            },
          ]"
        >
          <StaffSelect
            :key="`${staffSelectInstanceKey}-teacher`"
            v-model="formState.teacher"
            placeholder="请选择班主任"
            :width="'100%'"
            :multiple="true"
            :status="0"
            :allow-clear="true"
            :preset-staff="teacherPresetForSelect"
          />
        </a-form-item>
        <a-form-item
          label="默认上课老师"
          name="defaultTeacher"
        >
          <StaffSelect
            :key="`${staffSelectInstanceKey}-default`"
            v-model="formState.defaultTeacher"
            placeholder="可选，清空后不再自动带出"
            :width="'100%'"
            :multiple="false"
            :status="0"
            :allow-clear="true"
            :preset-staff="defaultTeacherPresetForSelect"
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
      </a-spin>
    </div>
    <template #footer>
      <a-button @click="closeFun"> 关闭 </a-button>
      <!-- 警告 -->
      <a-button @click="closeFun"> 保存并下一个 </a-button>
      <a-button type="primary" :loading="submitting" @click="handleSubmit">
        确定
      </a-button>
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
