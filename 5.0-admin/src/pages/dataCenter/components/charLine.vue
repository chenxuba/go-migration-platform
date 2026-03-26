<script setup>
import { Line } from '@antv/g2plot'

const container = ref(null)
const data = [
  {
    type: '家具家电',
    sales: 38,
  },
  {
    type: '粮油副食',
    sales: 52,
  },
  {
    type: '美容洗护',
    sales: 61,
  },
  {
    type: '母婴用品',
    sales: 43,
  },
  {
    type: '图书音像',
    sales: 35,
  },
  {
    type: '家居用品',
    sales: 30,
  },
  {
    type: '其他',
    sales: 25,
  },
  {
    type: '生鲜水果',
    sales: 18,
  },
  {
    type: '美妆个护',
    sales: 56,
  },
  {
    type: '食品饮料',
    sales: 40,
  },
  {
    type: '玩具乐器',
    sales: 20,
  },
  {
    type: '玩具乐器1',
    sales: 20,
  },
  {
    type: '玩具乐器2',
    sales: 20,
  },
  {
    type: '玩具乐器3',
    sales: 20,
  },
  {
    type: '玩具乐器4',
    sales: 20,
  },
  {
    type: '玩具乐器5',
    sales: 120,
  },
  {
    type: '玩具乐器6',
    sales: 20,
  },
  {
    type: '玩具乐器7',
    sales: 20,
  },
  {
    type: '玩具乐器8',
    sales: 20,
  },
  {
    type: '玩具乐器9',
    sales: 20,
  },
]

onMounted(() => {
  new Line(container.value, {
    height: 230,
    data,
    padding: 'auto',
    xField: 'type',
    yField: 'sales',
    xAxis: {
      // type: 'timeCat',
      // tickCount: 5,
    },
    // 设置折线为圆角
    smooth: true,
    // 设置折线颜色
    color: '#0371ff',
    // 设置折线样式
    lineStyle: {
      lineWidth: 3,
      // 可以设置虚线等
      // lineDash: [4, 4],
    },
  }).render()
})
</script>

<template>
  <div ref="container" />
</template>

<style scoped lang="less"></style>
