package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"talkabout/internal/service"
)

type PreferencesHandler struct {
	log  *slog.Logger
	svc  *service.PreferencesService
	mapE ErrorMapper
}

func NewPreferencesHandler(log *slog.Logger, svc *service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

func (h *PreferencesHandler) ListPairs(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		userID := chi.URLParam(r, "user_id")
		pairs, err := h.svc.ListPairs(r.Context(), userID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: pairs})
		return nil
	})(w, r)
}

func (h *PreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		userID := chi.URLParam(r, "user_id")
		pref, err := h.svc.Get(r.Context(), userID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: pref})
		return nil
	})(w, r)
}

type updatePreferencesRequest struct {
	CurrentPairID *string `json:"current_pair_id"`
}

func (h *PreferencesHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		userID := chi.URLParam(r, "user_id")
		var req updatePreferencesRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		pref, err := h.svc.SetCurrentPair(r.Context(), userID, req.CurrentPairID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: pref})
		return nil
	})(w, r)
}
