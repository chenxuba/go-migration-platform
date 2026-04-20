<script setup>
import { ref } from 'vue'
import { Rate } from 'ant-design-vue'

const scales = ref([
  {
    id: 1,
    name: '焦虑自评量表',
    desc: '焦虑自评量表是一种广泛应用于心理健康评估的工具，能够帮助用户快速了解自身的焦虑水平，适用于日常自查和专业筛查。通过科学的题项设计，量表可为用户提供个性化的情绪调节建议。',
    cover: 'https://picsum.photos/300/150?random=1',
    color: 'linear-gradient(135deg, #e0e7ff 0%, #c7d2fe 100%)',
    highlight: false,
    owned: true,
    isNew: false
  },
  {
    id: 2,
    name: '抑郁自评量表',
    desc: '抑郁自评量表可用于评估个体的抑郁症状严重程度，适合心理健康自测和临床辅助诊断。量表内容涵盖情绪、认知、行为等多个方面，帮助用户全面了解自身心理状态。',
    cover: 'https://picsum.photos/300/150?random=2',
    color: 'linear-gradient(135deg, #dbeafe 0%, #60a5fa 100%)',
    highlight: true,
    owned: false,
    isNew: false
  },
  {
    id: 3,
    name: '儿童注意力测评',
    desc: '儿童注意力测评专为儿童设计，能够科学评估注意力集中与分散情况，辅助家长和教师发现孩子的专注力优势与短板，为个性化教育提供数据支持。',
    cover: 'https://picsum.photos/300/150?random=3',
    color: 'linear-gradient(135deg, #f0fdf4 0%, #6ee7b7 100%)',
    highlight: false,
    owned: false,
    isNew: true
  },
  {
    id: 4,
    name: '社交能力评估',
    desc: '社交能力评估量表用于测量个体在不同社交场合下的沟通、表达、合作与人际互动能力，适用于青少年及成人。通过多维度分析，帮助用户发现自身社交优势与待提升领域。',
    cover: 'https://picsum.photos/300/150?random=4',
    color: 'linear-gradient(135deg, #fdf6b7 0%, #fbbf24 100%)',
    highlight: false,
    owned: false,
    isNew: false
  },
  {
    id: 5,
    name: '学习能力测评',
    desc: '学习能力测评量表聚焦于学习动机、专注力、记忆力、理解力等方面，适合学生及终身学习者。通过科学测评，助力个体制定更高效的学习策略。',
    cover: 'https://picsum.photos/300/150?random=5',
    color: 'linear-gradient(135deg, #e0ffe7 0%, #a7f3d0 100%)',
    highlight: false,
    owned: true,
    isNew: false
  },
  {
    id: 6,
    name: '情绪管理能力评估',
    desc: '情绪管理能力评估量表帮助用户了解自身在情绪识别、调节、表达等方面的能力水平，适用于希望提升自我调节能力的各类人群。',
    cover: 'https://picsum.photos/300/150?random=6',
    color: 'linear-gradient(135deg, #fce7f3 0%, #f472b6 100%)',
    highlight: true,
    owned: false,
    isNew: true
  },
  {
    id: 7,
    name: '压力应对能力测评',
    desc: '压力应对能力测评量表用于评估个体在面对压力时的应对方式和心理弹性，帮助用户发现自身的抗压优势与待提升空间，适用于职场人士及学生群体。',
    cover: 'https://picsum.photos/300/150?random=7',
    color: 'linear-gradient(135deg, #e0e7ff 0%, #a5b4fc 100%)',
    highlight: false,
    owned: false,
    isNew: false
  },
  {
    id: 8,
    name: '自信心测评',
    desc: '自信心测评量表通过多维度题项，帮助用户了解自我认知、社交自信、表达自信等方面的能力水平，适用于希望提升自信心的各类人群。',
    cover: 'https://picsum.photos/300/150?random=8',
    color: 'linear-gradient(135deg, #f0fdf4 0%, #fbbf24 100%)',
    highlight: false,
    owned: true,
    isNew: true
  }
])

function handleAction(item) {
  if (item.owned) {
    alert(`开始评估: ${item.name}`)
  } else {
    alert(`增购: ${item.name}`)
  }
}
</script>

<template>
  <div class="assessment-calendar">
    <!-- <h2 class="title">评估量表</h2> -->
    <div class="card-list">
      <div
        class="card"
        v-for="item in scales"
        :key="item.id"
        :style="{ background: item.color, boxShadow: item.highlight ? '0 8px 32px #60a5fa88, 0 0 0 4px #2d5cf6' : '0 2px 16px #e6e6e6' }"
        :class="{ highlight: item.highlight }"
      >
        <div v-if="item.highlight" class="badge">
          推荐
          <Rate :value="5" disabled allow-half style="font-size: 10px; margin-right: 6px; vertical-align: middle;" />
        </div>
        <div class="cover-wrap">
          <div v-if="item.isNew" class="new-badge">上新</div>
          <img class="cover" :src="item.cover" :alt="item.name" />
          <div class="cover-mask"></div>
        </div>
        <div class="card-header">
          <span class="card-title">{{ item.name }}</span>
        </div>
        <div class="card-body">
          <p class="desc">{{ item.desc }}</p>
        </div>
        <div class="card-divider"></div>
        <div class="card-footer">
          <button
            v-if="item.owned"
            class="btn"
            @click="handleAction(item)"
          >查看详情</button>
          <button
            v-else
            class="btn"
            @click="handleAction(item)"
          >去增购</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.assessment-calendar {
  // min-height: 100vh;
  padding: 48px 0 64px 0;
  .title {
    text-align: center;
    font-size: 32px;
    font-weight: 800;
    color: #2d5cf6;
    margin-bottom: 30px;
    letter-spacing: 2px;
    text-shadow: 0 2px 12px #c7d2fe88;
  }
  .card-list {
    display: grid;
    grid-template-columns: repeat(4, 280px);
    gap: 30px;
    width: 100%;
    justify-content: center;
    @media (max-width: 1410px) {
      grid-template-columns: repeat(3, 1fr);
    }
    @media (max-width: 1000px) {
      grid-template-columns: repeat(2, 1fr);
    }
    @media (max-width: 700px) {
      grid-template-columns: 1fr;
    }
    .card {
      position: relative;
      border-radius: 18px;
      width: 280px;
      min-width: 0;
      display: flex;
      flex-direction: column;
      transition: transform 0.25s, box-shadow 0.25s, border-color 0.25s;
      border: 3px solid transparent;
      overflow: hidden;
      background-clip: padding-box;
      box-shadow: 0 2px 16px #e6e6e6;
      &:hover {
        transform: translateY(-10px) scale(1.035);
        box-shadow: 0 16px 48px #60a5fa99, 0 0 0 4px #2d5cf6;
        border-color: #60a5fa;
      }
      &.highlight {
        border-color: #2d5cf6;
        .card-title {
          color: #2d5cf6;
        }
        .badge {
          background: linear-gradient(90deg, #2d5cf6 0%, #60a5fa 100%);
          color: #fff;
        }
      }
      .badge {
        display: flex;
        align-items: center;
        justify-content: center;
        position: absolute;
        top: 4px;
        left: 6px;
        z-index: 3;
        background: linear-gradient(90deg, #2d5cf6 0%, #60a5fa 100%);
        color: #fff;
        font-size: 13px;
        font-weight: 700;
        padding: 3px 10px;
        border-radius: 10px;
        box-shadow: 0 1px 4px #2d5cf655, 0 0 0 1px #fff;
        letter-spacing: 2px;
        user-select: none;
        text-shadow: 0 1px 2px #1746b8, 0 0 2px #fff;
        border: 1px solid #fff;
        .ant-rate {
          margin-bottom: 1px;
          font-size: 10px !important;
        }
      }
      .cover-wrap {
        width: 100%;
        height: 110px;
        border-radius: 14px 14px 0 0;
        overflow: hidden;
        background: #f4f6fa;
        display: flex;
        align-items: center;
        justify-content: center;
        position: relative;
        .cover {
          width: 100%;
          height: 100%;
          object-fit: cover;
          display: block;
          transition: transform 0.7s cubic-bezier(.23,1.02,.64,1.01);
        }
        .cover-mask {
          position: absolute;
          left: 0; top: 0; right: 0; bottom: 0;
          background: linear-gradient(180deg, rgba(255,255,255,0.12) 60%, rgba(45,92,246,0.10) 100%);
          pointer-events: none;
        }
      }
      &:hover .cover {
        transform: scale(1.06) translateY(-4px);
      }
      .card-header {
        padding: 16px 0 6px 0;
        text-align: center;
        .card-title {
          font-size: 18px;
          font-weight: 800;
          letter-spacing: 1px;
          color: #222;
          text-shadow: 0 2px 8px #c7d2fe44;
        }
      }
      .card-body {
        padding: 0 18px 10px 18px;
        .desc {
          color: #444;
          margin-bottom: 0;
          font-size: 13px;
          min-height: 50px;
          height: 90px;
          line-height: 1.8;
          text-align: justify;
          overflow: hidden;
          text-overflow: ellipsis;
          display: -webkit-box;
          -webkit-line-clamp: 4;
          -webkit-box-orient: vertical;
          position: relative;
          /* 滚动动画 */
          transition: transform 0.5s;
        }
      }
      .card:hover .desc {
        overflow-y: auto;
        animation: scroll-desc 4s linear 1;
      }
      @keyframes scroll-desc {
        0% {
          transform: translateY(0);
        }
        10% {
          transform: translateY(0);
        }
        90% {
          transform: translateY(calc(-100% + 108px));
        }
        100% {
          transform: translateY(calc(-100% + 108px));
        }
      }
      .card .desc::-webkit-scrollbar {
        display: none;
      }
      .card-divider {
        height: 2px;
        width: 80%;
        margin: 0 auto 10px auto;
        border-radius: 2px;
        background: linear-gradient(90deg, #2d5cf6 0%, #60a5fa 100%);
        opacity: 0.12;
      }
      .card-footer {
        padding: 10px 18px 18px 18px;
        text-align: center;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 10px;
        .owned-label {
          background: #f0f6ff;
          color: #2563eb;
          font-size: 14px;
          font-weight: 400;
          border-radius: 10px;
          padding: 4px 12px 4px 8px;
          display: inline-flex;
          align-items: center;
          line-height: 1;
          letter-spacing: 0.5px;
          box-shadow: none;
          &::before {
            content: '';
            display: inline-block;
            width: 14px;
            height: 14px;
            margin-right: 5px;
            background: url('data:image/svg+xml;utf8,<svg fill="none" stroke="%232563eb" stroke-width="2" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg"><path d="M4 8.5l3 3 5-6" stroke-linecap="round" stroke-linejoin="round"/></svg>') no-repeat center/14px 14px;
            vertical-align: middle;
          }
        }
        .purchase-label {
          background: #fef3c7;
          color: #f59e42;
          font-size: 15px;
          font-weight: 700;
          border-radius: 14px;
          padding: 6px 18px;
          box-shadow: 0 1px 4px #f59e4222;
        }
        .btn {
          background: linear-gradient(90deg, #2d5cf6 0%, #60a5fa 100%);
          color: #fff;
          border: none;
          border-radius: 24px;
          padding: 7px 20px;
          font-size: 15px;
          font-weight: 800;
          cursor: pointer;
          box-shadow: 0 2px 16px #60a5fa33, 0 0 0 2px #fff inset;
          transition: background 0.2s, box-shadow 0.2s, transform 0.2s;
          border: 2px solid #fff;
          outline: none;
          position: relative;
          &:hover {
            background: linear-gradient(90deg, #1746b8 0%, #2d5cf6 100%);
            box-shadow: 0 6px 24px #2d5cf655, 0 0 0 2px #2d5cf6 inset;
            transform: scale(1.04);
          }
        }
      }
    }
  }
}

/* 新增上新标记样式 */
.new-badge {
  position: absolute;
  top: 4px;
  right: 6px;
  z-index: 3;
  background: linear-gradient(90deg, #ff1a1a 0%, #ff7b7b 100%);
  color: #fff;
  font-size: 12px;
  font-weight: 900;
  padding: 3px 18px;
  border-radius: 10px;
  box-shadow: 0 2px 10px #ff1a1a55, 0 0 0 2px #fff;
  letter-spacing: 3px;
  user-select: none;
  text-shadow: 0 1px 2px #b80000, 0 0 2px #fff;
  border: 2px solid #fff;
}
</style>