package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ListSchoolHolidays(userID int64) ([]model.SchoolHolidayVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}

	ctx := context.Background()
	totalCount, err := svc.repo.CountSchoolHolidayRecords(ctx, instID)
	if err != nil {
		return nil, err
	}
	if totalCount == 0 {
		if err := svc.repo.SeedDefaultSchoolHolidays(ctx, instID, time.Now().Year()); err != nil {
			return nil, err
		}
	}
	return svc.repo.ListSchoolHolidays(ctx, instID)
}

func (svc *Service) SaveSchoolHoliday(userID int64, input model.SchoolHolidayMutation) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	return svc.repo.SaveSchoolHoliday(context.Background(), instID, input)
}

func (svc *Service) DeleteSchoolHoliday(userID, holidayID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if holidayID <= 0 {
		return errors.New("节假日ID不能为空")
	}
	if _, err := svc.repo.GetSchoolHolidayByID(context.Background(), instID, holidayID); err != nil {
		return err
	}
	return svc.repo.DeleteSchoolHoliday(context.Background(), instID, holidayID)
}

func (svc *Service) ResetSchoolHolidaysToDefault(userID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	return svc.repo.ResetSchoolHolidaysToDefault(context.Background(), instID, time.Now().Year())
}
