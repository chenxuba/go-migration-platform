package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) AddCourse(userID int64, input model.CourseProductSaveDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if err := validateCourseProduct(input); err != nil {
		return 0, err
	}
	used, err := svc.repo.CountCourseByName(context.Background(), instID, input.Name, nil)
	if err != nil {
		return 0, err
	}
	if used > 0 {
		return 0, errors.New("课程名称已存在")
	}
	return svc.repo.CreateCourse(context.Background(), instID, userID, input)
}

func (svc *Service) UpdateCourse(userID int64, input model.CourseProductSaveDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if input.ID == nil || *input.ID <= 0 {
		return errors.New("id is required")
	}
	if err := validateCourseProduct(input); err != nil {
		return err
	}
	used, err := svc.repo.CountCourseByName(context.Background(), instID, input.Name, input.ID)
	if err != nil {
		return err
	}
	if used > 0 {
		return errors.New("课程名称已存在")
	}
	return svc.repo.UpdateCourse(context.Background(), instID, userID, input)
}

func (svc *Service) PageCourseCategories(userID int64, query model.CourseCategoryQueryDTO) (model.PageResult[model.CourseCategory], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.CourseCategory]{}, errors.New("no institution context")
		}
		return model.PageResult[model.CourseCategory]{}, err
	}
	return svc.repo.PageCourseCategories(context.Background(), instID, query)
}

func (svc *Service) AddCourseCategory(userID int64, input model.CourseCategoryMutation) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if strings.TrimSpace(input.Name) == "" || input.Sort == nil {
		return 0, errors.New("name and sort are required")
	}
	count, err := svc.repo.CountCourseCategoryByName(context.Background(), instID, input.Name, nil)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("请勿创建相同名称的课程类别")
	}
	return svc.repo.CreateCourseCategory(context.Background(), instID, input)
}

func (svc *Service) UpdateCourseCategory(userID int64, input model.CourseCategoryMutation) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if input.ID == nil || *input.ID <= 0 || strings.TrimSpace(input.Name) == "" || input.Sort == nil {
		return errors.New("id, name and sort are required")
	}
	count, err := svc.repo.CountCourseCategoryByName(context.Background(), instID, input.Name, input.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请勿创建相同名称的课程类别")
	}
	return svc.repo.UpdateCourseCategory(context.Background(), instID, input)
}

func (svc *Service) DeleteCourseCategory(userID int64, categoryID int64) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	count, err := svc.repo.CountCoursesByCategory(context.Background(), instID, categoryID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("请移出分类内所有课程后再次尝试删除")
	}
	return svc.repo.DeleteCourseCategory(context.Background(), instID, categoryID)
}

func (svc *Service) PageCourses(userID int64, query model.CourseQueryDTO) (model.PageResult[model.Course], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.Course]{}, errors.New("no institution context")
		}
		return model.PageResult[model.Course]{}, err
	}
	return svc.repo.PageCourses(context.Background(), instID, query)
}

func (svc *Service) PageCourseIDNames(userID int64, query model.CourseQueryDTO) (model.PageResult[model.CourseIDName], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.CourseIDName]{}, errors.New("no institution context")
		}
		return model.PageResult[model.CourseIDName]{}, err
	}
	return svc.repo.PageCourseIDNames(context.Background(), instID, query)
}

func (svc *Service) GetCourseDetail(userID, courseID int64) (model.CourseDetail, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CourseDetail{}, errors.New("no institution context")
		}
		return model.CourseDetail{}, err
	}
	return svc.repo.GetCourseDetail(context.Background(), instID, courseID)
}

func (svc *Service) PageProcessContent(userID int64, query model.CourseQueryDTO) (model.PageResult[model.ProcessContentQueryVO], error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.ProcessContentQueryVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.ProcessContentQueryVO]{}, err
	}
	return svc.repo.PageProcessContent(context.Background(), instID, query)
}

func (svc *Service) BatchDeleteOrRestoreCourses(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.CourseIDs) == 0 || dto.DelFlag == nil {
		return errors.New("courseIds and delFlag are required")
	}
	return svc.repo.BatchDeleteOrRestoreCourses(context.Background(), instID, dto.CourseIDs, *dto.DelFlag)
}

func (svc *Service) BatchUpdateCourseSaleStatus(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.CourseIDs) == 0 || dto.SaleStatus == nil {
		return errors.New("courseIds and saleStatus are required")
	}
	return svc.repo.BatchUpdateCourseSaleStatus(context.Background(), instID, dto.CourseIDs, *dto.SaleStatus)
}

func (svc *Service) BatchOpenMicroSchoolShow(userID int64, dto model.BatchCommonDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if len(dto.CourseIDs) == 0 || dto.IsShowMicoSchool == nil {
		return errors.New("courseIds and isShowMicoSchool are required")
	}
	return svc.repo.BatchUpdateCourseMicroSchoolShow(context.Background(), instID, dto.CourseIDs, *dto.IsShowMicoSchool)
}

func (svc *Service) ListCourseProperties(userID int64) ([]model.CourseProperty, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListCourseProperties(context.Background(), instID)
}

func (svc *Service) UpdateCourseProperty(userID int64, dto model.CourseProperty) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetCoursePropertyByID(context.Background(), dto.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("课程属性不存在")
		}
		return err
	}
	if current.InstID != instID {
		return errors.New("课程属性不存在")
	}
	if dto.UUID != "" && (dto.UUID != current.UUID || dto.Version != current.Version) {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	if dto.Name == "" {
		dto.Name = current.Name
	}
	return svc.repo.UpdateCourseProperty(context.Background(), dto)
}

func (svc *Service) ListCoursePropertyOptions(userID, propertyID int64) ([]model.CoursePropertyOption, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	current, err := svc.repo.GetCoursePropertyByID(context.Background(), propertyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("课程属性不存在")
		}
		return nil, err
	}
	if current.InstID != instID {
		return nil, errors.New("课程属性不存在")
	}
	return svc.repo.ListCoursePropertyOptions(context.Background(), propertyID)
}

func (svc *Service) AddCoursePropertyOption(userID int64, dto model.CoursePropertyOption) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	current, err := svc.repo.GetCoursePropertyByID(context.Background(), dto.PropertyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("课程属性不存在")
		}
		return 0, err
	}
	if current.InstID != instID {
		return 0, errors.New("课程属性不存在")
	}
	if dto.Sort == 0 {
		options, _ := svc.repo.ListCoursePropertyOptions(context.Background(), dto.PropertyID)
		dto.Sort = len(options) + 1
	}
	return svc.repo.CreateCoursePropertyOption(context.Background(), dto)
}

func (svc *Service) UpdateCoursePropertyOption(userID int64, dto model.CoursePropertyOption) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetCoursePropertyOptionByID(context.Background(), dto.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("不存在该选项")
		}
		return err
	}
	prop, err := svc.repo.GetCoursePropertyByID(context.Background(), current.PropertyID)
	if err != nil {
		return err
	}
	if prop.InstID != instID {
		return errors.New("不存在该选项")
	}
	if dto.UUID != "" && (dto.UUID != current.UUID || dto.Version != current.Version) {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.UpdateCoursePropertyOption(context.Background(), dto)
}

func (svc *Service) DeleteCoursePropertyOption(userID int64, dto model.CoursePropertyOption) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	current, err := svc.repo.GetCoursePropertyOptionByID(context.Background(), dto.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("不存在该选项")
		}
		return err
	}
	prop, err := svc.repo.GetCoursePropertyByID(context.Background(), current.PropertyID)
	if err != nil {
		return err
	}
	if prop.InstID != instID {
		return errors.New("不存在该选项")
	}
	if dto.UUID != "" && (dto.UUID != current.UUID || dto.Version != current.Version) {
		return errors.New("当前信息已变更，请刷新后重试")
	}
	return svc.repo.DeleteCoursePropertyOption(context.Background(), dto.ID)
}

func (svc *Service) BatchUpdateCoursePropertyOptionSort(userID int64, items []model.CoursePropertyOption) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	for _, item := range items {
		current, err := svc.repo.GetCoursePropertyOptionByID(context.Background(), item.ID)
		if err != nil {
			return err
		}
		prop, err := svc.repo.GetCoursePropertyByID(context.Background(), current.PropertyID)
		if err != nil {
			return err
		}
		if prop.InstID != instID {
			return errors.New("非法操作")
		}
	}
	return svc.repo.BatchUpdateCoursePropertyOptionSort(context.Background(), items)
}

func validateCourseProduct(input model.CourseProductSaveDTO) error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("课程名称不能为空")
	}
	if input.SaleStatus == nil {
		return errors.New("课程售卖状态不能为空")
	}
	if input.CourseType == nil {
		return errors.New("通用课程不能为空")
	}
	if *input.CourseType == 1 && input.TeachMethod == nil {
		return errors.New("授课方式不能为空")
	}
	if *input.CourseType == 3 && len(input.CourseScope) == 0 {
		return errors.New("课程范围不能为空")
	}
	if len(input.ProductSku) == 0 {
		return errors.New("报价单不能为空")
	}
	return nil
}
