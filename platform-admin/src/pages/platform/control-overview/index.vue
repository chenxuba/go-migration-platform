<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CrownOutlined,
  GlobalOutlined,
  PlusOutlined,
  SafetyCertificateOutlined,
  ShopOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { getTenantBootstrapSummaryApi, listTenantsApi } from '@/api/platform/tenants'

interface ControlSummary {
  tenantId: string
  tenantName: string
  tenantType: string
  institutionCount: number
  menuCount: number
  adminUsernames: string[]
  domains: string[]
  status?: string
  edition?: string
}

const router = useRouter()
const userStore = useUserStore()
const isTenantAdmin = computed(() => userStore.userInfo?.tenantRole === 'tenant_admin')
const loading = ref(false)
const summary = ref<ControlSummary>()
const tenantItems = ref<ControlSummary[]>([])

const partnerTenants = computed(() => tenantItems.value.filter(item => item.tenantType !== 'platform'))
const activeTenantCount = computed(() => partnerTenants.value.filter(item => item.status !== 'disabled').length)
const totalInstitutionCount = computed(() => partnerTenants.value.reduce((total, item) => total + Number(item.institutionCount || 0), 0))
const totalDomainCount = computed(() => partnerTenants.value.reduce((total, item) => total + (item.domains?.length || 0), 0))
const totalAdminCount = computed(() => partnerTenants.value.reduce((total, item) => total + (item.adminUsernames?.length || 0), 0))
const recentTenants = computed(() => partnerTenants.value.slice(0, 6))
const pendingTenants = computed(() => partnerTenants.value.filter(item => !item.domains?.length || !item.adminUsernames?.length || !item.institutionCount))
const currentDomains = computed(() => isTenantAdmin.value ? (summary.value?.domains || []) : partnerTenants.value.flatMap(item => item.domains || []))

async function loadSummary() {
  loading.value = true
  try {
    if (isTenantAdmin.value) {
      const res = await getTenantBootstrapSummaryApi()
      summary.value = res.data || res.result
      tenantItems.value = summary.value ? [summary.value] : []
      return
    }

    const res = await listTenantsApi()
    const items = res.data || res.result || []
    tenantItems.value = items
    summary.value = items.find(item => item.tenantId === 'platform') || items[0]
  }
  finally {
    loading.value = false
  }
}

function go(path: string) {
  router.push(path)
}

function tenantStatusColor(status?: string) {
  return status === 'disabled' ? 'default' : 'success'
}

function tenantStatusText(status?: string) {
  return status === 'disabled' ? '停用' : '正常'
}

onMounted(loadSummary)
</script>

<template>
  <div class="control-overview">
    <section class="overview-hero">
      <div class="overview-hero__content">
        <a-tag color="blue" class="overview-hero__tag">
          {{ isTenantAdmin ? '客户子总控' : '公司平台总控' }}
        </a-tag>
        <h1>{{ isTenantAdmin ? '客户运营工作台' : '合作客户运营总览' }}</h1>
        <p>
          {{ isTenantAdmin
            ? '管理当前客户名下机构、版本开通和账号运营情况。'
            : '统一管理合作客户、子总控账号、机构归属、版本授权和独立域名。' }}
        </p>
        <div class="overview-hero__actions">
          <a-button v-if="!isTenantAdmin" type="primary" size="large" @click="go('/platform/tenants')">
            <template #icon><PlusOutlined /></template>
            开通合作客户
          </a-button>
          <a-button size="large" @click="go('/platform/organizations')">
            查看机构
          </a-button>
          <a-button size="large" @click="go('/platform/versions')">
            版本管理
          </a-button>
        </div>
      </div>
      <div class="overview-hero__panel">
        <CrownOutlined />
        <strong>{{ isTenantAdmin ? (summary?.tenantName || '客户子总控') : '平台运营中心' }}</strong>
        <span>{{ isTenantAdmin ? (summary?.tenantId || '--') : 'Multi-tenant Operation' }}</span>
      </div>
    </section>

    <a-spin :spinning="loading">
      <section class="metric-grid">
        <div class="metric-card metric-card--blue">
          <div class="metric-card__icon"><ShopOutlined /></div>
          <span>{{ isTenantAdmin ? '当前租户' : '合作客户' }}</span>
          <strong>{{ isTenantAdmin ? (summary?.tenantName || '--') : partnerTenants.length }}</strong>
          <small>{{ isTenantAdmin ? (summary?.status === 'disabled' ? '已停用' : '正常运营') : `${activeTenantCount} 个正常运营` }}</small>
        </div>
        <div class="metric-card metric-card--green">
          <div class="metric-card__icon"><ApartmentOutlined /></div>
          <span>下游机构</span>
          <strong>{{ isTenantAdmin ? (summary?.institutionCount || 0) : totalInstitutionCount }}</strong>
          <small>{{ isTenantAdmin ? '当前客户名下机构' : '所有客户机构总数' }}</small>
        </div>
        <div class="metric-card metric-card--purple">
          <div class="metric-card__icon"><GlobalOutlined /></div>
          <span>独立域名</span>
          <strong>{{ isTenantAdmin ? (summary?.domains?.length || 0) : totalDomainCount }}</strong>
          <small>{{ isTenantAdmin ? '当前客户域名' : '已配置客户域名' }}</small>
        </div>
        <div class="metric-card metric-card--orange">
          <div class="metric-card__icon"><SafetyCertificateOutlined /></div>
          <span>子总控账号</span>
          <strong>{{ isTenantAdmin ? (summary?.adminUsernames?.length || 0) : totalAdminCount }}</strong>
          <small>{{ isTenantAdmin ? '当前客户管理员' : '客户管理员账号数' }}</small>
        </div>
      </section>

      <section class="dashboard-grid">
        <div class="panel-card panel-card--large">
          <div class="panel-card__head">
            <div>
              <h3>{{ isTenantAdmin ? '租户开通状态' : '客户开通状态' }}</h3>
              <p>{{ isTenantAdmin ? '当前客户的机构、域名和账号配置情况。' : '关注客户是否完成机构绑定、域名配置和子总控账号开通。' }}</p>
            </div>
            <a-button v-if="!isTenantAdmin" type="link" @click="go('/platform/tenants')">
              管理租户
            </a-button>
          </div>

          <div v-if="!isTenantAdmin" class="tenant-list">
            <div v-for="item in recentTenants" :key="item.tenantId" class="tenant-row">
              <div class="tenant-row__main">
                <div class="tenant-row__avatar">
                  {{ item.tenantName?.slice(0, 1) || '租' }}
                </div>
                <div>
                  <strong>{{ item.tenantName }}</strong>
                  <span>{{ item.tenantId }}</span>
                </div>
              </div>
              <div class="tenant-row__stats">
                <span>{{ item.institutionCount || 0 }} 机构</span>
                <span>{{ item.domains?.length || 0 }} 域名</span>
                <span>{{ item.adminUsernames?.length || 0 }} 账号</span>
                <a-tag :color="tenantStatusColor(item.status)">{{ tenantStatusText(item.status) }}</a-tag>
              </div>
            </div>
            <a-empty v-if="!recentTenants.length" description="暂无合作客户" />
          </div>

          <div v-else class="tenant-self-card">
            <div>
              <span>租户名称</span>
              <strong>{{ summary?.tenantName || '--' }}</strong>
            </div>
            <div>
              <span>机构数量</span>
              <strong>{{ summary?.institutionCount || 0 }}</strong>
            </div>
            <div>
              <span>可用菜单</span>
              <strong>{{ summary?.menuCount || 0 }}</strong>
            </div>
            <div>
              <span>管理员</span>
              <strong>{{ summary?.adminUsernames?.join('、') || '--' }}</strong>
            </div>
          </div>
        </div>

        <div class="panel-card">
          <div class="panel-card__head">
            <div>
              <h3>运营待办</h3>
              <p>{{ isTenantAdmin ? '当前客户需要完善的运营事项。' : '需要平台侧继续处理的客户配置。' }}</p>
            </div>
            <ClockCircleOutlined />
          </div>

          <div class="todo-list">
            <div v-if="!isTenantAdmin" class="todo-item todo-item--warning">
              <ClockCircleOutlined />
              <div>
                <strong>{{ pendingTenants.length }} 个客户待完善</strong>
                <span>缺少域名、机构绑定或子总控账号</span>
              </div>
            </div>
            <div v-if="isTenantAdmin || !pendingTenants.length" class="todo-item todo-item--success">
              <CheckCircleOutlined />
              <div>
                <strong>{{ isTenantAdmin ? '租户环境已开通' : '核心配置正常' }}</strong>
                <span>{{ isTenantAdmin ? '如需增购版本或功能，请联系平台方。' : '暂无阻塞客户开通的事项。' }}</span>
              </div>
            </div>
            <div class="todo-item">
              <TeamOutlined />
              <div>
                <strong>机构运营</strong>
                <span>定期检查机构到期、停用和版本变更情况</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="panel-card domain-panel">
        <div class="panel-card__head">
          <div>
            <h3>{{ isTenantAdmin ? '当前客户域名' : '客户独立域名' }}</h3>
            <p>{{ isTenantAdmin ? '客户对外访问入口。' : '客户独立售卖时使用自己的访问域名。' }}</p>
          </div>
          <GlobalOutlined />
        </div>
        <div class="domain-list">
          <a-tag v-for="domain in currentDomains" :key="domain" color="blue">
            {{ domain }}
          </a-tag>
          <div v-if="!currentDomains.length" class="empty-box">
            暂未配置域名，请在租户管理中补充客户访问域名
          </div>
        </div>
      </section>
    </a-spin>
  </div>
</template>

<style scoped lang="less">
.control-overview {
  padding: 24px;
}

.overview-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  min-height: 220px;
  padding: 32px;
  border-radius: 20px;
  background:
    radial-gradient(circle at 82% 18%, rgba(255, 255, 255, 0.22), transparent 28%),
    linear-gradient(135deg, #12265a 0%, #1f5eff 100%);
  color: #fff;
  box-shadow: 0 18px 42px rgba(31, 94, 255, 0.22);

  &__content {
    max-width: 760px;
  }

  &__tag {
    margin-bottom: 14px;
  }

  h1 {
    margin: 0;
    color: #fff;
    font-size: 34px;
    font-weight: 700;
    letter-spacing: -0.02em;
  }

  p {
    max-width: 680px;
    margin: 12px 0 0;
    color: rgba(255, 255, 255, 0.78);
    font-size: 15px;
    line-height: 25px;
  }

  &__actions {
    display: flex;
    gap: 12px;
    margin-top: 26px;
  }

  &__panel {
    width: 220px;
    display: grid;
    place-items: center;
    align-content: center;
    gap: 10px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 22px;
    background: rgba(255, 255, 255, 0.12);
    text-align: center;

    .anticon {
      font-size: 40px;
    }

    strong {
      color: #fff;
      font-size: 18px;
    }

    span {
      color: rgba(255, 255, 255, 0.68);
      font-size: 12px;
    }
  }
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-top: 18px;
}

.metric-card,
.panel-card {
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 8px 26px rgba(15, 23, 42, 0.06);
}

.metric-card {
  padding: 20px;

  &__icon {
    width: 42px;
    height: 42px;
    display: grid;
    place-items: center;
    margin-bottom: 14px;
    border-radius: 14px;
    font-size: 20px;
  }

  span,
  small {
    display: block;
    color: rgba(0, 0, 0, 0.45);
  }

  strong {
    display: block;
    margin: 7px 0 4px;
    color: rgba(0, 0, 0, 0.88);
    font-size: 28px;
    line-height: 34px;
  }

  &--blue &__icon { color: #1677ff; background: #eaf3ff; }
  &--green &__icon { color: #13a86b; background: #eafaf3; }
  &--purple &__icon { color: #722ed1; background: #f5edff; }
  &--orange &__icon { color: #d46b08; background: #fff3e6; }
}

.dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(360px, 0.55fr);
  gap: 16px;
  margin-top: 18px;
}

.panel-card {
  padding: 20px;

  &__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 16px;

    h3 {
      margin: 0;
      color: rgba(0, 0, 0, 0.88);
      font-size: 18px;
      font-weight: 600;
    }

    p {
      margin: 5px 0 0;
      color: rgba(0, 0, 0, 0.45);
    }

    > .anticon {
      color: #1677ff;
      font-size: 22px;
    }
  }
}

.tenant-list {
  display: grid;
  gap: 10px;
}

.tenant-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 14px;
  background: #fff;

  &__main {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__avatar {
    width: 38px;
    height: 38px;
    display: grid;
    place-items: center;
    border-radius: 12px;
    background: #eaf3ff;
    color: #1677ff;
    font-weight: 700;
  }

  strong,
  span {
    display: block;
  }

  span {
    color: rgba(0, 0, 0, 0.45);
    font-size: 12px;
  }

  &__stats {
    display: flex;
    align-items: center;
    gap: 12px;
    color: rgba(0, 0, 0, 0.65);
  }
}

.tenant-self-card {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;

  > div {
    padding: 16px;
    border-radius: 14px;
    background: #f7f9fc;
  }

  span {
    display: block;
    color: rgba(0, 0, 0, 0.45);
  }

  strong {
    display: block;
    margin-top: 8px;
    color: rgba(0, 0, 0, 0.88);
    font-size: 18px;
  }
}

.todo-list {
  display: grid;
  gap: 12px;
}

.todo-item {
  display: flex;
  gap: 12px;
  padding: 14px;
  border-radius: 14px;
  background: #f7f9fc;

  .anticon {
    margin-top: 3px;
    color: #1677ff;
    font-size: 18px;
  }

  strong,
  span {
    display: block;
  }

  span {
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
  }

  &--warning .anticon { color: #fa8c16; }
  &--success .anticon { color: #13a86b; }
}

.domain-panel {
  margin-top: 18px;
}

.domain-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.empty-box {
  width: 100%;
  display: grid;
  place-items: center;
  min-height: 72px;
  border: 1px dashed rgba(0, 0, 0, 0.12);
  border-radius: 12px;
  color: rgba(0, 0, 0, 0.35);
  background: rgba(0, 0, 0, 0.015);
}

@media (max-width: 1200px) {
  .metric-grid,
  .dashboard-grid,
  .tenant-self-card {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .overview-hero {
    flex-direction: column;
  }
}
</style>
