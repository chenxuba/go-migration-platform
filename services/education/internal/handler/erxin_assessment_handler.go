package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/erxinscore"
	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/service"
)

type erxinScoreRequest struct {
	BirthDate      string                 `json:"birthDate"`
	AssessmentDate string                 `json:"assessmentDate"`
	ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
}

type erxinItemPassRequest struct {
	ItemNo int  `json:"itemNo"`
	Passed bool `json:"passed"`
}

type erxinAssessmentDraftSaveRequest struct {
	ID             int64                  `json:"id,omitempty"`
	StudentID      int64                  `json:"studentId,omitempty"`
	StudentName    string                 `json:"studentName,omitempty"`
	ExaminerName   string                 `json:"examinerName,omitempty"`
	Remark         string                 `json:"remark,omitempty"`
	BirthDate      string                 `json:"birthDate,omitempty"`
	AssessmentDate string                 `json:"assessmentDate,omitempty"`
	ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
}

type erxinAssessmentRecordCreateRequest struct {
	ID             int64                  `json:"id,omitempty"`
	StudentID      int64                  `json:"studentId,omitempty"`
	StudentName    string                 `json:"studentName,omitempty"`
	ExaminerName   string                 `json:"examinerName,omitempty"`
	Remark         string                 `json:"remark,omitempty"`
	BirthDate      string                 `json:"birthDate"`
	AssessmentDate string                 `json:"assessmentDate"`
	ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
	ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
}

type erxinAssessmentDraftItemSaveRequest struct {
	DraftID int64 `json:"draftId"`
	ItemNo  int   `json:"itemNo"`
	Passed  *bool `json:"passed"`
}

type erxinAssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
}

func (handler *Handler) scoreERXin(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScoreERXin(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinAssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SaveERXinAssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) saveERXinAssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req erxinAssessmentDraftItemSaveRequest
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
	if req.Passed == nil {
		httpx.WriteError(w, http.StatusBadRequest, "passed is required", ctx.RequestID)
		return
	}

	result, err := handler.service.SaveERXinAssessmentDraftItem(claims.UserID, service.ERXinAssessmentDraftItemSaveInput{
		DraftID: req.DraftID,
		ItemNo:  req.ItemNo,
		Passed:  req.Passed,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PageERXinAssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteERXinAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitERXinAssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitERXinAssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreateERXinAssessmentRecord(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updateERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	input, err := req.toRecordSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.UpdateERXinAssessmentRecord(claims.UserID, req.ID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordReport(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentReport(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := decodeAssessmentRecordPageQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.PageERXinAssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentRecordCategoryStats(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query, err := decodeAssessmentRecordPageQuery(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SummarizeERXinAssessmentRecordCategories(claims.UserID, query.QueryModel)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deleteERXinAssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req erxinAssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeleteERXinAssessmentRecord(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinAssessmentFormTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetERXinAssessmentFormTemplateSummary()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) erxinAssessmentFormTemplateItem(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetERXinAssessmentFormTemplateItem(itemNo)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (req erxinScoreRequest) toAssessmentInput() (erxinscore.AssessmentInput, error) {
	birthDate, err := parseERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	assessmentDate, err := parseERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	itemPasses, err := normalizeERXinItemPasses(req.ItemPasses, req.ItemPassList)
	if err != nil {
		return erxinscore.AssessmentInput{}, err
	}
	if len(itemPasses) == 0 {
		return erxinscore.AssessmentInput{}, fmt.Errorf("itemPasses or itemPassList is required")
	}
	return erxinscore.AssessmentInput{
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ItemPasses:     itemPasses,
	}, nil
}

func (req erxinAssessmentDraftSaveRequest) toDraftSaveInput() (service.ERXinAssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalERXinDate(req.BirthDate, "birthDate")
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalERXinDate(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	itemPasses, err := normalizeERXinItemPasses(req.ItemPasses, req.ItemPassList)
	if err != nil {
		return service.ERXinAssessmentDraftSaveInput{}, err
	}
	return service.ERXinAssessmentDraftSaveInput{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      birthDate,
		AssessmentDate: assessmentDate,
		ItemPasses:     itemPasses,
		InputSnapshot:  req.normalizedSnapshot(itemPasses),
	}, nil
}

func (req erxinAssessmentRecordCreateRequest) toRecordSaveInput() (service.ERXinAssessmentRecordSaveInput, error) {
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		return service.ERXinAssessmentRecordSaveInput{}, err
	}
	return service.ERXinAssessmentRecordSaveInput{
		StudentID:     req.StudentID,
		StudentName:   strings.TrimSpace(req.StudentName),
		ExaminerName:  strings.TrimSpace(req.ExaminerName),
		Remark:        strings.TrimSpace(req.Remark),
		ScoreInput:    scoreInput,
		InputSnapshot: req.normalizedSnapshot(scoreInput.ItemPasses),
	}, nil
}

func (req erxinAssessmentRecordCreateRequest) toScoreRequest() erxinScoreRequest {
	return erxinScoreRequest{
		BirthDate:      req.BirthDate,
		AssessmentDate: req.AssessmentDate,
		ItemPasses:     req.ItemPasses,
		ItemPassList:   req.ItemPassList,
	}
}

func (req erxinAssessmentDraftSaveRequest) normalizedSnapshot(itemPasses map[int]bool) any {
	return struct {
		ID             int64                  `json:"id,omitempty"`
		StudentID      int64                  `json:"studentId,omitempty"`
		StudentName    string                 `json:"studentName,omitempty"`
		ExaminerName   string                 `json:"examinerName,omitempty"`
		Remark         string                 `json:"remark,omitempty"`
		BirthDate      string                 `json:"birthDate,omitempty"`
		AssessmentDate string                 `json:"assessmentDate,omitempty"`
		ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
		ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
	}{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      strings.TrimSpace(req.BirthDate),
		AssessmentDate: strings.TrimSpace(req.AssessmentDate),
		ItemPasses:     itemPasses,
		ItemPassList:   erxinItemPassListFromMap(itemPasses),
	}
}

func (req erxinAssessmentRecordCreateRequest) normalizedSnapshot(itemPasses map[int]bool) any {
	return struct {
		ID             int64                  `json:"id,omitempty"`
		StudentID      int64                  `json:"studentId,omitempty"`
		StudentName    string                 `json:"studentName,omitempty"`
		ExaminerName   string                 `json:"examinerName,omitempty"`
		Remark         string                 `json:"remark,omitempty"`
		BirthDate      string                 `json:"birthDate"`
		AssessmentDate string                 `json:"assessmentDate"`
		ItemPasses     map[int]bool           `json:"itemPasses,omitempty"`
		ItemPassList   []erxinItemPassRequest `json:"itemPassList,omitempty"`
	}{
		ID:             req.ID,
		StudentID:      req.StudentID,
		StudentName:    strings.TrimSpace(req.StudentName),
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		Remark:         strings.TrimSpace(req.Remark),
		BirthDate:      strings.TrimSpace(req.BirthDate),
		AssessmentDate: strings.TrimSpace(req.AssessmentDate),
		ItemPasses:     itemPasses,
		ItemPassList:   erxinItemPassListFromMap(itemPasses),
	}
}

func parseERXinDate(raw, field string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}
	for _, layout := range []string{"2006-01-02", "2006/01/02", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("%s format should be YYYY-MM-DD", field)
}

func parseOptionalERXinDate(raw, field string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parsed, err := parseERXinDate(raw, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeERXinItemPasses(itemPasses map[int]bool, itemPassList []erxinItemPassRequest) (map[int]bool, error) {
	normalized := make(map[int]bool, len(itemPasses)+len(itemPassList))
	for itemNo, passed := range itemPasses {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemPasses contains invalid itemNo %d", itemNo)
		}
		normalized[itemNo] = passed
	}
	for _, item := range itemPassList {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("itemPassList contains invalid itemNo %d", item.ItemNo)
		}
		normalized[item.ItemNo] = item.Passed
	}
	return normalized, nil
}

func erxinItemPassListFromMap(itemPasses map[int]bool) []erxinItemPassRequest {
	if len(itemPasses) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemPasses))
	for itemNo := range itemPasses {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]erxinItemPassRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, erxinItemPassRequest{ItemNo: itemNo, Passed: itemPasses[itemNo]})
	}
	return out
}
