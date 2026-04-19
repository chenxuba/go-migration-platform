package service

import (
	"sort"
	"strings"

	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/services/iam/internal/model"
)

func buildVisibleInstitutionMenuTree(menus []model.Menu) []model.MenuTreeNode {
	if len(menus) == 0 {
		return nil
	}

	byID := make(map[int64]model.Menu, len(menus))
	childrenByPID := make(map[int64][]model.Menu, len(menus))
	codeIndex := make(map[string][]model.Menu)
	nameIndex := make(map[string][]model.Menu)
	for _, item := range menus {
		byID[item.ID] = item
		childrenByPID[item.PID] = append(childrenByPID[item.PID], item)

		code := strings.TrimSpace(item.MenuCode)
		if code != "" {
			codeIndex[code] = append(codeIndex[code], item)
		}
		name := strings.TrimSpace(item.MenuName)
		if name != "" {
			nameIndex[name] = append(nameIndex[name], item)
		}
	}

	result := make([]model.MenuTreeNode, 0, len(institutionmenu.VisibleRouteCatalog))
	for _, group := range institutionmenu.VisibleRouteCatalog {
		groupMenu, ok := matchVisibleMenuNode(0, group.Code, append([]string{group.Name}, group.MatchNames...), codeIndex, nameIndex)
		if !ok {
			continue
		}

		groupNode := model.MenuTreeNode{
			Menu:     groupMenu,
			Children: []model.MenuTreeNode{},
		}

		if group.UseAsLeaf {
			result = append(result, groupNode)
			continue
		}

		for _, child := range group.Children {
			childMenu, ok := matchVisibleMenuNode(groupMenu.ID, child.Code, append([]string{child.Name}, child.MatchNames...), codeIndex, nameIndex)
			if !ok {
				continue
			}

			childNode := model.MenuTreeNode{
				Menu:     childMenu,
				Children: []model.MenuTreeNode{},
			}

			leafMap := make(map[int64]model.Menu)
			if child.UseDirectChildren {
				appendDirectChildren(leafMap, childrenByPID[childMenu.ID])
			}
			appendPageUseChildren(leafMap, childrenByPID[childMenu.ID])

			for _, sourceCode := range child.AggregateNodeCodes {
				sourceMenu, matched := matchVisibleMenuNode(groupMenu.ID, sourceCode, nil, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendDirectChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, sourceName := range child.AggregateNodeNames {
				sourceMenu, matched := matchVisibleMenuNode(groupMenu.ID, "", []string{sourceName}, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendDirectChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, leafName := range child.AggregateLeafNames {
				appendAggregateLeavesByName(leafMap, nameIndex, childrenByPID, byID, groupMenu.ID, childMenu.ID, leafName)
			}
			removeExcludedLeaves(leafMap, child.ExcludeLeafCodes)

			leafItems := make([]model.Menu, 0, len(leafMap))
			for _, item := range leafMap {
				leafItems = append(leafItems, item)
			}
			sortVisibleLeafMenus(leafItems)

			for _, leaf := range leafItems {
				childNode.Children = append(childNode.Children, model.MenuTreeNode{
					Menu:     leaf,
					Children: []model.MenuTreeNode{},
				})
			}

			groupNode.Children = append(groupNode.Children, childNode)
		}

		result = append(result, groupNode)
	}

	return result
}

func matchVisibleMenuNode(parentID int64, code string, names []string, codeIndex map[string][]model.Menu, nameIndex map[string][]model.Menu) (model.Menu, bool) {
	code = strings.TrimSpace(code)
	if code != "" {
		for _, item := range codeIndex[code] {
			if (parentID == 0 && item.PID == 0) || (parentID != 0 && item.PID == parentID) {
				return item, true
			}
		}
	}

	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		for _, item := range nameIndex[trimmed] {
			if (parentID == 0 && item.PID == 0) || (parentID != 0 && item.PID == parentID) {
				return item, true
			}
		}
	}

	return model.Menu{}, false
}

func appendAggregateLeavesByName(target map[int64]model.Menu, nameIndex map[string][]model.Menu, childrenByPID map[int64][]model.Menu, byID map[int64]model.Menu, ancestorID, excludeID int64, leafName string) {
	for _, item := range nameIndex[strings.TrimSpace(leafName)] {
		if item.ID == excludeID || item.PID == 0 {
			continue
		}
		if len(childrenByPID[item.ID]) > 0 {
			continue
		}
		if !isMenuDescendantOf(item.ID, ancestorID, byID) {
			continue
		}
		target[item.ID] = item
	}
}

func isMenuDescendantOf(id, ancestorID int64, byID map[int64]model.Menu) bool {
	currentID := id
	for currentID > 0 {
		item, ok := byID[currentID]
		if !ok {
			return false
		}
		if item.PID == ancestorID {
			return true
		}
		currentID = item.PID
	}
	return false
}

func appendDirectChildren(target map[int64]model.Menu, items []model.Menu) {
	for _, item := range items {
		target[item.ID] = item
	}
}

func appendPageUseChildren(target map[int64]model.Menu, items []model.Menu) {
	for _, item := range items {
		code := institutionmenu.NormalizeCode(item.MenuCode)
		if strings.HasPrefix(code, "perm:") && strings.HasSuffix(code, "Use") {
			target[item.ID] = item
		}
	}
}

func removeExcludedLeaves(target map[int64]model.Menu, codes []string) {
	if len(target) == 0 || len(codes) == 0 {
		return
	}

	excluded := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		normalized := institutionmenu.NormalizeCode(code)
		if normalized == "" {
			continue
		}
		excluded[normalized] = struct{}{}
	}
	if len(excluded) == 0 {
		return
	}

	for id, item := range target {
		if _, ok := excluded[institutionmenu.NormalizeCode(item.MenuCode)]; ok {
			delete(target, id)
		}
	}
}

func sortVisibleLeafMenus(items []model.Menu) {
	sort.SliceStable(items, func(i, j int) bool {
		leftSort := 0
		if items[i].Sort != nil {
			leftSort = *items[i].Sort
		}
		rightSort := 0
		if items[j].Sort != nil {
			rightSort = *items[j].Sort
		}
		if leftSort == rightSort {
			leftWeight := 0
			if items[i].Weight != nil {
				leftWeight = *items[i].Weight
			}
			rightWeight := 0
			if items[j].Weight != nil {
				rightWeight = *items[j].Weight
			}
			if leftWeight == rightWeight {
				return items[i].ID < items[j].ID
			}
			return leftWeight > rightWeight
		}
		return leftSort < rightSort
	})
}
