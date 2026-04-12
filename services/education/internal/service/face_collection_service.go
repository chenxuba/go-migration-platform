package service

import (
	"context"
	"database/sql"
	"errors"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) PageFaceCollectionStudents(userID int64, query model.FaceCollectionStudentQueryDTO) (model.PageResult[model.FaceCollectionStudent], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.FaceCollectionStudent]{}, errors.New("no institution context")
		}
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}
	return svc.repo.PageFaceCollectionStudents(context.Background(), instID, query)
}

func (svc *Service) GetFaceCollectionProfile(userID, studentID int64) (model.FaceCollectionProfile, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceCollectionProfile{}, errors.New("no institution context")
		}
		return model.FaceCollectionProfile{}, err
	}
	item, err := svc.repo.GetFaceCollectionProfile(context.Background(), instID, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceCollectionProfile{}, errors.New("当前学员未采集人脸")
		}
		return model.FaceCollectionProfile{}, err
	}
	return item, nil
}

func (svc *Service) ListFaceCollectionProfiles(userID int64) ([]model.FaceCollectionProfile, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListFaceCollectionProfiles(context.Background(), instID)
}

func (svc *Service) CompareFaceCollectionProfile(userID int64, dto model.FaceCollectionCompareDTO) (model.FaceCollectionCompareResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceCollectionCompareResult{}, errors.New("no institution context")
		}
		return model.FaceCollectionCompareResult{}, err
	}
	return svc.repo.CompareFaceCollectionProfile(context.Background(), instID, dto.FaceDescriptor)
}

func (svc *Service) SaveFaceCollectionProfile(userID int64, dto model.FaceCollectionProfileSaveDTO) error {
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
	return svc.repo.SaveFaceCollectionProfile(context.Background(), instID, instUserID, dto)
}

func (svc *Service) DeleteFaceCollectionProfile(userID, studentID int64) error {
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
	return svc.repo.DeleteFaceCollectionProfile(context.Background(), instID, instUserID, studentID)
}

func (svc *Service) ListFaceAttendanceSessions(userID int64, limit int) ([]model.FaceAttendanceSession, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListFaceAttendanceSessions(context.Background(), instID, limit)
}

func (svc *Service) RecognizeFaceAttendanceSession(userID int64, dto model.FaceAttendanceSessionRecognizeDTO) (model.FaceAttendanceSessionRecognizeResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceSessionRecognizeResult{}, errors.New("no institution context")
		}
		return model.FaceAttendanceSessionRecognizeResult{}, err
	}
	return svc.repo.RecognizeFaceAttendanceSession(context.Background(), instID, dto)
}

func (svc *Service) CommitFaceAttendanceSession(userID int64, dto model.FaceAttendanceSessionCommitDTO) (model.FaceAttendanceSession, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceSession{}, errors.New("no institution context")
		}
		return model.FaceAttendanceSession{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceSession{}, errors.New("no institution user context")
		}
		return model.FaceAttendanceSession{}, err
	}
	return svc.repo.CommitFaceAttendanceSession(context.Background(), instID, instUserID, dto)
}

func (svc *Service) ListFaceAttendanceRecords(userID int64, limit int) ([]model.FaceAttendanceRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListFaceAttendanceRecords(context.Background(), instID, limit)
}

func (svc *Service) SaveFaceAttendanceRecord(userID int64, dto model.FaceAttendanceRecordSaveDTO) (model.FaceAttendanceRecord, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceRecord{}, errors.New("no institution context")
		}
		return model.FaceAttendanceRecord{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FaceAttendanceRecord{}, errors.New("no institution user context")
		}
		return model.FaceAttendanceRecord{}, err
	}
	return svc.repo.SaveFaceAttendanceRecord(context.Background(), instID, instUserID, dto)
}
