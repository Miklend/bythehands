package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"taalkbout/internal/service"
)

type RepeatsHandler struct {
	log  *slog.Logger
	svc  *service.IssueService
	mapE ErrorMapper
}

func NewRepeatsHandler(log *slog.Logger, svc *service.IssueService) *RepeatsHandler {
	return &RepeatsHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

func (h *RepeatsHandler) ListIssueRepeats(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		limit := parseInt(r.URL.Query().Get("limit"), 20)
		offset := parseInt(r.URL.Query().Get("offset"), 0)
		reps, err := h.svc.ListRepeats(r.Context(), issueID, limit, offset)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: reps})
		return nil
	})(w, r)
}

func (h *RepeatsHandler) GetRepeat(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		repeatID := chi.URLParam(r, "repeat_id")
		rep, err := h.svc.GetRepeat(r.Context(), repeatID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: rep})
		return nil
	})(w, r)
}

type disagreementRequest struct {
	UserID string `json:"user_id"`
	Note   string `json:"note"`
}

func (h *RepeatsHandler) AddDisagreement(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		repeatID := chi.URLParam(r, "repeat_id")
		var req disagreementRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		d, err := h.svc.AddRepeatDisagreement(r.Context(), repeatID, req.UserID, req.Note)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: d})
		return nil
	})(w, r)
}

func (h *RepeatsHandler) DeleteRepeat(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		repeatID := chi.URLParam(r, "repeat_id")
		if err := h.svc.DeleteRepeat(r.Context(), repeatID); err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: map[string]string{"status": "ok"}})
		return nil
	})(w, r)
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
