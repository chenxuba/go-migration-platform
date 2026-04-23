<script setup lang="ts">
import { reactive } from 'vue'

const teachingMethodRows = reactive([
  {
    key: 'classroom',
    label: '班级授课',
    mode: 'status',
    enabled: true,
    description: '开启后机构可开设班级授课，支持自定义学员数量，允许上课老师给多名学员上课',
  },
  {
    key: 'oneToOne',
    label: '1对1授课',
    mode: 'status',
    enabled: true,
    description: '开启后机构可开设1对1课程，学员数量有限制，只允许上课老师给1名学员上课',
  },
  {
    key: 'multiCourse',
    label: '一班多课',
    mode: 'switch',
    enabled: true,
    description: '开启后，机构可设置组合课程，创建班级时支持关联组合课程，实现报名不同课程的学员在同一个班级中上课',
    actionText: '查看组合课程列表',
  },
])
</script>

<template>
  <div class="tab-content">
    <div class="setting">
      <custom-title title="授课方式设置" font-size="18px" font-weight="800" before-height="14px" />

      <div class="table-wrap mt-2">
        <table border>
          <tbody>
            <tr v-for="row in teachingMethodRows" :key="row.key">
              <td class="td1">
                {{ row.label }}
              </td>
              <td>
                <div v-if="row.mode === 'status'" class="status-line">
                  <span class="status-dot" />
                  <span class="status-text">已开启</span>
                </div>

                <div v-else class="switch-line">
                  <a-switch v-model:checked="row.enabled" />
                </div>

                <div class="desc">
                  {{ row.description }}
                </div>

                <div v-if="row.actionText" class="action-wrap">
                  <a-button type="primary" ghost>
                    {{ row.actionText }}
                  </a-button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.tab-content {
  background: #fff;
  border-radius: 12px;
  padding: 18px 20px 12px;
}

.setting {
  .table-wrap {
    table {
      width: 100%;
      border: 1px solid #eee;
      border-collapse: collapse;
      border-radius: 8px;
    }

    tr,
    td {
      border: 1px solid #eee;
    }

    td {
      padding: 18px 24px;
      vertical-align: middle;
    }

    .td1 {
      width: 180px;
      color: #333;
      font-size: 14px;
      font-weight: 500;
      text-align: center;
    }
  }
}

.status-line {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #1fbe4f;
  border-radius: 999px;
  flex-shrink: 0;
}

.status-text {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.switch-line {
  display: inline-flex;
  align-items: center;
}

.desc {
  color: #222;
  font-size: 14px;
  line-height: 1.75;
}

.action-wrap {
  margin-top: 12px;
}
</style>
