package handlers

import (
	"log/slog"
	"net/http"

	"talkabout/internal/service"
)

type TestModeHandler struct {
	log  *slog.Logger
	svc  *service.TestModeService
	mapE ErrorMapper
}

func NewTestModeHandler(log *slog.Logger, svc *service.TestModeService) *TestModeHandler {
	return &TestModeHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type testModeRequest struct {
	UserID string `json:"user_id"`
}

func (h *TestModeHandler) Start(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		var req testModeRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		res, err := h.svc.Start(r.Context(), req.UserID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: res})
		return nil
	})(w, r)
}

func (h *TestModeHandler) Stop(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		var req testModeRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		res, err := h.svc.Stop(r.Context(), req.UserID)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: res})
		return nil
	})(w, r)
}
