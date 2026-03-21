package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetChannelCategoryList(userID int64) ([]model.ChannelCategoryVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.GetChannelCategories(context.Background(), instID)
}

func (svc *Service) GetChannelList(userID int64) ([]model.ChannelVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.GetChannels(context.Background(), instID)
}

func (svc *Service) GetChannelListWithChannels(userID int64) ([]model.CustomChannelVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.GetChannelListWithChannels(context.Background(), instID)
}

func (svc *Service) GetDefaultChannelList() ([]model.ChannelVO, error) {
	return svc.repo.GetDefaultChannels(context.Background())
}

func (svc *Service) AddChannelCategory(userID int64, input model.ChannelCategoryMutation) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(input.CategoryName) == "" {
		return 0, errors.New("categoryName is required")
	}
	count, err := svc.repo.CountChannelCategoriesByName(context.Background(), instID, input.CategoryName, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("请勿创建相同名称的渠道分类")
	}
	return svc.repo.CreateChannelCategory(context.Background(), instID, input)
}

func (svc *Service) UpdateChannelCategory(userID int64, input model.ChannelCategoryMutation) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if input.ID == nil || *input.ID <= 0 || strings.TrimSpace(input.CategoryName) == "" {
		return errors.New("id and categoryName are required")
	}
	count, err := svc.repo.CountChannelCategoriesByName(context.Background(), instID, input.CategoryName, input.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请勿创建相同名称的渠道分类")
	}
	return svc.repo.UpdateChannelCategory(context.Background(), instID, input)
}

func (svc *Service) DeleteChannelCategory(userID int64, categoryID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	count, err := svc.repo.CountChannelsByCategory(context.Background(), instID, categoryID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请移出分类内所有渠道后再次尝试删除")
	}
	return svc.repo.DeleteChannelCategory(context.Background(), instID, categoryID)
}

func (svc *Service) UpdateChannelStatus(userID int64, input model.ChannelStatusMutation) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if input.ID == nil || input.IsDisabled == nil {
		return errors.New("id and isDisabled are required")
	}
	return svc.repo.UpdateChannelStatus(context.Background(), instID, input)
}

func (svc *Service) AddChannel(userID int64, input model.ChannelMutation) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(input.ChannelName) == "" {
		return 0, errors.New("channelName is required")
	}
	count, err := svc.repo.CountCustomChannelsByName(context.Background(), instID, input.ChannelName, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("请勿创建相同名称的渠道")
	}
	defaultCount, err := svc.repo.CountDefaultChannelsByName(context.Background(), input.ChannelName)
	if err != nil {
		return 0, err
	}
	if defaultCount > 0 {
		return 0, errors.New("渠道名称与系统默认渠道重复")
	}
	return svc.repo.CreateChannel(context.Background(), instID, input)
}

func (svc *Service) UpdateChannel(userID int64, input model.ChannelMutation) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if input.ID == nil || *input.ID <= 0 || strings.TrimSpace(input.ChannelName) == "" {
		return errors.New("id and channelName are required")
	}
	count, err := svc.repo.CountCustomChannelsByName(context.Background(), instID, input.ChannelName, input.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请勿创建相同名称的渠道")
	}
	defaultCount, err := svc.repo.CountDefaultChannelsByName(context.Background(), input.ChannelName)
	if err != nil {
		return err
	}
	if defaultCount > 0 {
		return errors.New("渠道名称与系统默认渠道重复")
	}
	return svc.repo.UpdateChannel(context.Background(), instID, input)
}

func (svc *Service) AdjustChannels(userID int64, input model.AdjustChannelDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(input.ChannelIDs) == 0 || input.CategoryID == nil {
		return errors.New("channelIds and categoryId are required")
	}
	return svc.repo.AdjustChannels(context.Background(), instID, input)
}

func (svc *Service) GetChannelTree(userID int64) ([]model.ChannelTreeVO, error) {
	categories, err := svc.GetChannelCategoryList(userID)
	if err != nil {
		return nil, err
	}
	channels, err := svc.GetChannelList(userID)
	if err != nil {
		return nil, err
	}

	tree := make([]model.ChannelTreeVO, 0, len(categories)+1)
	uncategorized := make([]model.ChannelVO, 0)
	grouped := make(map[int64][]model.ChannelVO)

	for _, channel := range channels {
		if channel.CategoryID == 0 {
			uncategorized = append(uncategorized, channel)
			continue
		}
		grouped[channel.CategoryID] = append(grouped[channel.CategoryID], channel)
	}

	if len(uncategorized) > 0 {
		tree = append(tree, model.ChannelTreeVO{
			ID:          0,
			Name:        "无分类",
			IsDisabled:  false,
			Type:        0,
			ChannelList: uncategorized,
		})
	}

	for _, category := range categories {
		tree = append(tree, model.ChannelTreeVO{
			ID:          category.ID,
			Name:        category.CategoryName,
			IsDisabled:  false,
			Type:        1,
			ChannelList: grouped[category.ID],
		})
	}

	return tree, nil
}

func (svc *Service) PageChannelPC(userID int64, query model.ChannelPCQueryDTO) (model.PageResult[model.ChannelPCVO], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.ChannelPCVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.ChannelPCVO]{}, err
	}
	return svc.repo.PageChannelPC(context.Background(), instID, query)
}
