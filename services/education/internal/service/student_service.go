package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func (svc *Service) UpdateStudentStatus(userID int64, dto model.StudentStatusUpdateDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.ID <= 0 {
		return errors.New("id is required")
	}
	if dto.FollowUpStatus == nil && dto.IntentLevel == nil {
		return errors.New("followUpStatus or intentLevel is required")
	}
	if err := svc.repo.UpdateStudentStatus(context.Background(), instID, dto); err != nil {
		return err
	}
	go func(instID, studentID int64, dto model.StudentStatusUpdateDTO) {
		if svc.mqClient != nil {
			_ = svc.publishMQ("student_intent", "status_changed", map[string]any{
				"instId":         instID,
				"studentId":      studentID,
				"followUpStatus": dto.FollowUpStatus,
				"intentLevel":    dto.IntentLevel,
			})
		}
	}(instID, dto.ID, dto)
	return nil
}

func (svc *Service) BatchAssignSalesperson(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.SalespersonID == nil || len(dto.StudentIDs) == 0 {
		return errors.New("salespersonId and studentIds are required")
	}
	beforeSnapshots := make(map[int64]repository.StudentSnapshot)
	for _, id := range dto.StudentIDs {
		if snapshot, err := svc.repo.GetStudentSnapshot(context.Background(), instID, id); err == nil {
			beforeSnapshots[id] = snapshot
		}
	}
	if err := svc.repo.BatchAssignSalesperson(context.Background(), instID, *dto.SalespersonID, dto.StudentIDs); err != nil {
		return err
	}
	instUserID, _ := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	for _, id := range dto.StudentIDs {
		if before, ok := beforeSnapshots[id]; ok {
			if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, id); err == nil {
				_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, id, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
			}
		}
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_intent", "assigned_sales", map[string]any{
			"instId":        instID,
			"studentIds":    dto.StudentIDs,
			"salespersonId": *dto.SalespersonID,
		})
	}
	return nil
}

func (svc *Service) BatchTransferToPublicPool(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.StudentIDs) == 0 {
		return errors.New("studentIds are required")
	}
	beforeSnapshots := make(map[int64]repository.StudentSnapshot)
	for _, id := range dto.StudentIDs {
		if snapshot, err := svc.repo.GetStudentSnapshot(context.Background(), instID, id); err == nil {
			beforeSnapshots[id] = snapshot
		}
	}
	if err := svc.repo.BatchTransferToPublicPool(context.Background(), instID, dto.StudentIDs); err != nil {
		return err
	}
	instUserID, _ := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	for _, id := range dto.StudentIDs {
		if before, ok := beforeSnapshots[id]; ok {
			if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, id); err == nil {
				_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, id, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
			}
		}
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_intent", "transfer_public_pool", map[string]any{
			"instId":     instID,
			"studentIds": dto.StudentIDs,
		})
	}
	return nil
}

func (svc *Service) BatchDeleteIntentStudents(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.StudentIDs) == 0 {
		return errors.New("studentIds are required")
	}
	if err := svc.repo.BatchDeleteIntentStudents(context.Background(), instID, dto.StudentIDs); err != nil {
		return err
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_intent", "deleted", map[string]any{
			"instId":     instID,
			"studentIds": dto.StudentIDs,
		})
	}
	return nil
}

func (svc *Service) AddIntentStudent(userID int64, dto model.StudentSaveDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(dto.StuName) == "" || strings.TrimSpace(dto.Mobile) == "" {
		return 0, errors.New("stuName and mobile are required")
	}
	rule, count, err := svc.studentDuplicateCheck(context.Background(), instID, dto.StuName, dto.Mobile, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New(studentDuplicateMessage(rule))
	}
	limitSameWeChat, err := svc.repo.GetLimitSameWeChat(context.Background(), instID)
	if err != nil {
		return 0, err
	}
	if limitSameWeChat && strings.TrimSpace(dto.WeChatNumber) != "" {
		count, err := svc.repo.CountStudentByWeChat(context.Background(), instID, dto.WeChatNumber, nil)
		if err != nil {
			return 0, err
		}
		if count > 0 {
			return 0, errors.New(studentWeChatDuplicateMessage())
		}
	}
	return svc.createIntentStudentRecord(userID, instID, dto)
}

func (svc *Service) createIntentStudentRecord(userID int64, instID int64, dto model.StudentSaveDTO) (int64, error) {
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution user context")
		}
		return 0, err
	}
	dto.OperatorID = &instUserID
	return svc.repo.CreateIntentStudent(context.Background(), instID, instUserID, dto)
}

func (svc *Service) UpdateIntentStudent(userID int64, dto model.StudentSaveDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.StudentID == nil || *dto.StudentID <= 0 {
		return errors.New("studentId is required")
	}
	if strings.TrimSpace(dto.StuName) == "" || strings.TrimSpace(dto.Mobile) == "" {
		return errors.New("stuName and mobile are required")
	}
	rule, count, err := svc.studentDuplicateCheck(context.Background(), instID, dto.StuName, dto.Mobile, dto.StudentID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(studentDuplicateMessage(rule))
	}
	limitSameWeChat, err := svc.repo.GetLimitSameWeChat(context.Background(), instID)
	if err != nil {
		return err
	}
	if limitSameWeChat && strings.TrimSpace(dto.WeChatNumber) != "" {
		count, err := svc.repo.CountStudentByWeChat(context.Background(), instID, dto.WeChatNumber, dto.StudentID)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New(studentWeChatDuplicateMessage())
		}
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}
	dto.OperatorID = &instUserID
	before, err := svc.repo.GetStudentSnapshot(context.Background(), instID, *dto.StudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("学生信息不存在")
		}
		return err
	}
	if err := svc.repo.UpdateIntentStudent(context.Background(), instID, dto); err != nil {
		return err
	}
	if err == nil {
		if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, *dto.StudentID); err == nil {
			_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, *dto.StudentID, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
		}
	}
	return nil
}

func (svc *Service) CheckStudentRepeat(userID int64, req model.StudentDuplicateCheckRequest) (model.IntentStudentRepeatVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentStudentRepeatVO{}, errors.New("no institution context")
		}
		return model.IntentStudentRepeatVO{}, err
	}
	if strings.TrimSpace(req.StuName) == "" || strings.TrimSpace(req.Mobile) == "" {
		return model.IntentStudentRepeatVO{}, errors.New("stuName and mobile are required")
	}
	rule, count, err := svc.studentDuplicateCheck(context.Background(), instID, req.StuName, req.Mobile, req.ID)
	if err != nil {
		return model.IntentStudentRepeatVO{}, err
	}
	result := model.IntentStudentRepeatVO{}
	if count > 0 {
		result.AddStudentRepeatRuleEnum = rule
	}
	return result, nil
}

func (svc *Service) CheckStudentTips(userID int64, req model.StudentDuplicateCheckRequest) ([]string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if strings.TrimSpace(req.StuName) == "" || strings.TrimSpace(req.Mobile) == "" {
		return nil, errors.New("stuName and mobile are required")
	}
	rule, count, err := svc.studentDuplicateCheck(context.Background(), instID, req.StuName, req.Mobile, req.ID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return []string{}, nil
	}
	return []string{studentDuplicateMessage(rule)}, nil
}

func (svc *Service) GetStudentPhoneNumber(userID, studentID int64) (string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}
	return svc.repo.GetStudentPhone(context.Background(), instID, studentID)
}

func (svc *Service) PageRecommenders(userID int64, query model.RecommenderQueryDTO) (model.PageResult[model.RecommenderQueryVO], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.RecommenderQueryVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.RecommenderQueryVO]{}, err
	}
	return svc.repo.PageRecommenders(context.Background(), instID, query)
}

func (svc *Service) PageBirthdayStudents(userID int64, query model.BirthdayStudentQueryDTO) (model.PageResult[model.BirthdayStudentQueryVO], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.BirthdayStudentQueryVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.BirthdayStudentQueryVO]{}, err
	}
	return svc.repo.PageBirthdayStudents(context.Background(), instID, query)
}

func (svc *Service) ListStudentChangeRecords(userID, stuID int64) ([]model.StudentChangeRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListStudentChangeRecords(context.Background(), instID, stuID)
}
