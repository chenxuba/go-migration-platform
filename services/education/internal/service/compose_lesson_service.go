package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) CreateComposeLesson(userID int64, lessonName string, productIDs []string) (id string, name string, err error) {
	if strings.TrimSpace(lessonName) == "" {
		return "", "", errors.New("组合课程名称不能为空")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.New("no institution context")
		}
		return "", "", err
	}
	courseIDs, err := svc.repo.NormalizeComposeProductIDs(productIDs)
	if err != nil {
		return "", "", err
	}
	n, err := svc.repo.CountCoursesInInst(context.Background(), instID, courseIDs)
	if err != nil {
		return "", "", err
	}
	if n != len(courseIDs) {
		return "", "", errors.New("存在无效或非本机构的课程")
	}
	lid, err := svc.repo.CreateComposeLesson(context.Background(), instID, userID, lessonName, courseIDs)
	if err != nil {
		return "", "", err
	}
	return strconv.FormatInt(lid, 10), strings.TrimSpace(lessonName), nil
}

func (svc *Service) PageComposeLessonsForPC(userID int64, searchKey string, pageIndex, pageSize, skipCount int) (list []model.ComposeLessonListItem, total int, err error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, errors.New("no institution context")
		}
		return nil, 0, err
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	offset := (pageIndex - 1) * pageSize
	if skipCount > 0 {
		offset = skipCount
	}
	list, total, err = svc.repo.PageComposeLessonsForPC(context.Background(), instID, searchKey, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
