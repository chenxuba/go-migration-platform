<script setup>
const activeKey = ref('deduct-order')

const courseDeductOrder = ref('oldest')
const switches = reactive({
  consumeOver: true,
  autoRollCall: true,
  liveAutoRollCall: false,
  limitSingleOrder: false,
  leaveNormalByHour: false,
  absentNormalByHour: false,
  periodMakeup: true,
  periodAutoEnd: false,
  amountLeaveNormal: false,
  amountAbsentNormal: true,
  amountMakeup: false,
  faceDeduct: true,
  faceSignInNotice: true,
  faceSignOutNotice: true,
  voicePrompt: true,
  faceAdminNotice: false,
})

const orderExampleRows = [
  { no: 1, order: '订单 A', validUntil: '2025-06-08', accountType: '正价', createdAt: '2025-05-05 10:10' },
  { no: 2, order: '订单 B', validUntil: '2025-08-08', accountType: '正价', createdAt: '2025-06-05 10:10' },
  { no: 3, order: '订单 B', validUntil: '2025-08-08', accountType: '赠送', createdAt: '2025-06-05 10:10' },
  { no: 4, order: '订单 C', validUntil: '不限制', accountType: '正价', createdAt: '2025-07-05 10:10' },
]
</script>

<template>
  <div class="roll-call-settings">
    <a-tabs v-model:active-key="activeKey" destroy-inactive-tab-pane class="roll-call-settings__tabs">
      <a-tab-pane key="deduct-order" tab="点名课消顺序">
        <section class="settings-section">
          <div class="settings-section__title">
            点名课消顺序
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                课程课消顺序
              </div>
              <div class="settings-row__content">
                <a-radio-group v-model:value="courseDeductOrder" class="settings-radio-group custom-radio">
                  <a-radio value="oldest">
                    先进先出
                  </a-radio>
                  <a-radio value="account-first">
                    优先消通用账户
                  </a-radio>
                  <a-radio value="normal-first">
                    优先消普通账户
                  </a-radio>
                </a-radio-group>
                <div class="settings-desc">
                  按学员的学费账户生成时间优先消耗，适用于多数常规课消场景。
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                订单课消顺序
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  <span>智能排序：</span>
                  <span class="text-primary">先进先出</span>
                  <a-button type="link"  class="settings-link">
                    编辑
                  </a-button>
                </div>

                <div class="example-card">
                  <div class="example-card__title">
                    场景举例：
                  </div>
                  <div class="example-card__desc">
                    学员报读同课程有多个订单时，系统将按规则自动选择优先课消的订单。
                  </div>
                  <table class="example-table">
                    <thead>
                      <tr>
                        <th>序号</th>
                        <th>订单</th>
                        <th>有效期至</th>
                        <th>账户属性</th>
                        <th>生成时间</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="row in orderExampleRows" :key="row.no">
                        <td>{{ row.no }}</td>
                        <td>{{ row.order }}</td>
                        <td>{{ row.validUntil }}</td>
                        <td>{{ row.accountType }}</td>
                        <td>{{ row.createdAt }}</td>
                      </tr>
                    </tbody>
                  </table>
                  <div class="example-card__footer">
                    若设置智能排序规则：近有效期 &gt; 赠送 &gt; 先进先出，则订单课消顺序为：1 &gt; 3 &gt; 2 &gt; 4
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="deduct-rule" tab="点名课消规则">
        <section class="settings-section">
          <div class="settings-section__title">
            点名课消规则
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                消超点名
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.consumeOver" />
                <div class="settings-desc">
                  开启后，学员的上课点名支持消超记录（支持欠费点名）。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                自动点名
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.autoRollCall" />
                <div class="settings-desc">
                  开启后，未点名课程在结束后 30 分钟内，系统将自动点名并记录课消。
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                限制订单欠费课消
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.limitSingleOrder" />
                <div class="settings-desc">
                  开启后，学员订单对应的费用耗完后，限制继续课消。
                </div>
                <div class="example-card">
                  <div class="example-card__title">
                    场景举例：
                  </div>
                  <div class="example-card__desc">
                    开启限制订单欠费课消后，以下情况学员无法正常课消：
                  </div>
                  <ol class="example-list">
                    <li>学员报名 10 课时共计 100 元，实际支付 50 元，欠费 50 元，消耗 5 课时后，无法继续课消。</li>
                    <li>学员报名课程支付 0 元，如果是欠费报名，则无法课消。</li>
                    <li>学员报名的课程 A 有三笔订单，订单 1 的实收学费不足时，可手动调整课消顺序继续课消。</li>
                  </ol>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按课时收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline settings-inline--heading">
                  按“课时购买数”定价，以“课时”为单位计费
                </div>
                <div class="rule-box">
                  <div class="settings-inline">
                    <span><span class="settings-switch-line__label">默认记录课时：</span>创建班级 / 1 对 1 时仍可编辑调整</span>
                  </div>
                  <div class="settings-inline settings-inline--muted">
                    <span>按固定课时记录：默认记录学员 <span class="text-primary">1</span> 课时，教师 <span class="text-primary">0</span> 课时</span>
                    <a-button type="link"  class="settings-link">
                      编辑
                    </a-button>
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员请假正常记录：</span>
                    <a-switch v-model:checked="switches.leaveNormalByHour"  />
                  </div>
                  <div class="settings-desc">
                    开启后，学员请假正常记录课时
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员旷课正常记录：</span>
                    <a-switch v-model:checked="switches.absentNormalByHour"  />
                  </div>
                  <div class="settings-desc">
                    开启后，学员旷课正常记录课时
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按时段收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  按“天 / 自然月 / 自然年”定价，以“天”为单位计费，每日自动课消。
                </div>
                <div class="rule-box">
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员补课：</span>
                    <a-switch v-model:checked="switches.periodMakeup"  />
                  </div>
                  <div class="settings-desc">
                    开启后，按时段收费的课程，支持缺课学员补课
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">课消为 0 天后自动结课：</span>
                    <a-switch v-model:checked="switches.periodAutoEnd"  />
                  </div>
                  <div class="settings-desc">
                    开启后，当课程账户每日自动课消剩余为 0 天后，次日自动结课
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按金额收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline settings-inline--heading">
                  按“充值金额”定价，每次点名扣除对应金额数
                </div>
                <div class="rule-box">
                  <div class="settings-inline">
                    <span class="settings-switch-line__label">扣费规则：</span>
                    <a-button type="link"  class="settings-link">
                      编辑
                    </a-button>
                  </div>
                  <div class="settings-desc">
                    默认扣费：<span class="text-primary">100</span> 元（仅对未设置单课扣费的课程有效）
                  </div>
                  <div class="settings-desc">
                    单课扣费：已设置 <span class="text-primary">0</span> 门课程，点此上方编辑操作对单个课程进行扣费金额设置
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员请假正常记录：</span>
                    <a-switch v-model:checked="switches.amountLeaveNormal"  />
                  </div>
                  <div class="settings-desc">
                    开启后，学员请假正常记录金额
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员旷课正常记录：</span>
                    <a-switch v-model:checked="switches.amountAbsentNormal"  />
                  </div>
                  <div class="settings-desc">
                    开启后，学员旷课正常记录金额
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员补课：</span>
                    <a-switch v-model:checked="switches.amountMakeup"  />
                  </div>
                  <div class="settings-desc">
                    开启后，按金额收费的课程，支持缺课学员补课
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="face-attendance" tab="人脸考勤设置">
        <section class="settings-section">
          <div class="settings-section__title">
            人脸考勤设置
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                人脸考勤关联点名课消
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.faceDeduct" />
                <div class="settings-inline settings-inline--block">
                  <span>每日考勤时段：</span>
                  <span>00:00 ~ 12:00 ~ 24:00</span>
                  <a-button type="link"  class="settings-link">
                    编辑
                  </a-button>
                  <a-button type="link"  class="settings-link">
                    恢复默认
                  </a-button>
                </div>
                <div class="settings-inline settings-inline--block">
                  <span>人脸考勤课消规则：</span>
                  <span>多日程按序课消</span>
                  <a-button type="link"  class="settings-link">
                    编辑
                  </a-button>
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员考勤签到通知
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.faceSignInNotice" />
                <div class="settings-desc">
                  开启后，学员完成考勤签到或手动标记签到将自动发送微信通知给家长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员考勤签退通知
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.faceSignOutNotice" />
                <div class="settings-desc">
                  开启后，学员完成考勤签退或手动标记签退将自动发送微信通知给家长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                语音提示
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.voicePrompt" />
                <div class="settings-desc">
                  开启后，系统实时播报学员是否人脸考勤成功、签到签退成功。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                刷脸间隔设置
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  <span>刷脸间隔 <span class="text-primary">1</span> 分钟</span>
                  <a-button type="link"  class="settings-link">
                    编辑
                  </a-button>
                </div>
                <div class="settings-desc">
                  学员刷脸考勤成功后，允许下次刷脸考勤间隔时长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员刷脸成功提醒至管理员
              </div>
              <div class="settings-row__content">
                <a-switch v-model:checked="switches.faceAdminNotice" />
                <div class="settings-desc">
                  开启后，校区管理员将在 App 消息中心收到学员刷脸成功的推送提醒。
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style lang="less" scoped>
.roll-call-settings {
  min-height: 420px;
}

.roll-call-settings__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    background: #fff;
    border-radius: 0 0 16px 16px !important;

    &::before {
      display: none;
    }
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 10px;
    margin: 6px 0;
  }

  :deep(.ant-tabs-tab) {
    margin: 0 8px 0 0;
    padding: 6px 14px !important;
    font-size: 14px !important;
  }

  :deep(.ant-tabs-tab .ant-tabs-tab-btn) {
    font-size: 14px !important;
    line-height: 22px;
  }

  :deep(.ant-tabs-tab-active) {
    background: #e6f0ff;
    border-radius: 8px;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    color: var(--pro-ant-color-primary, #1677ff);
    font-size: 14px !important;
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    display: none;
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.settings-section {
  margin-top: 10px;
  padding: 16px 16px 22px;
  background: #fff;
  border-radius: 14px;
}

.settings-section__title {
  position: relative;
  padding-left: 12px;
  margin-bottom: 14px;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;

  &::before {
    position: absolute;
    top: 5px;
    left: 0;
    width: 3px;
    height: 14px;
    border-radius: 3px;
    background: var(--pro-ant-color-primary, #1677ff);
    content: '';
  }
}

.settings-table {
  overflow: hidden;
  border: 1px solid #edf0f5;
  background: #fff;
}

.settings-row {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  min-height: 84px;
  border-bottom: 1px solid #edf0f5;

  &:last-child {
    border-bottom: 0;
  }
}

.settings-row--top {
  align-items: stretch;
}

.settings-row__label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px 14px;
  border-right: 1px solid #edf0f5;
  color: #1f2937;
  font-size: 14px;
  font-weight: 500;
  text-align: center;
}

.settings-row__content {
  min-width: 0;
  padding: 18px 18px;
  color: #1f2937;
  font-size: 14px;
}

.settings-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 18px;
}

.custom-radio :deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio :deep(.ant-radio:hover .ant-radio-inner),
.custom-radio :deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary, #1677ff);
}

.custom-radio :deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio :deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary, #1677ff);
}

.custom-radio :deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary, #1677ff);
  transform: scale(0.5);
}

.settings-desc {
  margin-top: 8px;
  color: #333;
  font-size: 14px;
  line-height: 20px;
}

.settings-inline {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  color: #1f2937;
}

.settings-inline--muted {
  margin-top: 8px;
  color: #333;
}

.settings-inline--heading {
  margin-bottom: 12px;
  color: #1f2937;
  font-size: 14px;
  line-height: 22px;
}

.settings-inline--block {
  margin-top: 12px;
}

.settings-switch-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  color: #333;
  font-size: 14px;
}

.settings-switch-line__label {
  color: #1f2937;
  font-weight: 600;
}

.settings-link {
  padding: 0 4px;
  font-size: 14px;
}

.text-primary {
  color: var(--pro-ant-color-primary, #1677ff);
}

.example-card {
  margin-top: 14px;
  padding: 14px;
  border: 1px solid #edf0f5;
  background: #fafafa;
}

.rule-box {
  margin-top: 14px;
  padding: 12px 14px;
  border: 1px solid #edf0f5;
  background: #fafafa;
  border-radius: 8px;
}

.example-card__title {
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.example-card__desc,
.example-card__footer {
  margin-top: 8px;
  color: #333;
  font-size: 14px;
  line-height: 20px;
}

.example-table {
  width: 100%;
  margin-top: 12px;
  overflow: hidden;
  border-spacing: 0;
  border-collapse: separate;
  border: 1px solid #edf0f5;

  th,
  td {
    height: 44px;
    padding: 0 12px;
    border-right: 1px solid #edf0f5;
    border-bottom: 1px solid #edf0f5;
    color: #4b5563;
    font-size: 14px;
    text-align: center;

    &:last-child {
      border-right: 0;
    }
  }

  th {
    color: #1f2937;
    font-weight: 600;
  }

  tbody tr:last-child td {
    border-bottom: 0;
  }
}

.example-list {
  margin: 8px 0 0;
  padding-left: 18px;
  color: #333;
  font-size: 14px;
  line-height: 24px;
}

@media (max-width: 768px) {
  .settings-row {
    grid-template-columns: 1fr;
  }

  .settings-row__label {
    justify-content: flex-start;
    border-right: 0;
    border-bottom: 1px solid #edf0f5;
  }
}
</style>
