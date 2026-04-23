package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PagePendingRenewalStudents(userID int64, query model.PendingRenewalStudentPagedQueryDTO) (model.PendingRenewalStudentPagedResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalStudentPagedResult{}, errors.New("no institution context")
		}
		return model.PendingRenewalStudentPagedResult{}, err
	}
	return svc.repo.GetPendingRenewalStudentsPagedList(context.Background(), instID, query)
}
