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

func TestBuildVisibleInstitutionModuleTreeUsesDirectRoleChildren(t *testing.T) {
	selected := map[int64]struct{}{
		398: {},
		532: {},
		584: {},
		600: {},
	}

	tree := buildVisibleInstitutionModuleTree([]rawMenu{
		{ID: 398, PID: 0, Name: "内部管理", Code: "grp:intl", Sort: 10, Weight: 10},
		{ID: 532, PID: 398, Name: "角色管理", Code: "page:intlRole", Sort: 20, Weight: 10},
		{ID: 584, PID: 532, Name: "页面功能访问", Code: "perm:intlRoleUse", Sort: 5, Weight: 0},
		{ID: 600, PID: 532, Name: "角色管理", Code: "perm:orgMngRoleMng", Sort: 10, Weight: 0},
	}, selected)

	if len(tree) != 1 || len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 role page node, got %+v", tree)
	}
	if len(tree[0].Children[0].Children) != 2 {
		t.Fatalf("expected role page to expose 2 direct children, got %d", len(tree[0].Children[0].Children))
	}
	if tree[0].Children[0].Children[0].MenuID != "584" || tree[0].Children[0].Children[1].MenuID != "600" {
		t.Fatalf("expected role page to contain page-use and role-manage child, got %+v", tree[0].Children[0].Children)
	}
}
