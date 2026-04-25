package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taalkbout/internal/service"
)

type InvitesHandler struct {
	log  *slog.Logger
	svc  *service.PairService
	mapE ErrorMapper
}

func NewInvitesHandler(log *slog.Logger, svc *service.PairService) *InvitesHandler {
	return &InvitesHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type joinInviteRequest struct {
	UserID string `json:"user_id"`
}

func (h *InvitesHandler) JoinByToken(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		token := chi.URLParam(r, "token")
		var req joinInviteRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		p, members, err := h.svc.JoinByInvite(r.Context(), token, req.UserID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: map[string]any{"pair": p, "members": members}})
		return nil
	})(w, r)
}
