package customization

import (
	"encoding/json"
	"os"
)

type TenantProfile struct {
	TenantID       string   `json:"tenantId"`
	Name           string   `json:"name"`
	Edition        string   `json:"edition"`
	Features       []string `json:"features"`
	CustomFields   []string `json:"customFields"`
	WorkflowScheme string   `json:"workflowScheme"`
	RulePack       string   `json:"rulePack"`
	Integrations   []string `json:"integrations"`
}

type Store struct {
	profiles map[string]TenantProfile
}

func NewStore(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultStore(), nil
	}

	var items []TenantProfile
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	store := &Store{profiles: make(map[string]TenantProfile, len(items))}
	for _, item := range items {
		store.profiles[item.TenantID] = item
	}

	return store, nil
}

func (store *Store) Get(tenantID string) TenantProfile {
	if profile, ok := store.profiles[tenantID]; ok {
		return profile
	}
	return store.profiles["tenant-a"]
}

func defaultStore() *Store {
	return &Store{
		profiles: map[string]TenantProfile{
			"tenant-a": {
				TenantID:       "tenant-a",
				Name:           "A客户",
				Edition:        "enterprise",
				Features:       []string{"tenant-branding", "approval-flow"},
				CustomFields:   []string{"student.guardian_job"},
				WorkflowScheme: "government-approval-v1",
				RulePack:       "subsidy-rules-a",
				Integrations:   []string{"wechat", "qiniu"},
			},
		},
	}
}
