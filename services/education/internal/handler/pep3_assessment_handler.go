package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/httpx"
	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/pkg/tenant"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/service"
)

type pep3ScoreRequest struct {
	BirthDate         string                 `json:"birthDate"`
	AssessmentDate    string                 `json:"assessmentDate"`
	ItemScores        map[int]int            `json:"itemScores,omitempty"`
	ItemScoreList     []pep3ItemScoreRequest `json:"itemScoreList,omitempty"`
	RawScores         map[string]int         `json:"rawScores,omitempty"`
	RawScoreList      []pep3RawScoreRequest  `json:"rawScoreList,omitempty"`
	AllowMissingItems bool                   `json:"allowMissingItems,omitempty"`
}

type pep3ItemScoreRequest struct {
	ItemNo int `json:"itemNo"`
	Score  int `json:"score"`
}

type pep3RawScoreRequest struct {
	ScaleCode string `json:"scaleCode"`
	RawScore  int    `json:"rawScore"`
}

type pep3AssessmentRecordCreateRequest struct {
	ID                  int64                                `json:"id,omitempty"`
	StudentID           int64                                `json:"studentId,omitempty"`
	StudentName         string                               `json:"studentName,omitempty"`
	ExaminerName        string                               `json:"examinerName,omitempty"`
	Remark              string                               `json:"remark,omitempty"`
	BirthDate           string                               `json:"birthDate"`
	AssessmentDate      string                               `json:"assessmentDate"`
	ItemScores          map[int]int                          `json:"itemScores,omitempty"`
	ItemScoreList       []pep3ItemScoreRequest               `json:"itemScoreList,omitempty"`
	RawScores           map[string]int                       `json:"rawScores,omitempty"`
	RawScoreList        []pep3RawScoreRequest                `json:"rawScoreList,omitempty"`
	ItemRecordValues    map[int]map[string]any               `json:"itemRecordValues,omitempty"`
	ItemRecordValueList []pep3ItemRecordValueRequest         `json:"itemRecordValueList,omitempty"`
	AllowMissingItems   bool                                 `json:"allowMissingItems,omitempty"`
	CaregiverReport     *model.PEP3CaregiverReportSubmission `json:"caregiverReport,omitempty"`
}

type pep3AssessmentRecordConfigUpdateRequest struct {
	ID             int64  `json:"id"`
	ExaminerName   string `json:"examinerName"`
	AssessmentDate string `json:"assessmentDate"`
}

type pep3AssessmentDraftSaveRequest struct {
	ID                  int64                                `json:"id,omitempty"`
	StudentID           int64                                `json:"studentId,omitempty"`
	StudentName         string                               `json:"studentName,omitempty"`
	ExaminerName        string                               `json:"examinerName,omitempty"`
	Remark              string                               `json:"remark,omitempty"`
	BirthDate           string                               `json:"birthDate,omitempty"`
	AssessmentDate      string                               `json:"assessmentDate,omitempty"`
	ItemScores          map[int]int                          `json:"itemScores,omitempty"`
	ItemScoreList       []pep3ItemScoreRequest               `json:"itemScoreList,omitempty"`
	RawScores           map[string]int                       `json:"rawScores,omitempty"`
	RawScoreList        []pep3RawScoreRequest                `json:"rawScoreList,omitempty"`
	ItemRecordValues    map[int]map[string]any               `json:"itemRecordValues,omitempty"`
	ItemRecordValueList []pep3ItemRecordValueRequest         `json:"itemRecordValueList,omitempty"`
	AllowMissingItems   bool                                 `json:"allowMissingItems,omitempty"`
	CaregiverReport     *model.PEP3CaregiverReportSubmission `json:"caregiverReport,omitempty"`
}

type pep3AssessmentDraftDeleteRequest struct {
	ID int64 `json:"id"`
}

type pep3AssessmentDraftItemSaveRequest struct {
	DraftID      int64          `json:"draftId"`
	ItemNo       int            `json:"itemNo"`
	Score        *int           `json:"score,omitempty"`
	RecordValues map[string]any `json:"recordValues,omitempty"`
}

type pep3ItemRecordValueRequest struct {
	ItemNo   int    `json:"itemNo"`
	FieldKey string `json:"fieldKey"`
	Value    any    `json:"value"`
}

func (handler *Handler) scorePEP3(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3ScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.ScorePEP3(input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentFormTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3AssessmentFormTemplate()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) scaleLibrary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetScaleLibrary(claims.UserID, model.ScaleLibraryQuery{
		Keyword:  r.URL.Query().Get("keyword"),
		Category: r.URL.Query().Get("category"),
		Scenario: r.URL.Query().Get("scenario"),
		Status:   r.URL.Query().Get("status"),
		AgeScope: r.URL.Query().Get("ageScope"),
		Duration: r.URL.Query().Get("duration"),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) scaleCategories(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.ListScaleCategoryOptions(claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) scaleAssessmentStudentCandidates(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	query := r.URL.Query()
	pageIndex, _ := strconv.Atoi(query.Get("pageIndex"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	var studentStatus *int
	if rawStatus := strings.TrimSpace(query.Get("studentStatus")); rawStatus != "" {
		if parsedStatus, err := strconv.Atoi(rawStatus); err == nil {
			studentStatus = &parsedStatus
		}
	}
	result, err := handler.service.ListScaleAssessmentStudentCandidates(claims.UserID, model.ScaleAssessmentStudentCandidateQuery{
		ScaleCode:     query.Get("scaleCode"),
		Keyword:       query.Get("keyword"),
		StudentStatus: studentStatus,
		PageIndex:     pageIndex,
		PageSize:      pageSize,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentFormTemplateSummary(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	if _, ok := handler.requireAuth(w, r, ctx); !ok {
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3AssessmentFormTemplateSummary()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentFormTemplateItem(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentFormTemplateItem(itemNo)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) savePEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3AssessmentDraftSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	input, err := req.toDraftSaveInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.SavePEP3AssessmentDraft(claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentDraftDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentDraft(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) savePEP3AssessmentDraftItem(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3AssessmentDraftItemSaveRequest
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
	if req.Score != nil && (*req.Score < 0 || *req.Score > 2) {
		httpx.WriteError(w, http.StatusBadRequest, "score must be 0, 1, or 2", ctx.RequestID)
		return
	}

	result, err := handler.service.SavePEP3AssessmentDraftItem(claims.UserID, service.PEP3AssessmentDraftItemSaveInput{
		DraftID:       req.DraftID,
		ItemNo:        req.ItemNo,
		Score:         req.Score,
		RecordValues:  normalizePEP3RecordValueMap(req.RecordValues),
		RecordTouched: req.RecordValues != nil,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentDraftsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PagePEP3AssessmentDrafts(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deletePEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeletePEP3AssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) submitPEP3AssessmentDraft(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.SubmitPEP3AssessmentDraft(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) createPEP3AssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}

	var req pep3AssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	itemRecordValues, err := normalizePEP3ItemRecordValues(req.ItemRecordValues, req.ItemRecordValueList)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.CreatePEP3AssessmentRecord(claims.UserID, service.PEP3AssessmentRecordSaveInput{
		StudentID:        req.StudentID,
		StudentName:      strings.TrimSpace(req.StudentName),
		ExaminerName:     strings.TrimSpace(req.ExaminerName),
		Remark:           strings.TrimSpace(req.Remark),
		ScoreInput:       scoreInput,
		ItemRecordValues: itemRecordValues,
		InputSnapshot:    req.normalizedSnapshot(scoreInput.ItemScores, scoreInput.RawScores, itemRecordValues),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updatePEP3AssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentRecordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	scoreReq := req.toScoreRequest()
	scoreInput, err := scoreReq.toAssessmentInput()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	itemRecordValues, err := normalizePEP3ItemRecordValues(req.ItemRecordValues, req.ItemRecordValueList)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.UpdatePEP3AssessmentRecord(claims.UserID, req.ID, service.PEP3AssessmentRecordSaveInput{
		StudentID:        req.StudentID,
		StudentName:      strings.TrimSpace(req.StudentName),
		ExaminerName:     strings.TrimSpace(req.ExaminerName),
		Remark:           strings.TrimSpace(req.Remark),
		ScoreInput:       scoreInput,
		ItemRecordValues: itemRecordValues,
		InputSnapshot:    req.normalizedSnapshot(scoreInput.ItemScores, scoreInput.RawScores, itemRecordValues),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) updatePEP3AssessmentRecordConfig(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentRecordConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "id is required", ctx.RequestID)
		return
	}
	assessmentDate, err := parsePEP3Date(req.AssessmentDate, "assessmentDate")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	result, err := handler.service.UpdatePEP3AssessmentRecordConfig(claims.UserID, req.ID, service.PEP3AssessmentRecordConfigInput{
		ExaminerName:   strings.TrimSpace(req.ExaminerName),
		AssessmentDate: assessmentDate,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordDetail(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentRecord(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordReport(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentReport(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordReportInterpretation(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, http.StatusBadRequest, "invalid assessment record id", ctx.RequestID)
		return
	}
	result, err := handler.service.GetPEP3ReportInterpretation(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordReportInterpretationPDF(w http.ResponseWriter, r *http.Request) {
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
		httpx.WriteError(w, http.StatusBadRequest, "invalid assessment record id", ctx.RequestID)
		return
	}
	filename, content, err := handler.service.GeneratePEP3ReportInterpretationPDF(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) pep3AssessmentRecordReportInterpretationAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.ERXinReportInterpretationGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GeneratePEP3ReportInterpretation(r.Context(), claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordReportInterpretationAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.ERXinReportInterpretationGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, "streaming is not supported", ctx.RequestID)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(event string, payload any) error {
		return writePEP3IEPPlanSSE(w, flusher, event, payload)
	}
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取PEP-3评估结果"}); err != nil {
		return
	}
	result, err := handler.service.GeneratePEP3ReportInterpretationStream(r.Context(), claims.UserID, req.ID, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) pep3AssessmentRecordBooklet(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.GetPEP3AssessmentBooklet(claims.UserID, id)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordBookletPDF(w http.ResponseWriter, r *http.Request) {
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
	exportDimension := strings.TrimSpace(r.URL.Query().Get("dimension"))
	filename, content, err := handler.service.GeneratePEP3AssessmentBookletPDF(claims.UserID, id, exportDimension)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename*=UTF-8''"+url.QueryEscape(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) pep3AssessmentRecordIEPPlanWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var recordID int64
	var durationMonths int
	var plan *model.PEP3IEPPlanAIResult
	if r.Method == http.MethodPost {
		var req model.PEP3IEPPlanWordExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
			return
		}
		recordID = req.ID
		durationMonths = req.DurationMonths
		plan = req.Plan
	} else if rawID := strings.TrimSpace(r.URL.Query().Get("id")); rawID != "" {
		parsedID, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || parsedID < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
			return
		}
		recordID = parsedID
	}
	if r.Method == http.MethodGet {
		durationMonths, _ = strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("duration")))
	}
	var fileName string
	var contentType string
	var content []byte
	var err error
	if plan != nil {
		fileName, contentType, content, err = handler.service.ExportPEP3IEPPlanWordFromAIResult(claims.UserID, recordID, *plan, durationMonths)
	} else {
		fileName, contentType, content, err = handler.service.ExportPEP3IEPPlanWord(claims.UserID, recordID, durationMonths)
	}
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) pep3AssessmentRecordExecutionPlanWord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanWordExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	fileName, contentType, content, err := handler.service.ExportPEP3ExecutionPlanWord(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.QueryEscape(fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (handler *Handler) pep3AssessmentRecordExecutionPlanAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GeneratePEP3ExecutionPlanWithAI(r.Context(), claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordExecutionPlanAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, "streaming is not supported", ctx.RequestID)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(event string, payload any) error {
		return writePEP3IEPPlanSSE(w, flusher, event, payload)
	}
	planTypeLabel := "执行计划"
	switch strings.ToLower(strings.TrimSpace(req.PlanType)) {
	case "monthly":
		planTypeLabel = "月度计划"
	case "weekly":
		planTypeLabel = "周计划"
	}
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在准备" + planTypeLabel + "生成上下文"}); err != nil {
		return
	}
	result, err := handler.service.GeneratePEP3ExecutionPlanWithAIStream(r.Context(), claims.UserID, req, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func (handler *Handler) pep3AssessmentRecordExecutionPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	durationMonths := parsePEP3IEPPlanDurationQuery(r)
	result, err := handler.service.GetPEP3ExecutionPlans(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordExecutionPlanSave(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3ExecutionPlanSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SavePEP3ExecutionPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordIEPPlanAI(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.GeneratePEP3IEPPlanWithAI(claims.UserID, req.ID, req.DurationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordIEPPlanDetail(w http.ResponseWriter, r *http.Request) {
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
	durationMonths := parsePEP3IEPPlanDurationQuery(r)
	result, err := handler.service.GetPEP3IEPPlan(claims.UserID, id, durationMonths)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func parsePEP3IEPPlanDurationQuery(r *http.Request) int {
	raw := strings.TrimSpace(r.URL.Query().Get("durationMonths"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("duration"))
	}
	durationMonths, _ := strconv.Atoi(raw)
	return durationMonths
}

func (handler *Handler) pep3AssessmentRecordIEPPlanSave(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	result, err := handler.service.SavePEP3IEPPlan(claims.UserID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordIEPPlanAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req model.PEP3IEPPlanGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.WriteError(w, http.StatusInternalServerError, "streaming is not supported", ctx.RequestID)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(event string, payload any) error {
		return writePEP3IEPPlanSSE(w, flusher, event, payload)
	}
	if err := writeEvent("status", map[string]any{"type": "status", "message": "正在读取评估和训练记录"}); err != nil {
		return
	}
	result, err := handler.service.GeneratePEP3IEPPlanWithAIStream(r.Context(), claims.UserID, req.ID, req.DurationMonths, func(text string) error {
		return writeEvent("delta", map[string]any{"type": "delta", "text": text})
	})
	if err != nil {
		_ = writeEvent("error", map[string]any{"type": "error", "message": err.Error()})
		return
	}
	_ = writeEvent("done", map[string]any{"type": "done", "data": result})
}

func writePEP3IEPPlanSSE(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func (handler *Handler) pep3AssessmentRecordsPage(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.PagePEP3AssessmentRecords(claims.UserID, query)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) pep3AssessmentRecordCategoryStats(w http.ResponseWriter, r *http.Request) {
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
	result, err := handler.service.SummarizePEP3AssessmentRecordCategories(claims.UserID, query.QueryModel)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

func (handler *Handler) deletePEP3AssessmentRecord(w http.ResponseWriter, r *http.Request) {
	ctx := tenant.FromContext(r.Context())
	claims, ok := handler.requireAuth(w, r, ctx)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", ctx.RequestID)
		return
	}
	var req pep3AssessmentDraftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", ctx.RequestID)
		return
	}
	if req.ID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id", ctx.RequestID)
		return
	}
	result, err := handler.service.DeletePEP3AssessmentRecord(claims.UserID, req.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), ctx.RequestID)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result, ctx.RequestID)
}

type assessmentPageRequestPayload struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Current   int `json:"current"`
	Size      int `json:"size"`
	SkipCount int `json:"skipCount"`
}

type assessmentRecordPagePayload struct {
	PageRequestModel    assessmentPageRequestPayload `json:"pageRequestModel"`
	QueryModel          assessmentRecordQueryPayload `json:"queryModel"`
	PageIndex           int                          `json:"pageIndex"`
	PageSize            int                          `json:"pageSize"`
	Current             int                          `json:"current"`
	Size                int                          `json:"size"`
	AssessmentCode      string                       `json:"assessmentCode"`
	ScaleCategory       string                       `json:"scaleCategory"`
	StudentID           int64Payload                 `json:"studentId"`
	SearchKey           string                       `json:"searchKey"`
	AssessmentDateBegin string                       `json:"assessmentDateBegin"`
	AssessmentDateEnd   string                       `json:"assessmentDateEnd"`
}

type assessmentDraftPagePayload struct {
	PageRequestModel    assessmentPageRequestPayload `json:"pageRequestModel"`
	QueryModel          assessmentDraftQueryPayload  `json:"queryModel"`
	PageIndex           int                          `json:"pageIndex"`
	PageSize            int                          `json:"pageSize"`
	Current             int                          `json:"current"`
	Size                int                          `json:"size"`
	AssessmentCode      string                       `json:"assessmentCode"`
	StudentID           int64Payload                 `json:"studentId"`
	SearchKey           string                       `json:"searchKey"`
	AssessmentDateBegin string                       `json:"assessmentDateBegin"`
	AssessmentDateEnd   string                       `json:"assessmentDateEnd"`
	Status              string                       `json:"status"`
	LatestOnly          bool                         `json:"latestOnly"`
}

type assessmentRecordQueryPayload struct {
	AssessmentCode      string       `json:"assessmentCode"`
	ScaleCategory       string       `json:"scaleCategory"`
	StudentID           int64Payload `json:"studentId"`
	SearchKey           string       `json:"searchKey"`
	AssessmentDateBegin string       `json:"assessmentDateBegin"`
	AssessmentDateEnd   string       `json:"assessmentDateEnd"`
}

type assessmentDraftQueryPayload struct {
	AssessmentCode      string       `json:"assessmentCode"`
	StudentID           int64Payload `json:"studentId"`
	SearchKey           string       `json:"searchKey"`
	AssessmentDateBegin string       `json:"assessmentDateBegin"`
	AssessmentDateEnd   string       `json:"assessmentDateEnd"`
	Status              string       `json:"status"`
	LatestOnly          bool         `json:"latestOnly"`
}

type int64Payload struct {
	Value int64
	Valid bool
}

func (value *int64Payload) UnmarshalJSON(raw []byte) error {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" || text == `""` {
		return nil
	}
	if strings.HasPrefix(text, `"`) {
		var str string
		if err := json.Unmarshal(raw, &str); err != nil {
			return err
		}
		parsed, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err == nil && parsed > 0 {
			value.Value = parsed
			value.Valid = true
		}
		return nil
	}
	var number float64
	if err := json.Unmarshal(raw, &number); err != nil {
		return err
	}
	if number > 0 {
		value.Value = int64(number)
		value.Valid = true
	}
	return nil
}

func (value int64Payload) ptr() *int64 {
	if !value.Valid || value.Value <= 0 {
		return nil
	}
	out := value.Value
	return &out
}

func decodeAssessmentRecordPageQuery(r *http.Request) (model.AssessmentRecordPageQueryDTO, error) {
	var query model.AssessmentRecordPageQueryDTO
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return query, err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return query, nil
	}
	var payload assessmentRecordPagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return query, err
	}
	query.PageRequestModel = payload.PageRequestModel.toPageRequestModel()
	query.QueryModel = payload.QueryModel.toRecordQueryModel()
	mergeAssessmentRecordFlatQuery(&query, payload)
	return query, nil
}

func decodeAssessmentDraftPageQuery(r *http.Request) (model.AssessmentDraftPageQueryDTO, error) {
	var query model.AssessmentDraftPageQueryDTO
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return query, err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return query, nil
	}
	var payload assessmentDraftPagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return query, err
	}
	query.PageRequestModel = payload.PageRequestModel.toPageRequestModel()
	query.QueryModel = payload.QueryModel.toDraftQueryModel()
	mergeAssessmentDraftFlatQuery(&query, payload)
	return query, nil
}

func (payload assessmentPageRequestPayload) toPageRequestModel() model.PageRequestModel {
	return model.PageRequestModel{
		PageIndex: firstPositiveInt(payload.PageIndex, payload.Current),
		PageSize:  firstPositiveInt(payload.PageSize, payload.Size),
		SkipCount: payload.SkipCount,
	}
}

func (payload assessmentRecordQueryPayload) toRecordQueryModel() model.AssessmentRecordQueryModel {
	return model.AssessmentRecordQueryModel{
		AssessmentCode:      strings.TrimSpace(payload.AssessmentCode),
		ScaleCategory:       strings.TrimSpace(payload.ScaleCategory),
		StudentID:           payload.StudentID.ptr(),
		SearchKey:           strings.TrimSpace(payload.SearchKey),
		AssessmentDateBegin: strings.TrimSpace(payload.AssessmentDateBegin),
		AssessmentDateEnd:   strings.TrimSpace(payload.AssessmentDateEnd),
	}
}

func (payload assessmentDraftQueryPayload) toDraftQueryModel() model.AssessmentDraftQueryModel {
	return model.AssessmentDraftQueryModel{
		AssessmentCode:      strings.TrimSpace(payload.AssessmentCode),
		StudentID:           payload.StudentID.ptr(),
		SearchKey:           strings.TrimSpace(payload.SearchKey),
		Status:              strings.TrimSpace(payload.Status),
		AssessmentDateBegin: strings.TrimSpace(payload.AssessmentDateBegin),
		AssessmentDateEnd:   strings.TrimSpace(payload.AssessmentDateEnd),
		LatestOnly:          payload.LatestOnly,
	}
}

func mergeAssessmentRecordFlatQuery(query *model.AssessmentRecordPageQueryDTO, flat assessmentRecordPagePayload) {
	if query.PageRequestModel.PageIndex <= 0 {
		query.PageRequestModel.PageIndex = firstPositiveInt(flat.PageIndex, flat.Current)
	}
	if query.PageRequestModel.PageSize <= 0 {
		query.PageRequestModel.PageSize = firstPositiveInt(flat.PageSize, flat.Size)
	}
	if query.QueryModel.AssessmentCode == "" {
		query.QueryModel.AssessmentCode = strings.TrimSpace(flat.AssessmentCode)
	}
	if query.QueryModel.ScaleCategory == "" {
		query.QueryModel.ScaleCategory = strings.TrimSpace(flat.ScaleCategory)
	}
	if query.QueryModel.StudentID == nil {
		query.QueryModel.StudentID = flat.StudentID.ptr()
	}
	if query.QueryModel.SearchKey == "" {
		query.QueryModel.SearchKey = strings.TrimSpace(flat.SearchKey)
	}
	if query.QueryModel.AssessmentDateBegin == "" {
		query.QueryModel.AssessmentDateBegin = strings.TrimSpace(flat.AssessmentDateBegin)
	}
	if query.QueryModel.AssessmentDateEnd == "" {
		query.QueryModel.AssessmentDateEnd = strings.TrimSpace(flat.AssessmentDateEnd)
	}
}

func mergeAssessmentDraftFlatQuery(query *model.AssessmentDraftPageQueryDTO, flat assessmentDraftPagePayload) {
	if query.PageRequestModel.PageIndex <= 0 {
		query.PageRequestModel.PageIndex = firstPositiveInt(flat.PageIndex, flat.Current)
	}
	if query.PageRequestModel.PageSize <= 0 {
		query.PageRequestModel.PageSize = firstPositiveInt(flat.PageSize, flat.Size)
	}
	if query.QueryModel.AssessmentCode == "" {
		query.QueryModel.AssessmentCode = strings.TrimSpace(flat.AssessmentCode)
	}
	if query.QueryModel.StudentID == nil {
		query.QueryModel.StudentID = flat.StudentID.ptr()
	}
	if query.QueryModel.SearchKey == "" {
		query.QueryModel.SearchKey = strings.TrimSpace(flat.SearchKey)
	}
	if query.QueryModel.Status == "" {
		query.QueryModel.Status = strings.TrimSpace(flat.Status)
	}
	if query.QueryModel.AssessmentDateBegin == "" {
		query.QueryModel.AssessmentDateBegin = strings.TrimSpace(flat.AssessmentDateBegin)
	}
	if query.QueryModel.AssessmentDateEnd == "" {
		query.QueryModel.AssessmentDateEnd = strings.TrimSpace(flat.AssessmentDateEnd)
	}
	if !query.QueryModel.LatestOnly {
		query.QueryModel.LatestOnly = flat.LatestOnly
	}
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func (req pep3AssessmentRecordCreateRequest) toScoreRequest() pep3ScoreRequest {
	return pep3ScoreRequest{
		BirthDate:         req.BirthDate,
		AssessmentDate:    req.AssessmentDate,
		ItemScores:        req.ItemScores,
		ItemScoreList:     req.ItemScoreList,
		RawScores:         req.RawScores,
		RawScoreList:      req.RawScoreList,
		AllowMissingItems: req.AllowMissingItems,
	}
}

func (req pep3AssessmentDraftSaveRequest) toDraftSaveInput() (service.PEP3AssessmentDraftSaveInput, error) {
	birthDate, err := parseOptionalPEP3Date(req.BirthDate, "birthDate")
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	assessmentDate, err := parseOptionalPEP3Date(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	itemScores, err := normalizePEP3ItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	rawScores, err := normalizePEP3RawScores(req.RawScores, req.RawScoreList)
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	itemRecordValues, err := normalizePEP3ItemRecordValues(req.ItemRecordValues, req.ItemRecordValueList)
	if err != nil {
		return service.PEP3AssessmentDraftSaveInput{}, err
	}
	return service.PEP3AssessmentDraftSaveInput{
		ID:                req.ID,
		StudentID:         req.StudentID,
		StudentName:       strings.TrimSpace(req.StudentName),
		ExaminerName:      strings.TrimSpace(req.ExaminerName),
		Remark:            strings.TrimSpace(req.Remark),
		BirthDate:         birthDate,
		AssessmentDate:    assessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		ItemRecordValues:  itemRecordValues,
		AllowMissingItems: req.AllowMissingItems,
		InputSnapshot:     req.normalizedSnapshot(itemScores, rawScores, itemRecordValues),
	}, nil
}

func (req pep3AssessmentDraftSaveRequest) normalizedSnapshot(itemScores map[int]int, rawScores map[string]int, itemRecordValues map[int]map[string]any) any {
	return struct {
		ID                  int64                                `json:"id,omitempty"`
		StudentID           int64                                `json:"studentId,omitempty"`
		StudentName         string                               `json:"studentName,omitempty"`
		ExaminerName        string                               `json:"examinerName,omitempty"`
		Remark              string                               `json:"remark,omitempty"`
		BirthDate           string                               `json:"birthDate,omitempty"`
		AssessmentDate      string                               `json:"assessmentDate,omitempty"`
		ItemScores          map[int]int                          `json:"itemScores,omitempty"`
		ItemScoreList       []pep3ItemScoreRequest               `json:"itemScoreList,omitempty"`
		RawScores           map[string]int                       `json:"rawScores,omitempty"`
		RawScoreList        []pep3RawScoreRequest                `json:"rawScoreList,omitempty"`
		ItemRecordValues    map[int]map[string]any               `json:"itemRecordValues,omitempty"`
		ItemRecordValueList []pep3ItemRecordValueRequest         `json:"itemRecordValueList,omitempty"`
		AllowMissingItems   bool                                 `json:"allowMissingItems,omitempty"`
		CaregiverReport     *model.PEP3CaregiverReportSubmission `json:"caregiverReport,omitempty"`
	}{
		ID:                  req.ID,
		StudentID:           req.StudentID,
		StudentName:         strings.TrimSpace(req.StudentName),
		ExaminerName:        strings.TrimSpace(req.ExaminerName),
		Remark:              strings.TrimSpace(req.Remark),
		BirthDate:           strings.TrimSpace(req.BirthDate),
		AssessmentDate:      strings.TrimSpace(req.AssessmentDate),
		ItemScores:          itemScores,
		ItemScoreList:       itemScoreListFromMap(itemScores),
		RawScores:           rawScores,
		RawScoreList:        rawScoreListFromMap(rawScores),
		ItemRecordValues:    itemRecordValues,
		ItemRecordValueList: itemRecordValueListFromMap(itemRecordValues),
		AllowMissingItems:   req.AllowMissingItems,
		CaregiverReport:     req.CaregiverReport,
	}
}

func (req pep3AssessmentRecordCreateRequest) normalizedSnapshot(itemScores map[int]int, rawScores map[string]int, itemRecordValues map[int]map[string]any) any {
	return struct {
		ID                  int64                                `json:"id,omitempty"`
		StudentID           int64                                `json:"studentId,omitempty"`
		StudentName         string                               `json:"studentName,omitempty"`
		ExaminerName        string                               `json:"examinerName,omitempty"`
		Remark              string                               `json:"remark,omitempty"`
		BirthDate           string                               `json:"birthDate"`
		AssessmentDate      string                               `json:"assessmentDate"`
		ItemScores          map[int]int                          `json:"itemScores,omitempty"`
		ItemScoreList       []pep3ItemScoreRequest               `json:"itemScoreList,omitempty"`
		RawScores           map[string]int                       `json:"rawScores,omitempty"`
		RawScoreList        []pep3RawScoreRequest                `json:"rawScoreList,omitempty"`
		ItemRecordValues    map[int]map[string]any               `json:"itemRecordValues,omitempty"`
		ItemRecordValueList []pep3ItemRecordValueRequest         `json:"itemRecordValueList,omitempty"`
		AllowMissingItems   bool                                 `json:"allowMissingItems,omitempty"`
		CaregiverReport     *model.PEP3CaregiverReportSubmission `json:"caregiverReport,omitempty"`
	}{
		ID:                  req.ID,
		StudentID:           req.StudentID,
		StudentName:         strings.TrimSpace(req.StudentName),
		ExaminerName:        strings.TrimSpace(req.ExaminerName),
		Remark:              strings.TrimSpace(req.Remark),
		BirthDate:           strings.TrimSpace(req.BirthDate),
		AssessmentDate:      strings.TrimSpace(req.AssessmentDate),
		ItemScores:          itemScores,
		ItemScoreList:       itemScoreListFromMap(itemScores),
		RawScores:           rawScores,
		RawScoreList:        rawScoreListFromMap(rawScores),
		ItemRecordValues:    itemRecordValues,
		ItemRecordValueList: itemRecordValueListFromMap(itemRecordValues),
		AllowMissingItems:   req.AllowMissingItems,
		CaregiverReport:     req.CaregiverReport,
	}
}

func (req pep3ScoreRequest) toAssessmentInput() (pep3score.AssessmentInput, error) {
	birthDate, err := parsePEP3Date(req.BirthDate, "birthDate")
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}
	assessmentDate, err := parsePEP3Date(req.AssessmentDate, "assessmentDate")
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}

	itemScores, err := normalizePEP3ItemScores(req.ItemScores, req.ItemScoreList)
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}
	if len(itemScores) == 0 {
		itemScores = nil
	}

	rawScores, err := normalizePEP3RawScores(req.RawScores, req.RawScoreList)
	if err != nil {
		return pep3score.AssessmentInput{}, err
	}
	if len(rawScores) == 0 {
		rawScores = nil
	}

	if len(itemScores) == 0 && len(rawScores) == 0 {
		return pep3score.AssessmentInput{}, fmt.Errorf("itemScores or rawScores is required")
	}

	return pep3score.AssessmentInput{
		BirthDate:         birthDate,
		AssessmentDate:    assessmentDate,
		ItemScores:        itemScores,
		RawScores:         rawScores,
		AllowMissingItems: req.AllowMissingItems,
	}, nil
}

func parsePEP3Date(raw, field string) (time.Time, error) {
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

func parseOptionalPEP3Date(raw, field string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parsed, err := parsePEP3Date(raw, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizePEP3ItemScores(itemScores map[int]int, itemScoreList []pep3ItemScoreRequest) (map[int]int, error) {
	normalized := make(map[int]int, len(itemScores)+len(itemScoreList))
	for itemNo, score := range itemScores {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemScores contains invalid itemNo %d", itemNo)
		}
		normalized[itemNo] = score
	}
	for _, item := range itemScoreList {
		if item.ItemNo <= 0 {
			return nil, fmt.Errorf("itemScoreList contains invalid itemNo %d", item.ItemNo)
		}
		normalized[item.ItemNo] = item.Score
	}
	return normalized, nil
}

func normalizePEP3RawScores(rawScores map[string]int, rawScoreList []pep3RawScoreRequest) (map[string]int, error) {
	normalizedScores := make(map[string]int, len(rawScores)+len(rawScoreList))
	for scaleCode, rawScore := range rawScores {
		normalized := normalizePEP3ScaleCode(scaleCode)
		if normalized == "" {
			return nil, fmt.Errorf("rawScores contains empty scale code")
		}
		normalizedScores[normalized] = rawScore
	}
	for _, item := range rawScoreList {
		normalized := normalizePEP3ScaleCode(item.ScaleCode)
		if normalized == "" {
			return nil, fmt.Errorf("rawScoreList contains empty scaleCode")
		}
		normalizedScores[normalized] = item.RawScore
	}
	return normalizedScores, nil
}

func normalizePEP3ItemRecordValues(itemRecordValues map[int]map[string]any, itemRecordValueList []pep3ItemRecordValueRequest) (map[int]map[string]any, error) {
	normalized := make(map[int]map[string]any, len(itemRecordValues)+len(itemRecordValueList))
	for itemNo, values := range itemRecordValues {
		if itemNo <= 0 {
			return nil, fmt.Errorf("itemRecordValues contains invalid itemNo %d", itemNo)
		}
		for fieldKey, value := range values {
			if err := addPEP3ItemRecordValue(normalized, itemNo, fieldKey, value); err != nil {
				return nil, err
			}
		}
	}
	for _, item := range itemRecordValueList {
		if err := addPEP3ItemRecordValue(normalized, item.ItemNo, item.FieldKey, item.Value); err != nil {
			return nil, err
		}
	}
	return normalized, nil
}

func addPEP3ItemRecordValue(out map[int]map[string]any, itemNo int, fieldKey string, value any) error {
	if itemNo <= 0 {
		return fmt.Errorf("itemRecordValueList contains invalid itemNo %d", itemNo)
	}
	fieldKey = strings.TrimSpace(fieldKey)
	if fieldKey == "" {
		return fmt.Errorf("item %d contains empty item record fieldKey", itemNo)
	}
	if out[itemNo] == nil {
		out[itemNo] = map[string]any{}
	}
	out[itemNo][fieldKey] = normalizePEP3ItemRecordValue(value)
	return nil
}

func normalizePEP3RecordValueMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	normalized := make(map[string]any, len(values))
	for fieldKey, value := range values {
		key := strings.TrimSpace(fieldKey)
		if key == "" {
			continue
		}
		normalizedValue := normalizePEP3ItemRecordValue(value)
		if isEmptyPEP3ItemRecordValue(normalizedValue) {
			continue
		}
		normalized[key] = normalizedValue
	}
	return normalized
}

func isEmptyPEP3ItemRecordValue(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case []string:
		return len(typed) == 0
	default:
		return false
	}
}

func normalizePEP3ItemRecordValue(value any) any {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, normalizePEP3ItemRecordValue(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			out[key] = normalizePEP3ItemRecordValue(item)
		}
		return out
	default:
		return value
	}
}

func itemScoreListFromMap(itemScores map[int]int) []pep3ItemScoreRequest {
	if len(itemScores) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemScores))
	for itemNo := range itemScores {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]pep3ItemScoreRequest, 0, len(itemNos))
	for _, itemNo := range itemNos {
		out = append(out, pep3ItemScoreRequest{ItemNo: itemNo, Score: itemScores[itemNo]})
	}
	return out
}

func itemRecordValueListFromMap(itemRecordValues map[int]map[string]any) []pep3ItemRecordValueRequest {
	if len(itemRecordValues) == 0 {
		return nil
	}
	itemNos := make([]int, 0, len(itemRecordValues))
	for itemNo := range itemRecordValues {
		itemNos = append(itemNos, itemNo)
	}
	sort.Ints(itemNos)
	out := make([]pep3ItemRecordValueRequest, 0)
	for _, itemNo := range itemNos {
		fieldKeys := make([]string, 0, len(itemRecordValues[itemNo]))
		for fieldKey := range itemRecordValues[itemNo] {
			fieldKeys = append(fieldKeys, fieldKey)
		}
		sort.Strings(fieldKeys)
		for _, fieldKey := range fieldKeys {
			out = append(out, pep3ItemRecordValueRequest{
				ItemNo:   itemNo,
				FieldKey: fieldKey,
				Value:    itemRecordValues[itemNo][fieldKey],
			})
		}
	}
	return out
}

func rawScoreListFromMap(rawScores map[string]int) []pep3RawScoreRequest {
	if len(rawScores) == 0 {
		return nil
	}
	scaleCodes := make([]string, 0, len(rawScores))
	for scaleCode := range rawScores {
		scaleCodes = append(scaleCodes, scaleCode)
	}
	sort.Strings(scaleCodes)
	out := make([]pep3RawScoreRequest, 0, len(scaleCodes))
	for _, scaleCode := range scaleCodes {
		out = append(out, pep3RawScoreRequest{ScaleCode: scaleCode, RawScore: rawScores[scaleCode]})
	}
	return out
}

func normalizePEP3ScaleCode(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}
