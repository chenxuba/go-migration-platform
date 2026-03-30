package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetOrderList(userID int64, query model.OrderManageQueryDTO) (model.OrderManageResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderManageResultVO{}, errors.New("no institution context")
		}
		return model.OrderManageResultVO{}, err
	}
	return svc.repo.PageOrders(context.Background(), instID, query)
}

func (svc *Service) GetOrderDetailList(userID int64, query model.OrderDetailListQueryDTO) (model.OrderDetailListResultVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderDetailListResultVO{}, errors.New("no institution context")
		}
		return model.OrderDetailListResultVO{}, err
	}
	return svc.repo.PageOrderDetails(context.Background(), instID, query)
}

func (svc *Service) GetOrderDetail(userID, orderID int64) (model.OrderDetailVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OrderDetailVO{}, errors.New("no institution context")
		}
		return model.OrderDetailVO{}, err
	}
	return svc.repo.GetOrderDetail(context.Background(), instID, orderID)
}

func (svc *Service) SetBadDebt(userID int64, dto model.BadDebtDTO) error {
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
	orderID, err := strconv.ParseInt(strings.TrimSpace(dto.OrderID), 10, 64)
	if err != nil || orderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	return svc.repo.SetBadDebt(context.Background(), instID, orderID, instUserID, dto.Remark)
}

func (svc *Service) CancelBadDebt(userID int64, orderIDRaw string) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	orderID, err := strconv.ParseInt(strings.TrimSpace(orderIDRaw), 10, 64)
	if err != nil || orderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	return svc.repo.CancelBadDebt(context.Background(), instID, orderID)
}

func (svc *Service) CloseOrder(userID int64, orderIDRaw string) error {
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
	orderID, err := strconv.ParseInt(strings.TrimSpace(orderIDRaw), 10, 64)
	if err != nil || orderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	return svc.repo.CloseOrder(context.Background(), instID, instUserID, orderID)
}

func (svc *Service) CalcCourseEnrollType(userID int64, dto model.CourseEnrollTypeDTO) ([]model.CourseEnrollTypeVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if dto.StudentID <= 0 {
		return nil, errors.New("学生ID不能为空")
	}
	if len(dto.Courses) == 0 {
		return nil, errors.New("意向内容列表不能为空")
	}
	studentSnapshot, err := svc.repo.GetStudentSnapshot(context.Background(), instID, dto.StudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("学生不存在")
		}
		return nil, err
	}

	hasAnyPurchased, err := svc.repo.StudentHasCompletedOrders(context.Background(), instID, dto.StudentID)
	if err != nil {
		return nil, err
	}

	result := make([]model.CourseEnrollTypeVO, 0, len(dto.Courses))
	for _, course := range dto.Courses {
		item := model.CourseEnrollTypeVO{CourseID: course.CourseID}
		switch {
		case course.IsAudition:
			item.EnrollType = 0
		case studentSnapshot.StudentStatus == 0:
			item.EnrollType = 1
		case !hasAnyPurchased:
			item.EnrollType = 1
		default:
			active, err := svc.repo.StudentHasActiveCourseEnrollment(context.Background(), instID, dto.StudentID, course.CourseID)
			if err != nil {
				return nil, err
			}
			if active {
				item.EnrollType = 2
			} else {
				purchased, err := svc.repo.StudentHasCompletedOrderForCourse(context.Background(), instID, dto.StudentID, course.CourseID)
				if err != nil {
					return nil, err
				}
				if !purchased {
					item.EnrollType = 3
				} else {
					item.EnrollType = 1
				}
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (svc *Service) CheckQuoteInfo(userID int64, dto model.CheckQuoteDTO) ([]model.CheckQuoteVO, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	if len(dto.QuoteDetailList) == 0 {
		return nil, errors.New("请选择办理内容")
	}

	_ = instID
	quoteIDs := make([]int64, 0, len(dto.QuoteDetailList))
	for _, item := range dto.QuoteDetailList {
		if item.QuoteID > 0 {
			quoteIDs = append(quoteIDs, item.QuoteID)
		}
	}
	quotationMap, err := svc.repo.GetCourseQuotationsByIDs(context.Background(), quoteIDs)
	if err != nil {
		return nil, err
	}

	result := make([]model.CheckQuoteVO, 0)
	for _, detail := range dto.QuoteDetailList {
		quotation, ok := quotationMap[detail.QuoteID]
		if !ok {
			result = append(result, model.CheckQuoteVO{
				CourseID: detail.CourseID,
				Error:    1,
			})
			continue
		}
		priceMismatch := detail.Price != quotation.Price
		quantityMismatch := (detail.Quantity == nil) != (quotation.Quantity == nil)
		if !quantityMismatch && detail.Quantity != nil && quotation.Quantity != nil && *detail.Quantity != *quotation.Quantity {
			quantityMismatch = true
		}
		if priceMismatch || quantityMismatch {
			result = append(result, model.CheckQuoteVO{
				CourseID: detail.CourseID,
				Error:    1,
			})
		}
	}
	return result, nil
}

func (svc *Service) CreateOrder(userID int64, dto model.CreateOrderDTO) (int64, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution context")
		}
		return 0, err
	}
	if dto.StudentID <= 0 {
		return 0, errors.New("请选择学员!")
	}
	if len(dto.OrderDetail.QuoteDetailList) == 0 {
		return 0, errors.New("请选择办理内容!")
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("no institution user context")
		}
		return 0, err
	}
	return svc.repo.CreateOrder(context.Background(), instID, instUserID, dto)
}

func (svc *Service) PayOrder(userID int64, dto model.PayOrderDTO) error {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution context")
		}
		return err
	}
	if dto.OrderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("no institution user context")
		}
		return err
	}
	return svc.repo.PayOrder(context.Background(), instID, instUserID, dto)
}

func (svc *Service) GetRegistrationListPage(userID int64, query model.RegistrationListQueryDTO) (model.RegistrationListResultVO, error) {
	svc.SyncScheduledSuspendResumeTuitionAccountsOnce()
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RegistrationListResultVO{}, errors.New("no institution context")
		}
		return model.RegistrationListResultVO{}, err
	}
	return svc.repo.PageRegistrationList(context.Background(), instID, query)
}
