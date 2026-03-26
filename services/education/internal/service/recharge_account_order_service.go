package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetStudentDetailView(userID, studentID int64) (model.StudentDetailView, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StudentDetailView{}, errors.New("no institution context")
		}
		return model.StudentDetailView{}, err
	}
	return svc.repo.GetStudentDetailView(context.Background(), instID, studentID)
}

func (svc *Service) GetRechargeAccountByStudent(userID int64, studentIDRaw string) (model.RechargeAccountByStudent, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountByStudent{}, errors.New("no institution context")
		}
		return model.RechargeAccountByStudent{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountByStudent{}, errors.New("no institution user context")
		}
		return model.RechargeAccountByStudent{}, err
	}
	studentID, err := strconv.ParseInt(strings.TrimSpace(studentIDRaw), 10, 64)
	if err != nil || studentID <= 0 {
		return model.RechargeAccountByStudent{}, errors.New("studentId不能为空")
	}
	return svc.repo.GetRechargeAccountByStudent(context.Background(), instID, studentID, instUserID)
}

func (svc *Service) GetRechargeAccountByID(userID int64, rechargeAccountIDRaw string) (model.RechargeAccountByStudent, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountByStudent{}, errors.New("no institution context")
		}
		return model.RechargeAccountByStudent{}, err
	}
	rechargeAccountID, err := strconv.ParseInt(strings.TrimSpace(rechargeAccountIDRaw), 10, 64)
	if err != nil || rechargeAccountID <= 0 {
		return model.RechargeAccountByStudent{}, errors.New("rechargeAccountId不能为空")
	}
	item, err := svc.repo.GetRechargeAccountByID(context.Background(), instID, rechargeAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountByStudent{}, errors.New("储值账户不存在")
		}
		return model.RechargeAccountByStudent{}, err
	}
	return item, nil
}

func (svc *Service) CreateRechargeAccountOrder(userID int64, dto model.CreateRechargeAccountOrderDTO) (model.RechargeAccountOrderCreateResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountOrderCreateResult{}, errors.New("no institution context")
		}
		return model.RechargeAccountOrderCreateResult{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountOrderCreateResult{}, errors.New("no institution user context")
		}
		return model.RechargeAccountOrderCreateResult{}, err
	}
	orderID, err := svc.repo.CreateRechargeAccountOrder(context.Background(), instID, instUserID, dto)
	if err != nil {
		return model.RechargeAccountOrderCreateResult{}, err
	}
	return model.RechargeAccountOrderCreateResult{
		ID:   strconv.FormatInt(orderID, 10),
		Name: "",
	}, nil
}

func (svc *Service) CreateRechargeAccountRefundOrder(userID int64, dto model.CreateRechargeAccountOrderDTO) (model.RechargeAccountOrderCreateResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountOrderCreateResult{}, errors.New("no institution context")
		}
		return model.RechargeAccountOrderCreateResult{}, err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountOrderCreateResult{}, errors.New("no institution user context")
		}
		return model.RechargeAccountOrderCreateResult{}, err
	}
	orderID, err := svc.repo.CreateRechargeAccountRefundOrder(context.Background(), instID, instUserID, dto)
	if err != nil {
		return model.RechargeAccountOrderCreateResult{}, err
	}
	return model.RechargeAccountOrderCreateResult{
		ID:   strconv.FormatInt(orderID, 10),
		Name: "",
	}, nil
}

func (svc *Service) GetRechargeAccountOrderDetail(userID int64, query model.RechargeAccountOrderDetailQuery) (model.RechargeAccountOrderDetail, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RechargeAccountOrderDetail{}, errors.New("no institution context")
		}
		return model.RechargeAccountOrderDetail{}, err
	}
	orderID, _ := strconv.ParseInt(strings.TrimSpace(query.RechargeAccountOrderID), 10, 64)
	saleOrderID, _ := strconv.ParseInt(strings.TrimSpace(query.SaleOrderID), 10, 64)
	if orderID <= 0 && saleOrderID <= 0 {
		return model.RechargeAccountOrderDetail{}, errors.New("rechargeAccountOrderId或saleOrderId不能为空")
	}
	return svc.repo.GetRechargeAccountOrderDetail(context.Background(), instID, orderID, saleOrderID)
}

func (svc *Service) PayOrderBySchoolPal(userID int64, dto model.PayOrderBySchoolPalDTO) (string, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}
	instUserID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution user context")
		}
		return "", err
	}
	billFlowID, err := svc.repo.PayRechargeAccountOrderBySchoolPal(context.Background(), instID, instUserID, dto)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(billFlowID, 10), nil
}
