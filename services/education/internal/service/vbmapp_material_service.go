package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ListVBMAPPMaterialProfiles(userID int64) ([]model.VBMAPPMaterialProfile, error) {
	if svc == nil || svc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	if userID <= 0 {
		return nil, errors.New("用户未登录")
	}
	return svc.repo.ListVBMAPPMaterialProfiles(context.Background(), vbmappScaleVersion)
}

func (svc *Service) SaveVBMAPPMaterialItem(userID int64, item model.VBMAPPMaterialItem) (model.VBMAPPMaterialItem, error) {
	if svc == nil || svc.repo == nil {
		return model.VBMAPPMaterialItem{}, errors.New("repository is not configured")
	}
	if userID <= 0 {
		return model.VBMAPPMaterialItem{}, errors.New("用户未登录")
	}
	item.ProfileID = strings.TrimSpace(item.ProfileID)
	item.MaterialName = strings.TrimSpace(item.MaterialName)
	item.MaterialCode = strings.TrimSpace(item.MaterialCode)
	item.MaterialType = strings.TrimSpace(item.MaterialType)
	if item.ProfileID == "" {
		return model.VBMAPPMaterialItem{}, errors.New("素材分类不能为空")
	}
	if item.MaterialName == "" {
		return model.VBMAPPMaterialItem{}, errors.New("素材名称不能为空")
	}
	profiles, err := svc.repo.ListVBMAPPMaterialProfiles(context.Background(), vbmappScaleVersion)
	if err != nil {
		return model.VBMAPPMaterialItem{}, err
	}
	if !vbmappMaterialProfileExists(profiles, item.ProfileID) {
		return model.VBMAPPMaterialItem{}, errors.New("素材分类不存在")
	}
	return svc.repo.SaveVBMAPPMaterialItem(context.Background(), vbmappScaleVersion, userID, item)
}

func (svc *Service) DeleteVBMAPPMaterialItem(userID, id int64) error {
	if svc == nil || svc.repo == nil {
		return errors.New("repository is not configured")
	}
	if userID <= 0 {
		return errors.New("用户未登录")
	}
	if id <= 0 {
		return errors.New("素材ID不能为空")
	}
	if err := svc.repo.DeleteVBMAPPMaterialItem(context.Background(), vbmappScaleVersion, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("素材不存在或已删除")
		}
		return err
	}
	return nil
}

func vbmappMaterialProfileExists(profiles []model.VBMAPPMaterialProfile, profileID string) bool {
	for _, profile := range profiles {
		if profile.ProfileID == profileID {
			return true
		}
	}
	return false
}
