package careers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/workspace/ride-platform/internal/middleware"
	apperrors "github.com/workspace/ride-platform/pkg/errors"
	"github.com/workspace/ride-platform/pkg/respond"
)

var validate = validator.New()

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/v1/careers (Public candidate submission)
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var input CreateApplicationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, apperrors.ErrBadRequest)
		return
	}

	if err := validate.Struct(input); err != nil {
		respond.ErrorMsg(w, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	app, err := h.svc.Submit(r.Context(), input)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.Created(w, map[string]interface{}{
		"message":        "Thank you for submitting your application to Rides! Our team will review your submission.",
		"application_id": app.ID,
	})
}

// GET /api/v1/admin/careers (Admin list applications)
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ListFilter{
		Position:          q.Get("position"),
		ApplicationStatus: q.Get("status"),
		Search:            q.Get("search"),
		Limit:             limit,
		Offset:            offset,
	}

	apps, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.OK(w, map[string]interface{}{
		"applications": apps,
		"total":        total,
	})
}

// GET /api/v1/admin/careers/{id}
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := h.svc.Get(r.Context(), id)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, app)
}

// PATCH /api/v1/admin/careers/{id}/status
func (h *Handler) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	reviewerID := ""
	if claims != nil {
		reviewerID = claims.UserID
	}

	id := chi.URLParam(r, "id")
	var input UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, apperrors.ErrBadRequest)
		return
	}

	if err := validate.Struct(input); err != nil {
		respond.ErrorMsg(w, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	app, err := h.svc.UpdateStatus(r.Context(), id, input.ApplicationStatus, input.ReviewerNotes, reviewerID)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.OK(w, app)
}

// GET /api/v1/admin/careers/export (CSV file download for HR / Excel / Google Sheets)
func (h *Handler) AdminExportCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{
		Position:          q.Get("position"),
		ApplicationStatus: q.Get("status"),
		Search:            q.Get("search"),
	}

	csvBytes, err := h.svc.ExportCSV(r.Context(), filter)
	if err != nil {
		respond.Error(w, err)
		return
	}

	filename := fmt.Sprintf("rides_career_applications_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBytes)
}

// GET /api/v1/careers/config (Public & Admin configuration check)
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.GetSettings(r.Context())
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, settings)
}

// PUT /api/v1/admin/careers/settings (Admin update settings)
func (h *Handler) AdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var input UpdateSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respond.Error(w, apperrors.ErrBadRequest)
		return
	}

	settings, err := h.svc.UpdateSettings(r.Context(), input)
	if err != nil {
		respond.Error(w, err)
		return
	}

	respond.OK(w, settings)
}
