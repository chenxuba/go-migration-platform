<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ApartmentOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  GlobalOutlined,
  PlusOutlined,
  RightOutlined,
  SafetyCertificateOutlined,
  ShopOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  ToolOutlined,
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
  adminDomains?: string[]
  institutionDomains?: string[]
  moduleCount?: number
  moduleNames?: string[]
  status?: string
  edition?: string
}

interface QuickAction {
  title: string
  desc: string
  path: string
  icon: any
  primary?: boolean
}

const router = useRouter()
const userStore = useUserStore()
const isTenantAdmin = computed(() => userStore.userInfo?.tenantRole === 'tenant_admin')
const loading = ref(false)
const summary = ref<ControlSummary>()
const tenantItems = ref<ControlSummary[]>([])

const partnerTenants = computed(() => tenantItems.value.filter(item => item.tenantType !== 'platform'))
const activeTenantCount = computed(() => partnerTenants.value.filter(item => item.status !== 'disabled').length)
const disabledTenantCount = computed(() => partnerTenants.value.filter(item => item.status === 'disabled').length)
const totalInstitutionCount = computed(() => partnerTenants.value.reduce((total, item) => total + Number(item.institutionCount || 0), 0))
const totalDomainCount = computed(() => partnerTenants.value.reduce((total, item) => total + getTenantDomainCount(item), 0))
const totalAdminCount = computed(() => partnerTenants.value.reduce((total, item) => total + (item.adminUsernames?.length || 0), 0))
const pendingTenants = computed(() => partnerTenants.value.filter(item => getTenantMissingItems(item).length > 0))
const readyTenants = computed(() => partnerTenants.value.filter(item => getTenantMissingItems(item).length === 0))
const averageReadiness = computed(() => {
  const items = isTenantAdmin.value && summary.value ? [summary.value] : partnerTenants.value
  if (!items.length)
    return 0
  return Math.round(items.reduce((sum, item) => sum + getTenantReadiness(item), 0) / items.length)
})
const displayedPendingTenants = computed(() => pendingTenants.value.slice(0, 4))
const displayedTenantMatrix = computed(() => partnerTenants.value.slice(0, 6))
const currentDomains = computed(() => isTenantAdmin.value ? getTenantDomains(summary.value) : partnerTenants.value.flatMap(item => getTenantDomains(item)))
const quickActions = computed<QuickAction[]>(() => {
  if (isTenantAdmin.value) {
    return [
      { title: '机构管理', desc: '查看当前租户下属机构', path: '/platform/organizations', icon: ApartmentOutlined, primary: true },
      { title: '版本管理', desc: '维护租户自主管理版本', path: '/platform/versions', icon: SafetyCertificateOutlined },
      { title: '默认角色', desc: '查看裁剪后的默认角色权限', path: '/platform/roles', icon: TeamOutlined },
    ]
  }
  return [
    { title: '开通合作客户', desc: '创建租户、账号和登录入口', path: '/platform/tenants', icon: PlusOutlined, primary: true },
    { title: '机构列表', desc: '查看机构归属和到期状态', path: '/platform/organizations', icon: ApartmentOutlined },
    { title: '版本管理', desc: '维护平台可授权版本', path: '/platform/versions', icon: SafetyCertificateOutlined },
    { title: '云存储配置', desc: '配置租户独立存储资源', path: '/platform/storage', icon: ToolOutlined },
  ]
})

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

function getTenantDomains(item?: ControlSummary) {
  if (!item)
    return []
  const values = [
    ...(item.adminDomains || []),
    ...(item.institutionDomains || []),
    ...(item.domains || []),
  ]
  return Array.from(new Set(values.map(domain => String(domain || '').trim()).filter(Boolean)))
}

function getTenantDomainCount(item?: ControlSummary) {
  return getTenantDomains(item).length
}

function getTenantMissingItems(item?: ControlSummary) {
  if (!item)
    return ['租户资料']
  const missing: string[] = []
  if (!getTenantDomains(item).length)
    missing.push('登录域名')
  if (!item.adminUsernames?.length)
    missing.push('子总控账号')
  if (!Number(item.institutionCount || 0))
    missing.push('机构归属')
  if (!Number(item.menuCount || 0) && !Number(item.moduleCount || 0))
    missing.push('版本授权')
  return missing
}

function getTenantReadiness(item?: ControlSummary) {
  if (!item)
    return 0
  const checks = [
    getTenantDomains(item).length > 0,
    Boolean(item.adminUsernames?.length),
    Number(item.institutionCount || 0) > 0,
    Number(item.menuCount || 0) > 0 || Number(item.moduleCount || 0) > 0,
  ]
  return Math.round((checks.filter(Boolean).length / checks.length) * 100)
}

function tenantStatusText(status?: string) {
  return status === 'disabled' ? '停用' : '正常'
}

function tenantStatusClass(status?: string) {
  return status === 'disabled' ? 'status-pill--disabled' : 'status-pill--success'
}

function readinessClass(value: number) {
  if (value >= 90)
    return 'readiness--good'
  if (value >= 60)
    return 'readiness--warning'
  return 'readiness--danger'
}

function metricValue(type: 'tenants' | 'pending' | 'institutions' | 'domains') {
  if (isTenantAdmin.value) {
    if (type === 'tenants')
      return summary.value?.tenantName || '--'
    if (type === 'pending')
      return getTenantMissingItems(summary.value).length
    if (type === 'institutions')
      return summary.value?.institutionCount || 0
    return getTenantDomainCount(summary.value)
  }
  if (type === 'tenants')
    return partnerTenants.value.length
  if (type === 'pending')
    return pendingTenants.value.length
  if (type === 'institutions')
    return totalInstitutionCount.value
  return totalDomainCount.value
}

function metricSubText(type: 'tenants' | 'pending' | 'institutions' | 'domains') {
  if (isTenantAdmin.value) {
    if (type === 'tenants')
      return summary.value?.status === 'disabled' ? '当前租户已停用' : '当前租户正常运营'
    if (type === 'pending')
      return getTenantMissingItems(summary.value).length ? '仍有配置需要完善' : '核心配置完整'
    if (type === 'institutions')
      return '当前租户下属机构'
    return '子总控与机构端入口'
  }
  if (type === 'tenants')
    return `${activeTenantCount.value} 正常 / ${disabledTenantCount.value} 停用`
  if (type === 'pending')
    return pendingTenants.value.length ? '建议优先处理' : '暂无待完善客户'
  if (type === 'institutions')
    return '全平台已归属机构'
  return '子总控与机构端域名'
}

onMounted(loadSummary)
</script>

<template>
  <div class="control-overview">
    <a-spin :spinning="loading">
      <section class="overview-hero">
        <div class="overview-hero__main">
          <a-tag class="overview-hero__tag" color="blue">
            {{ isTenantAdmin ? '客户子总控' : '超级总控运营台' }}
          </a-tag>
          <h1>{{ isTenantAdmin ? `${summary?.tenantName || '客户'}运营工作台` : '合作客户交付驾驶舱' }}</h1>
          <p>
            {{ isTenantAdmin
              ? '聚焦当前租户的机构归属、版本授权、登录域名和账号状态。'
              : '用于跟进合作客户开通进度、域名入口、机构归属和授权完整度。' }}
          </p>
          <div class="overview-hero__actions">
            <a-button v-for="action in quickActions.slice(0, 3)" :key="action.title" :type="action.primary ? 'primary' : 'default'" size="large" @click="go(action.path)">
              <template #icon>
                <component :is="action.icon" />
              </template>
              {{ action.title }}
            </a-button>
          </div>
        </div>

        <div class="overview-health-card" :class="readinessClass(averageReadiness)">
          <div class="overview-health-card__ring" :style="{ '--score': `${averageReadiness * 3.6}deg` }">
            <strong>{{ averageReadiness }}</strong>
            <span>%</span>
          </div>
          <div>
            <h3>开通完整度</h3>
            <p>{{ isTenantAdmin ? '当前租户核心交付配置' : '合作客户平均交付进度' }}</p>
            <a-tag :color="averageReadiness >= 90 ? 'success' : averageReadiness >= 60 ? 'warning' : 'error'">
              {{ averageReadiness >= 90 ? '交付健康' : averageReadiness >= 60 ? '需要跟进' : '风险较高' }}
            </a-tag>
          </div>
        </div>
      </section>

      <section class="metric-grid">
        <div class="metric-card metric-card--blue">
          <div class="metric-card__icon"><ShopOutlined /></div>
          <span>{{ isTenantAdmin ? '当前租户' : '合作客户' }}</span>
          <strong>{{ metricValue('tenants') }}</strong>
          <small>{{ metricSubText('tenants') }}</small>
        </div>
        <div class="metric-card metric-card--red">
          <div class="metric-card__icon"><ExclamationCircleOutlined /></div>
          <span>{{ isTenantAdmin ? '待完善项' : '待完善客户' }}</span>
          <strong>{{ metricValue('pending') }}</strong>
          <small>{{ metricSubText('pending') }}</small>
        </div>
        <div class="metric-card metric-card--green">
          <div class="metric-card__icon"><ApartmentOutlined /></div>
          <span>下游机构</span>
          <strong>{{ metricValue('institutions') }}</strong>
          <small>{{ metricSubText('institutions') }}</small>
        </div>
        <div class="metric-card metric-card--purple">
          <div class="metric-card__icon"><GlobalOutlined /></div>
          <span>登录入口</span>
          <strong>{{ metricValue('domains') }}</strong>
          <small>{{ metricSubText('domains') }}</small>
        </div>
      </section>

      <section class="overview-layout">
        <div class="panel-card delivery-panel">
          <div class="panel-card__head">
            <div>
              <h3>{{ isTenantAdmin ? '租户配置检查' : '待处理客户' }}</h3>
              <p>{{ isTenantAdmin ? '检查当前租户是否具备完整运营条件。' : '优先处理缺域名、缺机构、缺账号或缺授权的客户。' }}</p>
            </div>
            <a-button type="link" @click="go(isTenantAdmin ? '/platform/organizations' : '/platform/tenants')">
              去处理 <RightOutlined />
            </a-button>
          </div>

          <div v-if="!isTenantAdmin" class="pending-list">
            <div v-for="item in displayedPendingTenants" :key="item.tenantId" class="pending-row">
              <div class="pending-row__identity">
                <div class="tenant-avatar">{{ item.tenantName?.slice(0, 1) || '租' }}</div>
                <div>
                  <strong>{{ item.tenantName }}</strong>
                  <span>{{ item.tenantId }}</span>
                </div>
              </div>
              <div class="pending-row__missing">
                <a-tag v-for="missing in getTenantMissingItems(item)" :key="missing" color="orange">
                  {{ missing }}
                </a-tag>
              </div>
              <div class="pending-row__score" :class="readinessClass(getTenantReadiness(item))">
                {{ getTenantReadiness(item) }}%
              </div>
            </div>
            <div v-if="!displayedPendingTenants.length" class="success-empty">
              <CheckCircleOutlined />
              <strong>所有合作客户核心配置已完成</strong>
              <span>暂无需要立即处理的开通阻塞项。</span>
            </div>
          </div>

          <div v-else class="tenant-check-grid">
            <div v-for="missing in getTenantMissingItems(summary)" :key="missing" class="tenant-check-card tenant-check-card--warning">
              <ExclamationCircleOutlined />
              <strong>{{ missing }}</strong>
              <span>建议联系平台方补齐配置</span>
            </div>
            <div v-if="!getTenantMissingItems(summary).length" class="tenant-check-card tenant-check-card--success">
              <CheckCircleOutlined />
              <strong>核心配置完整</strong>
              <span>当前租户可正常运营</span>
            </div>
          </div>
        </div>

        <div class="panel-card quick-panel">
          <div class="panel-card__head">
            <div>
              <h3>快捷处理</h3>
              <p>常用运营动作入口。</p>
            </div>
            <ThunderboltOutlined />
          </div>
          <div class="quick-list">
            <button v-for="action in quickActions" :key="action.title" class="quick-action" :class="{ 'quick-action--primary': action.primary }" @click="go(action.path)">
              <component :is="action.icon" />
              <span>
                <strong>{{ action.title }}</strong>
                <small>{{ action.desc }}</small>
              </span>
              <RightOutlined />
            </button>
          </div>
        </div>
      </section>

      <section class="overview-layout overview-layout--bottom">
        <div class="panel-card matrix-panel">
          <div class="panel-card__head">
            <div>
              <h3>{{ isTenantAdmin ? '当前租户概览' : '客户交付矩阵' }}</h3>
              <p>{{ isTenantAdmin ? '当前租户核心资源统计。' : '查看客户开通完整度和资源配置情况。' }}</p>
            </div>
            <a-button v-if="!isTenantAdmin" type="link" @click="go('/platform/tenants')">
              全部客户
            </a-button>
          </div>

          <div v-if="!isTenantAdmin" class="tenant-matrix">
            <div v-for="item in displayedTenantMatrix" :key="item.tenantId" class="tenant-matrix-row">
              <div class="tenant-matrix-row__name">
                <strong>{{ item.tenantName }}</strong>
                <span>{{ item.tenantId }}</span>
              </div>
              <div class="tenant-matrix-row__bar">
                <i :style="{ width: `${getTenantReadiness(item)}%` }" :class="readinessClass(getTenantReadiness(item))" />
              </div>
              <div class="tenant-matrix-row__meta">
                <span>{{ item.institutionCount || 0 }} 机构</span>
                <span>{{ getTenantDomainCount(item) }} 入口</span>
                <a-tag :class="tenantStatusClass(item.status)">{{ tenantStatusText(item.status) }}</a-tag>
              </div>
            </div>
            <a-empty v-if="!displayedTenantMatrix.length" description="暂无合作客户" />
          </div>

          <div v-else class="tenant-self-grid">
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

        <div class="panel-card domain-panel">
          <div class="panel-card__head">
            <div>
              <h3>{{ isTenantAdmin ? '当前登录入口' : '登录入口总览' }}</h3>
              <p>{{ isTenantAdmin ? '当前客户对外登录地址。' : '客户子总控和机构端域名。' }}</p>
            </div>
            <GlobalOutlined />
          </div>
          <div class="domain-list">
            <span v-for="domain in currentDomains.slice(0, 10)" :key="domain" class="domain-chip">
              {{ domain }}
            </span>
            <div v-if="!currentDomains.length" class="empty-box">
              暂未配置域名，请在租户管理中补充客户访问域名
            </div>
          </div>
        </div>
      </section>
    </a-spin>
  </div>
</template>

<style scoped lang="less">
.control-overview {
  padding: 22px 24px 28px;
  background: #f5f7fb;
}

.overview-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 22px;
  min-height: 236px;
  padding: 28px 30px;
  border-radius: 24px;
  background:
    radial-gradient(circle at 68% 8%, rgba(255, 255, 255, 0.2), transparent 30%),
    linear-gradient(135deg, #101f4f 0%, #175cff 62%, #4096ff 100%);
  color: #fff;
  overflow: hidden;
  box-shadow: 0 18px 42px rgba(22, 85, 255, 0.2);
}

.overview-hero__main {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.overview-hero__tag {
  width: max-content;
  margin-bottom: 14px;
  border: 0;
  background: rgba(255, 255, 255, 0.92);
}

.overview-hero h1 {
  margin: 0;
  color: #fff;
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.overview-hero p {
  max-width: 720px;
  margin: 12px 0 0;
  color: rgba(255, 255, 255, 0.78);
  font-size: 15px;
  line-height: 25px;
}

.overview-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 28px;
}

.overview-health-card {
  display: grid;
  grid-template-columns: 124px minmax(0, 1fr);
  gap: 18px;
  align-items: center;
  align-self: center;
  min-height: 166px;
  padding: 22px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.13);
  backdrop-filter: blur(10px);

  h3 {
    margin: 0;
    color: #fff;
    font-size: 20px;
    font-weight: 700;
  }

  p {
    margin: 8px 0 14px;
    color: rgba(255, 255, 255, 0.72);
    font-size: 13px;
    line-height: 20px;
  }
}

.overview-health-card__ring {
  width: 112px;
  height: 112px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background:
    radial-gradient(circle at center, rgba(255, 255, 255, 0.95) 0 58%, transparent 59%),
    conic-gradient(#67e8f9 var(--score), rgba(255, 255, 255, 0.22) 0);
  color: #0f172a;

  strong {
    grid-area: 1 / 1;
    transform: translateX(-6px);
    font-size: 30px;
    line-height: 1;
  }

  span {
    grid-area: 1 / 1;
    transform: translate(26px, 4px);
    color: #64748b;
    font-size: 13px;
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
  border: 1px solid rgba(15, 23, 42, 0.06);
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.05);
}

.metric-card {
  position: relative;
  min-height: 142px;
  padding: 20px;
  overflow: hidden;

  &::after {
    position: absolute;
    right: -26px;
    bottom: -32px;
    width: 92px;
    height: 92px;
    border-radius: 50%;
    background: currentColor;
    content: '';
    opacity: 0.07;
  }

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
    max-width: 100%;
    margin: 7px 0 4px;
    color: rgba(0, 0, 0, 0.88);
    font-size: 28px;
    line-height: 34px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &--blue { color: #1677ff; }
  &--green { color: #13a86b; }
  &--purple { color: #722ed1; }
  &--red { color: #f97316; }

  &--blue &__icon { background: #eaf3ff; }
  &--green &__icon { background: #eafaf3; }
  &--purple &__icon { background: #f5edff; }
  &--red &__icon { background: #fff7ed; }
}

.overview-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(360px, 0.55fr);
  gap: 16px;
  margin-top: 18px;
}

.overview-layout--bottom {
  grid-template-columns: minmax(0, 1.3fr) minmax(360px, 0.7fr);
}

.panel-card {
  padding: 20px;
}

.panel-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;

  h3 {
    margin: 0;
    color: rgba(0, 0, 0, 0.88);
    font-size: 18px;
    font-weight: 700;
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

.pending-list,
.quick-list,
.tenant-matrix,
.domain-list {
  display: grid;
  gap: 12px;
}

.pending-row {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) minmax(220px, 1.1fr) 72px;
  gap: 14px;
  align-items: center;
  padding: 14px;
  border: 1px solid rgba(15, 23, 42, 0.06);
  border-radius: 16px;
  background: #fbfcff;
}

.pending-row__identity {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;

  strong,
  span {
    display: block;
  }

  strong {
    overflow: hidden;
    color: rgba(0, 0, 0, 0.88);
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    margin-top: 3px;
    color: rgba(0, 0, 0, 0.42);
    font-size: 12px;
  }
}

.tenant-avatar {
  width: 40px;
  height: 40px;
  display: grid;
  flex: none;
  place-items: center;
  border-radius: 13px;
  background: #eaf3ff;
  color: #1677ff;
  font-weight: 800;
}

.pending-row__missing {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.pending-row__score {
  justify-self: end;
  padding: 6px 10px;
  border-radius: 999px;
  font-weight: 700;
}

.readiness--good { color: #13a86b; background: rgba(19, 168, 107, 0.1); }
.readiness--warning { color: #d97706; background: rgba(245, 158, 11, 0.12); }
.readiness--danger { color: #dc2626; background: rgba(239, 68, 68, 0.1); }

.success-empty,
.tenant-check-card {
  display: grid;
  place-items: center;
  min-height: 168px;
  padding: 18px;
  border: 1px dashed rgba(19, 168, 107, 0.28);
  border-radius: 18px;
  background: rgba(19, 168, 107, 0.04);
  text-align: center;

  .anticon {
    color: #13a86b;
    font-size: 28px;
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-size: 16px;
  }

  span {
    color: rgba(0, 0, 0, 0.45);
  }
}

.tenant-check-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.tenant-check-card {
  min-height: 132px;

  &--warning {
    border-color: rgba(245, 158, 11, 0.3);
    background: rgba(245, 158, 11, 0.05);

    .anticon { color: #d97706; }
  }
}

.quick-list {
  gap: 10px;
}

.quick-action {
  width: 100%;
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) 16px;
  gap: 12px;
  align-items: center;
  padding: 13px;
  border: 1px solid rgba(15, 23, 42, 0.06);
  border-radius: 15px;
  background: #fbfcff;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;

  > .anticon:first-child {
    width: 40px;
    height: 40px;
    display: grid;
    place-items: center;
    border-radius: 13px;
    background: #eef5ff;
    color: #1677ff;
    font-size: 18px;
  }

  strong,
  small {
    display: block;
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-weight: 700;
  }

  small {
    margin-top: 3px;
    color: rgba(0, 0, 0, 0.45);
  }

  > .anticon:last-child {
    color: rgba(0, 0, 0, 0.28);
  }

  &:hover {
    border-color: rgba(22, 119, 255, 0.28);
    background: #f7fbff;
    transform: translateY(-1px);
  }
}

.quick-action--primary {
  background: linear-gradient(135deg, #f7fbff, #eef6ff);
}

.tenant-matrix-row {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr) 220px;
  gap: 14px;
  align-items: center;
  padding: 13px 0;
  border-bottom: 1px solid rgba(15, 23, 42, 0.06);

  &:last-child {
    border-bottom: 0;
  }
}

.tenant-matrix-row__name {
  min-width: 0;

  strong,
  span {
    display: block;
  }

  strong {
    overflow: hidden;
    color: rgba(0, 0, 0, 0.88);
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    margin-top: 3px;
    color: rgba(0, 0, 0, 0.42);
    font-size: 12px;
  }
}

.tenant-matrix-row__bar {
  height: 9px;
  overflow: hidden;
  border-radius: 999px;
  background: #eef2f7;

  i {
    display: block;
    height: 100%;
    border-radius: inherit;
  }

  .readiness--good { background: #13a86b; }
  .readiness--warning { background: #f59e0b; }
  .readiness--danger { background: #ef4444; }
}

.tenant-matrix-row__meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: rgba(0, 0, 0, 0.55);
  font-size: 12px;
}

.status-pill--success { color: #15803d; background: rgba(22, 163, 74, 0.1); border: 0; }
.status-pill--disabled { color: #64748b; background: rgba(100, 116, 139, 0.12); border: 0; }

.tenant-self-grid {
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

.domain-list {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.domain-chip {
  display: block;
  min-width: 0;
  padding: 9px 10px;
  border: 1px solid rgba(22, 119, 255, 0.12);
  border-radius: 10px;
  background: #f7fbff;
  color: #1677ff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-box {
  grid-column: 1 / -1;
  display: grid;
  place-items: center;
  min-height: 92px;
  border: 1px dashed rgba(0, 0, 0, 0.12);
  border-radius: 14px;
  color: rgba(0, 0, 0, 0.35);
  background: rgba(0, 0, 0, 0.015);
}

@media (max-width: 1280px) {
  .overview-hero,
  .overview-layout,
  .overview-layout--bottom {
    grid-template-columns: 1fr;
  }

  .metric-grid,
  .tenant-self-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .tenant-matrix-row,
  .pending-row {
    grid-template-columns: 1fr;
  }

  .tenant-matrix-row__meta,
  .pending-row__score {
    justify-self: start;
  }
}
</style>
