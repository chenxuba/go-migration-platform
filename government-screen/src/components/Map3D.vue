<template>
  <div ref="containerRef" class="three-map" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import * as THREE from 'three'

type PointType = 'normal' | 'focus' | 'warn' | 'danger'

interface MapPoint {
  x: number
  y: number
  type: PointType
}

const containerRef = ref<HTMLDivElement | null>(null)
let renderer: THREE.WebGLRenderer | undefined
let scene: THREE.Scene | undefined
let camera: THREE.OrthographicCamera | undefined
let animationId = 0
let resizeObserver: ResizeObserver | undefined
const animatedGroups: THREE.Object3D[] = []

const areaShapes = [
  [[-360, 54], [-250, 142], [-92, 126], [-48, 28], [-206, -44], [-316, -12]],
  [[-92, 126], [68, 176], [224, 132], [178, 28], [-48, 28]],
  [[224, 132], [360, 34], [302, -92], [122, -106], [178, 28]],
  [[-316, -12], [-206, -44], [-112, -158], [-260, -208], [-392, -118]],
  [[-206, -44], [-48, 28], [178, 28], [122, -106], [-70, -84], [-112, -158]],
  [[-70, -84], [122, -106], [302, -92], [218, -210], [14, -178]],
]

const labels = [
  { text: '滨江区', x: -215, y: 48 },
  { text: '城北区', x: 90, y: 78 },
  { text: '城中区', x: -58, y: -46 },
  { text: '江南区', x: 182, y: -124 },
]

const points: MapPoint[] = [
  { x: -282, y: 48, type: 'danger' }, { x: -186, y: -18, type: 'normal' },
  { x: -98, y: 68, type: 'focus' }, { x: 18, y: 98, type: 'focus' },
  { x: 96, y: 46, type: 'normal' }, { x: 244, y: 56, type: 'normal' },
  { x: 288, y: -38, type: 'danger' }, { x: 166, y: -128, type: 'warn' },
  { x: -2, y: -116, type: 'normal' }, { x: -142, y: -106, type: 'danger' },
  { x: -244, y: -86, type: 'normal' }, { x: 250, y: -174, type: 'danger' },
  { x: 332, y: -136, type: 'normal' }, { x: -54, y: -4, type: 'focus' },
]

const colors: Record<PointType, number> = {
  normal: 0x43e28c,
  focus: 0xffc847,
  warn: 0xff8a35,
  danger: 0xff5757,
}

function createShape(points: number[][]) {
  const shape = new THREE.Shape()
  points.forEach(([x, y], index) => {
    if (index === 0) shape.moveTo(x, y)
    else shape.lineTo(x, y)
  })
  shape.closePath()
  return shape
}

function createTextSprite(text: string) {
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 64
  const ctx = canvas.getContext('2d')!
  ctx.font = 'bold 28px Microsoft YaHei, PingFang SC, sans-serif'
  ctx.fillStyle = 'rgba(210,238,255,.86)'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.shadowColor = '#1cb8ff'
  ctx.shadowBlur = 10
  ctx.fillText(text, 128, 32)
  const texture = new THREE.CanvasTexture(canvas)
  const material = new THREE.SpriteMaterial({ map: texture, transparent: true, depthWrite: false })
  const sprite = new THREE.Sprite(material)
  sprite.scale.set(78, 20, 1)
  return sprite
}

function createPoint(point: MapPoint, index: number) {
  const group = new THREE.Group()
  const color = colors[point.type]
  const base = new THREE.Mesh(
    new THREE.CylinderGeometry(10, 13, 18, 32),
    new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.92 }),
  )
  base.position.z = 26
  const core = new THREE.Mesh(
    new THREE.SphereGeometry(7, 24, 16),
    new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.88 }),
  )
  core.position.z = 42
  const ring = new THREE.Mesh(
    new THREE.RingGeometry(13, 17, 42),
    new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.42, side: THREE.DoubleSide }),
  )
  ring.position.z = 23
  group.add(base, core, ring)
  group.position.set(point.x, point.y, 0)
  group.userData.offset = index * 0.36
  animatedGroups.push(group)
  return group
}

function createFlowLine(points: THREE.Vector3[], color: number) {
  const curve = new THREE.CatmullRomCurve3(points)
  const geometry = new THREE.TubeGeometry(curve, 72, 1.4, 8, false)
  const material = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.62 })
  return new THREE.Mesh(geometry, material)
}

function init() {
  if (!containerRef.value) return
  scene = new THREE.Scene()
  scene.fog = new THREE.Fog(0x031a38, 720, 1180)

  const rect = containerRef.value.getBoundingClientRect()
  const aspect = rect.width / rect.height
  const frustum = 610
  camera = new THREE.OrthographicCamera(-frustum * aspect, frustum * aspect, frustum, -frustum, 1, 1800)
  camera.position.set(0, -450, 620)
  camera.lookAt(0, 0, 0)

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  renderer.setSize(rect.width, rect.height)
  containerRef.value.appendChild(renderer.domElement)

  const root = new THREE.Group()
  root.rotation.x = -0.72
  root.rotation.z = -0.03
  scene.add(root)

  const grid = new THREE.GridHelper(980, 28, 0x1f8cff, 0x0d376c)
  grid.rotation.x = Math.PI / 2
  grid.position.z = -8
  ;(grid.material as THREE.Material).transparent = true
  ;(grid.material as THREE.Material).opacity = 0.26
  root.add(grid)

  areaShapes.forEach((area, index) => {
    const geometry = new THREE.ExtrudeGeometry(createShape(area), { depth: 18 + index * 1.2, bevelEnabled: true, bevelThickness: 2, bevelSize: 2, bevelSegments: 1 })
    const material = new THREE.MeshBasicMaterial({ color: index % 2 ? 0x0d65c8 : 0x0a4fa4, transparent: true, opacity: 0.46 })
    const mesh = new THREE.Mesh(geometry, material)
    mesh.position.z = 0
    root.add(mesh)

    const linePoints = area.map(([x, y]) => new THREE.Vector3(x, y, 24 + index * 1.2))
    linePoints.push(linePoints[0].clone())
    const border = new THREE.Line(
      new THREE.BufferGeometry().setFromPoints(linePoints),
      new THREE.LineBasicMaterial({ color: 0x35cfff, transparent: true, opacity: 0.88 }),
    )
    root.add(border)
  })

  const flowA = createFlowLine([new THREE.Vector3(-318, -52, 38), new THREE.Vector3(-108, 42, 42), new THREE.Vector3(58, -18, 42), new THREE.Vector3(298, 34, 42)], 0x38eaff)
  const flowB = createFlowLine([new THREE.Vector3(-240, 112, 36), new THREE.Vector3(-140, -38, 40), new THREE.Vector3(112, 42, 42), new THREE.Vector3(238, 128, 36)], 0xffc547)
  const flowC = createFlowLine([new THREE.Vector3(-320, -154, 36), new THREE.Vector3(-126, -78, 44), new THREE.Vector3(94, -92, 42), new THREE.Vector3(342, -42, 36)], 0x38eaff)
  root.add(flowA, flowB, flowC)

  points.forEach((point, index) => root.add(createPoint(point, index)))

  labels.forEach((item) => {
    const sprite = createTextSprite(item.text)
    sprite.position.set(item.x, item.y, 58)
    root.add(sprite)
  })

  const halo = new THREE.Mesh(
    new THREE.RingGeometry(24, 46, 72),
    new THREE.MeshBasicMaterial({ color: 0x49e9ff, transparent: true, opacity: 0.38, side: THREE.DoubleSide }),
  )
  halo.position.set(24, -8, 44)
  animatedGroups.push(halo)
  root.add(halo)

  const resize = () => {
    if (!containerRef.value || !renderer || !camera) return
    const nextRect = containerRef.value.getBoundingClientRect()
    const nextAspect = nextRect.width / nextRect.height
    camera.left = -frustum * nextAspect
    camera.right = frustum * nextAspect
    camera.top = frustum
    camera.bottom = -frustum
    camera.updateProjectionMatrix()
    renderer.setSize(nextRect.width, nextRect.height)
  }
  resizeObserver = new ResizeObserver(resize)
  resizeObserver.observe(containerRef.value)

  const tick = () => {
    const time = performance.now() * 0.001
    animatedGroups.forEach((group) => {
      group.rotation.z += 0.01
      if ('offset' in group.userData) {
        group.position.z = Math.sin(time * 2.4 + group.userData.offset) * 4
      }
    })
    root.rotation.z = -0.03 + Math.sin(time * 0.28) * 0.01
    renderer?.render(scene!, camera!)
    animationId = requestAnimationFrame(tick)
  }
  tick()
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

onMounted(init)
onBeforeUnmount(() => {
  cancelAnimationFrame(animationId)
  resizeObserver?.disconnect()
  if (scene) disposeObject(scene)
  renderer?.dispose()
  renderer?.domElement.remove()
})
</script>
