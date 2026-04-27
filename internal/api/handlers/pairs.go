package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taalkbout/internal/service"
)

type PairsHandler struct {
	log  *slog.Logger
	svc  *service.PairService
	mapE ErrorMapper
}

func NewPairsHandler(log *slog.Logger, svc *service.PairService) *PairsHandler {
	return &PairsHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type createPairRequest struct {
	UserID string `json:"user_id"`
}

func (h *PairsHandler) CreatePair(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		var req createPairRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		res, err := h.svc.CreatePair(r.Context(), req.UserID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: res})
		return nil
	})(w, r)
}

func (h *PairsHandler) GetPair(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		p, members, err := h.svc.GetPair(r.Context(), pairID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: map[string]any{"pair": p, "members": members}})
		return nil
	})(w, r)
}

type createInviteRequest struct {
	UserID string `json:"user_id"`
}

func (h *PairsHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var req createInviteRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		token, err := h.svc.CreateInvite(r.Context(), pairID, req.UserID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: map[string]string{"invite_token": token}})
		return nil
	})(w, r)
}

type setWelcomeRequest struct {
	UserID string  `json:"user_id"`
	Text   *string `json:"text"`
}

func (h *PairsHandler) SetWelcome(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var req setWelcomeRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		p, err := h.svc.SetWelcomeMessage(r.Context(), pairID, req.UserID, req.Text)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: p})
		return nil
	})(w, r)
}

type setMemberNameRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func (h *PairsHandler) SetMemberName(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var req setMemberNameRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		m, err := h.svc.SetMemberName(r.Context(), pairID, req.UserID, req.Name)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: m})
		return nil
	})(w, r)
}
