import { computed, reactive, ref, watch } from 'vue'

export interface Authority {
  id: number
  name: string
  remark: string
  type: number
  mode: number
  groupCode: string
  weight: number
  checked: boolean
}

export interface AuthorityChild {
  id: string
  menuName: string
  checked: boolean
  indeterminate: boolean
  children: Authority[]
}

export interface AuthorityGroup {
  id: string
  menuName: string
  checked: boolean
  indeterminate: boolean
  children: AuthorityChild[]
}

export function useRolePermissions(initialData: AuthorityGroup[] = [], loading: Ref<boolean>) {
  // 基础数据
  const boxList = ref<AuthorityGroup[]>(initialData)
  const searchValue = ref('')

  // 表单状态
  const formState = reactive({
    roleId: undefined,
    roleName: '',
    description: '',
    menuIds: [],
  })

  // 展开状态管理
  const expandedParentIds = ref<string[]>([])
  const expandedChildIds = ref<string[]>([])
  const expandedAuthorityIds = ref<string[]>([])
  const isAllExpanded = ref(false)

  // 过滤后的权限列表
  const filteredBoxList = computed(() => {
    if (!searchValue.value.trim()) {
      return boxList.value
    }

    const keyword = searchValue.value.toLowerCase().trim()

    return boxList.value.filter((parent) => {
      // 检查一级（父级）是否匹配
      const parentMatches = parent.menuName?.toLowerCase().includes(keyword)

      // 检查是否有匹配的子级
      const hasMatchingChildren = parent.children?.some((child) => {
        // 检查二级（子级）是否匹配
        const childMatches = child.menuName?.toLowerCase().includes(keyword)

        // 检查三级（权限点）是否匹配
        const hasMatchingchildren = child.children?.some((authority) => {
          return (authority.name?.toLowerCase().includes(keyword))
            || (authority.remark?.toLowerCase().includes(keyword))
        })

        return childMatches || hasMatchingchildren
      })

      // 显示条件：父级匹配 或 有匹配的子级
      return parentMatches || hasMatchingChildren
    })
  })

  // 过滤子级显示
  const getFilteredChildren = (parent: AuthorityGroup) => {
    if (!searchValue.value.trim()) {
      return parent.children
    }

    const keyword = searchValue.value.toLowerCase().trim()
    const parentMatches = parent.menuName.toLowerCase().includes(keyword)

    // 如果父级匹配，显示所有子级
    if (parentMatches) {
      return parent.children
    }

    // 否则只显示匹配的子级
    return parent.children.filter((child) => {
      const childMatches = child.menuName?.toLowerCase().includes(keyword)
      const hasMatchingchildren = child.children?.some((authority) => {
        return (authority.name?.toLowerCase().includes(keyword))
          || (authority.remark?.toLowerCase().includes(keyword))
      })
      return childMatches || hasMatchingchildren
    })
  }

  // 过滤权限点显示
  const getFilteredchildren = (child: AuthorityChild, parent: AuthorityGroup) => {
    if (!searchValue.value.trim()) {
      return child.children
    }

    const keyword = searchValue.value.toLowerCase().trim()
    const parentMatches = parent.menuName?.toLowerCase().includes(keyword)
    const childMatches = child.menuName?.toLowerCase().includes(keyword)

    // 如果父级或子级匹配，显示所有权限点
    if (parentMatches || childMatches) {
      return child.children
    }

    // 否则只显示匹配的权限点
    return child.children.filter((authority) => {
      return (authority.name?.toLowerCase().includes(keyword))
        || (authority.remark?.toLowerCase().includes(keyword))
    })
  }

  // 高亮搜索关键字
  const highlightText = (text: string | undefined, keyword: string) => {
    if (!text || !keyword.trim())
      return text || ''

    const regex = new RegExp(`(${keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
    return text.replace(regex, '<mark style="background-color: #ffd700; padding: 0 2px;">$1</mark>')
  }

  // 展开状态检查函数
  const isParentExpanded = (id: string) => expandedParentIds.value.includes(id)
  const isChildExpanded = (id: string) => expandedChildIds.value.includes(id)
  const isAuthorityExpanded = (id: string) => expandedAuthorityIds.value.includes(id)

  // 切换展开状态
  const toggleParentExpand = (id: string) => {
    const index = expandedParentIds.value.indexOf(id)
    if (index === -1) {
      expandedParentIds.value.push(id)
    }
    else {
      expandedParentIds.value.splice(index, 1)
    }
  }

  const toggleChildExpand = (id: string) => {
    const index = expandedChildIds.value.indexOf(id)
    if (index === -1) {
      expandedChildIds.value.push(id)
    }
    else {
      expandedChildIds.value.splice(index, 1)
    }
  }

  const toggleAuthorityExpand = (id: string) => {
    const index = expandedAuthorityIds.value.indexOf(id)
    if (index === -1) {
      expandedAuthorityIds.value.push(id)
    }
    else {
      expandedAuthorityIds.value.splice(index, 1)
    }
  }

  // 检查是否有父级已展开
  const checkAllParentsExpanded = () => {
    const hasExpandedParent = filteredBoxList.value.some(item =>
      expandedParentIds.value.includes(item.id),
    )
    isAllExpanded.value = hasExpandedParent
  }

  // 一键展开/收起
  const toggleAllExpand = () => {
    if (isAllExpanded.value) {
      // 收起全部
      expandedParentIds.value = []
      expandedChildIds.value = []
      expandedAuthorityIds.value = []
      isAllExpanded.value = false
    }
    else {
      // 展开全部
      expandedParentIds.value = []
      expandedChildIds.value = []
      expandedAuthorityIds.value = []
      filteredBoxList.value.forEach((item) => {
        expandedParentIds.value.push(item.id)
        getFilteredChildren(item).forEach((child) => {
          expandedChildIds.value.push(child.id)
          getFilteredchildren(child, item).forEach((authority) => {
            expandedAuthorityIds.value.push(authority.id.toString())
          })
        })
      })
      isAllExpanded.value = true
    }
  }

  // 展开父级下的子级，不展开第三级
  const expandAllChildren = (parentId: string) => {
    const parent = filteredBoxList.value.find(item => item.id === parentId)
    if (!parent)
      return

    // 确保父级展开
    if (!expandedParentIds.value.includes(parentId)) {
      expandedParentIds.value.push(parentId)
    }

    // 检查并更新全局展开状态
    checkAllParentsExpanded()
  }

  // 收起父级下所有子级
  const collapseAllChildren = (parentId: string) => {
    const parent = filteredBoxList.value.find(item => item.id === parentId)
    if (parent) {
      // 从展开列表中移除父级
      const parentIndex = expandedParentIds.value.indexOf(parentId)
      if (parentIndex !== -1) {
        expandedParentIds.value.splice(parentIndex, 1)
      }

      // 收起所有子级
      getFilteredChildren(parent).forEach((child) => {
        const childIndex = expandedChildIds.value.indexOf(child.id)
        if (childIndex !== -1) {
          expandedChildIds.value.splice(childIndex, 1)
        }

        // 收起所有权限点
        getFilteredchildren(child, parent).forEach((authority) => {
          const authorityIndex = expandedAuthorityIds.value.indexOf(authority.id.toString())
          if (authorityIndex !== -1) {
            expandedAuthorityIds.value.splice(authorityIndex, 1)
          }
        })
      })

      // 检查并更新全局展开状态
      checkAllParentsExpanded()
    }
  }

  // 权限选择逻辑
  // 更新子级状态
  const updateChildStatus = (child: AuthorityChild) => {
    const checkedCount = child.children.filter(item => item.checked).length
    child.checked = checkedCount === child.children.length
    child.indeterminate = checkedCount > 0 && checkedCount < child.children.length
  }

  // 更新父级状态
  const updateParentStatus = (parent: AuthorityGroup) => {
    const checkedChildCount = parent.children.filter(item => item.checked).length
    const indeterminateChildCount = parent.children.filter(item => item.indeterminate).length

    parent.checked = checkedChildCount === parent.children.length
    parent.indeterminate = checkedChildCount > 0 && checkedChildCount < parent.children.length || indeterminateChildCount > 0
  }

  // 父级选择处理
  const handleParentChange = (item: AuthorityGroup) => {
    item.indeterminate = false

    if (item.checked) {
      // 如果是选中操作，需要处理每个子级的互斥情况
      item.children.forEach((child) => {
        child.checked = true
        child.indeterminate = false

        // 处理每个子级的互斥情况
        const groupCodeMap = new Map<string, Authority[]>()

        // 先按groupCode分组
        child.children.forEach((authority) => {
          if (authority.groupCode) {
            if (!groupCodeMap.has(authority.groupCode)) {
              groupCodeMap.set(authority.groupCode, [])
            }
            groupCodeMap.get(authority.groupCode)!.push(authority)
          }
        })

        // 对于每个groupCode组，只选中第一个
        groupCodeMap.forEach((children) => {
          if (children.length > 1) {
            children[0].checked = true
            for (let i = 1; i < children.length; i++) {
              children[i].checked = false
            }
          }
          else if (children.length === 1) {
            children[0].checked = true
          }
        })

        // 没有groupCode的authority全部选中
        child.children.forEach((authority) => {
          if (!authority.groupCode) {
            authority.checked = true
          }
        })

        // 更新子级状态
        updateChildStatus(child)
      })
    }
    else {
      // 如果是取消选中，则所有子项都取消选中
      item.children.forEach((child) => {
        child.checked = false
        child.indeterminate = false
        child.children.forEach((authority) => {
          authority.checked = false
        })
      })
    }
  }

  // 子级选择处理
  const handleChildChange = (child: AuthorityChild, parent: AuthorityGroup) => {
    // 如果是选中操作
    if (child.checked) {
      // 先处理互斥的情况：找出所有有groupCode的authority，按groupCode分组
      const groupCodeMap = new Map<string, Authority[]>()
      child.children.forEach((authority) => {
        if (authority.groupCode) {
          if (!groupCodeMap.has(authority.groupCode)) {
            groupCodeMap.set(authority.groupCode, [])
          }
          groupCodeMap.get(authority.groupCode)!.push(authority)
        }
      })

      // 对于每个groupCode组，只选中第一个
      groupCodeMap.forEach((children) => {
        if (children.length > 1) {
          // 选中第一个，其余取消选中
          children[0].checked = true
          for (let i = 1; i < children.length; i++) {
            children[i].checked = false
          }
        }
      })

      // 没有groupCode的authority全部选中
      child.children.forEach((authority) => {
        if (!authority.groupCode) {
          authority.checked = true
        }
      })
    }
    else {
      // 如果是取消选中，则所有子项都取消选中
      child.children.forEach((authority) => {
        authority.checked = false
      })
    }

    // 更新子级状态
    updateChildStatus(child)

    // 更新父级状态
    updateParentStatus(parent)
  }

  // 权限点选择处理
  const handleAuthorityChange = (authority: Authority, child: AuthorityChild, parent: AuthorityGroup) => {
    // 如果有相同的 groupCode，则互斥选择
    if (authority.groupCode && authority.checked) {
      child.children.forEach((item) => {
        if (item.groupCode === authority.groupCode && item.id !== authority.id) {
          item.checked = false
        }
      })
    }

    // 更新子级状态
    updateChildStatus(child)

    // 更新父级状态
    updateParentStatus(parent)
  }

  // 清空已选
  const clearAllSelected = () => {
    boxList.value.forEach((item) => {
      item.checked = false
      item.indeterminate = false
      item.children.forEach((child) => {
        child.checked = false
        child.indeterminate = false
        child.children.forEach((authority) => {
          authority.checked = false
        })
      })
    })
  }

  // 根据权限ID数组设置默认选中状态
  const setDefaultCheckedByIds = (menuIds: number[]) => {
    if (!menuIds || menuIds.length === 0)
      return

    // 先清空所有选择状态
    clearAllSelected()

    // 遍历所有权限数据，找到匹配的ID并设置为选中
    boxList.value.forEach((parent) => {
      parent.children.forEach((child) => {
        // 先收集所有在menuIds中的权限项
        const authoritiesToCheck = child.children.filter(authority => 
          menuIds.includes(authority.id)
        )

        // 按groupCode分组
        const groupCodeMap = new Map<string, Authority[]>()
        const ungroupedAuthorities: Authority[] = []

        authoritiesToCheck.forEach((authority) => {
          if (authority.groupCode) {
            if (!groupCodeMap.has(authority.groupCode)) {
              groupCodeMap.set(authority.groupCode, [])
            }
            groupCodeMap.get(authority.groupCode)!.push(authority)
          } else {
            ungroupedAuthorities.push(authority)
          }
        })

        // 先清空所有选中状态
        child.children.forEach((authority) => {
          authority.checked = false
        })

        // 对于有groupCode的组，只选中权重最高的
        groupCodeMap.forEach((authorities) => {
          if (authorities.length > 1) {
            // 按权重降序排序，选中权重最高的
            const sortedByWeight = authorities.sort((a, b) => (b.weight || 0) - (a.weight || 0))
            sortedByWeight[0].checked = true
          } else if (authorities.length === 1) {
            authorities[0].checked = true
          }
        })

        // 对于没有groupCode的权限，直接选中
        ungroupedAuthorities.forEach((authority) => {
          authority.checked = true
        })

        // 更新子级状态
        updateChildStatus(child)
      })

      // 更新父级状态
      updateParentStatus(parent)
    })
    loading.value = false
  }

  // 重置所有状态
  const resetAllStates = () => {
    // 重置表单状态
    // delete formState.roleId
    formState.roleName = ''
    formState.description = ''
    formState.menuIds = [] // 重置时保持默认值

    // 重置搜索值
    searchValue.value = ''

    // 重置展开状态
    expandedParentIds.value = []
    expandedChildIds.value = []
    expandedAuthorityIds.value = []
    isAllExpanded.value = false

    // 重置所有选择状态
    clearAllSelected()

    // 重新设置默认选中
    setDefaultCheckedByIds(formState.menuIds)
  }

  // 更新数据
  const updateData = (newData: AuthorityGroup[]) => {
    boxList.value = newData
    loading.value = false

    // 数据加载完成后，设置默认选中状态
    setDefaultCheckedByIds(formState.menuIds)
  }

  // 搜索时自动展开匹配的内容
  watch(searchValue, (newValue) => {
    if (newValue.trim()) {
      // 当有搜索内容时，展开所有匹配的父级和子级
      filteredBoxList.value.forEach((parent) => {
        if (!expandedParentIds.value.includes(parent.id)) {
          expandedParentIds.value.push(parent.id)
        }

        getFilteredChildren(parent).forEach((child) => {
          if (!expandedChildIds.value.includes(child.id)) {
            expandedChildIds.value.push(child.id)
          }
        })
      })
      // 更新一键展开状态
      checkAllParentsExpanded()
    }
  })

  return {
    // 数据
    boxList,
    formState,
    searchValue,
    filteredBoxList,

    // 展开状态
    expandedParentIds,
    expandedChildIds,
    expandedAuthorityIds,
    isAllExpanded,

    // 展开相关方法
    isParentExpanded,
    isChildExpanded,
    isAuthorityExpanded,
    toggleParentExpand,
    toggleChildExpand,
    toggleAuthorityExpand,
    toggleAllExpand,
    expandAllChildren,
    collapseAllChildren,

    // 过滤相关方法
    getFilteredChildren,
    getFilteredchildren,
    highlightText,

    // 权限选择相关方法
    handleParentChange,
    handleChildChange,
    handleAuthorityChange,
    clearAllSelected,
    resetAllStates,
    updateData,
    setDefaultCheckedByIds,
  }
}
