<script setup lang="ts">
const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open'])

const smsExamples = [
  {
    title: '课时不足提醒',
    content: '【IRTS】亲爱的家长您好，【学员姓名】在【常青藤培训学校】报读的【课程名称】仅剩余【X】课时，您可联系我们咨询续费事宜，期待继续为您提供优质的教学服务，感谢您的信赖与支持。',
  },
  {
    title: '课时到期提醒',
    content: '【IRTS】亲爱的家长您好，【学员姓名】在【常青藤培训学校】报读的【课程名称】即将于【20XX-XX-XX】到期，还剩余【X】课时。您可联系我们咨询续费事宜，期待继续为您提供优质的教学服务，感谢您的信赖与支持。',
  },
  {
    title: '课程到期提醒',
    content: '【IRTS】亲爱的家长您好，【学员姓名】在【常青藤培训学校】报读的【课程名称】即将于【20XX-XX-XX】到期。您可联系我们咨询续费事宜，期待继续为您提供优质的教学服务，感谢您的信赖与支持。',
  },
  {
    title: '金额不足提醒',
    content: '【IRTS】亲爱的家长您好，【学员姓名】在【常青藤培训学校】报读的【课程名称】仅剩余【X】元，您可联系我们咨询续费事宜，期待继续为您提供优质的教学服务，感谢您的信赖与支持。',
  },
  {
    title: '有效期到期提醒',
    content: '【IRTS】亲爱的家长您好，【学员姓名】在【常青藤培训学校】报读的【课程名称】即将于【20XX-XX-XX】到期，还剩余【X】元。您可联系我们咨询续费事宜，期待继续为您提供优质的教学服务，感谢您的信赖与支持。',
  },
]
</script>

<template>
  <a-modal
    :open="props.open"
    title="模板示例"
    :width="720"
    :footer="null"
    wrap-class-name="pending-renewal-template-example-modal"
    @cancel="emit('update:open', false)"
  >
    <div class="example-wrap">
      <section class="example-card">
        <div class="section-head">
          <a-tag color="blue">
            微信模板
          </a-tag>
          <span class="section-title">续费提醒</span>
        </div>
        <div class="example-block">
          学生姓名：[学员姓名]
        </div>
        <div class="example-block">
          缴费课程：[课程名称]，[剩余课时不足 / 剩余天数不足 / 剩余金额不足 / 剩余课时即将到期]
        </div>
        <div class="example-block">
          项目名称：[剩余 X 课时 / X 天 / X 元，请及时续费]
        </div>
        <div class="example-block">
          机构名称：[机构名称]
        </div>
        <div class="example-block">
          发布时间：[发送时间]
        </div>
        <div class="example-tip">
          系统统一使用一个公众号模板，按课时/天数/金额类别动态替换“缴费课程”和“项目名称”文案。
        </div>
      </section>

      <section class="example-card">
        <div class="section-head">
          <a-tag color="default">
            短信模板
          </a-tag>
          <span class="section-title">预留示例</span>
        </div>
        <div class="sms-list">
          <div v-for="item in smsExamples" :key="item.title" class="sms-item">
            <span class="sms-title">{{ item.title }}：</span>
            <span class="sms-content">{{ item.content }}</span>
          </div>
        </div>
      </section>
    </div>
  </a-modal>
</template>

<style lang="less">
.pending-renewal-template-example-modal {
  .ant-modal-content {
    border-radius: 12px;
    overflow: hidden;
  }

  .ant-modal-body {
    max-height: 70vh;
    overflow-y: auto;
    padding-top: 12px;
    padding-bottom: 20px;
    scrollbar-width: thin;
    scrollbar-color: #c6cfdd transparent;
  }

  .ant-modal-body::-webkit-scrollbar {
    width: 6px;
  }

  .ant-modal-body::-webkit-scrollbar-thumb {
    background: #c6cfdd;
    border-radius: 999px;
  }
}
</style>

<style lang="less" scoped>
.example-wrap {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.example-card {
  border: 1px solid #eef1f6;
  border-radius: 12px;
  background: #f8faff;
  padding: 16px;
}

.section-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.section-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
}

.example-block {
  border-radius: 10px;
  background: #fff;
  border: 1px solid #eef1f6;
  color: #303133;
  line-height: 22px;
  padding: 12px 14px;

  & + .example-block {
    margin-top: 10px;
  }
}

.example-tip {
  color: #7a8293;
  font-size: 12px;
  margin-top: 12px;
}

.sms-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sms-item {
  background: #fff;
  border: 1px solid #eef1f6;
  border-radius: 12px;
  box-shadow: 0 4px 14px rgb(15 23 42 / 3%);
  color: #303133;
  line-height: 22px;
  padding: 12px 14px;
}

.sms-title {
  color: #1f2329;
  font-weight: 600;
}

.sms-content {
  color: #303133;
}
</style>
