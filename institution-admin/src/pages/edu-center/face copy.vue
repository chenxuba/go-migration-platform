<script setup>
import { CheckCircleFilled, ExclamationCircleFilled } from '@ant-design/icons-vue'
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import * as faceapi from 'face-api.js'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

const route = useRoute()
const data = ref(null)
const student = ref(undefined)
const studentList = ref([
  { id: 1, name: '张晨', phone: '17601241636', status: 0 },
  { id: 2, name: '李元芳', phone: '18899238823', status: 0 },
  { id: 3, name: '潘金莲', phone: '15782827732', status: 0 },
  { id: 4, name: '武松', phone: '17866232253', status: 0 },
])
// 添加一个ref来跟踪每个学生的最后考勤时间
const lastAttendanceTimes = ref({})
// 添加一个ref来控制显示哪个提示
const showCooldownMessage = ref(false)
// formatDate 格式化时间 07-11 12:23
function formatDate(timestamp) {
  return dayjs(timestamp).format('MM-DD HH:mm')
}

// 考勤记录
const attendanceRecords = ref([])

// Face detection related refs
const videoStream = ref(null)
const videoRef = ref(null)
const canvasRef = ref(null)
const isFaceDetected = ref(false)
const faceDescriptor = ref(null)
const isLoading = ref(false)
const isModelLoaded = ref(false)
const capturedImageUrl = ref('')
const capturedTime = ref('')
const showVideoEndStream = ref(false)
// 人脸识别状态
const recognizingFace = ref(false)
const videoReady = ref(false)
// 添加一个新的状态来控制是否开始考勤
const isAttendanceStarted = ref(false)

// Mock faceapi if not available for testing purposes
if (typeof faceapi === 'undefined' || !faceapi) {
  console.warn('face-api.js not found, using mock implementation')
  // Simple mock implementation
  const mockFaceapi = {
    nets: {
      ssdMobilenetv1: { loadFromUri: async () => console.log('Mock SSD loaded') },
      faceLandmark68Net: { loadFromUri: async () => console.log('Mock landmarks loaded') },
      faceRecognitionNet: { loadFromUri: async () => console.log('Mock recognition loaded') },
    },
    detectAllFaces: () => ({
      withFaceLandmarks: () => ({
        withFaceDescriptors: () => {
          // Randomly decide if a face is detected or not
          const faceDetected = Math.random() > 0.3
          if (faceDetected) {
            return [{
              detection: {
                box: {
                  x: 150,
                  y: 100,
                  width: 200,
                  height: 200,
                },
              },
              landmarks: { positions: [] },
              descriptor: new Float32Array(128).fill(0.5),
            }]
          }
          return []
        },
      }),
    }),
    matchDimensions: () => { },
    resizeResults: detections => detections,
    draw: {
      drawDetections: () => { },
      drawFaceLandmarks: () => { },
    },
  }

  // Replace with mock if not available
  window.faceapi = mockFaceapi
}

function filterOption(input, option) {
  const name = option.label?.toString() || ''
  const value = option.data?.value?.toString() || ''
  const phone = option.data?.phone?.toString() || '' // 电话号码
  const course = option.data?.course?.toString() || ''

  const normalizedInput = input.toLowerCase().trim()
  const searchContent = [
    name.toLowerCase(),
    value.toLowerCase(),
    phone.toLowerCase(),
    course.toLowerCase(),
  ].join(' ')

  return searchContent.includes(normalizedInput)
}

// Load face-api.js models
async function loadModels() {
  try {
    isLoading.value = true
    message.loading('正在加载人脸识别模型...', 0)

    // Set the models path
    const MODEL_URL = '/models'

    // Load models required for face detection and recognition
    await faceapi.nets.ssdMobilenetv1.loadFromUri(MODEL_URL)
    await faceapi.nets.faceLandmark68Net.loadFromUri(MODEL_URL)
    await faceapi.nets.faceRecognitionNet.loadFromUri(MODEL_URL)

    isModelLoaded.value = true
    message.destroy()
    // message.success('模型加载成功');
    startVideo()
  }
  catch (error) {
    console.error('加载模型失败:', error)
    message.error('加载模型失败，请刷新重试')
  }
  finally {
    isLoading.value = false
  }
}

// Start video stream
async function startVideo() {
  if (!isModelLoaded.value) {
    message.warning('请等待模型加载完成')
    return
  }

  try {
    videoReady.value = false
    showVideoEndStream.value = false

    // Stop existing stream if any
    if (videoStream.value) {
      stopVideo()
    }

    // Get access to webcam
    const stream = await navigator.mediaDevices.getUserMedia({
      video: {
        width: 600,
        height: 450,
      },
    })

    videoStream.value = stream

    // Set video source
    if (videoRef.value) {
      videoRef.value.srcObject = stream

      // 添加视频加载完成事件监听
      videoRef.value.onloadedmetadata = () => {
        // 确保视频已经开始播放
        videoRef.value.play().then(() => {
          // 添加淡入效果
          videoRef.value.classList.add('loaded')
          if (canvasRef.value) {
            canvasRef.value.classList.add('loaded')
          }
          // 延迟设置 videoReady，确保过渡效果完成
          setTimeout(() => {
            videoReady.value = true
          }, 300)
        })
      }
    }

    // Start face detection
    startFaceDetection()
  }
  catch (error) {
    console.error('获取摄像头失败:', error)
    message.error('无法访问摄像头，请检查摄像头是否正常工作')
    showVideoEndStream.value = true
    videoReady.value = false
  }
}
// 结束考勤
function endAttendance() {
  // 重置考勤状态
  isAttendanceStarted.value = false

  // 重置人脸识别状态
  recognizingFace.value = false

  // 重置人脸检测状态
  isFaceDetected.value = false
  faceDescriptor.value = null

  // 显示提示信息
  message.success('考勤已结束')
  speakMessage('考勤已结束')

  // 停止视频流
  stopVideo()
  // 清空本地考勤记录
  attendanceRecords.value = []
  // 重新启动视频流
  setTimeout(() => {
    startVideo()
  }, 100)
}
// Stop video stream
function stopVideo() {
  videoReady.value = false
  if (videoStream.value) {
    const tracks = videoStream.value.getTracks()
    tracks.forEach((track) => {
      track.stop()
      track.enabled = false
    })
    videoStream.value = null
  }

  if (videoRef.value) {
    videoRef.value.srcObject = null
    videoRef.value.classList.remove('loaded')
    if (canvasRef.value) {
      canvasRef.value.classList.remove('loaded')
      // 清理 canvas 上下文
      const ctx = canvasRef.value.getContext('2d')
      if (ctx) {
        ctx.clearRect(0, 0, canvasRef.value.width, canvasRef.value.height)
      }
    }
  }
}

// Face detection loop
function startFaceDetection() {
  if (!videoRef.value || !canvasRef.value)
    return

  const canvas = canvasRef.value
  const video = videoRef.value
  const displaySize = { width: video.width, height: video.height }
  faceapi.matchDimensions(canvas, displaySize)

  // 清理旧的检测循环
  if (detectionInterval.value) {
    clearInterval(detectionInterval.value)
  }

  let isProcessing = false // 添加处理锁

  // 创建新的检测循环
  detectionInterval.value = setInterval(async () => {
    if (!videoRef.value || !canvasRef.value || isProcessing) {
      return
    }

    try {
      isProcessing = true // 设置处理锁

      // 根据当前模式选择合适的检测级别
      let detections
      if (data.value === 1 || (data.value === 2 && isAttendanceStarted.value)) {
        // 采集模式或正在考勤时，需要完整的人脸特征
        detections = await faceapi
          .detectAllFaces(video)
          .withFaceLandmarks()
          .withFaceDescriptors()
      }
      else {
        // 其他情况只需要基本的人脸检测
        detections = await faceapi.detectAllFaces(video)
      }

      if (detections.length === 1) {
        isFaceDetected.value = true
        // 只在需要时更新人脸特征描述符
        if (data.value === 1 || (data.value === 2 && isAttendanceStarted.value)) {
          faceDescriptor.value = detections[0].descriptor
        }

        // 只在考勤模式且已开始考勤时进行人脸识别
        if (data.value === 2 && !recognizingFace.value && isAttendanceStarted.value && faceDescriptor.value) {
          recognizeFace(faceDescriptor.value)
        }
      }
      else if (detections.length > 1) {
        isFaceDetected.value = false
        // 使用防抖处理警告消息
        if (!window.faceWarningTimeout) {
          window.faceWarningTimeout = setTimeout(() => {
            message.warning('请确保画面中只有一个人脸')
            window.faceWarningTimeout = null
          }, 2000)
        }
      }
      else {
        isFaceDetected.value = false
      }

      // 优化绘制操作
      requestAnimationFrame(() => {
        const ctx = canvas.getContext('2d')
        ctx.clearRect(0, 0, canvas.width, canvas.height)

        if (detections.length > 0) {
          const resizedDetections = faceapi.resizeResults(detections, displaySize)
          resizedDetections.forEach((detection) => {
            const box = detection.detection ? detection.detection.box : detection.box

            // 绘制人脸框
            ctx.strokeStyle = isFaceDetected.value ? '#00cc33' : '#ff3333'
            ctx.lineWidth = 3
            ctx.beginPath()
            ctx.rect(box.x, box.y, box.width, box.height)
            ctx.stroke()

            // 绘制角标
            const cornerSize = 20
            ctx.strokeStyle = '#ffffff'
            ctx.lineWidth = 4

            // 左上角
            ctx.beginPath()
            ctx.moveTo(box.x, box.y + cornerSize)
            ctx.lineTo(box.x, box.y)
            ctx.lineTo(box.x + cornerSize, box.y)
            ctx.stroke()

            // 右上角
            ctx.beginPath()
            ctx.moveTo(box.x + box.width - cornerSize, box.y)
            ctx.lineTo(box.x + box.width, box.y)
            ctx.lineTo(box.x + box.width, box.y + cornerSize)
            ctx.stroke()

            // 右下角
            ctx.beginPath()
            ctx.moveTo(box.x + box.width, box.y + box.height - cornerSize)
            ctx.lineTo(box.x + box.width, box.y + box.height)
            ctx.lineTo(box.x + box.width - cornerSize, box.y + box.height)
            ctx.stroke()

            // 左下角
            ctx.beginPath()
            ctx.moveTo(box.x + cornerSize, box.y + box.height)
            ctx.lineTo(box.x, box.y + box.height)
            ctx.lineTo(box.x, box.y + box.height - cornerSize)
            ctx.stroke()
          })
        }
      })
    }
    catch (error) {
      console.error('Face detection error:', error)
      clearInterval(detectionInterval.value)
    }
    finally {
      isProcessing = false // 释放处理锁
    }
  }, 100) // 提高检测频率到100ms
}

// Store detection interval for cleanup
const detectionInterval = ref(null)

// 人脸识别匹配
async function recognizeFace(currentFaceDescriptor) {
  if (recognizingFace.value)
    return

  recognizingFace.value = true

  try {
    const storedFaces = loadFaceData()

    if (storedFaces.length === 0) {
      showCooldownMessage.value = false // 确保显示"脸部与摄像头平视，识别中"
      message.warning('未找到已采集的人脸数据')
      setTimeout(() => {
        recognizingFace.value = false
      }, 2000)
      return
    }

    let bestMatch = null
    let minDistance = 0.6

    for (const face of storedFaces) {
      const storedDescriptor = new Float32Array(face.faceDescriptor)
      const distance = faceapi.euclideanDistance(currentFaceDescriptor, storedDescriptor)

      if (distance < minDistance) {
        minDistance = distance
        bestMatch = face
      }
    }

    if (bestMatch) {
      const matchedStudent = studentList.value.find(s => s.id === bestMatch.studentId)

      if (matchedStudent) {
        // 检查是否在1分钟内重复考勤
        const lastAttendanceTime = lastAttendanceTimes.value[matchedStudent.id]
        const now = Date.now()

        if (lastAttendanceTime && (now - lastAttendanceTime) < 60000) { // 60000ms = 1分钟
          showCooldownMessage.value = true
        }
        else {
          showCooldownMessage.value = false
          message.success(`人脸考勤成功: ${matchedStudent.name}`)
          speakMessage(`人脸考勤成功: ${matchedStudent.name}`)

          const timestamp = new Date().toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          })

          // 更新最后考勤时间
          lastAttendanceTimes.value[matchedStudent.id] = now
          saveAttendanceRecord(bestMatch.studentId, timestamp)
        }
      }
    }
    else {
      showCooldownMessage.value = false // 确保显示"脸部与摄像头平视，识别中"
      message.warning('未能识别该人脸，请确保已完成人脸采集')
    }
  }
  catch (error) {
    console.error('人脸识别失败:', error)
    showCooldownMessage.value = false // 确保显示"脸部与摄像头平视，识别中"
    message.error('人脸识别失败，请重试')
  }
  finally {
    setTimeout(() => {
      recognizingFace.value = false
    }, 2000)
  }
}

// Capture face
async function captureFace() {
  if (!student.value) {
    message.warning('请先选择一位学员进行人脸采集')
    return
  }

  if (!isFaceDetected.value) {
    message.warning('未检测到人脸，请将脸对准摄像头')
    return
  }

  try {
    // 先清除之前的状态
    capturedImageUrl.value = ''
    capturedTime.value = ''

    // Here you would typically save the face descriptor to your backend
    // For demonstration, we'll just update the local student list
    const studentIndex = studentList.value.findIndex(s => s.id === student.value)
    if (studentIndex !== -1) {
      studentList.value[studentIndex].status = 1 // Mark as collected

      // Take a snapshot of the current face with correct orientation
      const canvas = document.createElement('canvas')
      canvas.width = videoRef.value.videoWidth
      canvas.height = videoRef.value.videoHeight
      const ctx = canvas.getContext('2d')

      // Flip the image horizontally during capture to match what user sees
      // and to ensure the final image has the correct orientation
      ctx.translate(canvas.width, 0)
      ctx.scale(-1, 1)
      ctx.drawImage(videoRef.value, 0, 0, canvas.width, canvas.height)
      ctx.setTransform(1, 0, 0, 1, 0, 0)

      // Convert to data URL (this would be sent to server in a real app)
      const faceImageData = canvas.toDataURL('image/jpeg')
      capturedImageUrl.value = faceImageData // Store the image URL
      capturedTime.value = new Date().toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      }) // Store capture time

      // 保存到本地存储
      if (saveFaceData(student.value, faceDescriptor.value, faceImageData)) {
        message.success('人脸采集成功并已保存到本地')
        speakMessage('人脸采集成功')
      }
      else {
        message.warning('人脸采集成功但保存失败')
      }

      console.log('Face descriptor captured for student:', studentList.value[studentIndex].name)
      console.log('Face data:', faceDescriptor.value)
    }
  }
  catch (error) {
    console.error('人脸采集失败:', error)
    message.error('人脸采集失败，请重试')
  }
}

// 读取本地存储的人脸数据
function loadFaceData() {
  try {
    const faceData = localStorage.getItem('faceData')
    if (faceData) {
      const parsedData = JSON.parse(faceData)

      // 更新学生状态
      studentList.value.forEach((student) => {
        const storedFace = parsedData.find(item => item.studentId === student.id)
        if (storedFace) {
          student.status = 1 // 标记为已采集
        }
      })

      return parsedData
    }
    return []
  }
  catch (error) {
    console.error('读取本地人脸数据失败:', error)
    return []
  }
}

// 保存人脸数据到本地存储
function saveFaceData(studentId, faceDesc, faceImg) {
  try {
    const existingData = loadFaceData()

    // 检查是否已存在该学生的数据
    const existingIndex = existingData.findIndex(item => item.studentId === studentId)

    if (existingIndex !== -1) {
      // 更新现有数据
      existingData[existingIndex] = {
        studentId,
        faceDescriptor: Array.from(faceDesc), // 转换为普通数组以便JSON序列化
        faceImage: faceImg,
        timestamp: Date.now(),
      }
    }
    else {
      // 添加新数据
      existingData.push({
        studentId,
        faceDescriptor: Array.from(faceDesc), // 转换为普通数组以便JSON序列化
        faceImage: faceImg,
        timestamp: Date.now(),
      })
    }

    // 保存到本地存储
    localStorage.setItem('faceData', JSON.stringify(existingData))
    return true
  }
  catch (error) {
    console.error('保存人脸数据失败:', error)
    return false
  }
}

// 删除本地存储的人脸数据
function deleteFaceData(studentId) {
  try {
    const existingData = loadFaceData()
    const filteredData = existingData.filter(item => item.studentId !== studentId)
    localStorage.setItem('faceData', JSON.stringify(filteredData))

    // 更新学生状态
    const studentIndex = studentList.value.findIndex(s => s.id === studentId)
    if (studentIndex !== -1) {
      studentList.value[studentIndex].status = 0
    }

    return true
  }
  catch (error) {
    console.error('删除人脸数据失败:', error)
    return false
  }
}

// 加载本地存储的考勤记录
function loadAttendanceRecords() {
  try {
    const records = localStorage.getItem('attendanceRecords')
    if (records) {
      attendanceRecords.value = JSON.parse(records)
      // 加载记录后滚动到底部
      scrollToBottom()
    }
  }
  catch (error) {
    console.error('读取考勤记录失败:', error)
  }
}
// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    const container = document.querySelector('.scrollbar')
    if (container) {
      container.scrollTop = container.scrollHeight
    }
  })
}
// 保存考勤记录
function saveAttendanceRecord(studentId, timestamp) {
  try {
    const student = studentList.value.find(s => s.id === studentId)
    if (!student)
      return false

    // 捕获当前视频帧作为考勤图像
    const canvas = document.createElement('canvas')
    canvas.width = videoRef.value.videoWidth
    canvas.height = videoRef.value.videoHeight
    const ctx = canvas.getContext('2d')

    // 翻转图像以匹配用户看到的画面
    ctx.translate(canvas.width, 0)
    ctx.scale(-1, 1)
    ctx.drawImage(videoRef.value, 0, 0, canvas.width, canvas.height)
    ctx.setTransform(1, 0, 0, 1, 0, 0)

    // 转换为base64图像数据
    const faceImageData = canvas.toDataURL('image/jpeg')

    attendanceRecords.value.push({
      studentId,
      studentName: student.name,
      timestamp,
      faceImage: faceImageData, // 保存人脸图像
    })

    localStorage.setItem('attendanceRecords', JSON.stringify(attendanceRecords.value))

    // 滚动到底部
    scrollToBottom()
    return true
  }
  catch (error) {
    console.error('保存考勤记录失败:', error)
    return false
  }
}

// 语音提示方法
function speakMessage(message) {
  const speech = new SpeechSynthesisUtterance(message)
  speech.voice = speechSynthesis.getVoices()[0]
  speech.rate = 1.5 // 语速控制: 1.0是正常语速，小于1更慢，大于1更快，范围通常在0.1到10之间
  speech.pitch = 1.0 // 音调控制: 1.0是正常音调，2.0是高音调，0.0是低音调
  speech.volume = 1.0 // 音量控制: 0到1之间的值
  speechSynthesis.speak(speech)
}

// Delete captured face
function deleteFace() {
  if (!student.value)
    return

  try {
    const studentIndex = studentList.value.findIndex(s => s.id === student.value)
    if (studentIndex !== -1) {
      // 更新学生状态为未采集
      studentList.value[studentIndex].status = 0

      // 清除当前显示
      capturedImageUrl.value = ''
      capturedTime.value = ''

      // 从本地存储删除
      if (deleteFaceData(student.value)) {
        // 强制更新 studentList 以触发视图更新
        studentList.value = [...studentList.value]

        message.success('人脸已删除')
        speakMessage('人脸已删除')
      }
      else {
        message.warning('人脸删除失败')
      }
    }
  }
  catch (error) {
    console.error('删除人脸失败:', error)
    message.error('删除人脸失败，请重试')
  }
}

// Retake face photo
function retakeFace() {
  // We'll reuse the capture face functionality but first clear the current image
  capturedImageUrl.value = ''
  capturedTime.value = ''

  // Make sure camera is still running
  if (!videoStream.value) {
    startVideo()
  }

  message.info('请重新采集人脸')
  speakMessage('请重新采集人脸')
}

// Watch for student selection changes
watch(student, (newVal, oldVal) => {
  // 当切换学生时，重置采集状态
  if (newVal !== oldVal) {
    // 先重置状态
    capturedImageUrl.value = ''
    capturedTime.value = ''
    isFaceDetected.value = false
    faceDescriptor.value = null

    // 如果选择了新学生，尝试加载已存在的人脸数据
    if (newVal) {
      const storedFaces = loadFaceData()
      const existingFace = storedFaces.find(face => face.studentId === newVal)

      if (existingFace) {
        // 如果找到已存在的人脸数据，显示图像
        capturedImageUrl.value = existingFace.faceImage
        capturedTime.value = new Date(existingFace.timestamp).toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        })
      }
    }
  }

  if (newVal && isModelLoaded.value && !videoStream.value) {
    // Start camera when student is selected
    startVideo()
  }
})

onMounted(() => {
  // 直接获取参数
  const type = route.query.type
  console.log('参数变化:', type)
  document.title = type == 1 ? '人脸采集' : '人脸考勤'
  data.value = type

  // 加载本地存储的人脸数据
  loadFaceData()

  // 加载考勤记录
  loadAttendanceRecords()

  // 初始化最后考勤时间
  const records = localStorage.getItem('attendanceRecords')
  if (records) {
    const parsedRecords = JSON.parse(records)
    parsedRecords.forEach((record) => {
      lastAttendanceTimes.value[record.studentId] = new Date(record.timestamp).getTime()
    })
  }

  // Load face-api.js models
  loadModels()

  // 更新界面提示
  if (type == 2) {
    document.querySelector('.faceTips').innerText = '请面对摄像头，系统将自动进行人脸识别考勤'
    switchMode(2)
  }
  else {
    switchMode(1)
  }
})

onUnmounted(() => {
  // 清理所有资源
  if (detectionInterval.value) {
    clearInterval(detectionInterval.value)
    detectionInterval.value = null
  }

  // 清理警告定时器
  if (window.faceWarningTimeout) {
    clearTimeout(window.faceWarningTimeout)
    window.faceWarningTimeout = null
  }

  if (videoStream.value) {
    const tracks = videoStream.value.getTracks()
    tracks.forEach((track) => {
      track.stop()
      track.enabled = false
    })
    videoStream.value = null
  }

  if (videoRef.value) {
    videoRef.value.srcObject = null
  }

  if (canvasRef.value) {
    const ctx = canvasRef.value.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, canvasRef.value.width, canvasRef.value.height)
    }
  }

  // 重置所有状态
  videoReady.value = false
  isFaceDetected.value = false
  faceDescriptor.value = null
  capturedImageUrl.value = ''
  capturedTime.value = ''
  student.value = undefined
  recognizingFace.value = false
  isAttendanceStarted.value = false // 重置考勤状态
})

// 监听参数变化（重要！）
watch(
  () => route.query.type,
  (newType) => {
    console.log('参数变化:', newType)
    switchMode(newType)
  },
)

// 切换模式
function switchMode(mode) {
  data.value = Number(mode)
  document.title = data.value == 1 ? '人脸采集' : '人脸考勤'

  // 更新界面提示
  if (data.value == 2) {
    document.querySelector('.faceTips').textContent = '请面对摄像头，系统将自动进行人脸识别考勤'
  }
  else {
    document.querySelector('.faceTips').textContent = '请单人采集，人脸采集成功后，前往"人脸考勤"进行考勤'
  }

  // 重置状态
  recognizingFace.value = false
  capturedImageUrl.value = ''
  capturedTime.value = ''
  student.value = undefined
  isFaceDetected.value = false
  faceDescriptor.value = null
  isAttendanceStarted.value = false // 重置考勤状态
  faceDescriptor.value = null
  isFaceDetected.value = false
  isAttendanceStarted.value = false // 重置考勤状态

  // 清理旧的视频流和检测循环
  if (detectionInterval.value) {
    clearInterval(detectionInterval.value)
    detectionInterval.value = null
  }

  if (videoStream.value) {
    stopVideo()
  }

  // 等待一小段时间确保清理完成后再启动新的视频流
  setTimeout(() => {
    startVideo()
  }, 100)
}

// 添加开始考勤的处理函数
function startAttendance() {
  // if (!isFaceDetected.value) {
  //   message.warning('未检测到人脸，请将脸对准摄像头');
  //   return;
  // }

  // 设置状态以隐藏准备区域
  isAttendanceStarted.value = true
  // 直接开始人脸识别
  if (faceDescriptor.value) {
    speakMessage('开始考勤')
    recognizeFace(faceDescriptor.value)
  }
}
</script>

<template>
  <div class="face">
    <div class="faceInner">
      <div class="topTitle text-#000 text-6 font-800 pt-4" style="text-align: center;">
        智能人脸考勤系统
      </div>
      <div class="change-btn mb-3">
        <a-button :class="data == 1 ? 'active' : ''" class="mr-4 w-34 h-10" @click="switchMode(1)">
          人脸采集
        </a-button>
        <a-button :class="data == 2 ? 'active' : ''" class="w-34 h-10" @click="switchMode(2)">
          人脸考勤
        </a-button>
      </div>
      <div class="faceBody">
        <div class="faceTips">
          <!-- 今日待考勤 <span class="text-#0066ff mx-1">4</span>，今日考勤成功 <span class="text-#0066ff mx-1">1</span>，今日考勤成功未点名 <span
            class="text-#ff3333 mx-1">1</span> 人 -->
          请单人采集，人脸采集成功后，前往"人脸考勤"进行考勤
        </div>
        <div class="face-wrap flex">
          <!-- 摄像头采集区域 -->
          <div class="face-left-camera relative">
            <video
              id="video" ref="videoRef" width="600" height="450" autoplay muted playsinline
              style="transform: scaleX(-1);" class="video-element"
            />
            <canvas
              id="canvas" ref="canvasRef" width="600" height="450" class="absolute top-0 left-0 canvas-element"
              style="transform: scaleX(-1);"
            />
            <div v-if="student" class="faceMaskLine" />
            <div
              v-show="!videoReady"
              class="face-left absolute right-0 left-0 top-0 bottom-0 transition-opacity duration-300" :class="{ 'opacity-0': videoReady, 'opacity-100': !videoReady }"
            >
              <div class="cameraPic">
                <div class="moveLine" />
              </div>
              <div class="tips" :class="{ red: showVideoEndStream }">
                {{ showVideoEndStream ? '抱歉，未找到摄像头，请检查后重试' : '正在检测摄像头，请耐心等待' }}
              </div>
            </div>
            <!-- 人脸考勤准备区域 -->
            <div v-show="data == 2 && !isAttendanceStarted" class="face-card absolute right-0 left-0 top-0 bottom-0">
              <div class="flex-center absolute bottom-0 right-0 left-0 px-24px pb-50px">
                <a-button type="primary" class="w-250px h-40px font500 flex-center" @click="startAttendance">
                  <span class="startAttSpan1">
                    <img
                      class="animationImg w-10px"
                      src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                    >
                  </span>
                  <span class="startAttSpan2">
                    <img
                      class="animationImg w-10px"
                      src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                    >
                  </span>
                  <span class="mx-12px">开始考勤</span>
                  <span class="startAttSpan3">
                    <img
                      class="animationImg w-10px"
                      src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                    >
                  </span>
                  <span class="startAttSpan4">
                    <img
                      class="animationImg w-10px"
                      src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                    >
                  </span>
                </a-button>
              </div>
            </div>
            <!-- 脸部与摄像头平视，识别中 -->
            <div v-if="isAttendanceStarted && !showCooldownMessage" class="absolute top-0 right-0 left-0 z-200">
              <div
                class="flex flex-center h-40px w-100% text-#fff text-15px font500 "
                :style="isFaceDetected
                  ? 'background: linear-gradient(270deg, rgba(0, 103, 255, .1), rgba(0, 102, 255, .8) 49%, rgba(0, 102, 255, .1))'
                  : 'background: linear-gradient(270deg, rgba(255, 51, 50, .1), rgba(255, 51, 50, .8) 49%, rgba(255, 51, 50, .1))'"
              >
                <span class="startAttSpan1">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan2">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="mx-12px">{{ isFaceDetected ? '脸部与摄像头平视，识别中' : '未检测到人脸，请面对摄像头' }}</span>
                <span class="startAttSpan3">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan4">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
              </div>
            </div>
            <!-- 1分钟内不能重复刷脸 -->
            <div v-if="isAttendanceStarted && showCooldownMessage" class="absolute top-0 right-0 left-0 z-200">
              <div
                class="flex flex-center h-40px w-100% text-#fff text-15px font500 "
                :style="isFaceDetected
                  ? 'background: linear-gradient(90deg, rgba(255, 51, 50, .1), rgba(255, 51, 50, .8) 52%, rgba(255, 51, 50, .1))'
                  : 'background: linear-gradient(270deg, rgba(255, 51, 50, .1), rgba(255, 51, 50, .8) 49%, rgba(255, 51, 50, .1))'"
              >
                <span class="startAttSpan1">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan2">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="mx-12px">{{ isFaceDetected ? '1分钟内不能重复刷脸' : '未检测到人脸，请面对摄像头' }}</span>
                <span class="startAttSpan3">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan4">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
              </div>
            </div>
          </div>
          <div v-if="data == 1" class="face-right relative">
            <div class="t">
              学员人脸采集
            </div>
            <div class="con">
              <a-select
                v-model:value="student" allow-clear :filter-option="filterOption" show-search
                style="width: 100%;" placeholder="搜索学员姓名/手机号" option-label-prop="label"
              >
                <a-select-option v-for="(item, index) in studentList" :key="index" :value="item.id" :label="item.name">
                  <div class="flex justify-between flex-items-center">
                    <div>
                      <span>{{ item.name }}</span>
                      <span class="text-3 text-#888 ml-2 font-300">{{ item.phone }}</span>
                    </div>
                    <span v-if="item.status === 0" class="bg-#eee px-2.5 py-0.5 text-3 rounded-10">未采集</span>
                    <span v-else class="bg-#e6ffec px-2.5 py-0.5 text-3 rounded-10 text-#0c3">已采集</span>
                  </div>
                </a-select-option>
              </a-select>
              <div v-if="!student" class="faceNoDataBox">
                请先选择一位学员进行人脸采集
              </div>
              <div v-else class="faceDataBox rounded-2">
                <div class="flex bg-#fff rounded-2 rounded-lb-0 rounded-rb-0 justify-between mt2 px3 py2 flex-center">
                  <span class="text-3.5">
                    {{ studentList.find(s => s.id === student)?.name || '' }} {{ studentList.find(s => s.id
                      === student)?.phone || '' }}
                  </span>
                  <span
                    v-if="studentList.find(s => s.id === student)?.status === 1"
                    class="text-3 bg-#e6ffec text-#0c3 rounded-4 px3 font500" style="line-height:2;"
                  >
                    <CheckCircleFilled /> <span>已采集</span>
                  </span>
                </div>
                <div class="flex-center bg-#fff px-10px pb-16px">
                  <div v-if="!capturedImageUrl" class="bg-#f6f7f8 rounded-2 flex-center w-100% py-12px flex-col">
                    <div
                      class="w-64px h-64px bg-#dfdfdf rounded-20 flex-center text-#ffffff80 text-40px font800"
                      :class="{ 'bg-#00cc33': isFaceDetected }"
                      style="font-family: PingFangSC-Regular, PingFang SC, sans-serif;"
                    >
                      {{ isFaceDetected ? '✓' : '?' }}
                    </div>
                    <div class="text-10px flex-col flex-center text-#888 mt-6px">
                      <span v-if="!isFaceDetected">未检测到人脸，请将人脸对准左侧屏幕</span>
                      <span v-else>已检测到人脸</span>
                      <span>点击下方"确认采集"上传</span>
                    </div>
                  </div>
                  <!-- 图像回显区域 -->
                  <div v-else class="face-right-image">
                    <div class="captured-image-container">
                      <img :src="capturedImageUrl" alt="已采集人脸" class="captured-image">
                      <div class="text-3 text-#888 mt-2 pb-2 flex">
                        采集时间：{{ capturedTime }}
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="capturedImageUrl" class="btn flex justify-between mt-4 w-100%">
                  <a-button class="text-#ff3333 border-0 flex-1 mr-5" @click="deleteFace">
                    删除人脸
                  </a-button>
                  <a-button class="text-#06f border-0 flex-1" @click="retakeFace">
                    重新采集
                  </a-button>
                </div>
                <!-- 确认采集按钮 -->
                <div v-if="!capturedImageUrl" class="flex-center absolute bottom-0 right-0 left-0 px-24px pb-24px">
                  <a-button
                    type="primary" class="w-100% h-40px font500 flex-center" :disabled="!isFaceDetected"
                    @click="captureFace"
                  >
                    <span class="startAttSpan1">
                      <img
                        class="animationImg w-10px"
                        src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                      >
                    </span>
                    <span class="startAttSpan2">
                      <img
                        class="animationImg w-10px"
                        src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                      >
                    </span>
                    <span class="mx-12px">确认采集</span>
                    <span class="startAttSpan3">
                      <img
                        class="animationImg w-10px"
                        src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                      >
                    </span>
                    <span class="startAttSpan4">
                      <img
                        class="animationImg w-10px"
                        src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                      >
                    </span>
                  </a-button>
                </div>
              </div>
            </div>
          </div>
          <div v-if="data == 2" class="face-right relative">
            <div class="t">
              学员考勤
            </div>
            <div class="con2 con scrollbar">
              <div
                v-for="(item, index) in attendanceRecords"
                :key="index" class="flex flex-items-center mb-12px pb-12px border-x-0 border-t-0 border-b border-color-#e6e6e6 border-solid "
              >
                <div class="left w-40px h-40px">
                  <img width="40px" height="40px" class="rounded-20 object-cover" :src="item.faceImage" alt="">
                </div>
                <div class="center mx-10px flex-1">
                  <div class="name flex flex-items-center">
                    <span class="text-16px font500 text-#222">{{ item.studentName }}</span>
                    <span class="bg-#e6f0ff rounded-20 px-10px py-2px text-12px text-#0066ff font500">自动签到</span>
                  </div>
                  <div class="tips text-#ff9900 text-3 font-500">
                    <ExclamationCircleFilled /> 考勤当日无排课计划
                  </div>
                </div>
                <div class="right flex flex-items-end flex-col">
                  <div class="icon">
                    <CheckCircleFilled class="text-#01c38f text-22px" />
                  </div>
                  <div class="time text-3 text-#7b889d">
                    <!-- 格式化成 这样的格式 05-11 18:33 -->
                    {{ formatDate(item.timestamp) }}
                  </div>
                </div>
              </div>
            </div>
            <div v-if="isAttendanceStarted" class="flex-center absolute bottom-0 right-0 left-0 px-24px pb-24px">
              <a-button type="primary" class="w-100% h-40px font500 flex-center" @click="endAttendance">
                <span class="startAttSpan1">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan2">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="mx-12px">结束考勤</span>
                <span class="startAttSpan3">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
                <span class="startAttSpan4">
                  <img
                    class="animationImg w-10px"
                    src="https://pcsys.admin.ybc365.com/3551cca2-7ab0-4d9f-bb52-902a88b8cdbd.png"
                  >
                </span>
              </a-button>
            </div>
          </div>
        </div>
      </div>
      <div style="padding-right: 20px;">
        <div class="faceBottom">
          <div class="t">
            注意事项
          </div>
          <div class="li">
            <div>面部平视摄像</div>
            <div>被遮挡</div>
            <div>面部平视摄像</div>
            <div>面部平视摄像</div>
          </div>
          <ul>
            <li>若人脸考勤无反应，请刷新浏览器再次尝试；如无法正常使用，请切换为谷歌浏览器</li>
            <li>未注册的学员，不能完成考勤，需要 1 秒人脸采集</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.face {
  width: 100%;
  min-height: 100vh;
  background: #f0f0fb;
  padding-bottom: 30px;

  .faceInner {
    width: 960px;
    margin: 0 auto;

    .active {
      background: #06f;
      border-color: #06f;
      box-shadow: 0 2px 0 rgba(0, 0, 0, .045);
      color: #fff;
      text-shadow: 0 -1px 0 rgba(0, 0, 0, .12);
      font-weight: bold;
    }

    .faceBody {
      width: 100%;
      height: 545px;
      background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/face-bg-new.d77f45f0.png");
      background-size: 100% 100%;
      margin-left: -11px;
      padding: 16px 34px 32px;

      .faceTips {
        color: #222;
        font-size: 12px;
        display: flex;
        align-items: center;
        height: 36px;
        margin-bottom: 14px;
        border: 1px solid;
        border-radius: 4px;
        border-image: linear-gradient(94deg, #fff, hsla(0, 0%, 100%, .13)) 1 1;
        color: #222;
        background: rgba(159, 196, 253, .28);
        box-shadow: 0 1px 4px 0 rgba(142, 185, 230, .35), inset 0 -1px 1px 0 #fcfffc, inset 1px 1px 0 0 hsla(0, 0%, 100%, .74), inset 0 1px 0 0 rgba(153, 208, 255, .45);
        backdrop-filter: blur(25.085829px);
        padding-left: 20px;
      }

      .face-wrap {
        display: flex;
        justify-content: space-between;

        .face-left-camera {
          width: 600px;
          height: 450px;
          position: relative;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 14px;

          canvas,
          video {
            background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12171/static/face-add.78ea9a4d.png");
            background-size: 100% 100%;
            border-radius: 6px;

          }

          .faceMaskLine {
            position: absolute;
            top: 213px;
            width: 301px;
            height: 19px;
            background: url(https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12171/static/face-att-line.3b7ae3a4.png) no-repeat;
            background-size: contain;
            transform: translateZ(0);
            animation: moveLineDown 3s cubic-bezier(0.4, 0, 0.2, 1) infinite;
            z-index: 10;
          }

          .face-card {
            backdrop-filter: blur(30px);
            background: rgba(0, 23, 58, .45);
            border-radius: 8px;
          }
        }

        .face-left {
          width: 600px;
          height: 450px;
          margin-right: 14px;
          background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/face-init.611d0d29.png");
          background-size: 100% 100%;
          display: flex;
          align-items: center;
          flex-direction: column;
          padding-top: 118px;
          transition: opacity 0.3s ease-in-out;
          z-index: 10;
          transform: translateZ(0);
          backface-visibility: hidden;
          perspective: 1000px;

          .cameraPic {
            width: 67px;
            height: 65px;
            background: url("https://pcsys.admin.ybc365.com//172d2f4e-dd0a-40a6-8278-766575e19367.png");
            background-size: 100% 100%;
            position: relative;
            transform: translateZ(0);
            will-change: transform;

            .moveLine {
              position: absolute;
              left: -5px;
              width: 77px;
              height: 22px;
              background: url("https://pcsys.admin.ybc365.com//0fead903-f008-4633-83dc-eb12b5333452.png") no-repeat;
              background-size: contain;
              animation: moveUpDown 1.5s cubic-bezier(0.4, 0, 0.2, 1) infinite;
              transform: translateZ(0);
              will-change: transform;
            }
          }

          @keyframes moveLineDown {
            0% {
              transform: translateY(-120px);
            }

            50% {
              transform: translateY(100px);
            }

            100% {
              transform: translateY(-120px);
            }
          }

          @keyframes moveUpDown {
            0% {
              transform: translateY(0);
            }

            50% {
              transform: translateY(60px);
            }

            100% {
              transform: translateY(0);
            }
          }

          .tips {
            color: #fff;
            font-size: 12px;
            font-weight: bold;
            line-height: 4;
            transform: translateZ(0);
            will-change: opacity;

            &::before {
              display: inline-block;
              content: "";
              width: 8px;
              height: 8px;
              margin-right: 5px;
              border-radius: 4px;
              background: #57c7ff;
              transform: translateZ(0);
            }
          }

          .red {
            &::before {
              background: red;
            }
          }
        }

        .face-right {
          flex: 1;
          height: 450px;
          background: red;
          background: url('https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/face-title-bg-new.3c96e3f1.png');
          background-size: 100% 100%;

          .t {
            height: 38px;
            line-height: 44px;
            text-align: center;
            font-size: 16px;
            font-weight: 500;
          }

          .con {
            padding: 12px;
            margin: 10px 0;

            .faceNoDataBox {
              background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/face-no-data-new.f42af9dd.png") no-repeat center;
              background-size: 160px 100px;
              height: 290px;
              font-size: 12px;
              display: flex;
              align-items: center;
              justify-content: center;
              padding-top: 55px;
              color: #888;
            }
          }

          .con2 {
            overflow-y: scroll;
            max-height: calc(100% - 120px);
          }
        }
      }
    }

    .faceBottom {
      width: 100%;
      height: 218px;
      background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/face-bottom-bg-new.bb4e8d81.png");
      background-size: 100% 100%;
      padding: 30px 26px 0;
      margin-right: 20px;

      .t {
        margin-bottom: 24px;
        font-family: PingFangSC-Medium, PingFang SC, sans-serif;
        font-size: 18px;
        font-weight: 500;
        color: #222;
      }

      .li {
        display: flex;
        justify-content: space-between;

        div {
          text-align: right;
          flex: 1;
          height: 48px;
          display: flex;
          align-items: center;
          justify-content: center;
          padding-left: 60px;

          &:first-child {
            background: url('https://pcsys.admin.ybc365.com//4f4e1526-6335-45df-b9ea-80fd5ddc0d67.png');
            background-size: 100% 100%;
            padding-top: 4px;
          }

          &:nth-child(2) {
            background: url('https://pcsys.admin.ybc365.com//67fc388d-8075-4be0-82f8-56b19323886f.png');
            background-size: 100% 100%;
            padding-top: 2px;
            padding-right: 28px;
          }

          &:nth-child(3) {
            background: url('https://pcsys.admin.ybc365.com//d692d98c-24f6-44b8-bd25-9942076f46dd.png');
            background-size: 100% 100%;
            padding-right: 44px;
          }

          &:nth-child(4) {
            background: url('https://pcsys.admin.ybc365.com//f2e8bf11-e7b3-4cd2-95b9-c81c0bf25910.png');
            background-size: 100% 100%;
            padding-left: 66px;
          }
        }
      }

      ul {
        padding-left: 18px;
        font-size: 14px;
        color: #888;
        margin-top: 18px;
      }
    }
  }

  .startAttSpan1 {
    animation: backgroundAnimation 1.6s infinite;
  }

  .startAttSpan2 {
    animation: backgroundAnimation2 1.6s infinite;
  }

  .startAttSpan3 {
    animation: backgroundAnimation 1.6s .8s infinite;
  }

  .startAttSpan4 {
    animation: backgroundAnimation2 1.6s .8s infinite;
  }
}

@keyframes backgroundAnimation {

  0%,
  50%,
  75%,
  to {
    opacity: .4
  }

  25% {
    opacity: 1
  }
}

@keyframes backgroundAnimation2 {

  0%,
  25%,
  75%,
  to {
    opacity: .4
  }

  50% {
    opacity: 1
  }
}

.face-right-image {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;

  .captured-image-container {
    display: flex;
    flex-direction: column;
    // align-items: center;
  }

  .captured-image {
    width: 100%;
    height: 173px;
    object-fit: cover;
    border-radius: 8px;
  }

  .no-image-placeholder {
    height: 100px;
    background: #f6f7f8;
    border-radius: 8px;
  }
}

.video-element {
  opacity: 0;
  transition: opacity 0.3s ease-in-out;

  &.loaded {
    opacity: 1;
  }
}

.canvas-element {
  opacity: 0;
  transition: opacity 0.3s ease-in-out;

  &.loaded {
    opacity: 1;
  }
}

.face-card {
  backdrop-filter: blur(30px);
  background: rgba(0, 23, 58, .45);
  border-radius: 8px;
  z-index: 10;
}

.attendance-result {
  background: rgba(0, 0, 0, 0.5);
  padding: 20px 40px;
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.scrollbar {
  &::-webkit-scrollbar {
    width: 5px;
    height: 10px;
  }

  &::-webkit-scrollbar-thumb {
    border-radius: 5px;
    -webkit-box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);
    background: rgba(190, 190, 190, 0.2);
  }

  &::-webkit-scrollbar-track {
    -webkit-box-shadow: inset 0 0 5px rgba(227, 227, 227, 0.2);
    border-radius: 0;
    background: rgba(0, 0, 0, 0.1);
  }
}
</style>
