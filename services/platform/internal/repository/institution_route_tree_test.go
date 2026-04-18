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
