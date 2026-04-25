package handlers

import (
	"log/slog"
	"net/http"

	"taalkbout/internal/service"
)

type UsersHandler struct {
	log  *slog.Logger
	svc  *service.UserService
	mapE ErrorMapper
}

func NewUsersHandler(log *slog.Logger, svc *service.UserService) *UsersHandler {
	return &UsersHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type upsertTelegramUserRequest struct {
	TelegramID  int64   `json:"telegram_id"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
}

func (h *UsersHandler) UpsertTelegramUser(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		var req upsertTelegramUserRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		u, err := h.svc.UpsertTelegramUser(r.Context(), req.TelegramID, req.Username, req.DisplayName)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: u})
		return nil
	})(w, r)
}
