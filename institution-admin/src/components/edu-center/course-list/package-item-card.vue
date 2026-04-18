<script setup>
import { InfoCircleOutlined } from '@ant-design/icons-vue'
import { ref } from 'vue'

const props = defineProps({
  item: {
    type: Object,
    required: true,
  },
  total: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['cancel'])

const itemFormRef = ref(null)

function getCourseTagList(item) {
  const tags = []
  if (item.teachMethod === 1) {
    tags.push({ text: '班级授课', color: '#e6f0ff', textColor: '#0066ff', type: 'normal' })
  }
  if (item.teachMethod === 2) {
    tags.push({ text: '1v1授课', color: '#e6f0ff', textColor: '#0066ff', type: 'normal' })
  }
  if (item.chargeMethods) {
    tags.push({ text: item.chargeMethods, color: '#e6f0ff', textColor: '#0066ff', type: 'normal' })
  }
  if (item.categoryName) {
    tags.push({ text: item.categoryName, color: '#e6f0ff', textColor: '#0066ff', type: 'normal' })
  }
  if (Array.isArray(item.properties)) {
    item.properties.forEach((property) => {
      tags.push({
        text: property.lessonPropertyOptionName,
        color: '#e6f0ff',
        textColor: '#0066ff',
        type: 'normal',
        key: property.lessonPropertyOptionId,
      })
    })
  }
  if (item.hasExperiencePrice) {
    tags.push({ text: '体验价', color: '#fff5e6', textColor: '#ff9900', type: 'normal' })
  }
  return tags
}

function getTagStyle(type = 'normal') {
  const baseStyle = {
    borderRadius: '20px',
    marginRight: '0',
    height: '20px',
  }
  if (type === 'primary') {
    return {
      ...baseStyle,
      color: '#fff',
    }
  }
  return {
    ...baseStyle,
    color: '#0066ff',
  }
}

function getSelectedSku() {
  return (props.item.productSku || []).find(sku => String(sku.id) === String(props.item.productSkuId))
}

function handlePriceListChange(value) {
  props.item.productSkuId = value
  props.item.skuCount = 0
  props.item.freeQuantity = 0
  props.item.discountType = '0'
  props.item.discountNumber = undefined
  props.item.discountRate = undefined
}

function getGiftLabel() {
  const sku = getSelectedSku()
  switch (sku?.lessonModel) {
    case 2:
      return '赠送天数'
    case 3:
      return '赠送金额'
    default:
      return '赠送课时'
  }
}

function getGiftPlaceholder() {
  const sku = getSelectedSku()
  switch (sku?.lessonModel) {
    case 2:
      return '请输入天数'
    case 3:
      return '请输入金额'
    default:
      return '请输入课时'
  }
}

function getGiftPrecision() {
  const sku = getSelectedSku()
  switch (sku?.lessonModel) {
    case 2:
      return 0
    case 3:
      return 2
    default:
      return 2
  }
}

function numRules() {
  return [
    {
      validator: async () => {
        const total = Number(props.item?.skuCount || 0) + Number(props.item?.freeQuantity || 0)
        if (total <= 0) {
          throw new Error('购买数+赠送数的总和不可为0')
        }
      },
    },
  ]
}

function giftNumRules() {
  return numRules()
}

function discountNumberRules() {
  return [
    {
      required: props.item?.discountType === '1',
      type: 'number',
      min: 0.01,
      message: '请输入优惠金额',
    },
  ]
}

function discountRateRules() {
  return [
    {
      required: props.item?.discountType === '2',
      type: 'number',
      min: 0.1,
      max: 9.9,
      message: '请输入折扣（0.1-9.9）',
    },
  ]
}

async function validate() {
  await itemFormRef.value?.validate?.()
}

defineExpose({
  validate,
})
</script>

<template>
  <div class="container-box mb-4">
    <div class="container-box-top">
      <div class="flex flex-items-center">
        <span class="font-600 text-#222 text-4">{{ item.name }}</span>
        <span class="ml-4">
          <a-space :size="5" class="flex flex-wrap">
            <template v-for="tag in getCourseTagList(item)" :key="tag.key || tag.text">
              <a-tag v-if="tag.type === 'tooltip'" class="font500" :style="getTagStyle(tag.type)" :color="tag.color">
                {{ tag.text }}
                <a-tooltip>
                  <template #title>
                    {{ tag.tooltipTitle }}
                  </template>
                  <InfoCircleOutlined class="ml-1" />
                </a-tooltip>
              </a-tag>
              <a-tag v-else class="font500" :style="getTagStyle(tag.type)" :color="tag.color">
                <span :style="{ color: tag.textColor }">{{ tag.text }}</span>
              </a-tag>
            </template>
          </a-space>
        </span>
      </div>
      <div class="right">
        <span class="price">合计：¥ {{ total.toFixed(2) }}</span>
        <span class="line" />
        <span class="cancel" @click="emit('cancel')">取消选择</span>
      </div>
    </div>

    <div class="container-box-bottom scrollbar relative">
      <a-form ref="itemFormRef" layout="vertical" :model="item">
        <a-space :size="36" class="pr-30px">
          <a-form-item name="productSkuId" label="报价单" :rules="[{ required: true, message: '请选择报价单' }]">
            <a-select
              v-model:value="item.productSkuId"
              placeholder="请选择报价单"
              class="quote-select"
              style="width: 220px"
              popup-class-name="auto-width-dropdown"
              @change="handlePriceListChange"
            >
              <a-select-opt-group v-for="group in item.priceList" :key="group.label" :label="group.label">
                <a-select-option v-for="option in group.options" :key="option.value" :value="option.value">
                  <div class="flex flex-items-center ">
                    <span class="tagCustom mr-6px" v-if="option.lessonAudition">体验价</span>
                    <span>{{ option.label }}</span>
                  </div>
                </a-select-option>
              </a-select-opt-group>
            </a-select>
          </a-form-item>

          <a-form-item name="skuCount" label="购买份数" :rules="numRules()">
            <a-input-number v-model:value="item.skuCount" style="width: 120px" :precision="0" :min="0" placeholder="0" />
          </a-form-item>

          <a-form-item name="freeQuantity" :label="getGiftLabel()" :rules="giftNumRules()">
            <a-input-number
              v-model:value="item.freeQuantity"
              :placeholder="getGiftPlaceholder()"
              style="width: 120px"
              :precision="getGiftPrecision()"
              :min="0"
            />
          </a-form-item>

          <a-form-item label="单课优惠">
            <div class="flex flex-items-center ">
              <a-radio-group v-model:value="item.discountType" class="custom-radio whitespace-nowrap">
                <a-radio value="0">
                  无
                </a-radio>
                <a-radio value="1">
                  金额
                </a-radio>
                <a-radio value="2">
                  折扣
                </a-radio>
              </a-radio-group>

              <template v-if="item.discountType === '1'">
                <a-form-item name="discountNumber" :rules="discountNumberRules()" class="ml-2 mgnone">
                  <div class="flex flex-center styleCss relative">
                    <a-input-number
                      v-model:value="item.discountNumber"
                      style="width: 120px"
                      :precision="2"
                      :min="0.01"
                      placeholder="金额"
                    />
                    <span class="ml-1">元</span>
                  </div>
                </a-form-item>
              </template>

              <template v-if="item.discountType === '2'">
                <a-form-item name="discountRate" :rules="discountRateRules()" class="ml-2 mgnone">
                  <div class="flex flex-center styleCss relative">
                    <a-input-number
                      v-model:value="item.discountRate"
                      style="width: 120px"
                      :precision="1"
                      :min="0.1"
                      :max="9.9"
                      placeholder="折扣"
                    />
                    <span class="ml-1">折</span>
                  </div>
                </a-form-item>
              </template>
            </div>
          </a-form-item>
        </a-space>
      </a-form>
    </div>
  </div>
</template>

<style scoped lang="less">
.quote-select {
  :deep(.ant-select-selector) {
    padding-right: 36px !important;
  }

  :deep(.ant-select-selection-item) {
    overflow: hidden;
  }

  :deep(.ant-select-selection-item > div) {
    display: flex;
    align-items: center;
    min-width: 0;
    overflow: hidden;
  }

  :deep(.ant-select-selection-item > div > span:last-child) {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.tagCustom {
  display: inline-block;
  padding: 2px 5px;
  background: linear-gradient(212deg, #fad961, #f76b1c);
  border-radius: 8px;
  font-size: 10px;
  font-weight: 500;
  color: #fff;
  line-height: 12px;
}

.container-box {
  background: #fafafa;

  .container-box-top {
    height: 44px;
    background: #f0f5fe;
    padding: 10px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-top-left-radius: 16px;
    border-top-right-radius: 16px;

    .right {
      display: flex;
      align-items: center;

      span.price {
        font-weight: 500;
        color: #222;
        white-space: nowrap;
        font-size: 14px;
      }

      .line {
        height: 12px;
        width: 1px;
        background: #ccc;
        display: inline-block;
        margin: 0 18px;
      }

      .cancel {
        color: var(--pro-ant-color-primary);
        font-weight: bold;
        cursor: pointer;
      }
    }
  }

  .container-box-bottom {
    padding: 10px 8px 0 24px;
    overflow-x: auto;

    :deep(.ant-form-item-label label) {
      white-space: nowrap;
    }
  }
}

.mgnone {
  margin: 0 !important;

  :deep(.ant-form-item-explain-error) {
    position: absolute;
  }
}

:deep(.ant-form-item-explain-error) {
  font-size: 12px !important;
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}
</style>
