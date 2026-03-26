<script setup>
import { MobileOutlined, LockOutlined, EyeOutlined, EyeInvisibleOutlined, SafetyOutlined } from '@ant-design/icons-vue';
import { AxiosError } from 'axios';
import QRCode from 'qrcode';
import { loginApi } from '~/api/common/login';
import { getQueryParam } from '~/utils/tools';
import { useMessage, useNotification } from '@/composables/global-config';
import { useRouter } from 'vue-router';
import { useAuthorization } from '@/composables/authorization';

const activeKey = ref(0); // 0: 密码登录, 1: 短信登录
const mode = ref(true); // input: 密码登录, qrcode: 扫码登录
const formRef = ref();
const formState = reactive({
  username: '',
  password: '',
  verifyCode: '',
});
const agreeToTerms = ref(true);
const message = useMessage();
const notification = useNotification();
const router = useRouter();
const token = useAuthorization();
const submitLoading = ref(false);
const errorAlert = ref(false);
const qrCodeUrl = ref(''); // 存储生成的二维码URL

// 生成二维码
async function generateQRCode() {
  try {
    const currentUrl = window.location.href;
    const qrCodeDataUrl = await QRCode.toDataURL(currentUrl, {
      width: 200,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      }
    });
    qrCodeUrl.value = qrCodeDataUrl;
  } catch (error) {
    console.error('生成二维码失败:', error);
    message.error('生成二维码失败');
  }
}

// 监听mode变化，当切换到二维码模式时生成二维码
watch(mode, (newMode) => {
  if (!newMode) {
    generateQRCode();
  }
});

// 组件挂载时，如果当前是二维码模式则生成二维码
onMounted(() => {
  if (!mode.value) {
    generateQRCode();
  }
});

function getLoginParams() {
  if (activeKey.value === 0) {
    return {
      username: formState.username,
      password: formState.password,
      type: 'account',
    };
  } else {
    return {
      mobile: formState.username,
      code: formState.verifyCode,
      type: 'mobile',
    };
  }
}

async function onSubmit() {
  if (!agreeToTerms.value) {
    message.warning('请先阅读并同意《用户协议》和《隐私条款》');
    return;
  }
  submitLoading.value = true;
  try {
    await formRef.value?.validate();
    const params = getLoginParams();
    const { result } = await loginApi(params);
    if (result) {
      token.value = result?.token;
      notification.success({
        message: '登录成功',
        description: '欢迎回来！',
        duration: 1,
      });
      router.push({ path: "/" })
    } else {
      submitLoading.value = false;
    }
  } catch (e) {
    if (e instanceof AxiosError) errorAlert.value = true;
    submitLoading.value = false;
  }
}

function changeQrCode() {
  mode.value = !mode.value
}

</script>

<template>
  <div class="login-page flex">
    <div class="main-content flex">
      <div class="content flex">
        <div class="left">
          <div class="top flex">
            <div class="e-logo"></div>
          </div>
          <div class="bg-box"></div>
        </div>
        <div class="right flex" :class="mode ? 'phone' : 'qrcode'">
          <div class="switchLogin">
            <div class="switchBtn" v-if="mode" @click="changeQrCode"></div>
            <div class="switchBtn other" v-if="!mode" @click="changeQrCode"></div>
            <div class="switchTooltip" v-if="mode">App 扫码登录</div>
            <div class="switchTooltip other" v-if="!mode">其他登录方式</div>
          </div>
          <a-form ref='formRef' v-if="mode" :model="formState" autocomplete="off">
            <div class="login-phone-wrap">
              <div class="phoneBox">
                <div class="phoneBg"></div>
                <div class="loginTabContainer">
                  <a-tabs v-model:activeKey="activeKey" :tab-bar-style="{
                    'border-bottom-left-radius': '0px',
                    'border-bottom-right-radius': '0px',
                  }">

                    <a-tab-pane :key="0" tab="密码登录">
                    </a-tab-pane>
                    <a-tab-pane :key="1" tab="短信登陆">
                    </a-tab-pane>
                  </a-tabs>
                </div>
                <a-form-item name="username" :rules="[{ required: true, message: '请输入手机号' }]">
                  <div class="inputBox">
                    <a-input :bordered="false" autocomplete="off" class="w-300px bg-#f2f3f9"
                      v-model:value="formState.username" placeholder="请输入手机号">
                      <template #prefix>
                        <MobileOutlined />
                      </template>
                    </a-input>
                  </div>
                </a-form-item>
                <a-form-item name="password" :rules="[{ required: true, message: '请输入密码' }]" v-if="activeKey === 0">
                  <div class="inputBox">
                    <a-input :bordered="false" type="password" autocomplete='off' class="w-300px bg-#f2f3f9"
                      v-model:value="formState.password" placeholder="请输入密码">
                      <template #prefix>
                        <LockOutlined />
                      </template>
                      <template #suffix>
                        <EyeInvisibleOutlined />
                      </template>
                    </a-input>
                  </div>
                </a-form-item>
                <!-- 验证码 -->
                <a-form-item name="verifyCode" :rules="[{ required: true, message: '请输入验证码' }]" v-else>
                  <div class="inputBox">
                    <a-input :bordered="false" class="w-300px bg-#f2f3f9" v-model:value="formState.verifyCode"
                      placeholder="请输入验证码">
                      <template #prefix>
                        <SafetyOutlined />
                      </template>
                      <template #suffix>
                        <span class="text-#06f cursor-pointer">获取验证码</span>
                      </template>
                    </a-input>
                  </div>
                </a-form-item>
                <div class="submitBox">
                  <a-button type="primary" class="w-300px" :loading="submitLoading" @click="onSubmit">立即登录</a-button>
                  <a-checkbox v-model:checked="agreeToTerms" class="agreement-checkbox">
                    <span class="agreement-text">
                      已阅读并同意校宝的
                      <a href="#" class="link">《用户协议》</a>
                      与
                      <a href="#" class="link">《隐私条款》</a>
                    </span>
                  </a-checkbox>
                </div>
              </div>
            </div>
          </a-form>

          <div class="top2 flex" v-if="!mode">
            <div class="title">扫码登录</div>
            <div class="login-qrcode-wrap">
              <div class="image">
                <img :src="qrCodeUrl" alt="二维码" v-if="qrCodeUrl" draggable="false"
                  style="width: 100%; height: 100%; object-fit: contain; user-select: none;">
                <div v-else
                  style="display: flex; align-items: center; justify-content: center; width: 100%; height: 100%; color: #999; font-size: 14px;">
                  正在生成二维码...
                </div>
              </div>
            </div>
            <div class="desc">
              打开「云校 App」扫描二维码登录
            </div>
          </div>

          <div class="bottom2 flex" v-if="mode">
            <div class="download flex">
              <div class="e-icon"></div>
              下载「云校 App」
            </div>
            <div class="tel flex">
              咨询热线:
              <span class="number">021-80392253</span>
            </div>
          </div>
          <div class="bottom2 flex" v-if="!mode">
            <div class="download flex other">
              <div class="e-icon"></div>
              下载「云校 App」
            </div>
            <div class="tel other flex">
              咨询热线:
              <span class="number">021-80392253</span>
            </div>
          </div>
        </div>
      </div>
      <div class="more-link"></div>
    </div>
    <div class="footer">
      <span><a target="_blank" href="https://beian.miit.gov.cn"
          rel="noreferrer"><span class="beian-icon"></span>沪ICP备15044463号-1 &nbsp; 型号:YBC-IRTS-DE &nbsp;&nbsp;</a>已通过 ISO27001:2013
        信息安全认证</span>
    </div>
  </div>
</template>


<style lang="less" scoped>
.login-page {
  flex-direction: column;
  justify-content: center;
  align-content: flex-start;
  width: 100%;
  min-width: unset;
  height: 100%;
  min-height: 100vh;
  background-color: #eff5ff;

  .main-content {
    flex: 1 1;
    flex-direction: column;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    background-color: #eff5ff;

    .content {
      justify-content: space-between;
      width: 100%;
      max-width: 1300px;
      height: auto;
      min-height: 536px;
      border-radius: 32px;
      overflow: hidden;

      .left {
        flex: 1 1;
        flex-direction: column;
        flex-wrap: wrap;
        background-color: #fff;
        position: relative;

        .top {
          justify-content: space-between;
          align-items: center;
          padding: 13px 32px;
          position: absolute;
          top: 11px;

          .e-logo {
            width: 200px;
            height: 28px;
            // background: url("data:image/png;base64,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") 0 100% no-repeat;
            background-size: contain;
          }
        }

        .bg-box {
          width: 100%;
          height: 100%;
          background-repeat: no-repeat;
          background-size: cover;
          background-image: url("https://pcsys.admin.ybc365.com/b4d8b1e2-220f-4374-8f06-d59b10e70b6f.png");
          background-position: 0 75%;
        }
      }

      .right {
        position: relative;
        flex: 0 0 500px;
        max-width: 100%;
        width: 100%;
        flex-direction: column;
        justify-content: space-between;
        align-content: center;
        background-color: #005ce6;

        :deep(.ant-form-item) {
          margin-bottom: 12px;

          .ant-form-item-explain-connected {
            width: 300px;
            margin: auto;
          }
        }

        .switchLogin {
          position: absolute;
          top: -2px;
          right: 16px;
          z-index: 2;

          .switchBtn {
            background-image: url("https://pcsys.admin.ybc365.com/e50126bb-32f6-4a0a-8630-81415e5310ea.png");
            background-position: 7px 3px;
            background-size: 62px auto;
            width: 56px;
            height: 69px;
            cursor: pointer;
            background-repeat: no-repeat;
          }

          .switchBtn.other {
            background-image: url("https://pcsys.admin.ybc365.com/a10de00d-fc6f-4d82-a016-aa084cccd8d4.png");
            background-position: 50%;
            background-size: 84px auto;
          }

          .switchTooltip {
            position: absolute;
            display: flex;
            justify-content: center;
            align-items: center;
            width: 99px;
            height: 78px;
            font-family: PingFangSC-Regular, PingFang SC, sans-serif;
            font-size: 12px;
            font-weight: bold;
            color: #222;
            text-indent: -1px;
            background-repeat: no-repeat;
            background-position: 50%;
            top: -5px;
            left: -92px;
            color: #fff;
            background-image: url("https://pcsys.admin.ybc365.com/4dc3a53a-29f0-4c92-94e4-b36614e2c027.png");
            background-size: 133px auto;
          }

          .switchTooltip.other {
            background-image: url("https://pcsys.admin.ybc365.com/7a5ed862-50cb-4c1f-bb1c-b936aed76828.png");
            color: #333;
            top: -4px;
            left: -88px;
            padding-right: 4px;
            color: #222;
            background-size: 90px 30px;
            font-weight: normal;
          }
        }

        .login-phone-wrap {
          .phoneBox {
            padding-top: 5vw;
            max-width: 500px;
            margin: 0 auto;

            .phoneBg {
              width: 100%;
              height: 68px;
              background-position: 50%;
              background-size: contain;
            }

            .loginTabContainer {
              position: relative;
              display: flex;
              flex-direction: row;
              justify-content: center;
              width: 100%;
              height: 48px;
              margin-bottom: 12px;
              border-radius: 8px;
              width: 100%;
              border-radius: 10px;
              line-height: 40px;

              :deep(.ant-tabs-nav) {
                background: #fff;
                border-radius: 16px;
                margin: 0;
              }

              :deep(.ant-tabs-nav-wrap) {
                padding-left: 36px;
              }

              :deep(.ant-tabs-nav::before) {
                border: none;
              }

              :deep(.ant-tabs-tab) {
                font-size: 16px;
                font-weight: 500;
                color: #222;
              }

              :deep(.ant-tabs-ink-bar) {
                text-align: center;
                height: 9px !important;
                background: transparent;
                bottom: 1px !important;

                &::after {
                  position: absolute;
                  top: 0;
                  left: calc(50% - 12px);
                  width: 24px !important;
                  height: 4px !important;
                  border-radius: 2px;
                  background-color: var(--pro-ant-color-primary);
                  content: "";
                }
              }
            }

            .inputBox {
              position: relative;
              display: flex;
              justify-content: center;
              width: 100%;
              max-width: 300px;
              height: 48px;
              border-radius: 8px;
              margin: 0 auto;

              :deep(input) {
                background: transparent;
              }

              :deep(.ant-input-prefix) {
                font-size: 18px;
              }

              :deep(.ant-input-affix-wrapper:hover) {
                background-color: #f2f3f9;
              }

            }

            .submitBox {
              display: flex;
              flex-direction: column;
              justify-content: center;
              align-items: center;
              margin-top: 30px;
              margin-bottom: 8px;
              user-select: none;

              button {
                width: 100%;
                max-width: 300px;
                height: 40px;
                font-family: PingFangSC-Medium, PingFang SC, sans-serif;
                font-size: 16px;
                font-weight: 500;
                letter-spacing: 1px;
              }

              .agreement-checkbox {
                padding-left: 14px;
                margin-top: 8px;

                :deep(.ant-checkbox-wrapper) {
                  font-size: 12px;
                  color: #666;
                  line-height: 1.4;
                }

                :deep(.ant-checkbox) {
                  top: 0;
                }

                :deep(.ant-checkbox-checked .ant-checkbox-inner) {
                  background-color: #1890ff;
                  border-color: #1890ff;
                }

                .agreement-text {
                  font-size: 12px;
                  color: #666;
                  line-height: 1.4;

                  .link {
                    color: #1890ff;
                    text-decoration: none;
                    transition: color 0.2s ease;

                    &:hover {
                      color: #40a9ff;
                    }
                  }
                }
              }
            }

            .submitTips {}
          }

        }

        .top2 {
          flex-direction: column;
          flex-wrap: wrap;
          align-content: center;
          align-items: center;
          padding-top: 53px;

          .title {
            line-height: 40px;
            font-family: PingFangSC-Medium, PingFang SC, sans-serif;
            font-size: 32px;
            font-weight: bold;
            color: #fff;
            text-align: center;
          }

          .login-qrcode-wrap {
            position: relative;
            box-sizing: border-box;
            width: 276px;
            max-width: 80vw;
            height: 276px;
            max-height: 80vw;
            padding: 16px;
            margin-top: 17px;
            border-radius: 24px;
            overflow: hidden;
            background-color: #fff;

            .image {
              display: block;
              width: 100%;
              height: 100%;
            }
          }

          .desc {
            margin-top: 8px;
            line-height: 24px;
            font-size: 16px;
            color: #fff;
            text-align: center;
            font-weight: bold;
          }
        }

        .bottom2 {
          justify-content: space-between;
          padding: 0 56px 0 40px;
          margin-bottom: 15px;
          flex-wrap: wrap;
          gap: 10px;

          .download {
            color: #222;
            padding: 8px 16px;
            line-height: 22px;
            border-radius: 8px;
            font-size: 16px;
            align-items: center;



            .e-icon {
              background: url("https://pcsys.admin.ybc365.com/572acdf9-e563-48b0-8517-d5779448bad7.png") 100% 100% no-repeat;
              background-size: contain;
              width: 24px;
              height: 24px;
              margin-right: 4px;
              border-radius: 8px;
            }

            &:hover {
              color: #fff;
              cursor: pointer;
              background: #06f;

              .e-icon {
                background: url("https://pcsys.admin.ybc365.com/a3eb8cc1-34fb-40d1-8bd5-fe7112d522de.png");
                background-color: #fff;
                background-size: contain;
              }
            }
          }

          .download.other {
            color: #fff;
            cursor: pointer;
            background: hsla(0, 0%, 100%, .15);

            .e-icon {
              background: url("https://pcsys.admin.ybc365.com/a3eb8cc1-34fb-40d1-8bd5-fe7112d522de.png");
              background-color: #fff;
              background-size: contain;
            }
          }

          .tel {
            color: #222;
            line-height: 40px;
            font-size: 16px;
            font-weight: 400;

            .number {
              line-height: 40px;
              font-size: 18px;
              font-weight: bold;
              text-indent: 4px;
            }
          }

          .tel.other {
            color: #fff;
          }
        }
      }

      .right.phone {
        background-color: #fff;
      }
    }
  }
  .footer {
    height: 40px;
    line-height: 40px;
    font-size: 12px;
    color: #888;
    text-align: center;
    background: #fff;
    a{
      color:#888;
    }
    .beian-icon {
      position: relative;
    top: 2px;
    display: inline-block;
    width: 18px;
    height: 18px;
    margin-right: 8px;
    background: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12728/static/beian.0c577066.png") 100% 100% no-repeat;
    background-size: contain;
    }
  }
}

// 响应式媒体查询
@media (max-width: 1200px) {
  .login-page .main-content .content {
    max-width: 1000px;
    margin: 0 20px;
  }
}

@media (max-width: 992px) {
  .login-page .main-content .content {
    flex-direction: column;
    width: 100%;
    max-width: 100%;
    height: auto;
    min-height: auto;
    margin: 0;

    .left,
    .right {
      width: 100%;
      min-width: unset;
      max-width: 100%;
      height: auto;
    }

    .left {
      min-height: 300px;
    }

    .right {
      min-height: 400px;
      flex: 0 0 auto;
    }
  }
}

@media (max-width: 768px) {
  .login-page {
    min-height: 100vh;

    .main-content {
      padding: 0;

      .content {
        flex-direction: column;
        border-radius: 0;

        .left,
        .right {
          width: 100%;
          min-width: unset;
          max-width: 100%;
          height: auto;
          border-radius: 0;
        }

        .left {
          min-height: 200px;
        }

        .right {
          min-height: 300px;
        }
      }
    }
  }

  .login-page .main-content .content .right {
    .login-phone-wrap .phoneBox {
      padding-top: 24px;
      padding-left: 20px;
      padding-right: 20px;

    }

    .top2 {
      padding-top: 30px;

      .title {
        font-size: 24px;
      }

      .login-qrcode-wrap {
        width: 200px;
        height: 200px;
      }

      .desc {
        font-size: 14px;
      }
    }

    .bottom2 {
      padding: 0 20px 0 20px;
      justify-content: center;
      text-align: center;

      .download,
      .tel {
        width: 100%;
        justify-content: center;
      }
    }
  }
}

@media (max-width: 480px) {
  .login-page {
    .main-content .content {
      .left {
        min-height: 150px;
      }

      .right {
        min-height: 250px;
      }
    }
  }

  .login-page .main-content .content .right {
    .login-phone-wrap .phoneBox {
      padding: 12px;

      .loginTabContainer {
        :deep(.ant-tabs-nav-wrap) {
          padding-left: 20px;
        }
      }
    }

    .top2 {
      padding-top: 20px;

      .title {
        font-size: 20px;
      }

      .login-qrcode-wrap {
        width: 150px;
        height: 150px;
      }

      .desc {
        font-size: 12px;
      }
    }

    .bottom2 {
      padding: 0 15px 0 15px;

      .download,
      .tel {
        font-size: 14px;

        .number {
          font-size: 16px;
        }
      }
    }
  }
}
</style>