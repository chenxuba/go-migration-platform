<script setup>
const steps = [
  {
    icon: '👤',
    title: '创建员工',
    status: '已完成',
    statusClass: 'done',
    desc: '填写员工信息，配置权限',
  },
  {
    icon: '📄',
    title: '添加课程',
    status: '已完成',
    statusClass: 'done',
    desc: '完善课程体系，设置价格',
  },
  {
    icon: '💳',
    title: '开通/绑定收银宝',
    status: '待完成',
    statusClass: 'pending',
    desc: '免费开通，提高到账效率',
  },
  {
    icon: '👥',
    title: '导入学员',
    status: '已完成',
    statusClass: 'done',
    desc: '批量快速录入学员信息',
  },
];
</script>

<template>
  <div class="init-tasks-wrapper">
    <div class="init-tasks-header">
      <div>
        <span class="init-tasks-title">初始化任务(3/4)</span>
        <span class="init-tasks-desc">一起來完成初始化，快速开始教学新体验~</span>
      </div>
      <div class="init-tasks-actions">
        <button class="app-btn">
          <span class="icon">📱</span> 下载App
        </button>
        <a class="collapse-link" href="#">收起 &gt;</a>
      </div>
    </div>
    <div class="init-tasks-steps">
      <div v-for="(step, idx) in steps" :key="idx" class="step-card">
        <div class="step-icon">{{ step.icon }}</div>
        <div class="step-info">
          <div class="step-title">
            {{ step.title }}
            <span :class="['step-status', step.statusClass]">{{ step.status }}</span>
          </div>
          <div class="step-desc">{{ step.desc }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.init-tasks-wrapper {
  width: 100%;
  background: #fff;
  border-radius: 12px;
  padding: 24px 16px 32px 16px;
  box-shadow: 0 2px 8px #0000000d;
  margin: 0 auto;
}
.init-tasks-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 14px;
}
.init-tasks-title {
  font-size: 16px;
  font-weight: 600;
  color: #222;
  margin-right: 8px;
}
.init-tasks-desc {
  color: #888;
  font-size: 12px;
}
.init-tasks-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.app-btn {
  background: #fff;
  border: 1px solid #e5e6eb;
  border-radius: 8px;
  padding: 4px 16px;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.collapse-link {
  color: #165dff;
  font-size: 14px;
  text-decoration: none;
  cursor: pointer;
}
.init-tasks-steps {
  display: flex;
  gap: 32px;
  justify-content: space-between;
}
.step-card {
  background: #fbfbfb;
  border-radius: 16px;
  box-shadow: 0 2px 8px #0000000d;
  padding: 10px 24px;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}
.step-icon {
  font-size: 28px;
  margin-top: 2px;
}
.step-info {
  flex: 1;
}
.step-title {
  font-size: 14px;
  font-weight: 500;
  color: #222;
  display: flex;
  align-items: center;
  gap: 8px;
}
.step-status {
  font-size: 12px;
  border-radius: 6px;
  padding: 1px 8px;
  margin-left: 2px;
  font-weight: 400;
}
.step-status.done {
  background: #e8f7e5;
  color: #3fc06d;
}
.step-status.pending {
  background: #fff7e6;
  color: #faad14;
}
.step-desc {
  color: #888;
  font-size: 12px;
  margin-top: 2px;
}
</style>
