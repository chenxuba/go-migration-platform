package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetLedgerList(userID int64, query model.LedgerListQueryDTO) (model.LedgerListResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LedgerListResultVO{}, errors.New("no institution context")
		}
		return model.LedgerListResultVO{}, err
	}
	return svc.repo.PageLedgers(context.Background(), instID, query)
}

func (svc *Service) GetLedgerStatistics(userID int64, query model.LedgerListQueryDTO) (model.LedgerStatisticsVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LedgerStatisticsVO{}, errors.New("no institution context")
		}
		return model.LedgerStatisticsVO{}, err
	}
	return svc.repo.GetLedgerStatistics(context.Background(), instID, query)
}

func (svc *Service) ExportLedgers(userID int64, query model.LedgerListQueryDTO) ([]byte, string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("no institution context")
		}
		return nil, "", err
	}

	ledgerIDs := make([]string, 0, len(query.QueryModel.LedgerIDs))
	seen := make(map[string]struct{}, len(query.QueryModel.LedgerIDs))
	for _, item := range query.QueryModel.LedgerIDs {
		ledgerID := strings.TrimSpace(item)
		if ledgerID == "" {
			continue
		}
		if _, ok := seen[ledgerID]; ok {
			continue
		}
		seen[ledgerID] = struct{}{}
		ledgerIDs = append(ledgerIDs, ledgerID)
	}
	if len(ledgerIDs) == 0 {
		return nil, "", errors.New("请选择需要导出的账单")
	}
	if len(ledgerIDs) > ledgerExportMaxRows {
		return nil, "", errors.New("当前最多支持批量导出10000条账单，请减少勾选数量后重试")
	}

	exportQuery := query
	exportQuery.QueryModel.LedgerIDs = ledgerIDs
	exportQuery.PageRequestModel.PageIndex = 1
	exportQuery.PageRequestModel.PageSize = ledgerExportMaxRows

	result, err := svc.repo.PageLedgers(context.Background(), instID, exportQuery)
	if err != nil {
		return nil, "", err
	}
	if result.Total == 0 || len(result.List) == 0 {
		return nil, "", errors.New("没有符合条件的账单可以导出")
	}
	if result.Total > ledgerExportMaxRows {
		return nil, "", errors.New("当前最多支持批量导出10000条账单，请减少勾选数量后重试")
	}

	content, err := buildLedgerExportWorkbook(result.List)
	if err != nil {
		return nil, "", err
	}
	fileName := fmt.Sprintf("系统账单批量导出-%s.xlsx", time.Now().Format("20060102150405"))
	return content, fileName, nil
}

func (svc *Service) ConfirmLedger(userID int64, dto model.LedgerOperateDTO) error {
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
	ledgerID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || ledgerID <= 0 {
		return errors.New("账单ID不能为空")
	}
	operatorName := svc.repo.GetStaffNameByID(context.Background(), &instUserID)
	return svc.repo.ConfirmLedgerWithRemark(context.Background(), instID, ledgerID, instUserID, operatorName, dto.ConfirmRemark)
}

func (svc *Service) CancelConfirmLedger(userID int64, dto model.LedgerOperateDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	ledgerID, err := strconv.ParseInt(strings.TrimSpace(dto.ID), 10, 64)
	if err != nil || ledgerID <= 0 {
		return errors.New("账单ID不能为空")
	}
	return svc.repo.CancelConfirmLedger(context.Background(), instID, ledgerID)
}
