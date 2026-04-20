<script setup lang="ts">
import { accountRoles } from '../shared/mock-data'

const columns = [
  { title: '角色名称', dataIndex: 'roleName', key: 'roleName', width: 140 },
  { title: '数据范围', dataIndex: 'dataScope', key: 'dataScope', width: 220 },
  { title: '典型权限', dataIndex: 'typicalPermissions', key: 'typicalPermissions' },
]
</script>

<template>
  <div class="gov-page">
    <a-card :bordered="false">
      <div class="page-head__title">
        层级权限
      </div>
      <div class="page-head__desc">
        G 端省、市、区三级继续共用这一套代码，区别只放在账号角色、行政区划范围和可见菜单上。当前账号来源建议复用总控端
        <code>manage</code>
        账号体系，在总控端创建监管账号并分配省市区监管角色。
      </div>
    </a-card>

    <div class="content-grid">
      <a-card title="推荐拆分方式" :bordered="false">
        <div class="bullet-card">
          <div class="bullet-card__item">
            <code>institution-admin</code> 继续承载机构端业务。
          </div>
          <div class="bullet-card__item">
            <code>government-admin</code> 单独承载 G 端监管业务。
          </div>
          <div class="bullet-card__item">
            省、市、区三类监管账号共用一个前端项目，不再复制三套代码。
          </div>
          <div class="bullet-card__item">
            监管账号建议在总控端统一创建，再通过角色和区域范围控制省、市、区数据权限。
          </div>
        </div>
      </a-card>

      <a-card title="三级监管角色" :bordered="false">
        <a-table :columns="columns" :data-source="accountRoles" :pagination="false" size="small" />
      </a-card>
    </div>
  </div>
</template>

<style scoped lang="less">
.gov-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-head__title {
  margin-bottom: 8px;
  color: #1f2329;
  font-size: 24px;
  font-weight: 700;
}

.page-head__desc {
  color: #6b7280;
  line-height: 24px;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(320px, 0.9fr) minmax(0, 1.4fr);
  gap: 16px;
}

.bullet-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.bullet-card__item {
  padding: 14px 16px;
  border-radius: 12px;
  background: #f8fafc;
  color: #4b5563;
  line-height: 24px;
}

@media (max-width: 1200px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
