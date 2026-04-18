package service

import (
	"testing"

	"go-migration-platform/services/iam/internal/model"
)

func TestBuildVisibleInstitutionMenuTreeIncludesPageUseChild(t *testing.T) {
	sort10 := 10
	sort5 := 5
	weight10 := 10
	weight0 := 0

	tree := buildVisibleInstitutionMenuTree([]model.Menu{
		{ID: 1, PID: 0, MenuName: "招生中心", MenuCode: "grp:enr", Sort: &sort10, Weight: &weight10},
		{ID: 2, PID: 1, MenuName: "招生自测", MenuCode: "page:enrSelfTst", Sort: &sort10, Weight: &weight10},
		{ID: 3, PID: 2, MenuName: "页面功能访问", MenuCode: "perm:enrSelfTstUse", Sort: &sort5, Weight: &weight0},
	})

	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 page node, got %d", len(tree[0].Children))
	}
	if len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("expected page use child to be visible, got %d children", len(tree[0].Children[0].Children))
	}
	if tree[0].Children[0].Children[0].MenuCode != "perm:enrSelfTstUse" {
		t.Fatalf("expected page use code, got %q", tree[0].Children[0].Children[0].MenuCode)
	}
}
