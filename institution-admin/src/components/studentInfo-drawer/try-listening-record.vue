<script setup>

</script>

<template>
  <div class="register-for-courses p3 pr0">
    <a-space class="flex flex-wrap">
      <div v-for="(item, index) in 3" :key="index" class=" bg-white w-92.5 h-34 rounded-2">
        <div class="t flex justify-between p4 pb0 pt3 flex-items-center">
          <span class="text-4 text-#222 font-500 flex-1 line-clamp-1 mr-10">一对一认知课</span>
        </div>
        <div class="tag px3 mt1">
          <a-space>
            <span class="bg-#eee text-#888 text-3 rounded-10 px3 py1">已取消</span>
          </a-space>
        </div>
        <div class="remaining-class-hours px3 text-3 text-#888 mt-5 flex justify-between">
          <span>上课时间：
            2025-04-02（周三）10:00 ~ 10:30</span>
        </div>
        <div class="validity px3 text-#888 flex justify-start flex-items-center">
          <span class="text-3 ">上课教师：郭杨</span>
        </div>
      </div>
    </a-space>
  </div>
</template>

<style lang="less" scoped></style>
