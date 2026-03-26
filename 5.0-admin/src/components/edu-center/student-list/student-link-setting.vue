<script setup>
const state = reactive({
  checked1: false,
  checked2: false,
  checked3: false,
  checked4: false,
  checked5: false,
  checked6: false,
})
</script>

<template>
  <div class="tab-content">
    <!-- 意向学员录入设置 -->
    <div class="setting">
      <div class="title">
        售前人员
      </div>
      <div class="table-wrap mt-3 mb-3">
        <table border>
          <tbody>
            <tr>
              <td class="td1">
                销售员
              </td>
              <td>
                负责向家长介绍课程和服务，跟进报名意向，促成学员成功入学
              </td>
            </tr>
            <tr>
              <td class="td1">
                采单员
              </td>
              <td>
                <a-switch v-model:checked="state.checked1" /> <span class="ml-2">走访社区、学校等地，与家长沟通，收集学员信息</span>
              </td>
            </tr>
            <tr>
              <td class="td1">
                电话销售
              </td>
              <td>
                <a-switch v-model:checked="state.checked2" /> <span class="ml-2">通过电话与家长沟通，介绍课程优势，解答疑问</span>
              </td>
            </tr>
            <tr>
              <td class="td1">
                前台
              </td>
              <td>
                <a-switch v-model:checked="state.checked3" /> <span class="ml-2">负责接待家长与学员，处理日常咨询、课务安排</span>
              </td>
            </tr>
            <tr>
              <td class="td1">
                副销售员
              </td>
              <td>
                <a-switch v-model:checked="state.checked4" /> <span class="ml-2">协助销售员跟进学员报名，处理家长咨询及意向登记</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div class="setting">
      <div class="title">
        售后人员
      </div>
      <div class="table-wrap mt-3 mb-3">
        <table border>
          <tbody>
            <tr>
              <td class="td1">
                学管师
              </td>
              <td>
                <a-switch v-model:checked="state.checked5" /> <span class="ml-2">负责深入了解学生，及时解决学生学习中的疑难问题与学习态度等方面的问题</span>
              </td>
            </tr>
            <tr>
              <td class="td1">
                顾问
              </td>
              <td>
                <a-switch v-model:checked="state.checked6" /> <span class="ml-2">专业的咨询服务人员，主要为学生，家长和教育机构提供个性化的咨询服务和建议</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div class="setting">
      <div class="title">
        班级人员
      </div>
      <div class="table-wrap mt-3 mb-3">
        <table border>
          <tbody>
            <tr>
              <td class="td1">
                班主任
              </td>
              <td>
                班课/1v1的班主任，负责调查研究学生情况，组织管理班集体，教育指导学生全面发展
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div class="setting">
      <div class="title">
        课程人员
      </div>
      <div class="table-wrap mt-3 mb-3">
        <table border>
          <tbody>
            <tr>
              <td class="td1">
                上课教师
              </td>
              <td>
                负责特定的教学工作，传授知识，进行教学活动和评估学生的学习成果
              </td>
            </tr>
            <tr>
              <td class="td1">
                上课助教
              </td>
              <td>
                协助教师完成教学准备工作，如课程设计，教学资料的收集和整理等
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.tab-content {
  background: #fff;
  border-radius: 12px;
  border-top-right-radius: 0 ;
  border-top-left-radius: 0;
  padding: 12px 20px;

  .setting {

    .title {
      font-size: 18px;
      color: #222;
      font-weight: 600;
      display: flex;
      align-items: center;

      &::before {
        display: inline-block;
        content: '';
        width: 4px;
        height: 14px;
        background: var(--pro-ant-color-primary);
        border-radius: 100px;
        margin-right: 6px;
      }
    }

    .tips {
      height: 40px;
      display: flex;
      align-items: center;
      background: var(--pro-ant-color-primary-bg-hover);
      color: var(--pro-ant-color-primary);
      padding: 0 24px;
      font-size: 14px;
      margin-bottom: 8px;
      border-radius: 4px;
      justify-content: flex-start;
    }

    .table-wrap {
      a {
        display: flex;
        align-items: center;
      }

      table {
        width: 100%;
        border-collapse: collapse;
        border: 1px solid #eee;
        border-radius: 8px;

        tr,
        td {
          border: 1px solid #eee;
        }

        td {
          padding: 18px 24px;
        }

        .td1 {
          width: 200px;
          text-align: center;
          font-size: 14px;
          font-family: PingFangSC-Regular, PingFang SC;
          font-weight: 400;
          color: #222;
        }
      }
      .day{
        color: var(--pro-ant-color-primary);
        font-weight: bold;
        font-size: 14px;
      }
      .icon{
        color: var(--pro-ant-color-primary);
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }
}
</style>
