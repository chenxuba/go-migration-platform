import { defineStore } from 'pinia'

export const useModalStore = defineStore('modal', {
  state: () => ({
    // 当前激活的模态框
    activeModal: null as string | null,
    // 模态框的参数
    modalParams: {} as Record<string, any>,
    // 模态框的标题
    modalTitle: '业务设置' as string,
    // 模态框的返回按钮
    modalBackButton: false as boolean,

  }),
  actions: {
    // 打开模态框
    openModal(name: string, params = {}) {
      this.activeModal = name
      this.modalParams = params
    },
    // 关闭模态框
    closeModal() {
      this.activeModal = null
      this.modalParams = {}
    },
    // 判断指定模态框是否激活
    isModalActive(name: string) {
      return this.activeModal === name
    },
    // 更新模态框标题
    updateModalTitle(title: string) {
      console.log('updateModalTitle', title)
      this.modalTitle = title
    },
    // 更新模态框返回按钮
    updateModalBackButton(backButton: boolean) {
      this.modalBackButton = backButton
    },
  },
})
