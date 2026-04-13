package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetRechargeAccountItemPage(userID int64, query model.RechargeAccountItemPageQueryDTO) (model.RechargeAccountItemPageResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountItemPageResult{}, errors.New("no institution context")
		}
		return model.RechargeAccountItemPageResult{}, err
	}
	return svc.repo.PageRechargeAccountItems(context.Background(), instID, query)
}

func (svc *Service) GetRechargeAccountStatistics(userID int64) (model.RechargeAccountStatistics, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountStatistics{}, errors.New("no institution context")
		}
		return model.RechargeAccountStatistics{}, err
	}
	return svc.repo.GetRechargeAccountStatistics(context.Background(), instID)
}

func (svc *Service) ExportRechargeAccountItems(userID int64, query model.RechargeAccountItemPageQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	exportQuery := query
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = rechargeAccountExportMaxRows

	result, err := svc.repo.PageRechargeAccountItems(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的储值账户可以导出")
	}
	if result.Total > rechargeAccountExportMaxRows {
		return nil, "", errors.New("当前列表最多支持导出10000条数据，请缩小筛选范围后重试")
	}

	content, err := buildRechargeAccountExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("储值账户-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func (svc *Service) UpdateRechargeAccount(userID int64, dto model.UpdateRechargeAccountDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}

	name := strings.TrimSpace(dto.RechargeAccountName)
	if strings.TrimSpace(dto.RechargeAccountID) == "" {
		return errors.New("rechargeAccountId不能为空")
	}
	if name == "" {
		return errors.New("rechargeAccountName不能为空")
	}
	if len([]rune(name)) > 30 {
		return errors.New("rechargeAccountName长度不能超过30")
	}

	dto.RechargeAccountName = name
	return svc.repo.UpdateRechargeAccount(context.Background(), instID, instUserID, dto)
}
