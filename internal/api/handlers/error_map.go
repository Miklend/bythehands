package handlers

import (
	"errors"
	"net/http"

	"talkabout/internal/repository"
	"talkabout/internal/service"
)

func DefaultErrorMapper(err error) (int, string, string) {
	if err == nil {
		return http.StatusInternalServerError, "internal", "internal error"
	}

	var se *service.Error
	if errors.As(err, &se) {
		switch se.Kind {
		case service.KindValidation:
			return http.StatusBadRequest, "validation", se.Message
		case service.KindNotFound:
			return http.StatusNotFound, "not_found", se.Message
		case service.KindConflict:
			return http.StatusConflict, "conflict", se.Message
		case service.KindForbidden:
			return http.StatusForbidden, "forbidden", se.Message
		default:
			return http.StatusInternalServerError, "internal", "internal error"
		}
	}

	switch {
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, "not_found", "not found"
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, "conflict", "conflict"
	case errors.Is(err, repository.ErrForbidden):
		return http.StatusForbidden, "forbidden", "forbidden"
	default:
		return http.StatusInternalServerError, "internal", "internal error"
	}
}
