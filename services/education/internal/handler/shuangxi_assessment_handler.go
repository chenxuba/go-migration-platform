package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/service"
)

type shuangxiAssessmentDraftSaveRequest struct {
	ID             int64                      `json:"id,omitempty"`
	StudentID      int64                      `json:"studentId,omitempty"`
	StudentName    string                     `json:"studentName,omitempty"`
	ExaminerName   string                     `json:"examinerName,omitempty"`
	Remark         string                     `json:"remark,omitempty"`
	BirthDate      string                     `json:"birthDate,omitempty"`
	AssessmentDate string                     `json:"assessmentDate,omitempty"`
	ItemScores     map[int]int                `json:"itemScores,omitempty"`
	ItemScoreList  []shuangxiItemScoreRequest `json:"itemScoreList,omitempty"`
}

type shuangxiItemScoreRequest struct {
	ItemNo int `json:"itemNo"`
	Score  int `json:"score"`
}

type shuangxiAssessmentDraftItemSaveRequest struct {
	DraftID int64 `json:"draftId"`
	ItemNo  int   `json:"itemNo"`
	Score   *int  `json:"score"`
}

type shuangxiAssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
}

func (handler *Handler) shuangxiAAssessmentFormTemplateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetShuangxiAAssessmentFormTemplateSummary(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveShuangxiAAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req shuangxiAssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SaveShuangxiAAssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveShuangxiAAssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req shuangxiAssessmentDraftItemSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.DraftID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "draftId is required", ctx.RequestID)
		return
	}
	if req.ItemNo <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "itemNo is required", ctx.RequestID)
		return
	}
	if req.Score == nil {
		httpx.WriteError(w, http.StatusBadRequest, "score is required", ctx.RequestID)
		return
	}
	result, err := handler.service.SaveShuangxiAAssessmentDraftItem(claims.UserID, service.ShuangxiAAssessmentDraftItemSaveInput{
		DraftID: req.DraftID,
		ItemNo:  req.ItemNo,
		Score:   req.Score,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetShuangxiAAssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) shuangxiAAssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := decodeAssessmentDraftPageQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageShuangxiAAssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitShuangxiAAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req shuangxiAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitShuangxiAAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req shuangxiAssessmentDraftSaveRequest) toDraftSaveInput() (service.ShuangxiAAssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.ShuangxiAAssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.ShuangxiAAssessmentDraftSaveInput{}, err
	}
	itemScores, err := normalizeShuangxiItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return service.ShuangxiAAssessmentDraftSaveInput{}, err
	}
	return service.ShuangxiAAssessmentDraftSaveInput{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ItemScores:     itemScores,
		InputSnapshot:  req.normalizedSnapshot(itemScores),
	}, nil
}

func normalizeShuangxiItemScores(itemScores map[int]int, itemScoreList []shuangxiItemScoreRequest) (map[int]int, error) {
	normalized := make(map[int]int, len(itemScores)+len(itemScoreList))
	for itemNo, score := range itemScores {
		if itemNo <= 0 {
			return nil, strconv.ErrSyntax
		}
		normalized[itemNo] = score
	}
	for _, item := range itemScoreList {
		if item.ItemNo <= 0 {
			return nil, strconv.ErrSyntax
		}
		normalized[item.ItemNo] = item.Score
	}
	return normalized, nil
}

func (req shuangxiAssessmentDraftSaveRequest) normalizedSnapshot(itemScores map[int]int) any {
	return struct {
		ID             int64                      `json:"id,omitempty"`
		StudentID      int64                      `json:"studentId,omitempty"`
		StudentName    string                     `json:"studentName,omitempty"`
		ExaminerName   string                     `json:"examinerName,omitempty"`
		Remark         string                     `json:"remark,omitempty"`
		BirthDate      string                     `json:"birthDate,omitempty"`
		AssessmentDate string                     `json:"assessmentDate,omitempty"`
		ItemScores     map[int]int                `json:"itemScores,omitempty"`
		ItemScoreList  []shuangxiItemScoreRequest `json:"itemScoreList,omitempty"`
	}{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      strings.TrimSpace(req.BirthDate),
		AssessmentDate: strings.TrimSpace(req.AssessmentDate),
		ItemScores:     itemScores,
		ItemScoreList:  shuangxiItemScoreListFromMap(itemScores),
	}
}

func shuangxiItemScoreListFromMap(itemScores map[int]int) []shuangxiItemScoreRequest {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]shuangxiItemScoreRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, shuangxiItemScoreRequest{
			ItemNo: itemNo,
			Score:  itemScores[itemNo],
		})
	}
	return out
}

func (handler *Handler) shuangxiAAssessmentFormTemplateItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	itemNo, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("itemNo")))
	if err != nil || itemNo <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid itemNo", ctx.RequestID)
		return
	}
	result, err := handler.service.GetShuangxiAAssessmentFormTemplateItem(r.Context(), itemNo)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}
