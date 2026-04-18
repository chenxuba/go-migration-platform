package repository

import (
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/services/platform/internal/model"
)

func buildVisibleInstitutionModuleTree(items []rawMenu, selected map[int64]struct{}) []model.ModuleMenu {
	if len(items) == 0 {
		return nil
	}

	byID := make(map[int64]rawMenu, len(items))
	childrenByPID := make(map[int64][]rawMenu, len(items))
	codeIndex := make(map[string][]rawMenu)
	nameIndex := make(map[string][]rawMenu)
	for _, item := range items {
		byID[item.ID] = item
		childrenByPID[item.PID] = append(childrenByPID[item.PID], item)

		code := strings.TrimSpace(item.Code)
		if code != "" {
			codeIndex[code] = append(codeIndex[code], item)
		}
		name := strings.TrimSpace(item.Name)
		if name != "" {
			nameIndex[name] = append(nameIndex[name], item)
		}
	}

	result := make([]model.ModuleMenu, 0, len(institutionmenu.VisibleRouteCatalog))
	for _, group := range institutionmenu.VisibleRouteCatalog {
		groupMenu, ok := matchVisibleRawMenu(0, group.Code, append([]string{group.Name}, group.MatchNames...), codeIndex, nameIndex)
		if !ok {
			continue
		}

		groupNode := model.ModuleMenu{
			MenuID:    formatMenuID(groupMenu.ID),
			MenuName:  groupMenu.Name,
			Introduce: groupMenu.Introduce,
			MenuType:  groupMenu.MenuType,
			GroupCode: groupMenu.GroupCode,
			Weight:    groupMenu.Weight,
			Children:  []model.ModuleMenu{},
		}

		if group.UseAsLeaf {
			_, groupNode.IsSelect = selected[groupMenu.ID]
			result = append(result, groupNode)
			continue
		}

		for _, child := range group.Children {
			childMenu, ok := matchVisibleRawMenu(groupMenu.ID, child.Code, append([]string{child.Name}, child.MatchNames...), codeIndex, nameIndex)
			if !ok {
				continue
			}

			childNode := model.ModuleMenu{
				MenuID:    formatMenuID(childMenu.ID),
				MenuName:  childMenu.Name,
				Introduce: childMenu.Introduce,
				MenuType:  childMenu.MenuType,
				GroupCode: childMenu.GroupCode,
				Weight:    childMenu.Weight,
				Children:  []model.ModuleMenu{},
			}

			leafMap := make(map[int64]rawMenu)
			if child.UseDirectChildren {
				appendDirectRawChildren(leafMap, childrenByPID[childMenu.ID])
			}

			for _, sourceCode := range child.AggregateNodeCodes {
				sourceMenu, matched := matchVisibleRawMenu(groupMenu.ID, sourceCode, nil, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendDirectRawChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, sourceName := range child.AggregateNodeNames {
				sourceMenu, matched := matchVisibleRawMenu(groupMenu.ID, "", []string{sourceName}, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendDirectRawChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, leafName := range child.AggregateLeafNames {
				appendAggregateRawLeavesByName(leafMap, nameIndex, childrenByPID, byID, groupMenu.ID, childMenu.ID, leafName)
			}

			leafItems := make([]rawMenu, 0, len(leafMap))
			for _, item := range leafMap {
				leafItems = append(leafItems, item)
			}
			sortVisibleRawLeaves(leafItems)

			for _, leaf := range leafItems {
				leafNode := model.ModuleMenu{
					MenuID:    formatMenuID(leaf.ID),
					MenuName:  leaf.Name,
					Introduce: leaf.Introduce,
					MenuType:  leaf.MenuType,
					GroupCode: leaf.GroupCode,
					Weight:    leaf.Weight,
				}
				_, leafNode.IsSelect = selected[leaf.ID]
				childNode.Children = append(childNode.Children, leafNode)
			}

			if len(childNode.Children) == 0 {
				_, childNode.IsSelect = selected[childMenu.ID]
			} else {
				allSelected := true
				for _, leaf := range childNode.Children {
					if !leaf.IsSelect {
						allSelected = false
						break
					}
				}
				childNode.IsSelect = allSelected
			}

			groupNode.Children = append(groupNode.Children, childNode)
		}

		allSelected := len(groupNode.Children) > 0
		for _, childNode := range groupNode.Children {
			if !childNode.IsSelect {
				allSelected = false
				break
			}
		}
		groupNode.IsSelect = allSelected
		result = append(result, groupNode)
	}

	return result
}

func matchVisibleRawMenu(parentID int64, code string, names []string, codeIndex map[string][]rawMenu, nameIndex map[string][]rawMenu) (rawMenu, bool) {
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

	return rawMenu{}, false
}

func appendAggregateRawLeavesByName(target map[int64]rawMenu, nameIndex map[string][]rawMenu, childrenByPID map[int64][]rawMenu, byID map[int64]rawMenu, ancestorID, excludeID int64, leafName string) {
	for _, item := range nameIndex[strings.TrimSpace(leafName)] {
		if item.ID == excludeID || item.PID == 0 {
			continue
		}
		if len(childrenByPID[item.ID]) > 0 {
			continue
		}
		if !isRawMenuDescendantOf(item.ID, ancestorID, byID) {
			continue
		}
		target[item.ID] = item
	}
}

func isRawMenuDescendantOf(id, ancestorID int64, byID map[int64]rawMenu) bool {
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

func appendDirectRawChildren(target map[int64]rawMenu, items []rawMenu) {
	for _, item := range items {
		target[item.ID] = item
	}
}

func sortVisibleRawLeaves(items []rawMenu) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			if items[i].Weight == items[j].Weight {
				return items[i].ID < items[j].ID
			}
			return items[i].Weight > items[j].Weight
		}
		return items[i].Sort < items[j].Sort
	})
}

func formatMenuID(id int64) string {
	return strconv.FormatInt(id, 10)
}
