package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PageFollowUpRecords(userID int64, query model.StudentFollowUpQueryDTO) (model.PageResult[model.StudentFollowUpRecord], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.StudentFollowUpRecord]{}, errors.New("no institution context")
		}
		return model.PageResult[model.StudentFollowUpRecord]{}, err
	}
	return svc.repo.PageFollowUpRecords(context.Background(), instID, query)
}

func (svc *Service) CreateFollowUp(userID int64, dto model.CreateFollowUpDTO) error {
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
	if dto.StudentID <= 0 || strings.TrimSpace(dto.Content) == "" {
		return errors.New("studentId and content are required")
	}
	before, err := svc.repo.GetStudentSnapshot(context.Background(), instID, dto.StudentID)
	if err != nil {
		return err
	}
	if err := svc.repo.CreateFollowUp(context.Background(), instID, instUserID, dto); err != nil {
		return err
	}
	if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, dto.StudentID); err == nil {
		_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, dto.StudentID, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_followup", "created", map[string]any{
			"instId":       instID,
			"studentId":    dto.StudentID,
			"followMethod": dto.FollowMethod,
			"content":      dto.Content,
		})
	}
	return nil
}

func (svc *Service) GetFollowUpCount(userID int64, dto model.FollowUpCountDTO) (model.FollowUpCountVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FollowUpCountVO{}, errors.New("no institution context")
		}
		return model.FollowUpCountVO{}, err
	}
	return svc.repo.GetFollowUpCount(context.Background(), instID)
}

func (svc *Service) UpdateFollowUpRecord(userID int64, dto model.UpdateFollowUpDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.ID <= 0 || strings.TrimSpace(dto.Content) == "" {
		return errors.New("id and content are required")
	}
	studentID, err := svc.repo.GetFollowRecordStudentID(context.Background(), instID, dto.ID)
	if err != nil {
		return err
	}
	before, err := svc.repo.GetStudentSnapshot(context.Background(), instID, studentID)
	if err != nil {
		return err
	}
	if err := svc.repo.UpdateFollowUpRecord(context.Background(), instID, dto); err != nil {
		return err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err == nil {
		if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, studentID); err == nil {
			_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, studentID, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
		}
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_followup", "updated", map[string]any{
			"instId":         instID,
			"followRecordId": dto.ID,
			"followUpStatus": dto.FollowUpStatus,
			"intentLevel":    dto.IntentLevel,
		})
	}
	return nil
}

func (svc *Service) GetFollowUpRecordStatistics(userID int64, query model.StudentFollowUpQueryDTO) (model.FollowVisitCountVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FollowVisitCountVO{}, errors.New("no institution context")
		}
		return model.FollowVisitCountVO{}, err
	}
	return svc.repo.GetFollowUpRecordStatistics(context.Background(), instID, query)
}

func (svc *Service) UpdateVisitStatus(userID int64, dto model.VisitStatusUpdateDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.ID <= 0 || dto.VisitStatus == nil {
		return errors.New("id and visitStatus are required")
	}
	studentID, err := svc.repo.GetFollowRecordStudentID(context.Background(), instID, dto.ID)
	if err != nil {
		return err
	}
	before, err := svc.repo.GetStudentSnapshot(context.Background(), instID, studentID)
	if err != nil {
		return err
	}
	if err := svc.repo.UpdateVisitStatus(context.Background(), instID, dto); err != nil {
		return err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err == nil {
		if after, err := svc.repo.GetStudentSnapshot(context.Background(), instID, studentID); err == nil {
			_ = svc.repo.InsertStudentChangeRecord(context.Background(), instID, studentID, instUserID, svc.buildStudentSnapshotChangeText(context.Background(), before, after))
		}
	}
	if svc.mqClient != nil {
		_ = svc.publishMQ("student_followup", "visit_status", map[string]any{
			"instId":         instID,
			"followRecordId": dto.ID,
			"visitStatus":    *dto.VisitStatus,
		})
	}
	return nil
}
