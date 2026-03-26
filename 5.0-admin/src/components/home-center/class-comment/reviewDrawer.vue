<script setup>
import Icon, { CloseOutlined } from '@ant-design/icons-vue'

const open = defineModel({
  type: Boolean,
  default: false,
})
</script>

<template>
  <div>
    <a-drawer
      v-model:open="open" :body-style="{ padding: '0', background: '#f7f7fd' }" :closable="false"
      placement="right" width="1165px" :keyboard="false" :mask-closable="false"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            课堂点评详情
          </div>
          <a-button type="text" class="close-btn" @click="open = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <!-- 自定义内容 -->
      <div class="py-10px">
        <div class="unfollowContainer">
          <span
            role="img" aria-label="exclamation-circle" class="anticon anticon-exclamation-circle"
            style="color: rgb(0, 102, 255); font-size: 18px;"
          >
            <Icon>
              <template #component>
                <svg
                  viewBox="64 64 896 896" focusable="false" data-icon="exclamation-circle" width="1em" height="1em"
                  fill="currentColor" aria-hidden="true"
                >
                  <path
                    d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm-32 232c0-4.4 3.6-8 8-8h48c4.4 0 8 3.6 8 8v272c0 4.4-3.6 8-8 8h-48c-4.4 0-8-3.6-8-8V296zm32 440a48.01 48.01 0 010-96 48.01 48.01 0 010 96z"
                  />
                </svg>
              </template>
            </Icon>
          </span>
          <span class="unfollowHintText">此位学员家长尚未关注家校平台</span>
          <div class="unfollowButtonContainer">
            <span class="unfollowButtonText">点击邀请关注，接收课评提醒</span>
            <img
              src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAYCAYAAADzoH0MAAAAAXNSR0IArs4c6QAAAERlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAA6ABAAMAAAABAAEAAKACAAQAAAABAAAAEKADAAQAAAABAAAAGAAAAAAEJTozAAABqklEQVQ4EaVUzUrDQBCeqeAb6NWDPoHgqwgeFEFNYrVoEUFBUdGTiApatKGiUnwZ776B4MUXkGrGb7OdTbZpSaALYf6//XZ2NkS66jJHdVlQs6qspYmRTNMvfeB7p0hOqxabPAuQUEJMkhYmdEKhHFUFsQAxfwNgBUV/aaHQOQVyUAWEvaRQlsCjC99E6mfapzZfeTkDhg9ggpEsA+QFn2VH1KSYbwfqnKlJzkGP3MVx1uCwPSG6Aeh2luBrRQYaD2QdaoxPc+pg8qBhlUUGGom5g0NswlQmLTDZ0LBKRVe7KAPZgvO+H5D0eG1+1sTRDDQj5haKdvum2bCDOTFXnq5yBpoZyB5UvdIe9Bn05KucgQLkJbvGurvOh4t6KDtw2t0ZY0+4ZuxuEsuPMFYTQ4mwyZ3ZCcvMZojRdjdgnKMZjDVIkawCPJvCGjWGTaFhULwF+5ieEFN2TbwPHSRT4y37bNVln/MrzOw5x3yt4WEyA4hkEW16c8VEh6B9Oawo77M0GzJFP/SJwGQaZDpGty/yiaN024OeO6/pylnVYh80kFn8B+d9Z7n1D6Ubbvlzv5iIAAAAAElFTkSuQmCC"
              class="arrowAnimation" style="margin-left: 3px;"
            >
          </div>
        </div>
        <div class="mx-15px my-10px p-20px bg-white rounded-15px">
          <custom-title class="mb-12px" title="上课信息" font-size="18px" font-weight="500" />
          <div class="flex flex-col gap-30px mt-20px">
            <div class="flex ">
              <div class="w-120px text-right">
                上课内容：
              </div>
              <div class="text-#888">
                发发发 地方
              </div>
            </div>
            <div class="flex">
              <div class="w-120px text-right">
                课堂点评模版：
              </div>
              <div class="flex items-center">
                <span>综合总分</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.unfollowContainer {
  background: #e6f0ff;
  padding: 15px;
  display: flex;
  align-items: center;
  margin: 0 0 8px;

  .unfollowHintText {
    font-weight: 400;
    font-size: 14px;
    color: #222;
    line-height: 20px;
    text-align: left;
    font-style: normal;
    margin-left: 6px;
  }

  .unfollowButtonContainer {
    word-wrap: break-word;
    font-size: 14px;
    line-height: 1.5715;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    margin-left: 10px;
    align-items: center;
    cursor: pointer;

    .unfollowButtonText {
      word-wrap: break-word;
      cursor: pointer;
      box-sizing: border-box;
      height: 22px;
      font-weight: 400;
      font-size: 14px;
      color: #06f;
      line-height: 22px;
      text-align: left;
      font-style: normal;
    }

    .arrowAnimation {
      width: 8px;
      height: 12px;
    }
  }
}
</style>
