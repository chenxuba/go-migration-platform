package repository

import "testing"

func TestSanitizeInstPeriodPayloadBoundTeachers(t *testing.T) {
	payload := &instPeriodFilePayload{
		Version: 1,
		Groups: []instPeriodGroupJSON{
			{
				ID: "group-a",
				BoundTeachers: []instPeriodBoundTeacherJSON{
					{ID: "30000212", Name: "历史脏数据"},
					{ID: "11", Name: "旧名字"},
					{ID: "11", Name: "重复老师"},
					{ID: "12", Name: "离职旧名字"},
					{ID: "bad", Name: "非法老师"},
				},
			},
		},
	}

	sanitizeInstPeriodPayloadBoundTeachers(payload, map[int64]instPeriodBoundTeacherMeta{
		11: {Name: "张老师"},
		12: {Name: "李老师", Disabled: true},
	})

	got := payload.Groups[0].BoundTeachers
	if len(got) != 2 {
		t.Fatalf("expected only valid bound teachers to remain, got %#v", got)
	}
	if got[0].ID != "11" || got[0].Name != "张老师" {
		t.Fatalf("expected first teacher to be refreshed from inst_user, got %#v", got[0])
	}
	if got[1].ID != "12" || got[1].Name != "李老师（离职）" {
		t.Fatalf("expected disabled teacher to keep current disabled marker, got %#v", got[1])
	}
}
