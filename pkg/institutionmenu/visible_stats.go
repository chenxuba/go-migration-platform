package institutionmenu

import (
	"sort"
	"strings"
)

type VisibleStatMenu struct {
	ID       int64
	PID      int64
	MenuCode string
	MenuName string
	MenuType int
	Sort     int
	Weight   int
}

// CountVisibleLeafAuthorities returns the institution-side visible leaf permission counts
// using the same catalog aggregation rules as the permission tree UI.
func CountVisibleLeafAuthorities(menus []VisibleStatMenu, selectedMenuIDs map[int64]struct{}) (int, int) {
	if len(menus) == 0 || len(selectedMenuIDs) == 0 {
		return 0, 0
	}

	byID := make(map[int64]VisibleStatMenu, len(menus))
	childrenByPID := make(map[int64][]VisibleStatMenu)
	codeIndex := make(map[string][]VisibleStatMenu)
	nameIndex := make(map[string][]VisibleStatMenu)
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

	functionCount := 0
	dataCount := 0
	counted := make(map[int64]struct{}, len(selectedMenuIDs))

	countLeaf := func(item VisibleStatMenu) {
		if _, exists := selectedMenuIDs[item.ID]; !exists {
			return
		}
		if _, exists := counted[item.ID]; exists {
			return
		}
		counted[item.ID] = struct{}{}
		if item.MenuType == 1 {
			dataCount++
			return
		}
		functionCount++
	}

	for _, group := range VisibleRouteCatalog {
		groupMenu, ok := matchVisibleStatMenuNode(0, group.Code, append([]string{group.Name}, group.MatchNames...), codeIndex, nameIndex)
		if !ok {
			continue
		}

		if group.UseAsLeaf {
			countLeaf(groupMenu)
			continue
		}

		for _, child := range group.Children {
			childMenu, ok := matchVisibleStatMenuNode(groupMenu.ID, child.Code, append([]string{child.Name}, child.MatchNames...), codeIndex, nameIndex)
			if !ok {
				continue
			}

			leafMap := make(map[int64]VisibleStatMenu)
			if child.UseDirectChildren {
				appendVisibleStatDirectChildren(leafMap, childrenByPID[childMenu.ID])
			}
			appendVisibleStatPageUseChildren(leafMap, childrenByPID[childMenu.ID])

			for _, sourceCode := range child.AggregateNodeCodes {
				sourceMenu, matched := matchVisibleStatMenuNode(groupMenu.ID, sourceCode, nil, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendVisibleStatDirectChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, sourceName := range child.AggregateNodeNames {
				sourceMenu, matched := matchVisibleStatMenuNode(groupMenu.ID, "", []string{sourceName}, codeIndex, nameIndex)
				if !matched {
					continue
				}
				appendVisibleStatDirectChildren(leafMap, childrenByPID[sourceMenu.ID])
			}

			for _, leafName := range child.AggregateLeafNames {
				appendVisibleStatAggregateLeavesByName(leafMap, nameIndex, childrenByPID, byID, groupMenu.ID, childMenu.ID, leafName)
			}

			leafItems := make([]VisibleStatMenu, 0, len(leafMap))
			for _, item := range leafMap {
				leafItems = append(leafItems, item)
			}
			sortVisibleStatLeafMenus(leafItems)

			for _, leaf := range leafItems {
				countLeaf(leaf)
			}
		}
	}

	return functionCount, dataCount
}

func matchVisibleStatMenuNode(parentID int64, code string, names []string, codeIndex map[string][]VisibleStatMenu, nameIndex map[string][]VisibleStatMenu) (VisibleStatMenu, bool) {
	code = strings.TrimSpace(code)
	if code != "" {
		normalized := NormalizeCode(code)
		candidates := codeIndex[code]
		if normalized != code {
			candidates = append(candidates, codeIndex[normalized]...)
		}
		for _, item := range candidates {
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

	return VisibleStatMenu{}, false
}

func appendVisibleStatDirectChildren(target map[int64]VisibleStatMenu, items []VisibleStatMenu) {
	for _, item := range items {
		target[item.ID] = item
	}
}

func appendVisibleStatPageUseChildren(target map[int64]VisibleStatMenu, items []VisibleStatMenu) {
	for _, item := range items {
		code := NormalizeCode(item.MenuCode)
		if strings.HasPrefix(code, "perm:") && strings.HasSuffix(code, "Use") {
			target[item.ID] = item
		}
	}
}

func appendVisibleStatAggregateLeavesByName(target map[int64]VisibleStatMenu, nameIndex map[string][]VisibleStatMenu, childrenByPID map[int64][]VisibleStatMenu, byID map[int64]VisibleStatMenu, ancestorID, excludeID int64, leafName string) {
	for _, item := range nameIndex[strings.TrimSpace(leafName)] {
		if item.ID == excludeID || item.PID == 0 {
			continue
		}
		if len(childrenByPID[item.ID]) > 0 {
			continue
		}
		if !isVisibleStatMenuDescendantOf(item.ID, ancestorID, byID) {
			continue
		}
		target[item.ID] = item
	}
}

func isVisibleStatMenuDescendantOf(id, ancestorID int64, byID map[int64]VisibleStatMenu) bool {
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

func sortVisibleStatLeafMenus(items []VisibleStatMenu) {
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
