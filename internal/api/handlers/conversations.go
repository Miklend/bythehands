package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taalkbout/internal/domain/conversation"
	"taalkbout/internal/service"
)

type ConversationsHandler struct {
	log  *slog.Logger
	svc  *service.ConversationService
	mapE ErrorMapper
}

func NewConversationsHandler(log *slog.Logger, svc *service.ConversationService) *ConversationsHandler {
	return &ConversationsHandler{log: log, svc: svc, mapE: DefaultErrorMapper}
}

type startConversationRequest struct {
	PairID             string  `json:"pair_id"`
	Goal               *string `json:"goal"`
	Questions          *string `json:"questions"`
	StartState         *string `json:"start_state"`
	RuleViolationLimit int     `json:"rule_violation_limit"`
}

func (h *ConversationsHandler) StartConversation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		issueID := chi.URLParam(r, "issue_id")
		var req startConversationRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		sess, err := h.svc.Start(r.Context(), service.StartConversationInput{
			IssueID:            issueID,
			PairID:             req.PairID,
			Goal:               req.Goal,
			Questions:          req.Questions,
			StartState:         req.StartState,
			RuleViolationLimit: req.RuleViolationLimit,
		})
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: sess})
		return nil
	})(w, r)
}

type finishConversationRequest struct {
	ResultStatus    conversation.ResultStatus `json:"result_status"`
	ResultText      *string                   `json:"result_text"`
	EndState        *string                   `json:"end_state"`
	EndedEarly      bool                      `json:"ended_early"`
	EndedByUserID   *string                   `json:"ended_by_user_id"`
	EndedInitiative *string                   `json:"ended_initiative"`
	EndReason       *string                   `json:"end_reason"`
}

func (h *ConversationsHandler) FinishConversation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		var req finishConversationRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		sess, err := h.svc.Finish(r.Context(), id, service.FinishConversationInput{
			ResultStatus:    req.ResultStatus,
			ResultText:      req.ResultText,
			EndState:        req.EndState,
			EndedEarly:      req.EndedEarly,
			EndedByUserID:   req.EndedByUserID,
			EndedInitiative: req.EndedInitiative,
			EndReason:       req.EndReason,
		})
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: sess})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		sess, err := h.svc.Get(r.Context(), id)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: sess})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) PauseConversation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		sess, err := h.svc.Pause(r.Context(), id)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: sess})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ResumeConversation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		sess, err := h.svc.Resume(r.Context(), id)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: sess})
		return nil
	})(w, r)
}

type addNoteRequest struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func (h *ConversationsHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		var req addNoteRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		if err := h.svc.AddNote(r.Context(), id, req.UserID, req.Text); err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: map[string]string{"status": "ok"}})
		return nil
	})(w, r)
}

type addSideIssueRequest struct {
	CreatedByUserID string `json:"created_by_user_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
}

func (h *ConversationsHandler) AddSideIssue(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		var req addSideIssueRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		issueID, err := h.svc.AddSideIssue(r.Context(), service.AddSideIssueInput{
			ConversationID:  id,
			CreatedByUserID: req.CreatedByUserID,
			Title:           req.Title,
			Description:     req.Description,
		})
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: map[string]string{"issue_id": issueID}})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ListByPair(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		var st *conversation.Status
		if qs := r.URL.Query().Get("status"); qs != "" {
			s := conversation.Status(qs)
			st = &s
		}
		limit := parseInt(r.URL.Query().Get("limit"), 20)
		offset := parseInt(r.URL.Query().Get("offset"), 0)
		items, err := h.svc.ListByPair(r.Context(), pairID, st, limit, offset)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		limit := parseInt(r.URL.Query().Get("limit"), 20)
		offset := parseInt(r.URL.Query().Get("offset"), 0)
		items, err := h.svc.ListNotes(r.Context(), id, limit, offset)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ListSideIssues(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		items, err := h.svc.ListSideIssues(r.Context(), id)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ListNotesByPair(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		pairID := chi.URLParam(r, "pair_id")
		limit := parseInt(r.URL.Query().Get("limit"), 20)
		offset := parseInt(r.URL.Query().Get("offset"), 0)
		items, err := h.svc.ListNotesByPair(r.Context(), pairID, limit, offset)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		noteID := chi.URLParam(r, "note_id")
		if err := h.svc.DeleteNote(r.Context(), noteID); err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: map[string]string{"status": "ok"}})
		return nil
	})(w, r)
}

type addRuleViolationRequest struct {
	UserID   string `json:"user_id"`
	RuleCode string `json:"rule_code"`
	Note     string `json:"note"`
}

func (h *ConversationsHandler) AddRuleViolation(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		var req addRuleViolationRequest
		if err := DecodeJSON(r, &req); err != nil {
			return &service.Error{Kind: service.KindValidation, Message: "invalid json", Err: err}
		}
		if err := h.svc.AddRuleViolation(r.Context(), id, req.UserID, req.RuleCode, req.Note); err != nil {
			return err
		}
		WriteJSON(w, http.StatusCreated, Envelope{Data: map[string]string{"status": "ok"}})
		return nil
	})(w, r)
}

func (h *ConversationsHandler) ListRuleViolations(w http.ResponseWriter, r *http.Request) {
	Wrap(h.log, h.mapE, func(w http.ResponseWriter, r *http.Request) error {
		id := chi.URLParam(r, "conversation_id")
		limit := parseInt(r.URL.Query().Get("limit"), 20)
		offset := parseInt(r.URL.Query().Get("offset"), 0)
		items, err := h.svc.ListRuleViolations(r.Context(), id, limit, offset)
		if err != nil {
			return err
		}
		WriteJSON(w, http.StatusOK, Envelope{Data: items})
		return nil
	})(w, r)
}
