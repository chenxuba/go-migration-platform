<script setup>
import { computed, ref } from 'vue'
import { useStudentStore } from '@/stores/student'

const props = defineProps({
  name: {
    type: String,
    required: true,
  },
  gender: {
    type: String,
    default: '',
  },
  age: {
    type: String,
    default: '',
  },
  avatarUrl: {
    type: String,
    default: 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png',
  },
  autoWidth: {
    type: Boolean,
    default: true,
  },
  showGender: {
    type: Boolean,
    default: true,
  },
  showAge: {
    type: Boolean,
    default: true,
  },
  defaultActiveKey: {
    type: String,
    default: '0',
  },
  phone: {
    type: String,
    default: '',
  },
  relation: {
    type: String,
    default: '',
  },
  id: {
    type: [String, Number],
    default: '',
  },
})

const studentStore = useStudentStore()
const openDrawer = ref(false)

function handleSeeStuData() {
  studentStore.setStudentId(props.id)
  openDrawer.value = true
}

// 手机号脱敏处理
const maskedPhone = computed(() => {
  if (!props.phone)
    return ''
  return props.phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
})
</script>

<template>
  <a-tooltip>
    <template #title>
      查看学员档案
    </template>
    <div
      :class="{ 'w-30': autoWidth }" class="flex cursor-pointer hover flex-items-center h-4  my-14px"
      @click="handleSeeStuData"
    >
      <img width="36" height="36" class="mr-2" style="border-radius: 100%;" :src="avatarUrl" alt="">
      <div class="name mt-1">
        <div class="text-#222 name">
          {{ name }}
        </div>
        <div v-if="showGender || showAge || phone" class="text-3 text-#888 flex flex-items-center name ">
          <template v-if="showGender">
            <span class="whitespace-nowrap">{{ gender }}</span>
          </template>
          <span v-if="showGender && showAge" class="inline-block w-0.2 h-2.5 bg-#ccc ml-1.5 mr-1.5 name" />
          <template v-if="showAge">
            <span class="whitespace-nowrap">{{ age }}</span>
          </template>
          <template v-if="phone">
            <span class="whitespace-nowrap">{{ maskedPhone }}</span>
          </template>
        </div>
      </div>
    </div>
  </a-tooltip>
  <student-info-drawer v-model:open="openDrawer" :default-active-key="defaultActiveKey" />
</template>

<style lang="less" scoped>
.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
