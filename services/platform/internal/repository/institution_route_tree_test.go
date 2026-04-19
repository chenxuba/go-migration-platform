package repository

import "testing"

func TestBuildVisibleInstitutionModuleTreeIncludesPageUseChild(t *testing.T) {
	selected := map[int64]struct{}{
		1: {},
		2: {},
		3: {},
	}

	tree := buildVisibleInstitutionModuleTree([]rawMenu{
		{ID: 1, PID: 0, Name: "招生中心", Code: "grp:enr", Sort: 30, Weight: 30},
		{ID: 2, PID: 1, Name: "招生自测", Code: "page:enrSelfTst", Sort: 10, Weight: 10},
		{ID: 3, PID: 2, Name: "页面功能访问", Code: "perm:enrSelfTstUse", Sort: 5, Weight: 0},
	}, selected)

	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 page node, got %d", len(tree[0].Children))
	}
	if len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("expected page use child to be visible, got %d children", len(tree[0].Children[0].Children))
	}
	if tree[0].Children[0].Children[0].MenuName != "页面功能访问" {
		t.Fatalf("expected page use child, got %q", tree[0].Children[0].Children[0].MenuName)
	}
}

func TestBuildVisibleInstitutionModuleTreeDeduplicatesAggregatedLeaf(t *testing.T) {
	selected := map[int64]struct{}{
		398: {},
		297: {},
		532: {},
		464: {},
		583: {},
		584: {},
	}

	tree := buildVisibleInstitutionModuleTree([]rawMenu{
		{ID: 398, PID: 0, Name: "内部管理", Code: "grp:intl", Sort: 10, Weight: 10},
		{ID: 297, PID: 398, Name: "员工管理", Code: "page:intlStf", Sort: 10, Weight: 10},
		{ID: 532, PID: 398, Name: "角色管理", Code: "page:intlRole", Sort: 20, Weight: 10},
		{ID: 464, PID: 297, Name: "角色管理", Code: "perm:orgMngRoleMng", Sort: 50, Weight: 0},
		{ID: 583, PID: 297, Name: "页面功能访问", Code: "perm:intlStfUse", Sort: 5, Weight: 0},
		{ID: 584, PID: 532, Name: "页面功能访问", Code: "perm:intlRoleUse", Sort: 5, Weight: 0},
	}, selected)

	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("expected 2 page nodes, got %d", len(tree[0].Children))
	}

	employeeChildren := tree[0].Children[0].Children
	if len(employeeChildren) != 1 {
		t.Fatalf("expected employee page to keep only 1 direct child, got %d", len(employeeChildren))
	}
	if employeeChildren[0].MenuID != "583" {
		t.Fatalf("expected employee page to keep only its page-use child, got %s", employeeChildren[0].MenuID)
	}

	roleChildren := tree[0].Children[1].Children
	if len(roleChildren) != 2 {
		t.Fatalf("expected role page to expose 2 permissions, got %d", len(roleChildren))
	}
	if roleChildren[0].MenuID != "584" || roleChildren[1].MenuID != "464" {
		t.Fatalf("expected role page to include page-use and aggregated role-manage permission, got %+v", roleChildren)
	}
}
