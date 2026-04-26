package routes

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"taalkbout/internal/api/handlers"
	apimw "taalkbout/internal/api/middleware"
	"taalkbout/internal/service"
)

type Services struct {
	Users         *service.UserService
	Pairs         *service.PairService
	Issues        *service.IssueService
	Conversations *service.ConversationService
	Preferences   *service.PreferencesService
	TestMode      *service.TestModeService
}

func NewRouter(log *slog.Logger, svc Services) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(apimw.Recoverer(log))
	r.Use(apimw.RequestLogger(log))
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, handlers.Envelope{Data: map[string]string{"status": "ok"}})
	})

	r.Route("/api/v1", func(r chi.Router) {
		usersH := handlers.NewUsersHandler(log, svc.Users)
		pairsH := handlers.NewPairsHandler(log, svc.Pairs)
		invH := handlers.NewInvitesHandler(log, svc.Pairs)
		issuesH := handlers.NewIssuesHandler(log, svc.Issues)
		convH := handlers.NewConversationsHandler(log, svc.Conversations)
		prefH := handlers.NewPreferencesHandler(log, svc.Preferences)
		testH := handlers.NewTestModeHandler(log, svc.TestMode)
		repeatsH := handlers.NewRepeatsHandler(log, svc.Issues)

		r.Post("/users/telegram", usersH.UpsertTelegramUser)

		r.Get("/users/{user_id}/pairs", prefH.ListPairs)
		r.Get("/users/{user_id}/preferences", prefH.GetPreferences)
		r.Patch("/users/{user_id}/preferences", prefH.UpdatePreferences)

		r.Post("/test-mode/start", testH.Start)
		r.Post("/test-mode/stop", testH.Stop)

		r.Post("/pairs", pairsH.CreatePair)
		r.Get("/pairs/{pair_id}", pairsH.GetPair)
		r.Post("/pairs/{pair_id}/invite", pairsH.CreateInvite)
		r.Patch("/pairs/{pair_id}/welcome", pairsH.SetWelcome)

		r.Post("/invites/{token}/join", invH.JoinByToken)

		r.Post("/pairs/{pair_id}/issues", issuesH.CreateIssue)
		r.Get("/pairs/{pair_id}/issues", issuesH.ListIssues)

		r.Get("/issues/{issue_id}", issuesH.GetIssue)
		r.Patch("/issues/{issue_id}", issuesH.UpdateIssue)
		r.Post("/issues/{issue_id}/repeat", issuesH.RepeatIssue)
		r.Patch("/issues/{issue_id}/status", issuesH.UpdateStatus)
		r.Delete("/issues/{issue_id}", issuesH.DeleteIssue)
		r.Get("/issues/{issue_id}/repeats", repeatsH.ListIssueRepeats)

		r.Get("/repeats/{repeat_id}", repeatsH.GetRepeat)
		r.Post("/repeats/{repeat_id}/disagreement", repeatsH.AddDisagreement)
		r.Delete("/repeats/{repeat_id}", repeatsH.DeleteRepeat)

		r.Post("/issues/{issue_id}/conversations", convH.StartConversation)
		r.Patch("/conversations/{conversation_id}/finish", convH.FinishConversation)
		r.Get("/conversations/{conversation_id}", convH.GetConversation)
		r.Patch("/conversations/{conversation_id}/pause", convH.PauseConversation)
		r.Patch("/conversations/{conversation_id}/resume", convH.ResumeConversation)
		r.Post("/conversations/{conversation_id}/notes", convH.AddNote)
		r.Get("/conversations/{conversation_id}/notes", convH.ListNotes)
		r.Post("/conversations/{conversation_id}/side-issues", convH.AddSideIssue)
		r.Get("/conversations/{conversation_id}/side-issues", convH.ListSideIssues)
		r.Get("/pairs/{pair_id}/conversations", convH.ListByPair)
		r.Get("/pairs/{pair_id}/notes", convH.ListNotesByPair)
		r.Delete("/notes/{note_id}", convH.DeleteNote)
	})

	return r
}
