package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taalkbout/internal/domain/issue"
	"taalkbout/internal/service"
)

type IssuesHandler struct {
	log  *slog.Logger
	svc  *service.IssueService
	mapE ErrorMapper
}

func NewIssuesHandler(log *slog.Logger, svc *service.IssueService) *IssuesHandler {
	return &IssuesHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type createIssueRequest struct {
	CreatedByUserID string           `json:"created_by_user_id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Priority        issue.Priority   `json:"priority"`
	Visibility      issue.Visibility `json:"visibility"`
	RepeatThreshold int              `json:"repeat_threshold"`
	RepeatLimit     int              `json:"repeat_limit"`
}

func (h *IssuesHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var req createIssueRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		it, err := h.svc.CreateIssue(r.Context(), service.CreateIssueInput{
			PairID:          pairID,
			CreatedByUserID: req.CreatedByUserID,
			Title:           req.Title,
			Description:     req.Description,
			Priority:        req.Priority,
			Visibility:      req.Visibility,
			RepeatThreshold: req.RepeatThreshold,
			RepeatLimit:     req.RepeatLimit,
		})
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: it})
		return nil
	})(w, r)
}

func (h *IssuesHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var st *issue.Status
		if qs := r.URL.Query().Get("status"); qs != "" {
			s := issue.Status(qs)
			st = &s
		}
		items, err := h.svc.ListIssues(r.Context(), pairID, st)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}

func (h *IssuesHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		it, err := h.svc.GetIssue(r.Context(), issueID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: it})
		return nil
	})(w, r)
}

type updateIssueRequest struct {
	Title           *string `json:"title"`
	RepeatThreshold *int    `json:"repeat_threshold"`
	RepeatLimit     *int    `json:"repeat_limit"`
}

func (h *IssuesHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		var req updateIssueRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		it, err := h.svc.UpdateIssue(r.Context(), issueID, service.UpdateIssueInput{
			Title:           req.Title,
			RepeatThreshold: req.RepeatThreshold,
			RepeatLimit:     req.RepeatLimit,
		})
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: it})
		return nil
	})(w, r)
}

type repeatIssueRequest struct {
	UserID string  `json:"user_id"`
	Note   *string `json:"note"`
}

func (h *IssuesHandler) RepeatIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		var req repeatIssueRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		it, err := h.svc.Repeat(r.Context(), issueID, req.UserID, req.Note)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: it})
		return nil
	})(w, r)
}

type updateIssueStatusRequest struct {
	Status issue.Status `json:"status"`
}

func (h *IssuesHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		var req updateIssueStatusRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		it, err := h.svc.UpdateStatus(r.Context(), issueID, req.Status)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: it})
		return nil
	})(w, r)
}

func (h *IssuesHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		if err := h.svc.DeleteIssue(r.Context(), issueID); err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: map[string]string{"status": "ok"}})
		return nil
	})(w, r)
}
