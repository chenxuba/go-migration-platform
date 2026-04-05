import { computed, h, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import messageService from '@/utils/messageService'

interface PickerOptionItem {
  value: string
  label: string
  mobile?: string
}

interface UseSmartTimetablePickerOptions {
  activeGroupLabel: ComputedRef<string>
  currentModel: Ref<string>
  displayedGroupKey: Ref<string>
  detectOneToOneAvailability: (value: string | number | undefined) => void | Promise<void>
  periodGroupForKey: (key: string) => any
  resetAssistantConflicts: () => void
}

export function useSmartTimetablePicker(options: UseSmartTimetablePickerOptions) {
  const oneToOneRecordId = ref<string | undefined>(undefined)
  const oneToOnePickerOpen = ref(false)
  const selectedAssistantIds = ref<string[]>([])
  const assistantKeyword = ref('')
  const oneToOneData = ref<any[]>([])
  const oneToOneListLoading = ref(false)
  const assistantOptions = ref<PickerOptionItem[]>([])
  const assistantOptionsLoading = ref(false)

  let lastHandledOneToOneId = ''
  let preserveOneToOnePickerOpen = false

  function normalizeStringArray(values: unknown) {
    return Array.from(
      new Set(
        (Array.isArray(values) ? values : [])
          .map(value => String(value || '').trim())
          .filter(Boolean),
      ),
    )
  }

  const normalizedSelectedAssistantIds = computed(() => normalizeStringArray(selectedAssistantIds.value))

  const assistantOptionMap = computed(() => {
    const map = new Map<string, PickerOptionItem>()
    assistantOptions.value.forEach((item) => {
      map.set(String(item.value), item)
    })
    return map
  })

  function assistantNameById(id: unknown) {
    const normalized = String(id || '').trim()
    if (!normalized)
      return ''
    return assistantOptionMap.value.get(normalized)?.label || normalized
  }

  const selectedAssistantText = computed(() => {
    const names = normalizedSelectedAssistantIds.value.map(id => assistantNameById(id)).filter(Boolean)
    return names.length ? names.join('、') : '未安排'
  })

  const oneToOneDropdownStyle = computed(() => ({
    width: '520px',
    minWidth: '520px',
    '--assistant-render-tick': `${assistantKeyword.value}|${normalizedSelectedAssistantIds.value.join(',')}|${String(oneToOneRecordId.value || '')}|${options.activeGroupLabel.value}`,
  }))

  const currentDisplayedGroupTeacherIds = computed(() => {
    const bound = options.periodGroupForKey(options.displayedGroupKey.value)?.boundTeachers
    return Array.isArray(bound)
      ? bound.map((item: any) => String(item.id ?? '').trim()).filter(Boolean)
      : []
  })

  function isAssistantAllowedInDisplayedGroup(id: unknown) {
    const normalized = String(id || '').trim()
    if (!normalized)
      return false
    const allowed = currentDisplayedGroupTeacherIds.value
    if (!allowed.length)
      return true
    return allowed.includes(normalized)
  }

  const assistantOptionsInPicker = computed(() => {
    const keyword = String(assistantKeyword.value || '').trim().toLowerCase()
    return assistantOptions.value.filter((item) => {
      if (!isAssistantAllowedInDisplayedGroup(item.value))
        return false
      if (!keyword)
        return true
      const blob = `${item.label || ''} ${item.mobile || ''} ${item.value || ''}`.toLowerCase()
      return blob.includes(keyword)
    })
  })

  watch(
    [options.displayedGroupKey, assistantOptions],
    () => {
      if (!normalizedSelectedAssistantIds.value.length)
        return
      const bound = options.periodGroupForKey(options.displayedGroupKey.value)?.boundTeachers
      const allowed = Array.isArray(bound)
        ? bound.map((item: any) => String(item.id ?? '').trim()).filter(Boolean)
        : []
      const next = normalizedSelectedAssistantIds.value.filter((id) => {
        if (!allowed.length)
          return true
        return allowed.includes(String(id || '').trim())
      })
      if (next.length === normalizedSelectedAssistantIds.value.length)
        return
      const removedCount = normalizedSelectedAssistantIds.value.length - next.length
      handleAssistantSelectChange(next)
      if (removedCount > 0) {
        messageService.warning(`已切换到${options.activeGroupLabel.value || '当前组'}，自动移除 ${removedCount} 位非本组助教`, { duration: 4500 })
      }
    },
    { immediate: true },
  )

  function mapRowToOneToOneOption(row: any) {
    const id = String(row.id || '').trim()
    const studentId = String(row.studentId || '').trim()
    const studentName = String(row.studentName || '').trim()
    const lessonName = String(row.lessonName || '').trim()
    const name = String(row.name || '').trim()
      || (studentName && lessonName ? `${studentName}-${lessonName}` : studentName || lessonName || id)
    return {
      id,
      studentId,
      studentName,
      courseId: row.lessonId != null ? String(row.lessonId) : '',
      courseName: lessonName,
      name,
    }
  }

  function mapStaffToAssistantOption(row: any): PickerOptionItem {
    const value = String(row.id ?? '').trim()
    const label = String(row.nickName || row.name || value).trim()
    return {
      value,
      label: label || value,
      mobile: String(row.mobile ?? '').trim(),
    }
  }

  async function fetchOneToOneOptionsForTimetable() {
    oneToOneListLoading.value = true
    try {
      const res = await getOneToOneListApi({
        pageRequestModel: {
          needTotal: false,
          pageSize: 500,
          pageIndex: 1,
          skipCount: 0,
        },
        queryModel: {
          status: [1],
        },
      })
      if (res.code === 200 && res.result) {
        const list = Array.isArray(res.result.list) ? res.result.list : []
        oneToOneData.value = list.map(mapRowToOneToOneOption).filter(item => item.id)
      }
      else {
        oneToOneData.value = []
        messageService.error(res.message || '获取1对1列表失败')
      }
    }
    catch (error) {
      console.error('fetchOneToOneOptionsForTimetable', error)
      oneToOneData.value = []
      messageService.error('获取1对1列表失败')
    }
    finally {
      oneToOneListLoading.value = false
    }
  }

  async function fetchAssistantOptions() {
    assistantOptionsLoading.value = true
    try {
      const res = await getUserListApi({
        pageRequestModel: {
          needTotal: false,
          pageSize: 500,
          pageIndex: 1,
          skipCount: 0,
        },
        queryModel: {
          status: 0,
        },
      })
      if (res.code !== 200) {
        assistantOptions.value = []
        messageService.error(res.message || '获取助教列表失败')
        return
      }

      const rows = Array.isArray(res.result) ? res.result : []
      assistantOptions.value = rows
        .map(mapStaffToAssistantOption)
        .filter(item => item.value)
    }
    catch (error: any) {
      console.error('fetchAssistantOptions failed', error)
      assistantOptions.value = []
      messageService.error(error?.response?.data?.message || error?.message || '获取助教列表失败')
    }
    finally {
      assistantOptionsLoading.value = false
    }
  }

  function filterOneToOneOption(input: string, option: any) {
    const q = (input || '').trim().toLowerCase()
    if (!q)
      return true
    const id = option?.value != null ? String(option.value) : ''
    const item = oneToOneData.value.find(r => r.id === id)
    if (!item)
      return true
    const blob = `${item.name} ${item.studentName} ${item.courseName} ${item.studentId}`.toLowerCase()
    return blob.includes(q)
  }

  function handle1v1(value: string | number | undefined) {
    const nextId = String(value || '').trim()
    if (nextId !== lastHandledOneToOneId && normalizedSelectedAssistantIds.value.length)
      selectedAssistantIds.value = []
    assistantKeyword.value = ''
    lastHandledOneToOneId = nextId
    if (nextId) {
      requestAnimationFrame(() => {
        oneToOnePickerOpen.value = true
      })
    }
    else {
      oneToOnePickerOpen.value = false
    }
    void options.detectOneToOneAvailability(value)
  }

  function handleAssistantSelectChange(value: unknown) {
    selectedAssistantIds.value = normalizeStringArray(value)
    if (options.currentModel.value === '1' && oneToOneRecordId.value) {
      void options.detectOneToOneAvailability(oneToOneRecordId.value)
    }
    else {
      options.resetAssistantConflicts()
    }
  }

  function resetOneToOnePickerState() {
    oneToOneRecordId.value = undefined
    oneToOnePickerOpen.value = false
    selectedAssistantIds.value = []
    assistantKeyword.value = ''
    lastHandledOneToOneId = ''
    preserveOneToOnePickerOpen = false
  }

  function requestKeepOneToOnePickerOpen() {
    preserveOneToOnePickerOpen = true
    requestAnimationFrame(() => {
      oneToOnePickerOpen.value = true
      preserveOneToOnePickerOpen = false
    })
  }

  function toggleAssistantOption(value: unknown, checked: boolean) {
    const normalized = String(value || '').trim()
    if (!normalized)
      return
    const next = new Set(normalizedSelectedAssistantIds.value)
    if (checked)
      next.add(normalized)
    else
      next.delete(normalized)
    handleAssistantSelectChange([...next])
    requestKeepOneToOnePickerOpen()
  }

  function handleOneToOneDropdownVisibleChange(open: boolean) {
    if (!open && preserveOneToOnePickerOpen) {
      oneToOnePickerOpen.value = true
      return
    }
    oneToOnePickerOpen.value = open
  }

  function renderOneToOneDropdown({ menuNode }: { menuNode: any }) {
    const sideChildren = [
      h('div', {
        class: 'st-top-1v1-dropdown__section-head',
        style: {
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
          marginBottom: '12px',
        },
      }, [
        h('span', {
          class: 'st-top-1v1-dropdown__section-title',
          style: {
            color: '#262626',
            fontSize: '14px',
            fontWeight: 700,
            lineHeight: 1,
          },
        }, '选择助教'),
        h('span', {
          class: 'st-top-1v1-dropdown__section-hint',
          style: {
            color: '#8c8c8c',
            fontSize: '12px',
            lineHeight: 1,
          },
        }, oneToOneRecordId.value ? '多选，可不选' : '先选1v1后配置'),
      ]),
    ]

    if (oneToOneRecordId.value) {
      sideChildren.push(
        h('input', {
          class: 'st-top-1v1-dropdown__search-input',
          value: assistantKeyword.value,
          placeholder: '搜索助教',
          style: {
            width: '100%',
            height: '30px',
            padding: '0 10px',
            color: '#262626',
            fontSize: '12px',
            background: '#fff',
            border: '1px solid #d9d9d9',
            borderRadius: '8px',
            outline: 'none',
            boxSizing: 'border-box',
            marginBottom: '4px',
          },
          onInput: (event: any) => {
            assistantKeyword.value = event?.target?.value || ''
          },
          onFocus: () => {
            requestKeepOneToOnePickerOpen()
          },
          onClick: () => {
            requestKeepOneToOnePickerOpen()
          },
        }),
      )

      if (normalizedSelectedAssistantIds.value.length) {
        sideChildren.push(
          h('div', {
            class: 'st-top-1v1-dropdown__summary',
            style: {
              marginBottom: '2px',
              color: '#5b6475',
              fontSize: '12px',
              lineHeight: '1.5',
            },
          }, `已选助教：${selectedAssistantText.value}`),
        )
      }

      if (assistantOptionsInPicker.value.length) {
        sideChildren.push(
          h('div', {
            class: 'st-top-1v1-dropdown__assistant-list',
            style: {
              display: 'flex',
              flexDirection: 'column',
              gap: '0px',
              flex: 1,
              overflowY: 'auto',
              paddingRight: '4px',
            },
          }, assistantOptionsInPicker.value.map((item) => {
            const checked = normalizedSelectedAssistantIds.value.includes(String(item.value))
            return h('div', {
              class: 'st-top-1v1-dropdown__assistant-item',
              key: item.value,
              style: {
                display: 'flex',
                alignItems: 'center',
                gap: '6px',
                minHeight: '30px',
                padding: '2px 0px',
                borderRadius: '10px',
                cursor: 'pointer',
                boxSizing: 'border-box',
                userSelect: 'none',
              },
              onMousedown: (event: MouseEvent) => {
                event.preventDefault()
                event.stopPropagation()
              },
              onClick: () => {
                toggleAssistantOption(item.value, !checked)
              },
            }, [
              h('span', {
                class: 'st-top-1v1-dropdown__assistant-checkbox',
                style: {
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: '16px',
                  height: '16px',
                  borderRadius: '4px',
                  border: checked ? '1px solid #1677ff' : '1px solid #8c8c8c',
                  background: checked ? '#1677ff' : '#fff',
                  color: '#fff',
                  flex: '0 0 auto',
                  fontSize: '11px',
                  fontWeight: 700,
                  lineHeight: 1,
                },
              }, checked ? '✓' : ''),
              h('span', {
                class: 'st-top-1v1-dropdown__assistant-name',
                style: {
                  flex: 1,
                  minWidth: 0,
                  color: '#262626',
                  fontSize: '12px',
                  fontWeight: 600,
                  lineHeight: '20px',
                },
              }, item.label),
              item.mobile
                ? h('span', {
                    class: 'st-top-1v1-dropdown__assistant-mobile',
                    style: {
                      color: '#8c8c8c',
                      fontSize: '11px',
                      lineHeight: '20px',
                      flex: '0 0 auto',
                    },
                  }, item.mobile)
                : null,
            ])
          })),
        )
      }
      else {
        sideChildren.push(h('div', {
          class: 'st-top-1v1-dropdown__empty',
          style: {
            padding: '14px 0 4px',
            color: '#8c8c8c',
            fontSize: '12px',
            lineHeight: '18px',
          },
        }, '暂无匹配助教'))
      }
    }
    else {
      sideChildren.push(h('div', {
        class: 'st-top-1v1-dropdown__empty',
        style: {
          padding: '14px 0 4px',
          color: '#8c8c8c',
          fontSize: '12px',
          lineHeight: '18px',
        },
      }, '先选 1v1，再在右侧勾选助教。'))
    }

    return h('div', {
      class: 'st-top-1v1-dropdown',
      style: {
        display: 'flex',
        width: '520px',
        minWidth: '520px',
        maxWidth: '520px',
        minHeight: '280px',
        maxHeight: '280px',
        background: '#fff',
        borderRadius: '12px',
        overflow: 'hidden',
      },
    }, [
      h('div', {
        class: 'st-top-1v1-dropdown__list',
        style: {
          flex: '0 0 278px',
          minWidth: '278px',
          maxWidth: '278px',
          overflowY: 'auto',
          borderRight: '1px solid #f0f0f0',
        },
      }, [menuNode]),
      h('div', {
        class: 'st-top-1v1-dropdown__side',
        style: {
          display: 'flex',
          flex: 1,
          flexDirection: 'column',
          minWidth: 0,
          padding: '14px 16px 16px',
          background: 'linear-gradient(180deg, #fcfdff 0%, #fff 100%)',
        },
        onMousedown: (event: MouseEvent) => event.stopPropagation(),
      }, sideChildren),
    ])
  }

  return {
    assistantNameById,
    assistantKeyword,
    assistantOptionsLoading,
    fetchAssistantOptions,
    fetchOneToOneOptionsForTimetable,
    filterOneToOneOption,
    handle1v1,
    handleOneToOneDropdownVisibleChange,
    normalizedSelectedAssistantIds,
    oneToOneData,
    oneToOneDropdownStyle,
    oneToOneListLoading,
    oneToOnePickerOpen,
    oneToOneRecordId,
    renderOneToOneDropdown,
    resetOneToOnePickerState,
    selectedAssistantIds,
    selectedAssistantText,
  }
}
