<template>
  <section ref="moduleRef" class="district-map-module">
    <div class="map-frame">
      <header class="map-toolbar">
        <div class="map-title">
          <span>辖区机构分布地图</span>
          <b>杭州市主城区 · {{ filteredPoints.length }} 家机构</b>
        </div>

        <div class="risk-legend" aria-label="风险等级图例">
          <button
            v-for="item in riskOptions"
            :key="item.value"
            type="button"
            :class="['legend-item', item.value, { active: riskFilter === item.value }]"
            @click="riskFilter = item.value"
          >
            <i />
            <span>{{ item.label }}</span>
          </button>
        </div>

        <div class="map-actions">
          <select v-model="activeDistrict" aria-label="选择区县">
            <option value="">全部机构</option>
            <option v-for="district in districtModels" :key="district.name" :value="district.name">
              {{ district.name }}
            </option>
          </select>
          <button type="button" @click="resetView">复位</button>
        </div>
      </header>

      <div ref="containerRef" class="three-map" @pointermove="handlePointerMove" @pointerleave="clearHover" @click="handleClick" />

      <Transition name="tooltip">
        <div
          v-if="activePopupPoint"
          class="institution-card"
          :class="activePopupPoint.risk"
          :style="{ left: `${tooltipPosition.left}px`, top: `${tooltipPosition.top}px` }"
          @click.stop
        >
          <strong>
            {{ activePopupPoint.name }}
            <em>{{ riskLabel(activePopupPoint.risk) }}</em>
          </strong>
          <p>所属区县 <b>{{ activePopupPoint.district }}</b></p>
          <p>今日课程 <b>{{ activePopupPoint.courses }}</b> 节</p>
          <p>到课率 <b>{{ activePopupPoint.attendance }}%</b></p>
          <p class="risk-text">风险：{{ activePopupPoint.riskText }}</p>
          <button type="button" @click="activeDistrict = activePopupPoint.district">下钻区县</button>
        </div>
      </Transition>

      <svg v-if="connectorLine.visible" class="popup-connector" aria-hidden="true">
        <defs>
          <filter id="connectorGlow" x="-80%" y="-80%" width="260%" height="260%">
            <feGaussianBlur stdDeviation="3" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <path class="connector-glow" :class="activePopupPoint?.risk" :d="connectorLine.path" />
        <path class="connector-core" :class="activePopupPoint?.risk" :d="connectorLine.path" />
        <circle class="connector-dot" :class="activePopupPoint?.risk" :cx="connectorLine.x1" :cy="connectorLine.y1" r="4" />
      </svg>

      <aside class="risk-panel">
        <p>风险机构统计</p>
        <span v-for="item in statRows" :key="item.value" :class="item.value">
          <i />
          {{ item.label }} <b>{{ item.count }}</b> 家
        </span>
        <strong>机构总数 {{ filteredPoints.length }} 家</strong>
      </aside>

      <footer class="map-feed">
        <div class="feed-heading">
          <span>实时动态</span>
          <b>{{ activeDistrict || '全市' }}</b>
        </div>
        <div class="feed-tabs">
          <span>排课</span>
          <span>点名</span>
          <span>收费</span>
          <span>审批</span>
        </div>
        <div class="feed-list">
          <article v-for="event in liveEvents" :key="event.id" :class="event.risk">
            <i>{{ event.icon }}</i>
            <strong>{{ event.name }}</strong>
            <span>{{ event.text }}</span>
            <em>{{ event.time }}</em>
          </article>
        </div>
      </footer>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js'
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js'
import { UnrealBloomPass } from 'three/examples/jsm/postprocessing/UnrealBloomPass.js'
import { OutputPass } from 'three/examples/jsm/postprocessing/OutputPass.js'
import hangzhouGeo from '../assets/hangzhou-330100-full.json'

type RiskType = 'normal' | 'focus' | 'warn' | 'danger'
type RiskFilter = RiskType | 'all'
type LonLat = [number, number]
type Ring = LonLat[]
type Polygon = Ring[]

interface GeoFeature {
  type: 'Feature'
  properties: {
    name: string
    center?: LonLat
    centroid?: LonLat
  }
  geometry: {
    type: 'Polygon' | 'MultiPolygon'
    coordinates: Polygon | Polygon[]
  }
}

interface DistrictModel {
  name: string
  color: number
  polygons: Array<Array<[number, number][]>>
  center: THREE.Vector3
}

interface InstitutionPoint {
  id: number
  name: string
  district: string
  x: number
  y: number
  risk: RiskType
  courses: number
  attendance: number
  riskText: string
}

interface GovernmentAgency {
  name: string
  x: number
  y: number
}

interface InstitutionFlow {
  pointId: number
  line: THREE.Line
  glow: THREE.Mesh
  comet: THREE.Sprite[]
  curve: THREE.CatmullRomCurve3
}

const moduleRef = ref<HTMLElement | null>(null)
const containerRef = ref<HTMLDivElement | null>(null)
const riskFilter = ref<RiskFilter>('all')
const activeDistrict = ref('')
const hoveredPoint = ref<InstitutionPoint | null>(null)
const selectedPoint = ref<InstitutionPoint | null>(null)
const autoPopupPoint = ref<InstitutionPoint | null>(null)
const tooltipPosition = ref({ left: 0, top: 0 })
const connectorLine = ref({ visible: false, path: '', x1: 0, y1: 0 })

const displayDistrictNames = ['上城区', '拱墅区', '西湖区', '滨江区', '萧山区', '余杭区', '钱塘区', '临平区']
const cityFeatures = ((hangzhouGeo as { features: GeoFeature[] }).features)
  .filter(feature => displayDistrictNames.includes(feature.properties.name))

const riskOptions: Array<{ value: RiskFilter; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'normal', label: '正常' },
  { value: 'focus', label: '关注' },
  { value: 'warn', label: '预警' },
  { value: 'danger', label: '高风险' },
]

const riskColors: Record<RiskType, number> = {
  normal: 0x45e58c,
  focus: 0xffc94d,
  warn: 0xff8a3d,
  danger: 0xff535a,
}

const districtColors = [0x0c58b3, 0x0b63c6, 0x0a4fa3, 0x0f72cf, 0x0a58b1, 0x0a4898, 0x0c5bb8, 0x0b66c2]
const projection = createProjection(cityFeatures)
const districtModels = cityFeatures.map((feature, index) => createDistrictModel(feature, index))
const governmentAgency = createGovernmentAgency()
const institutions = createInstitutionPoints()
const institutionMap = new Map(institutions.map(item => [item.id, item]))

const filteredPoints = computed(() =>
  institutions.filter(point =>
    (riskFilter.value === 'all' || point.risk === riskFilter.value)
    && (!activeDistrict.value || point.district === activeDistrict.value),
  ),
)

const statRows = computed(() =>
  (['danger', 'warn', 'focus', 'normal'] as RiskType[]).map(value => ({
    value,
    label: riskLabel(value),
    count: filteredPoints.value.filter(point => point.risk === value).length,
  })),
)

const activePopupPoint = computed(() => hoveredPoint.value || selectedPoint.value || autoPopupPoint.value)

const liveEvents = computed(() => {
  const source = filteredPoints.value.length ? filteredPoints.value : institutions
  return source.slice(0, 4).map((point, index) => ({
    id: `${point.id}-${index}`,
    name: point.name,
    text: index % 4 === 0 ? `新增排课 ${point.courses} 节` : index % 4 === 1 ? `完成点名 ${Math.round(point.attendance)}%` : index % 4 === 2 ? '收费完成' : '审批待处理',
    time: `09:${String(29 - index).padStart(2, '0')}:${String(45 - index * 8).padStart(2, '0')}`,
    icon: index % 4 === 0 ? '课' : index % 4 === 1 ? '点' : index % 4 === 2 ? '收' : '审',
    risk: point.risk,
  }))
})

let scene: THREE.Scene | undefined
let camera: THREE.PerspectiveCamera | undefined
let renderer: THREE.WebGLRenderer | undefined
let composer: EffectComposer | undefined
let controls: OrbitControls | undefined
let root: THREE.Group | undefined
let resizeObserver: ResizeObserver | undefined
let animationId = 0
let introStart = 0
let popupTimer = 0
let popupIndex = 0
let manualHoldUntil = 0
let cameraTween: {
  startTime: number
  duration: number
  fromPosition: THREE.Vector3
  toPosition: THREE.Vector3
  fromTarget: THREE.Vector3
  toTarget: THREE.Vector3
} | null = null

const pointer = new THREE.Vector2()
const raycaster = new THREE.Raycaster()
const clock = new THREE.Clock()
const districtMeshes: THREE.Mesh[] = []
const pointGroups = new Map<number, THREE.Group>()
const pointHitMeshes: THREE.Object3D[] = []
const dashedBorders: THREE.Line[] = []
const institutionFlows: InstitutionFlow[] = []
const disposableTextures: THREE.Texture[] = []
const defaultCameraPosition = new THREE.Vector3(0, -520, 600)
const defaultCameraTarget = new THREE.Vector3(0, 8, 0)
const cardWidth = 224
const cardHeight = 196
const cardOffsetX = 84
const cardOffsetYRatio = -0.52

function riskLabel(value: RiskFilter) {
  return {
    all: '全部',
    normal: '正常',
    focus: '关注',
    warn: '预警',
    danger: '高风险',
  }[value]
}

function resetView() {
  riskFilter.value = 'all'
  activeDistrict.value = ''
  selectedPoint.value = null
  hoveredPoint.value = null
  manualHoldUntil = 0
  flyTo(defaultCameraPosition, defaultCameraTarget, 720)
}

function initThree() {
  if (!containerRef.value) return

  const rect = containerRef.value.getBoundingClientRect()
  scene = new THREE.Scene()
  scene.fog = new THREE.Fog(0x04152e, 850, 1500)

  camera = new THREE.PerspectiveCamera(33, rect.width / rect.height, 1, 2200)
  camera.up.set(0, 0, 1)
  camera.position.copy(defaultCameraPosition)
  camera.lookAt(defaultCameraTarget)

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(rect.width, rect.height)
  renderer.setClearColor(0x000000, 0)
  renderer.outputColorSpace = THREE.SRGBColorSpace
  containerRef.value.appendChild(renderer.domElement)

  composer = new EffectComposer(renderer)
  composer.addPass(new RenderPass(scene, camera))
  composer.addPass(new UnrealBloomPass(new THREE.Vector2(rect.width, rect.height), 0.16, 0.16, 0.82))
  composer.addPass(new OutputPass())

  controls = new OrbitControls(camera, renderer.domElement)
  controls.enableDamping = true
  controls.dampingFactor = 0.08
  controls.enableRotate = false
  controls.enablePan = true
  controls.screenSpacePanning = true
  controls.minDistance = 390
  controls.maxDistance = 980
  controls.target.copy(defaultCameraTarget)
  controls.mouseButtons.LEFT = THREE.MOUSE.PAN
  controls.mouseButtons.MIDDLE = THREE.MOUSE.DOLLY
  controls.mouseButtons.RIGHT = THREE.MOUSE.PAN

  scene.add(new THREE.AmbientLight(0x6abfff, 1.25))
  const light = new THREE.DirectionalLight(0x9bdcff, 1.7)
  light.position.set(120, -360, 520)
  scene.add(light)

  root = new THREE.Group()
  root.position.z = -34
  root.scale.setScalar(0.95)
  scene.add(root)

  buildMap(root)
  refreshVisibleObjects()

  resizeObserver = new ResizeObserver(resizeRenderer)
  resizeObserver.observe(containerRef.value)
  introStart = performance.now()
  tick()
}

function buildMap(targetRoot: THREE.Group) {
  const floor = new THREE.Mesh(
    new THREE.PlaneGeometry(1040, 660),
    new THREE.MeshBasicMaterial({ color: 0x03152f, transparent: true, opacity: 0.34, side: THREE.DoubleSide }),
  )
  floor.position.z = -16
  targetRoot.add(floor)

  const grid = new THREE.GridHelper(1040, 28, 0x1b83f5, 0x0b315f)
  grid.rotation.x = Math.PI / 2
  grid.position.z = -14
  ;(grid.material as THREE.Material).transparent = true
  ;(grid.material as THREE.Material).opacity = 0.08
  targetRoot.add(grid)

  targetRoot.add(createRoadNetwork())

  districtModels.forEach((district, districtIndex) => {
    district.polygons.forEach((polygon) => {
      const shape = createShapeFromRings(polygon)
      const mesh = new THREE.Mesh(
        new THREE.ExtrudeGeometry(shape, {
          depth: 12,
          bevelEnabled: true,
          bevelThickness: 1.5,
          bevelSize: 1.5,
          bevelSegments: 1,
        }),
        new THREE.MeshStandardMaterial({
          color: district.color,
          emissive: 0x0b5cc6,
          emissiveIntensity: 0.1,
          metalness: 0.12,
          roughness: 0.72,
          transparent: true,
          opacity: 0.38,
        }),
      )
      mesh.userData.district = district.name
      targetRoot.add(mesh)
      districtMeshes.push(mesh)

      const border = createBorderLine(polygon[0], 18 + districtIndex * 0.25)
      targetRoot.add(border)
      dashedBorders.push(border)

      const glowBorder = new THREE.Line(
        new THREE.BufferGeometry().setFromPoints(polygon[0].map(([x, y]) => new THREE.Vector3(x, y, 18))),
        new THREE.LineBasicMaterial({ color: 0x35cfff, transparent: true, opacity: 0.2, blending: THREE.AdditiveBlending }),
      )
      targetRoot.add(glowBorder)
    })

    const label = createTextSprite(district.name, 32, 'rgba(217,239,255,.76)', 'rgba(48,172,255,.72)')
    label.position.copy(district.center)
    label.position.z = 36
    label.scale.set(86, 25, 1)
    targetRoot.add(label)
  })

  createInstitutionFlows(targetRoot)

  institutions.forEach((point, index) => {
    const group = createPointMarker(point, index)
    targetRoot.add(group)
    pointGroups.set(point.id, group)
  })

  targetRoot.add(createGovernmentMarker())
}

function createRoadNetwork() {
  const group = new THREE.Group()
  const segments: THREE.Vector3[] = []
  const bounds = projection.projectedBounds

  for (let y = bounds.minY + 20; y <= bounds.maxY - 20; y += 23) {
    let prev: THREE.Vector3 | null = null
    for (let x = bounds.minX; x <= bounds.maxX; x += 18) {
      const point = new THREE.Vector3(x, y + Math.sin((x + y) * 0.025) * 4, 16)
      if (pointInsideCity(point.x, point.y)) {
        if (prev) segments.push(prev, point)
        prev = point
      } else {
        prev = null
      }
    }
  }

  for (let x = bounds.minX + 18; x <= bounds.maxX - 18; x += 30) {
    let prev: THREE.Vector3 | null = null
    for (let y = bounds.minY; y <= bounds.maxY; y += 20) {
      const point = new THREE.Vector3(x + Math.cos((x - y) * 0.023) * 3, y, 16.5)
      if (pointInsideCity(point.x, point.y)) {
        if (prev) segments.push(prev, point)
        prev = point
      } else {
        prev = null
      }
    }
  }

  group.add(new THREE.LineSegments(
    new THREE.BufferGeometry().setFromPoints(segments),
    new THREE.LineBasicMaterial({ color: 0x48a9ff, transparent: true, opacity: 0.13, blending: THREE.AdditiveBlending }),
  ))
  return group
}

function createInstitutionFlows(targetRoot: THREE.Group) {
  institutions.forEach((point, index) => {
    const start = new THREE.Vector3(point.x, point.y, 25)
    const end = new THREE.Vector3(governmentAgency.x, governmentAgency.y, 46)
    const distance = start.distanceTo(end)
    const mid = new THREE.Vector3(
      point.x + (governmentAgency.x - point.x) * 0.52,
      point.y + (governmentAgency.y - point.y) * 0.52,
      48 + Math.min(58, distance * 0.16) + Math.sin(index * 1.7) * 6,
    )
    const curve = new THREE.CatmullRomCurve3([start, mid, end])
    const line = new THREE.Line(
      new THREE.BufferGeometry().setFromPoints(curve.getPoints(128)),
      new THREE.LineBasicMaterial({
        color: 0x49e6ff,
        transparent: true,
        opacity: 0.18,
        blending: THREE.AdditiveBlending,
      }),
    )
    line.userData.pointId = point.id
    line.userData.baseOpacity = 0.18
    targetRoot.add(line)

    const glow = new THREE.Mesh(
      new THREE.TubeGeometry(curve, 96, 0.42, 6, false),
      new THREE.MeshBasicMaterial({
        color: 0x48e7ff,
        transparent: true,
        opacity: 0.055,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
      }),
    )
    glow.userData.pointId = point.id
    targetRoot.add(glow)

    const comet = createFlowComet(point, index)
    comet.forEach(sprite => targetRoot.add(sprite))

    institutionFlows.push({ pointId: point.id, line, glow, comet, curve })
  })
}

function createFlowComet(point: InstitutionPoint, flowIndex: number) {
  const sprites: THREE.Sprite[] = []
  const texture = getFlowCometTexture()
  const warmColor = point.risk === 'danger' || point.risk === 'warn' ? 0xff9b3d : 0xffd76b

  for (let index = 0; index < 8; index += 1) {
    const material = new THREE.SpriteMaterial({
      map: texture,
      color: index === 0 ? 0xffffff : warmColor,
      transparent: true,
      opacity: index === 0 ? 0.92 : Math.max(0.1, 0.52 - index * 0.055),
      depthWrite: false,
      blending: THREE.AdditiveBlending,
    })
    const sprite = new THREE.Sprite(material)
    sprite.userData.pointId = point.id
    sprite.userData.flowOffset = flowIndex * 0.073
    sprite.userData.speed = 0.105 + (flowIndex % 5) * 0.008
    sprite.userData.tailIndex = index
    sprite.userData.baseSize = index === 0 ? 12 : Math.max(4.4, 10 - index * 0.78)
    sprite.visible = false
    sprites.push(sprite)
  }

  return sprites
}

function createGovernmentMarker() {
  const group = new THREE.Group()
  group.position.set(governmentAgency.x, governmentAgency.y, 32)

  const base = new THREE.Mesh(
    new THREE.RingGeometry(13, 29, 72),
    new THREE.MeshBasicMaterial({
      color: 0x52eaff,
      transparent: true,
      opacity: 0.26,
      side: THREE.DoubleSide,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
    }),
  )
  base.userData.govHalo = true
  group.add(base)

  const pillar = new THREE.Mesh(
    new THREE.CylinderGeometry(7, 10, 28, 24),
    new THREE.MeshStandardMaterial({
      color: 0x38dfff,
      emissive: 0x38dfff,
      emissiveIntensity: 0.28,
      transparent: true,
      opacity: 0.88,
      metalness: 0.18,
      roughness: 0.42,
    }),
  )
  pillar.rotation.x = Math.PI / 2
  pillar.position.z = 16
  group.add(pillar)

  const icon = new THREE.Sprite(new THREE.SpriteMaterial({
    map: getGovernmentTexture(),
    transparent: true,
    opacity: 0.94,
    depthWrite: false,
  }))
  icon.scale.set(36, 36, 1)
  icon.position.z = 52
  group.add(icon)

  const label = createTextSprite(governmentAgency.name, 26, 'rgba(232,250,255,.92)', 'rgba(66,226,255,.84)')
  label.position.set(0, -34, 58)
  label.scale.set(120, 30, 1)
  group.add(label)

  return group
}

function createPointMarker(point: InstitutionPoint, index: number) {
  const group = new THREE.Group()
  const color = riskColors[point.risk]
  group.position.set(point.x, point.y, 25)
  group.userData.pointId = point.id
  group.userData.offset = index * 0.55

  const pointer = new THREE.Sprite(new THREE.SpriteMaterial({
    map: getMarkerTexture(point.risk),
    transparent: true,
    opacity: 0.94,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
  }))
  pointer.scale.set(18, 23, 1)
  pointer.center.set(0.5, 0.0625)
  pointer.position.z = 0
  pointer.userData.pointer = true
  group.add(pointer)
  group.userData.pinHoleZ = 27

  const icon = new THREE.Sprite(new THREE.SpriteMaterial({
    map: getIconTexture(point.risk),
    transparent: true,
    opacity: 0.86,
    depthWrite: false,
  }))
  icon.scale.set(13, 13, 1)
  icon.position.z = 27.5
  icon.visible = false
  group.add(icon)

  const hit = new THREE.Mesh(
    new THREE.SphereGeometry(14, 12, 12),
    new THREE.MeshBasicMaterial({ transparent: true, opacity: 0, depthWrite: false }),
  )
  hit.position.z = 26
  hit.userData.pointId = point.id
  group.add(hit)
  pointHitMeshes.push(hit)

  return group
}

function handlePointerMove(event: PointerEvent) {
  if (!renderer || !camera) return
  const rect = renderer.domElement.getBoundingClientRect()
  pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
  pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
  raycaster.setFromCamera(pointer, camera)

  const pointHit = raycaster.intersectObjects(pointHitMeshes.filter(mesh => mesh.visible && mesh.parent?.visible), false)[0]
  if (pointHit) {
    const point = getPointFromObject(pointHit.object)
    if (point) {
      hoveredPoint.value = point
      updateTooltipPosition(point.id)
      containerRef.value?.classList.add('is-hovering')
      return
    }
  }

  hoveredPoint.value = null
  containerRef.value?.classList.remove('is-hovering')
}

function handleClick() {
  if (!camera) return
  const pointHit = raycaster.intersectObjects(pointHitMeshes.filter(mesh => mesh.visible && mesh.parent?.visible), false)[0]
  if (pointHit) {
    const point = getPointFromObject(pointHit.object)
    if (point) {
      selectedPoint.value = point
      autoPopupPoint.value = null
      manualHoldUntil = Date.now() + 8000
      syncCarouselIndex(point)
      updateTooltipPosition(point.id)
      return
    }
  }

  const districtHit = raycaster.intersectObjects(districtMeshes, false)[0]
  if (districtHit?.object.userData.district) {
    activeDistrict.value = districtHit.object.userData.district
    selectedPoint.value = null
    manualHoldUntil = 0
  } else {
    selectedPoint.value = null
    manualHoldUntil = 0
  }
}

function clearHover() {
  hoveredPoint.value = null
  containerRef.value?.classList.remove('is-hovering')
}

function refreshVisibleObjects() {
  pointGroups.forEach((group, id) => {
    const point = institutionMap.get(id)
    group.visible = !!point
      && (riskFilter.value === 'all' || point.risk === riskFilter.value)
      && (!activeDistrict.value || point.district === activeDistrict.value)
  })

  districtMeshes.forEach((mesh) => {
    const districtName = mesh.userData.district as string
    const active = !activeDistrict.value || activeDistrict.value === districtName
    mesh.visible = active
    const material = mesh.material as THREE.MeshStandardMaterial
    material.opacity = activeDistrict.value && active ? 0.48 : 0.38
  })

  institutionFlows.forEach(({ pointId, line, glow, comet }) => {
    const point = institutionMap.get(pointId)
    const visible = !!point
      && (riskFilter.value === 'all' || point.risk === riskFilter.value)
      && (!activeDistrict.value || point.district === activeDistrict.value)
    line.visible = visible
    glow.visible = visible
    comet.forEach(sprite => {
      sprite.visible = visible
    })
  })
}

function startPopupCarousel() {
  stopPopupCarousel()
  resetCarouselPoint()
  popupTimer = window.setInterval(advanceCarouselPoint, 3600)
}

function stopPopupCarousel() {
  if (!popupTimer) return
  window.clearInterval(popupTimer)
  popupTimer = 0
}

function resetCarouselPoint() {
  popupIndex = 0
  selectedPoint.value = null
  manualHoldUntil = 0
  const nextPoint = getCarouselPoints()[0] || null
  autoPopupPoint.value = nextPoint
  if (nextPoint) requestAnimationFrame(() => updateTooltipPosition(nextPoint.id))
}

function advanceCarouselPoint() {
  if (hoveredPoint.value) return

  if (selectedPoint.value) {
    if (Date.now() < manualHoldUntil) return
    selectedPoint.value = null
  }

  const source = getCarouselPoints()
  if (!source.length) {
    autoPopupPoint.value = null
    return
  }

  popupIndex = (popupIndex + 1) % source.length
  autoPopupPoint.value = source[popupIndex]
  requestAnimationFrame(() => updateTooltipPosition(source[popupIndex].id))
}

function getCarouselPoints() {
  return filteredPoints.value
}

function syncCarouselIndex(point: InstitutionPoint) {
  const index = getCarouselPoints().findIndex(item => item.id === point.id)
  popupIndex = index >= 0 ? index : 0
}

function updateTooltipPosition(pointId: number) {
  if (!camera || !renderer || !moduleRef.value) return
  const group = pointGroups.get(pointId)
  if (!group) return

  const world = new THREE.Vector3()
  group.getWorldPosition(world)
  world.z += group.userData.pinHoleZ || 26
  const projected = world.project(camera)
  const canvasRect = renderer.domElement.getBoundingClientRect()
  const moduleRect = moduleRef.value.getBoundingClientRect()
  const left = canvasRect.left - moduleRect.left + (projected.x * 0.5 + 0.5) * canvasRect.width
  const top = canvasRect.top - moduleRect.top + (-projected.y * 0.5 + 0.5) * canvasRect.height
  const pointScreen = { x: left - 1, y: top - 2 }
  tooltipPosition.value = {
    left: Math.max(140, Math.min(left, moduleRect.width - cardWidth - cardOffsetX - 24)),
    top: Math.max(196, Math.min(top, moduleRect.height - 190)),
  }
  updateConnectorLine(pointScreen, tooltipPosition.value)
}

function updateConnectorLine(pointScreen: { x: number; y: number }, cardAnchor: { left: number; top: number }) {
  const cardLeft = cardAnchor.left + cardOffsetX
  const cardTop = cardAnchor.top + cardHeight * cardOffsetYRatio
  const cardRight = cardLeft + cardWidth
  const cardMinY = cardTop + 38
  const cardMaxY = cardTop + cardHeight - 38
  const targetX = pointScreen.x <= cardLeft ? cardLeft : pointScreen.x >= cardRight ? cardRight : pointScreen.x
  const targetY = Math.max(cardMinY, Math.min(pointScreen.y, cardMaxY))
  const midX = pointScreen.x + (targetX - pointScreen.x) * 0.58
  const midY = pointScreen.y + (targetY - pointScreen.y) * 0.22

  connectorLine.value = {
    visible: true,
    x1: pointScreen.x,
    y1: pointScreen.y,
    path: `M ${pointScreen.x.toFixed(1)} ${pointScreen.y.toFixed(1)} Q ${midX.toFixed(1)} ${midY.toFixed(1)} ${targetX.toFixed(1)} ${targetY.toFixed(1)}`,
  }
}

function tick() {
  const elapsed = clock.getElapsedTime()
  const now = performance.now()

  if (root) {
    const progress = easeOutCubic(Math.min(1, (now - introStart) / 900))
    root.position.z = -34 + progress * 34
    root.scale.setScalar(0.95 + progress * 0.05)
  }

  if (cameraTween && camera && controls) {
    const progress = easeInOutCubic(Math.min(1, (now - cameraTween.startTime) / cameraTween.duration))
    camera.position.lerpVectors(cameraTween.fromPosition, cameraTween.toPosition, progress)
    controls.target.lerpVectors(cameraTween.fromTarget, cameraTween.toTarget, progress)
    if (progress >= 1) cameraTween = null
  }

  controls?.update()

  dashedBorders.forEach((border, index) => {
    ;(border.material as THREE.LineDashedMaterial).dashOffset = -elapsed * (18 + index * 1.5)
  })

  institutionFlows.forEach(({ line, glow, comet, curve }, index) => {
    if (!line.visible) return
    const material = line.material as THREE.LineBasicMaterial
    material.opacity = (line.userData.baseOpacity || 0.18) * (0.78 + Math.sin(elapsed * 2.1 + index) * 0.18)
    const glowMaterial = glow.material as THREE.MeshBasicMaterial
    glowMaterial.opacity = 0.045 + Math.sin(elapsed * 1.7 + index) * 0.012

    comet.forEach((sprite) => {
      const tailIndex = sprite.userData.tailIndex as number
      const progress = (elapsed * sprite.userData.speed + sprite.userData.flowOffset) % 1
      const tailProgress = progress - tailIndex * 0.012
      if (tailProgress <= 0 || tailProgress >= 1) {
        sprite.visible = false
        return
      }

      sprite.visible = true
      sprite.position.copy(curve.getPointAt(1 - tailProgress))
      const baseSize = sprite.userData.baseSize as number
      const fade = 1 - tailIndex / Math.max(1, comet.length - 1)
      sprite.scale.setScalar(baseSize * (0.86 + Math.sin(elapsed * 5.4 + index) * 0.08) * (0.78 + fade * 0.28))
    })
  })

  pointGroups.forEach((group) => {
    if (!group.visible) return
    const point = institutionMap.get(group.userData.pointId)
    const offset = group.userData.offset || 0
    const isActive = point?.id === activePopupPoint.value?.id
    const pulse = point?.risk === 'danger' ? 1 + Math.sin(elapsed * 5 + offset) * 0.08 : 1 + Math.sin(elapsed * 2 + offset) * 0.035
    group.scale.setScalar(pulse * (isActive ? 1.16 : 1))
    group.position.z = 25
    group.children.forEach((child) => {
      if (child.userData.halo) child.scale.setScalar((isActive ? 1.28 : 1) + Math.sin(elapsed * 2.4 + offset) * 0.08)
    })
  })

  const popup = activePopupPoint.value
  if (popup) updateTooltipPosition(popup.id)
  else connectorLine.value.visible = false
  composer?.render()
  animationId = requestAnimationFrame(tick)
}

function createProjection(features: GeoFeature[]) {
  const coords: LonLat[] = []
  features.forEach(feature => forEachCoordinate(feature, coord => coords.push(coord)))
  const minLon = Math.min(...coords.map(coord => coord[0]))
  const maxLon = Math.max(...coords.map(coord => coord[0]))
  const minLat = Math.min(...coords.map(coord => coord[1]))
  const maxLat = Math.max(...coords.map(coord => coord[1]))
  const centerLon = (minLon + maxLon) / 2
  const centerLat = (minLat + maxLat) / 2
  const latScale = Math.cos((centerLat * Math.PI) / 180)
  const scale = Math.min(880 / ((maxLon - minLon) * latScale), 530 / (maxLat - minLat))
  const project = ([lon, lat]: LonLat): [number, number] => [
    (lon - centerLon) * latScale * scale,
    (lat - centerLat) * scale,
  ]
  const projected = coords.map(project)
  return {
    project,
    projectedBounds: {
      minX: Math.min(...projected.map(coord => coord[0])),
      maxX: Math.max(...projected.map(coord => coord[0])),
      minY: Math.min(...projected.map(coord => coord[1])),
      maxY: Math.max(...projected.map(coord => coord[1])),
    },
  }
}

function createDistrictModel(feature: GeoFeature, index: number): DistrictModel {
  const polygons = normalizePolygons(feature).map(polygon =>
    polygon.map(ring => ring.map(coord => projection.project(coord))),
  )
  const centerCoord = feature.properties.centroid || feature.properties.center || polygonCenter(normalizePolygons(feature)[0][0])
  const [x, y] = projection.project(centerCoord)
  return {
    name: feature.properties.name,
    color: districtColors[index % districtColors.length],
    polygons,
    center: new THREE.Vector3(x, y, 0),
  }
}

function createGovernmentAgency(): GovernmentAgency {
  const shangcheng = districtModels.find(item => item.name === '上城区')
  const gongshu = districtModels.find(item => item.name === '拱墅区')
  const binjiang = districtModels.find(item => item.name === '滨江区')
  const centerX = ((shangcheng?.center.x || 0) + (gongshu?.center.x || 0) + (binjiang?.center.x || 0)) / 3
  const centerY = ((shangcheng?.center.y || 0) + (gongshu?.center.y || 0) + (binjiang?.center.y || 0)) / 3
  return {
    name: '政府监管中心',
    x: centerX + 34,
    y: centerY + 20,
  }
}

function createInstitutionPoints(): InstitutionPoint[] {
  const plans: Array<Omit<InstitutionPoint, 'id' | 'x' | 'y'> & { offset: [number, number] }> = [
    { name: '星启康复中心', district: '上城区', risk: 'focus', courses: 46, attendance: 89, riskText: '课消异常', offset: [8, -12] },
    { name: '启智康复中心', district: '拱墅区', risk: 'normal', courses: 32, attendance: 94, riskText: '运行正常', offset: [-18, 8] },
    { name: '阳光康复中心', district: '西湖区', risk: 'warn', courses: 38, attendance: 82, riskText: '到课率下降', offset: [-34, -16] },
    { name: '未来星康复中心', district: '滨江区', risk: 'danger', courses: 24, attendance: 76, riskText: '退费集中', offset: [10, 6] },
    { name: '希望之家', district: '萧山区', risk: 'normal', courses: 29, attendance: 93, riskText: '数据稳定', offset: [4, -18] },
    { name: '蓝湾康复中心', district: '余杭区', risk: 'focus', courses: 35, attendance: 88, riskText: '续费待跟进', offset: [-16, 12] },
    { name: '晨星融合教育', district: '钱塘区', risk: 'warn', courses: 27, attendance: 84, riskText: '师资负载偏高', offset: [18, -2] },
    { name: '同心训练中心', district: '临平区', risk: 'normal', courses: 31, attendance: 95, riskText: '服务达标', offset: [-10, 4] },
    { name: '康桥儿童中心', district: '西湖区', risk: 'normal', courses: 22, attendance: 92, riskText: '运行正常', offset: [42, 18] },
    { name: '新声康复中心', district: '萧山区', risk: 'danger', courses: 18, attendance: 71, riskText: '资金监管异常', offset: [48, 14] },
    { name: '瑞康训练中心', district: '余杭区', risk: 'normal', courses: 26, attendance: 96, riskText: '运行正常', offset: [-78, 34] },
    { name: '云帆康复中心', district: '余杭区', risk: 'focus', courses: 33, attendance: 87, riskText: '续费待跟进', offset: [-56, -32] },
    { name: '明德儿童中心', district: '余杭区', risk: 'normal', courses: 21, attendance: 94, riskText: '数据稳定', offset: [70, -45] },
    { name: '海辰融合教育', district: '西湖区', risk: 'normal', courses: 28, attendance: 91, riskText: '运行正常', offset: [-35, -45] },
    { name: '启明康复中心', district: '西湖区', risk: 'focus', courses: 30, attendance: 86, riskText: '到课率下降', offset: [0, 60] },
    { name: '嘉禾训练中心', district: '萧山区', risk: 'normal', courses: 25, attendance: 93, riskText: '服务达标', offset: [-48, -86] },
    { name: '新桥康复中心', district: '萧山区', risk: 'warn', courses: 34, attendance: 81, riskText: '排课异常', offset: [28, -88] },
    { name: '星湾儿童中心', district: '钱塘区', risk: 'normal', courses: 20, attendance: 95, riskText: '运行正常', offset: [-92, -34] },
    { name: '东岸康复中心', district: '钱塘区', risk: 'danger', courses: 19, attendance: 73, riskText: '资金监管异常', offset: [56, -4] },
    { name: '临康训练中心', district: '临平区', risk: 'normal', courses: 24, attendance: 92, riskText: '数据稳定', offset: [35, -45] },
  ]

  const placedPoints: Array<[number, number]> = []
  return plans.map((plan, index) => {
    const district = districtModels.find(item => item.name === plan.district)!
    const [x, y] = resolveInstitutionPosition(district, plan.offset, placedPoints)
    placedPoints.push([x, y])
    return {
      id: index + 1,
      name: plan.name,
      district: plan.district,
      risk: plan.risk,
      courses: plan.courses,
      attendance: plan.attendance,
      riskText: plan.riskText,
      x,
      y,
    }
  })
}

function resolveInstitutionPosition(district: DistrictModel, offset: [number, number], placedPoints: Array<[number, number]>) {
  const center: [number, number] = [district.center.x, district.center.y]
  const preferred: [number, number] = [center[0] + offset[0], center[1] + offset[1]]
  const candidates = createDistrictCandidates(district, 22)
  if (!candidates.length) return center
  const candidatePool = selectCandidatePool(candidates, placedPoints)

  let bestPoint = candidatePool[0]
  let bestScore = Number.NEGATIVE_INFINITY

  for (const point of candidatePool) {
    const minPlacedDistance = placedPoints.length
      ? Math.min(...placedPoints.map(placed => pointDistance(point, placed)))
      : 160
    const nearestDistrictLabelDistance = Math.min(...districtModels.map(item => pointDistance(point, [item.center.x, item.center.y])))
    const governmentDistance = pointDistance(point, [governmentAgency.x, governmentAgency.y])
    const centerDistance = pointDistance(point, center)
    const preferredDistance = pointDistance(point, preferred)
    const overlapPenalty = minPlacedDistance < 72 ? (72 - minPlacedDistance) * 14 : 0
    const labelPenalty = nearestDistrictLabelDistance < 50 ? (50 - nearestDistrictLabelDistance) * 6 : 0
    const governmentPenalty = governmentDistance < 86 ? (86 - governmentDistance) * 10 : 0
    const score = Math.min(minPlacedDistance, 170) * 4.1
      + Math.min(centerDistance, 130) * 0.55
      - preferredDistance * 0.32
      - overlapPenalty
      - labelPenalty
      - governmentPenalty

    if (score > bestScore) {
      bestScore = score
      bestPoint = point
    }
  }

  return bestPoint
}

function selectCandidatePool(candidates: Array<[number, number]>, placedPoints: Array<[number, number]>) {
  const attempts = [
    { placed: 70, label: 56, government: 90 },
    { placed: 62, label: 48, government: 78 },
    { placed: 54, label: 42, government: 66 },
    { placed: 44, label: 34, government: 56 },
  ]

  for (const attempt of attempts) {
    const pool = candidates.filter((point) => {
      const minPlacedDistance = placedPoints.length
        ? Math.min(...placedPoints.map(placed => pointDistance(point, placed)))
        : Number.POSITIVE_INFINITY
      const nearestDistrictLabelDistance = Math.min(...districtModels.map(item => pointDistance(point, [item.center.x, item.center.y])))
      const governmentDistance = pointDistance(point, [governmentAgency.x, governmentAgency.y])
      return minPlacedDistance >= attempt.placed
        && nearestDistrictLabelDistance >= attempt.label
        && governmentDistance >= attempt.government
    })
    if (pool.length) return pool
  }

  return candidates
}

function createDistrictCandidates(district: DistrictModel, safeDistance: number) {
  const ring = district.polygons[0][0]
  const xs = ring.map(point => point[0])
  const ys = ring.map(point => point[1])
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const minY = Math.min(...ys)
  const maxY = Math.max(...ys)
  const candidates: Array<[number, number]> = []

  for (let x = minX + safeDistance; x <= maxX - safeDistance; x += 20) {
    for (let y = minY + safeDistance; y <= maxY - safeDistance; y += 20) {
      const point: [number, number] = [x, y]
      if (isSafeDistrictPoint(district, point, safeDistance)) candidates.push(point)
    }
  }

  return candidates.length ? candidates : [[district.center.x, district.center.y] as [number, number]]
}

function isSafeDistrictPoint(district: DistrictModel, point: [number, number], safeDistance = 24) {
  return district.polygons.some((polygon) => (
    pointInRing(point, polygon[0])
    && distanceToRing(point, polygon[0]) >= safeDistance
  ))
}

function createShapeFromRings(polygon: Array<[number, number][]>) {
  const shape = new THREE.Shape()
  polygon[0].forEach(([x, y], index) => {
    if (index === 0) shape.moveTo(x, y)
    else shape.lineTo(x, y)
  })
  shape.closePath()

  polygon.slice(1).forEach((ring) => {
    const hole = new THREE.Path()
    ring.forEach(([x, y], index) => {
      if (index === 0) hole.moveTo(x, y)
      else hole.lineTo(x, y)
    })
    hole.closePath()
    shape.holes.push(hole)
  })
  return shape
}

function createBorderLine(ring: Array<[number, number]>, z: number) {
  const points = ring.map(([x, y]) => new THREE.Vector3(x, y, z))
  return new THREE.Line(
    new THREE.BufferGeometry().setFromPoints(points),
    new THREE.LineDashedMaterial({
      color: 0x48dfff,
      dashSize: 16,
      gapSize: 9,
      transparent: true,
      opacity: 0.58,
    }),
  ).computeLineDistances()
}

const iconTextures = new Map<RiskType, THREE.CanvasTexture>()
const markerTextures = new Map<RiskType, THREE.CanvasTexture>()
let governmentTexture: THREE.CanvasTexture | undefined
let flowCometTexture: THREE.CanvasTexture | undefined

function getIconTexture(risk: RiskType) {
  const existing = iconTextures.get(risk)
  if (existing) return existing
  const color = `#${riskColors[risk].toString(16).padStart(6, '0')}`
  const canvas = document.createElement('canvas')
  canvas.width = 128
  canvas.height = 128
  const ctx = canvas.getContext('2d')!
  ctx.shadowColor = color
  ctx.shadowBlur = 9
  ctx.fillStyle = color
  roundRect(ctx, 31, 31, 66, 66, 16)
  ctx.fill()
  ctx.shadowBlur = 0
  ctx.strokeStyle = 'rgba(230,246,255,.66)'
  ctx.lineWidth = 3
  ctx.stroke()
  ctx.fillStyle = 'rgba(240,250,255,.9)'
  ctx.fillRect(49, 44, 9, 32)
  ctx.fillRect(63, 36, 9, 40)
  ctx.fillRect(77, 52, 9, 24)
  ctx.fillRect(45, 82, 46, 7)
  const texture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(texture)
  iconTextures.set(risk, texture)
  return texture
}

function getMarkerTexture(risk: RiskType) {
  const existing = markerTextures.get(risk)
  if (existing) return existing

  const color = `#${riskColors[risk].toString(16).padStart(6, '0')}`
  const canvas = document.createElement('canvas')
  canvas.width = 128
  canvas.height = 160
  const ctx = canvas.getContext('2d')!

  ctx.shadowColor = color
  ctx.shadowBlur = 10
  ctx.beginPath()
  ctx.moveTo(64, 150)
  ctx.bezierCurveTo(53, 133, 23, 105, 23, 80)
  ctx.bezierCurveTo(23, 53, 42, 34, 64, 34)
  ctx.bezierCurveTo(86, 34, 105, 53, 105, 80)
  ctx.bezierCurveTo(105, 105, 75, 133, 64, 150)
  ctx.closePath()
  ctx.fillStyle = color
  ctx.fill()

  ctx.shadowBlur = 0
  ctx.strokeStyle = 'rgba(214,255,251,.7)'
  ctx.lineWidth = 4
  ctx.stroke()

  ctx.beginPath()
  ctx.arc(64, 80, 21, 0, Math.PI * 2)
  ctx.fillStyle = 'rgba(3,38,64,.72)'
  ctx.fill()
  ctx.strokeStyle = 'rgba(228,255,255,.86)'
  ctx.lineWidth = 5
  ctx.stroke()

  const texture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(texture)
  markerTextures.set(risk, texture)
  return texture
}

function getGovernmentTexture() {
  if (governmentTexture) return governmentTexture

  const canvas = document.createElement('canvas')
  canvas.width = 160
  canvas.height = 160
  const ctx = canvas.getContext('2d')!
  ctx.shadowColor = '#52eaff'
  ctx.shadowBlur = 12
  ctx.fillStyle = '#1bd8ff'
  roundRect(ctx, 36, 34, 88, 88, 18)
  ctx.fill()
  ctx.shadowBlur = 0
  ctx.strokeStyle = 'rgba(236,251,255,.78)'
  ctx.lineWidth = 4
  ctx.stroke()

  ctx.fillStyle = 'rgba(244,252,255,.95)'
  ctx.beginPath()
  ctx.moveTo(80, 48)
  ctx.lineTo(112, 70)
  ctx.lineTo(48, 70)
  ctx.closePath()
  ctx.fill()
  ctx.fillRect(54, 76, 52, 8)
  ctx.fillRect(58, 88, 8, 22)
  ctx.fillRect(76, 88, 8, 22)
  ctx.fillRect(94, 88, 8, 22)
  ctx.fillRect(50, 114, 60, 8)

  governmentTexture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(governmentTexture)
  return governmentTexture
}

function getFlowCometTexture() {
  if (flowCometTexture) return flowCometTexture

  const canvas = document.createElement('canvas')
  canvas.width = 128
  canvas.height = 128
  const ctx = canvas.getContext('2d')!
  const gradient = ctx.createRadialGradient(64, 64, 0, 64, 64, 64)
  gradient.addColorStop(0, 'rgba(255,255,255,1)')
  gradient.addColorStop(0.18, 'rgba(255,230,142,.96)')
  gradient.addColorStop(0.48, 'rgba(255,158,48,.42)')
  gradient.addColorStop(1, 'rgba(255,120,24,0)')
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, 128, 128)

  flowCometTexture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(flowCometTexture)
  return flowCometTexture
}

function createGlowSprite(color: string, size: number, opacity: number) {
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')!
  const gradient = ctx.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2)
  gradient.addColorStop(0, color)
  gradient.addColorStop(0.42, `${color}88`)
  gradient.addColorStop(1, `${color}00`)
  ctx.fillStyle = gradient
  ctx.fillRect(0, 0, size, size)
  const texture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(texture)
  return new THREE.Sprite(new THREE.SpriteMaterial({ map: texture, transparent: true, opacity, depthWrite: false, blending: THREE.AdditiveBlending }))
}

function createTextSprite(text: string, fontSize: number, fill: string, glow: string) {
  const canvas = document.createElement('canvas')
  canvas.width = 260
  canvas.height = 80
  const ctx = canvas.getContext('2d')!
  ctx.font = `700 ${fontSize}px "Microsoft YaHei", "PingFang SC", sans-serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.shadowColor = glow
  ctx.shadowBlur = 10
  ctx.fillStyle = fill
  ctx.fillText(text, 130, 40)
  const texture = new THREE.CanvasTexture(canvas)
  disposableTextures.push(texture)
  return new THREE.Sprite(new THREE.SpriteMaterial({ map: texture, transparent: true, depthWrite: false }))
}

function pointInsideCity(x: number, y: number) {
  return districtModels.some(district =>
    district.polygons.some(polygon => pointInRing([x, y], polygon[0])),
  )
}

function getPointFromObject(object: THREE.Object3D) {
  let current: THREE.Object3D | null = object
  while (current) {
    const id = current.userData.pointId as number | undefined
    if (id) return institutionMap.get(id) || null
    current = current.parent
  }
  return null
}

function flyTo(toPosition: THREE.Vector3, toTarget: THREE.Vector3, duration: number) {
  if (!camera || !controls) return
  cameraTween = {
    startTime: performance.now(),
    duration,
    fromPosition: camera.position.clone(),
    toPosition,
    fromTarget: controls.target.clone(),
    toTarget,
  }
}

function focusDistrict(name: string) {
  const district = districtModels.find(item => item.name === name)
  if (!district) return
  flyTo(
    new THREE.Vector3(district.center.x, district.center.y - 330, 470),
    new THREE.Vector3(district.center.x, district.center.y, 4),
    720,
  )
}

function resizeRenderer() {
  if (!containerRef.value || !renderer || !composer || !camera) return
  const rect = containerRef.value.getBoundingClientRect()
  camera.aspect = rect.width / rect.height
  camera.updateProjectionMatrix()
  renderer.setSize(rect.width, rect.height)
  composer.setSize(rect.width, rect.height)
}

function disposeObject(object: THREE.Object3D) {
  object.traverse((child) => {
    const mesh = child as THREE.Mesh
    mesh.geometry?.dispose?.()
    const material = mesh.material as THREE.Material | THREE.Material[] | undefined
    if (Array.isArray(material)) material.forEach(item => item.dispose())
    else material?.dispose?.()
  })
}

function normalizePolygons(feature: GeoFeature): Polygon[] {
  return feature.geometry.type === 'Polygon'
    ? [feature.geometry.coordinates as Polygon]
    : feature.geometry.coordinates as Polygon[]
}

function forEachCoordinate(feature: GeoFeature, callback: (coord: LonLat) => void) {
  normalizePolygons(feature).forEach(polygon => polygon.forEach(ring => ring.forEach(callback)))
}

function polygonCenter(ring: Ring): LonLat {
  const total = ring.reduce((acc, coord) => [acc[0] + coord[0], acc[1] + coord[1]] as LonLat, [0, 0])
  return [total[0] / ring.length, total[1] / ring.length]
}

function pointInRing(point: [number, number], ring: Array<[number, number]>) {
  const [x, y] = point
  let inside = false
  for (let i = 0, j = ring.length - 1; i < ring.length; j = i++) {
    const xi = ring[i][0]
    const yi = ring[i][1]
    const xj = ring[j][0]
    const yj = ring[j][1]
    const intersect = ((yi > y) !== (yj > y)) && (x < ((xj - xi) * (y - yi)) / (yj - yi) + xi)
    if (intersect) inside = !inside
  }
  return inside
}

function distanceToRing(point: [number, number], ring: Array<[number, number]>) {
  let minDistance = Number.POSITIVE_INFINITY
  for (let index = 0; index < ring.length; index += 1) {
    const start = ring[index]
    const end = ring[(index + 1) % ring.length]
    minDistance = Math.min(minDistance, distanceToSegment(point, start, end))
  }
  return minDistance
}

function distanceToSegment(point: [number, number], start: [number, number], end: [number, number]) {
  const dx = end[0] - start[0]
  const dy = end[1] - start[1]
  if (dx === 0 && dy === 0) return Math.hypot(point[0] - start[0], point[1] - start[1])
  const t = Math.max(0, Math.min(1, ((point[0] - start[0]) * dx + (point[1] - start[1]) * dy) / (dx * dx + dy * dy)))
  return Math.hypot(point[0] - (start[0] + dx * t), point[1] - (start[1] + dy * t))
}

function pointDistance(a: [number, number], b: [number, number]) {
  return Math.hypot(a[0] - b[0], a[1] - b[1])
}

function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
  ctx.beginPath()
  ctx.moveTo(x + radius, y)
  ctx.arcTo(x + width, y, x + width, y + height, radius)
  ctx.arcTo(x + width, y + height, x, y + height, radius)
  ctx.arcTo(x, y + height, x, y, radius)
  ctx.arcTo(x, y, x + width, y, radius)
  ctx.closePath()
}

function easeOutCubic(value: number) {
  return 1 - Math.pow(1 - value, 3)
}

function easeInOutCubic(value: number) {
  return value < 0.5 ? 4 * value * value * value : 1 - Math.pow(-2 * value + 2, 3) / 2
}

watch([riskFilter, activeDistrict], () => {
  selectedPoint.value = null
  hoveredPoint.value = null
  autoPopupPoint.value = null
  manualHoldUntil = 0
  refreshVisibleObjects()
  requestAnimationFrame(resetCarouselPoint)
})

watch(activeDistrict, (name) => {
  if (name) focusDistrict(name)
  else flyTo(defaultCameraPosition, defaultCameraTarget, 720)
})

onMounted(async () => {
  await nextTick()
  initThree()
  startPopupCarousel()
})

onBeforeUnmount(() => {
  stopPopupCarousel()
  cancelAnimationFrame(animationId)
  resizeObserver?.disconnect()
  controls?.dispose()
  if (scene) disposeObject(scene)
  disposableTextures.forEach(texture => texture.dispose())
  composer?.dispose()
  renderer?.dispose()
  renderer?.domElement.remove()
})
</script>

<style scoped>
.district-map-module {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 620px;
}

.map-frame {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border: 1px solid rgba(55, 160, 255, 0.58);
  border-radius: 8px;
  background:
    radial-gradient(circle at 50% 42%, rgba(20, 132, 255, 0.22), transparent 40%),
    linear-gradient(135deg, rgba(8, 45, 95, 0.92), rgba(4, 18, 43, 0.96));
  box-shadow:
    inset 0 0 46px rgba(37, 136, 255, 0.22),
    0 0 28px rgba(11, 92, 205, 0.2);
}

.map-frame::before,
.map-frame::after {
  position: absolute;
  z-index: 4;
  content: "";
  pointer-events: none;
}

.map-frame::before {
  inset: 0;
  border: 1px solid rgba(112, 205, 255, 0.14);
  clip-path: polygon(0 0, 22% 0, 24% 11px, 76% 11px, 78% 0, 100% 0, 100% 100%, 78% 100%, 76% calc(100% - 11px), 24% calc(100% - 11px), 22% 100%, 0 100%);
}

.map-frame::after {
  inset: 58px 12px 92px;
  background:
    linear-gradient(rgba(75, 178, 255, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(75, 178, 255, 0.045) 1px, transparent 1px);
  background-size: 34px 34px;
  opacity: 0.74;
}

.map-toolbar {
  position: relative;
  z-index: 8;
  height: 58px;
  display: grid;
  grid-template-columns: minmax(260px, 370px) 1fr auto;
  align-items: center;
  gap: 18px;
  padding: 0 18px 0 24px;
  border-bottom: 1px solid rgba(65, 169, 255, 0.32);
  background:
    linear-gradient(90deg, rgba(13, 78, 158, 0.64), transparent 36%, transparent 64%, rgba(13, 78, 158, 0.34)),
    rgba(3, 19, 45, 0.66);
}

.map-title {
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: 14px;
}

.map-title span {
  color: #eef9ff;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 1px;
  text-shadow: 0 0 12px rgba(82, 199, 255, 0.72);
}

.map-title b {
  color: #63e8ff;
  font-size: 13px;
  font-weight: 600;
}

.risk-legend {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.legend-item,
.map-actions button,
.map-actions select,
.institution-card button {
  height: 30px;
  color: #cfeeff;
  border: 1px solid rgba(76, 170, 255, 0.38);
  border-radius: 4px;
  background: rgba(7, 35, 78, 0.72);
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 0 10px;
  cursor: pointer;
}

.legend-item i,
.risk-panel i {
  width: 9px;
  height: 9px;
  border-radius: 2px;
  box-shadow: 0 0 8px currentColor;
}

.legend-item.all i { background: #58dfff; color: #58dfff; }
.legend-item.normal i,
.risk-panel .normal i { background: #45e58c; color: #45e58c; }
.legend-item.focus i,
.risk-panel .focus i { background: #ffc94d; color: #ffc94d; }
.legend-item.warn i,
.risk-panel .warn i { background: #ff8a3d; color: #ff8a3d; }
.legend-item.danger i,
.risk-panel .danger i { background: #ff535a; color: #ff535a; }

.legend-item.active {
  color: #ffffff;
  border-color: rgba(86, 212, 255, 0.74);
  background: linear-gradient(180deg, rgba(23, 118, 216, 0.58), rgba(7, 42, 91, 0.74));
  box-shadow: inset 0 0 14px rgba(44, 171, 255, 0.18);
}

.map-actions {
  display: flex;
  align-items: center;
  gap: 9px;
}

.map-actions select {
  width: 132px;
  padding: 0 11px;
  outline: none;
}

.map-actions button {
  min-width: 58px;
  padding: 0 12px;
  cursor: pointer;
}

.three-map {
  position: absolute;
  z-index: 2;
  inset: 58px 12px 92px;
  overflow: hidden;
  cursor: grab;
}

.three-map.is-hovering {
  cursor: pointer;
}

.three-map:deep(canvas) {
  display: block;
  width: 100%;
  height: 100%;
}

.popup-connector {
  position: absolute;
  z-index: 8;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.connector-glow,
.connector-core {
  fill: none;
  stroke-linecap: round;
}

.connector-glow {
  stroke-width: 7;
  opacity: 0.32;
  filter: url(#connectorGlow);
}

.connector-core {
  stroke-width: 1.6;
  stroke-dasharray: 9 7;
  animation: connectorDash 1.4s linear infinite;
}

.connector-dot {
  filter: url(#connectorGlow);
}

.connector-glow.normal,
.connector-core.normal,
.connector-dot.normal { stroke: #45e58c; fill: #45e58c; }
.connector-glow.focus,
.connector-core.focus,
.connector-dot.focus { stroke: #ffc94d; fill: #ffc94d; }
.connector-glow.warn,
.connector-core.warn,
.connector-dot.warn { stroke: #ff8a3d; fill: #ff8a3d; }
.connector-glow.danger,
.connector-core.danger,
.connector-dot.danger { stroke: #ff535a; fill: #ff535a; }

@keyframes connectorDash {
  to { stroke-dashoffset: -16; }
}

.risk-panel,
.institution-card,
.map-feed {
  position: absolute;
  z-index: 9;
  border: 1px solid rgba(76, 170, 255, 0.4);
  background: linear-gradient(180deg, rgba(7, 35, 75, 0.88), rgba(4, 22, 53, 0.92));
  box-shadow: inset 0 0 18px rgba(37, 134, 255, 0.16), 0 0 16px rgba(5, 77, 164, 0.14);
  backdrop-filter: blur(8px);
}

.risk-panel {
  right: 24px;
  bottom: 116px;
  width: 172px;
  padding: 13px 14px 12px;
  border-radius: 6px;
}

.risk-panel p {
  margin: 0 0 8px;
  color: #ffffff;
  font-size: 15px;
  font-weight: 700;
}

.risk-panel span,
.risk-panel strong {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  color: #c9e8ff;
  font-size: 13px;
  line-height: 1.9;
}

.risk-panel span i {
  flex: none;
  margin-right: 2px;
}

.risk-panel span b {
  margin-left: auto;
  color: #f2fbff;
}

.risk-panel strong {
  margin-top: 6px;
  padding-top: 6px;
  border-top: 1px solid rgba(78, 170, 255, 0.2);
  color: #67eaff;
}

.institution-card {
  width: 224px;
  padding: 14px 14px 13px;
  border-radius: 7px;
  transform: translate(84px, -52%);
}

.institution-card.focus {
  border-color: rgba(255, 201, 77, 0.78);
  box-shadow: inset 0 0 16px rgba(255, 201, 77, 0.12), 0 0 18px rgba(255, 180, 40, 0.18);
}

.institution-card.warn {
  border-color: rgba(255, 138, 61, 0.76);
}

.institution-card.danger {
  border-color: rgba(255, 83, 90, 0.8);
  box-shadow: inset 0 0 16px rgba(255, 83, 90, 0.12), 0 0 18px rgba(255, 83, 90, 0.2);
}

.institution-card strong {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  color: #ffffff;
  font-size: 16px;
}

.institution-card em {
  flex: none;
  padding: 1px 7px;
  color: #ffe07b;
  font-size: 12px;
  font-style: normal;
  border-radius: 3px;
  background: rgba(255, 201, 67, 0.14);
}

.institution-card p {
  display: flex;
  justify-content: space-between;
  margin: 8px 0 0;
  color: #bfe1ff;
  font-size: 13px;
}

.institution-card b {
  color: #5ce9ff;
}

.institution-card .risk-text {
  display: block;
  color: #ff8379;
}

.institution-card button {
  width: 100%;
  margin-top: 11px;
  cursor: pointer;
}

.map-feed {
  left: 12px;
  right: 12px;
  bottom: 12px;
  height: 72px;
  display: grid;
  grid-template-columns: 132px 220px 1fr;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 7px;
}

.feed-heading {
  display: grid;
  align-content: center;
  border-right: 1px solid rgba(84, 172, 255, 0.24);
}

.feed-heading span {
  color: #ffffff;
  font-size: 17px;
  font-weight: 800;
}

.feed-heading b {
  margin-top: 4px;
  color: #58e8ff;
  font-size: 13px;
}

.feed-tabs {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  align-items: center;
  gap: 8px;
  color: #9fd4ff;
}

.feed-tabs span {
  text-align: center;
  border-right: 1px solid rgba(77, 168, 255, 0.22);
}

.feed-list {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}

.feed-list article {
  min-width: 0;
  display: grid;
  grid-template-columns: 28px 1fr;
  grid-template-rows: 18px 18px 16px;
  column-gap: 8px;
  align-items: center;
  padding: 2px 8px;
  border-left: 1px solid rgba(77, 168, 255, 0.22);
  color: #cfeaff;
}

.feed-list i {
  grid-row: 1 / 4;
  width: 28px;
  height: 28px;
  display: grid;
  place-items: center;
  color: white;
  font-style: normal;
  border-radius: 5px;
  background: #247bd6;
  box-shadow: 0 0 10px rgba(45, 151, 255, 0.25);
}

.feed-list article.focus i { background: #a9781c; }
.feed-list article.warn i { background: #b96021; }
.feed-list article.danger i { background: #b84249; }
.feed-list article.normal i { background: #1c8f5d; }

.feed-list strong,
.feed-list span,
.feed-list em {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.feed-list strong {
  color: #f2fbff;
  font-size: 13px;
}

.feed-list span,
.feed-list em {
  color: rgba(195, 225, 255, 0.76);
  font-size: 12px;
  font-style: normal;
}

.tooltip-enter-active,
.tooltip-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.tooltip-enter-from,
.tooltip-leave-to {
  opacity: 0;
  transform: translate(84px, -46%) scale(0.96);
}

@media (max-width: 1280px) {
  .map-toolbar {
    grid-template-columns: 300px 1fr auto;
  }

  .legend-item {
    padding: 0 8px;
  }

  .legend-item span {
    font-size: 12px;
  }
}
</style>
