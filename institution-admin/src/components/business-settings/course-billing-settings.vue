<script setup lang="ts">
const billingRows = [
  {
    key: 'byHours',
    label: '按课时收费',
    description: '按“课时购买数”定价，以“课时”为单位计费',
  },
  {
    key: 'byPeriod',
    label: '按时段收费',
    description: '按“天/自然月/自然年”定价，以“天”为单位计费',
  },
  {
    key: 'byAmount',
    label: '按金额收费',
    description: '开启后机构支持按“充值金额”定价，每次点名扣除对应金额数',
  },
]
</script>

<template>
  <div class="tab-content">
    <div class="setting">
      <div class="title">
        收费方式设置
      </div>

      <div class="table-wrap mt-2">
        <table border>
          <tbody>
            <tr v-for="row in billingRows" :key="row.key">
              <td class="td1">
                {{ row.label }}
              </td>
              <td>
                <div class="status-line">
                  <span class="status-dot" />
                  <span class="status-text">已开启</span>
                </div>

                <div class="desc">
                  {{ row.description }}
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
  .title {
    display: flex;
    align-items: center;
    color: #222;
    font-size: 18px;
    font-weight: 600;

    &::before {
      display: inline-block;
      width: 4px;
      height: 14px;
      margin-right: 6px;
      content: '';
      background: var(--pro-ant-color-primary);
      border-radius: 100px;
    }
  }

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
      color: #222;
      font-size: 14px;
      font-weight: 400;
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

.desc {
  color: #222;
  font-size: 14px;
}
</style>
