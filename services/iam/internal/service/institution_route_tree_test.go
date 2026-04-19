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

func TestBuildVisibleInstitutionMenuTreeDeduplicatesAggregatedLeaf(t *testing.T) {
	sort10 := 10
	sort20 := 20
	sort50 := 50
	sort5 := 5
	weight10 := 10
	weight0 := 0

	tree := buildVisibleInstitutionMenuTree([]model.Menu{
		{ID: 398, PID: 0, MenuName: "内部管理", MenuCode: "grp:intl", Sort: &sort10, Weight: &weight10},
		{ID: 297, PID: 398, MenuName: "员工管理", MenuCode: "page:intlStf", Sort: &sort10, Weight: &weight10},
		{ID: 532, PID: 398, MenuName: "角色管理", MenuCode: "page:intlRole", Sort: &sort20, Weight: &weight10},
		{ID: 464, PID: 297, MenuName: "角色管理", MenuCode: "perm:orgMngRoleMng", Sort: &sort50, Weight: &weight0},
		{ID: 583, PID: 297, MenuName: "页面功能访问", MenuCode: "perm:intlStfUse", Sort: &sort5, Weight: &weight0},
		{ID: 584, PID: 532, MenuName: "页面功能访问", MenuCode: "perm:intlRoleUse", Sort: &sort5, Weight: &weight0},
	})

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
	if employeeChildren[0].ID != 583 {
		t.Fatalf("expected employee page to keep only its page-use child, got %d", employeeChildren[0].ID)
	}

	roleChildren := tree[0].Children[1].Children
	if len(roleChildren) != 2 {
		t.Fatalf("expected role page to expose 2 permissions, got %d", len(roleChildren))
	}
	if roleChildren[0].ID != 584 || roleChildren[1].ID != 464 {
		t.Fatalf("expected role page to include page-use and aggregated role-manage permission, got %+v", roleChildren)
	}
}
