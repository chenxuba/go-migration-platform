package service

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetStudentTeachingRecordPagedList(userID int64, dto model.StudentTeachingRecordPagedQueryDTO) (model.StudentTeachingRecordPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.StudentTeachingRecordPagedResult{}, err
	}
	return svc.repo.GetStudentTeachingRecordPagedList(context.Background(), instID, dto)
}

func (svc *Service) GetScheduleTeachingRecordPagedList(userID int64, dto model.ScheduleTeachingRecordPagedQueryDTO) (model.ScheduleTeachingRecordPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.ScheduleTeachingRecordPagedResult{}, err
	}
	return svc.repo.GetScheduleTeachingRecordPagedList(context.Background(), instID, dto)
}

func (svc *Service) GetClassCommentPagedList(userID int64, dto model.ClassCommentPagedQueryDTO) (model.ClassCommentPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.ClassCommentPagedResult{}, err
	}

	queryModel := model.StudentTeachingRecordQueryModel{
		BeginStartTime:       dto.QueryModel.TeachingStartTime,
		EndStartTime:         dto.QueryModel.TeachingEndTime,
		TeacherIDs:           dto.QueryModel.TeacherIDs,
		LessonIDs:            nil,
		ClassIDs:             nil,
		One2OneIDs:           nil,
		TimetableSourceTypes: dto.QueryModel.TeachingRecordTypes,
		ScheduleCallStatus:   intPtr(2),
	}
	if lessonID := dto.QueryModel.LessonID; lessonID != "" {
		queryModel.LessonIDs = []string{lessonID}
	}
	if classID := dto.QueryModel.ClassID; classID != "" {
		queryModel.ClassIDs = []string{classID}
	}
	if one2OneID := dto.QueryModel.One2OneID; one2OneID != "" {
		queryModel.One2OneIDs = []string{one2OneID}
	}

	pageResult, err := svc.repo.GetScheduleTeachingRecordPagedList(context.Background(), instID, model.ScheduleTeachingRecordPagedQueryDTO{
		PageRequestModel: dto.PageRequestModel,
		SortModel: model.ScheduleTeachingRecordSortModel{
			StartTime: dto.SortModel.StartTime,
		},
		QueryModel: queryModel,
	})
	if err != nil {
		return model.ClassCommentPagedResult{}, err
	}

	result := model.ClassCommentPagedResult{
		List:  make([]model.ClassCommentPagedItem, 0, len(pageResult.List)),
		Total: pageResult.Total,
	}
	for _, item := range pageResult.List {
		result.List = append(result.List, model.ClassCommentPagedItem{
			TeachingRecordID: item.TeachingRecordID,
			SourceName:       item.SourceName,
			SourceType:       item.SourceType,
			SourceID:         item.SourceID,
			LessonID:         item.LessonID,
			LessonName:       item.LessonName,
			CreatedTime:      item.CreatedTime,
			TeacherID:        item.TeacherID,
			TeacherName:      item.TeacherName,
			StartTime:        item.StartTime,
			EndTime:          item.EndTime,
			ReadCount:        item.ReadCount,
			UnReadCount:      item.UnReadCount,
			CommentCount:     item.CommentCount,
			UnCommentCount:   item.UnCommentCount,
			Assistants:       item.Assistants,
			ClassRoomName:    item.ClassRoomName,
		})
	}
	return result, nil
}

func (svc *Service) GetClassCommentStudentPagedList(userID int64, dto model.ClassCommentStudentPagedQueryDTO) (model.ClassCommentStudentPagedResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.ClassCommentStudentPagedResult{}, err
	}

	if dto.QueryModel.IsRead != nil && *dto.QueryModel.IsRead {
		return model.ClassCommentStudentPagedResult{
			List:  []model.ClassCommentStudentPagedItem{},
			Total: 0,
		}, nil
	}
	if dto.QueryModel.IsParentFeedback != nil && *dto.QueryModel.IsParentFeedback {
		return model.ClassCommentStudentPagedResult{
			List:  []model.ClassCommentStudentPagedItem{},
			Total: 0,
		}, nil
	}
	if dto.QueryModel.IsRead != nil && !*dto.QueryModel.IsRead {
		if dto.QueryModel.IsComment != nil && !*dto.QueryModel.IsComment {
			return model.ClassCommentStudentPagedResult{
				List:  []model.ClassCommentStudentPagedItem{},
				Total: 0,
			}, nil
		}
		dto.QueryModel.IsComment = boolPtr(true)
	}

	return svc.repo.GetClassCommentStudentPagedList(context.Background(), instID, dto)
}

func (svc *Service) GetTeachingRecordDetail(userID int64, query model.TeachingRecordDetailQueryDTO) (model.TeachingRecordDetailResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.TeachingRecordDetailResult{}, err
	}
	return svc.repo.GetTeachingRecordDetail(context.Background(), instID, query)
}

func (svc *Service) GetStudentRehabRecordDetail(userID int64, query model.StudentRehabRecordQueryDTO) (model.StudentRehabRecordDetailResult, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return model.StudentRehabRecordDetailResult{}, err
	}
	return svc.repo.GetStudentRehabRecordDetail(context.Background(), instID, query)
}

func (svc *Service) UpdateTeachingRecordClassInfo(userID int64, dto model.UpdateTeachingRecordClassInfoDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.UpdateTeachingRecordClassInfo(context.Background(), instID, operatorID, dto)
}

func (svc *Service) ExportClassRecords(userID int64, req model.ClassRecordExportCreateRequest) (model.ClassRecordExportRecord, error) {
	return svc.exportClassRecords(userID, req)
}

func (svc *Service) ListClassRecordExportRecords(userID int64, exportType string) ([]model.ClassRecordExportRecord, error) {
	return svc.listClassRecordExportRecords(userID, exportType)
}

func (svc *Service) LoadClassRecordExportRecord(userID int64, recordIDRaw, exportType string) (string, string, []byte, error) {
	return svc.loadClassRecordExportRecord(userID, recordIDRaw, exportType)
}

func (svc *Service) UpdateStudentTeachingRecord(userID int64, dto model.UpdateStudentTeachingRecordDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.UpdateStudentTeachingRecord(context.Background(), instID, operatorID, dto)
}

func (svc *Service) SaveStudentRehabRecordDraft(userID int64, dto model.SaveStudentRehabRecordDraftDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.SaveStudentRehabRecordDraft(context.Background(), instID, operatorID, dto)
}

func (svc *Service) PublishStudentRehabRecord(userID int64, dto model.PublishStudentRehabRecordDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.PublishStudentRehabRecord(context.Background(), instID, operatorID, dto)
}

func (svc *Service) DeleteStudentTeachingRecord(userID int64, dto model.DeleteStudentTeachingRecordDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteStudentTeachingRecord(context.Background(), instID, operatorID, dto)
}

func (svc *Service) DeleteTeachingRecord(userID int64, dto model.DeleteTeachingRecordDTO) (bool, error) {
	instID, err := svc.rollCallInstID(userID)
	if err != nil {
		return false, err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		return false, err
	}
	return svc.repo.DeleteTeachingRecord(context.Background(), instID, operatorID, dto)
}

func boolPtr(v bool) *bool {
	return &v
}
