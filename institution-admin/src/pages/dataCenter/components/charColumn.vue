<script setup>
import { Column } from '@antv/g2plot'

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
  new Column(container.value, {
    height: 230,
    maxColumnWidth: 30, // 最大宽度
    minColumnWidth: 30, // 最小宽度
    columnStyle: {
      // 渐变
      fill: 'l(90) 0:#0371ff 1:#a3c6f3',
    },
    padding: 'auto',
    data,
    xField: 'type',
    yField: 'sales',
    xAxis: {
      label: {
        autoRotate: false,
      },
      max: 6,
    },
    slider: {
      start: 0,
      end: 7 / data.length,
      handlerStyle: {
        width: 0,
        height: 0,
        opacity: 0,
      },
      backgroundStyle: {
        stroke: '#f0f0f0',
        height: 15,
        opacity: 0.2,
      },
      foregroundStyle: {
        height: 15,
        fillOpacity: 0.5,
      },
      // textStyle: {
      //   opacity: 0, // 隐藏文字
      // },
      // formatter: () => '', // 返回空字符串，不显示文字
    },
    xField: 'type',
    yField: 'sales',
    meta: {
      type: {
        alias: '类别',
      },
      sales: {
        alias: '销售额(万)',
      },
    },
    yAxis: {
      nice: false,
      min: 0,
      max: Math.max(...data.map(item => item.sales)) * 1.1, // 最大值的1.1倍，留点空间
    },
  }).render()
})
</script>

<template>
  <div ref="container" />
</template>

<style scoped lang="less"></style>
