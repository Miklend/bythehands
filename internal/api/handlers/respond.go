package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Envelope struct {
	Data  any  `json:"data,omitempty"`
	Error *Err `json:"error,omitempty"`
}

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, Envelope{Error: &Err{Code: code, Message: message}})
}

type AppHandler func(w http.ResponseWriter, r *http.Request) error

type ErrorMapper func(err error) (status int, code string, message string)

func Wrap(log *slog.Logger, mapErr ErrorMapper, h AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			status, code, msg := mapErr(err)
			log.Error("request failed", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("status", status), slog.Any("err", err))
			WriteError(w, status, code, msg)
			return
		}
	}
}
