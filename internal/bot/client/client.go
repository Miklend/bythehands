package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	hc      *http.Client
}

func New(baseURL string, timeout time.Duration) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("empty baseURL")
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		hc:      &http.Client{Timeout: timeout},
	}, nil
}

type UpsertTelegramUserRequest struct {
	TelegramID  int64   `json:"telegram_id"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
}

func (c *Client) UpsertTelegramUser(ctx context.Context, req UpsertTelegramUserRequest) (User, error) {
	var env Envelope[User]
	if err := c.post(ctx, "/api/v1/users/telegram", req, &env); err != nil {
		return User{}, err
	}
	if env.Error != nil {
		return User{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) CreatePair(ctx context.Context, userID string) (CreatePairResult, error) {
	var env Envelope[CreatePairResult]
	if err := c.post(ctx, "/api/v1/pairs", map[string]string{"user_id": userID}, &env); err != nil {
		return CreatePairResult{}, err
	}
	if env.Error != nil {
		return CreatePairResult{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) CreatePairInvite(ctx context.Context, pairID string, userID string) (string, error) {
	var env Envelope[map[string]string]
	if err := c.post(ctx, "/api/v1/pairs/"+pairID+"/invite", map[string]string{"user_id": userID}, &env); err != nil {
		return "", err
	}
	if env.Error != nil {
		return "", fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data["invite_token"], nil
}

func (c *Client) SetPairWelcomeMessage(ctx context.Context, pairID string, userID string, text *string) (Pair, error) {
	var env Envelope[Pair]
	body := map[string]any{"user_id": userID, "text": text}
	if err := c.patch(ctx, "/api/v1/pairs/"+pairID+"/welcome", body, &env); err != nil {
		return Pair{}, err
	}
	if env.Error != nil {
		return Pair{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) GetPair(ctx context.Context, pairID string) (Pair, []PairMember, error) {
	var env Envelope[map[string]any]
	if err := c.get(ctx, "/api/v1/pairs/"+pairID, &env); err != nil {
		return Pair{}, nil, err
	}
	if env.Error != nil {
		return Pair{}, nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	rawPair, _ := json.Marshal(env.Data["pair"])
	var p Pair
	_ = json.Unmarshal(rawPair, &p)

	rawMembers, _ := json.Marshal(env.Data["members"])
	var members []PairMember
	_ = json.Unmarshal(rawMembers, &members)
	return p, members, nil
}

func (c *Client) SetMemberName(ctx context.Context, pairID, userID, name string) (PairMember, error) {
	var env Envelope[PairMember]
	body := map[string]any{"user_id": userID, "name": name}
	if err := c.patch(ctx, "/api/v1/pairs/"+pairID+"/member-name", body, &env); err != nil {
		return PairMember{}, err
	}
	if env.Error != nil {
		return PairMember{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) JoinInvite(ctx context.Context, token string, userID string) (Pair, []PairMember, error) {
	var env Envelope[map[string]any]
	if err := c.post(ctx, "/api/v1/invites/"+token+"/join", map[string]string{"user_id": userID}, &env); err != nil {
		return Pair{}, nil, err
	}
	if env.Error != nil {
		return Pair{}, nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	rawPair, _ := json.Marshal(env.Data["pair"])
	var p Pair
	_ = json.Unmarshal(rawPair, &p)

	rawMembers, _ := json.Marshal(env.Data["members"])
	var members []PairMember
	_ = json.Unmarshal(rawMembers, &members)
	return p, members, nil
}

func (c *Client) GetPreferences(ctx context.Context, userID string) (Preferences, error) {
	var env Envelope[Preferences]
	if err := c.get(ctx, "/api/v1/users/"+userID+"/preferences", &env); err != nil {
		return Preferences{}, err
	}
	if env.Error != nil {
		return Preferences{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) SetPreferences(ctx context.Context, userID string, currentPairID *string) (Preferences, error) {
	var env Envelope[Preferences]
	body := map[string]any{"current_pair_id": currentPairID}
	if err := c.patch(ctx, "/api/v1/users/"+userID+"/preferences", body, &env); err != nil {
		return Preferences{}, err
	}
	if env.Error != nil {
		return Preferences{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ListPairs(ctx context.Context, userID string) ([]Pair, error) {
	var env Envelope[[]Pair]
	if err := c.get(ctx, "/api/v1/users/"+userID+"/pairs", &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) StartTestMode(ctx context.Context, userID string) (TestModeStartResult, error) {
	var env Envelope[TestModeStartResult]
	if err := c.post(ctx, "/api/v1/test-mode/start", map[string]string{"user_id": userID}, &env); err != nil {
		return TestModeStartResult{}, err
	}
	if env.Error != nil {
		return TestModeStartResult{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) StopTestMode(ctx context.Context, userID string) (TestModeStopResult, error) {
	var env Envelope[TestModeStopResult]
	if err := c.post(ctx, "/api/v1/test-mode/stop", map[string]string{"user_id": userID}, &env); err != nil {
		return TestModeStopResult{}, err
	}
	if env.Error != nil {
		return TestModeStopResult{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

type CreateIssueRequest struct {
	CreatedByUserID string `json:"created_by_user_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Priority        string `json:"priority"`
	Visibility      string `json:"visibility"`
	RepeatThreshold int    `json:"repeat_threshold"`
	RepeatLimit     int    `json:"repeat_limit"`
}

func (c *Client) CreateIssue(ctx context.Context, pairID string, req CreateIssueRequest) (Issue, error) {
	var env Envelope[Issue]
	if err := c.post(ctx, "/api/v1/pairs/"+pairID+"/issues", req, &env); err != nil {
		return Issue{}, err
	}
	if env.Error != nil {
		return Issue{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ListIssues(ctx context.Context, pairID string, status string) ([]Issue, error) {
	path := "/api/v1/pairs/" + pairID + "/issues"
	if strings.TrimSpace(status) != "" {
		path += "?status=" + status
	}
	var env Envelope[[]Issue]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) GetIssue(ctx context.Context, issueID string) (Issue, error) {
	var env Envelope[Issue]
	if err := c.get(ctx, "/api/v1/issues/"+issueID, &env); err != nil {
		return Issue{}, err
	}
	if env.Error != nil {
		return Issue{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

type UpdateIssueRequest struct {
	Title           *string `json:"title,omitempty"`
	RepeatThreshold *int    `json:"repeat_threshold,omitempty"`
	RepeatLimit     *int    `json:"repeat_limit,omitempty"`
}

func (c *Client) UpdateIssue(ctx context.Context, issueID string, req UpdateIssueRequest) (Issue, error) {
	var env Envelope[Issue]
	if err := c.patch(ctx, "/api/v1/issues/"+issueID, req, &env); err != nil {
		return Issue{}, err
	}
	if env.Error != nil {
		return Issue{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) RepeatIssue(ctx context.Context, issueID, userID string, note string) (Issue, error) {
	var env Envelope[Issue]
	body := map[string]any{"user_id": userID, "note": note}
	if err := c.post(ctx, "/api/v1/issues/"+issueID+"/repeat", body, &env); err != nil {
		return Issue{}, err
	}
	if env.Error != nil {
		return Issue{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) UpdateIssueStatus(ctx context.Context, issueID string, status string) (Issue, error) {
	var env Envelope[Issue]
	body := map[string]any{"status": status}
	if err := c.patch(ctx, "/api/v1/issues/"+issueID+"/status", body, &env); err != nil {
		return Issue{}, err
	}
	if env.Error != nil {
		return Issue{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) DeleteIssue(ctx context.Context, issueID string) error {
	var env Envelope[map[string]string]
	if err := c.delete(ctx, "/api/v1/issues/"+issueID, &env); err != nil {
		return err
	}
	if env.Error != nil {
		return fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return nil
}

func (c *Client) ListRepeats(ctx context.Context, issueID string, limit int) ([]IssueRepeat, error) {
	path := fmt.Sprintf("/api/v1/issues/%s/repeats?limit=%d", issueID, limit)
	var env Envelope[[]IssueRepeat]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) GetRepeat(ctx context.Context, repeatID string) (IssueRepeat, error) {
	var env Envelope[IssueRepeat]
	if err := c.get(ctx, "/api/v1/repeats/"+repeatID, &env); err != nil {
		return IssueRepeat{}, err
	}
	if env.Error != nil {
		return IssueRepeat{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) AddRepeatDisagreement(ctx context.Context, repeatID, userID, note string) (IssueRepeatDisagreement, error) {
	var env Envelope[IssueRepeatDisagreement]
	body := map[string]any{"user_id": userID, "note": note}
	if err := c.post(ctx, "/api/v1/repeats/"+repeatID+"/disagreement", body, &env); err != nil {
		return IssueRepeatDisagreement{}, err
	}
	if env.Error != nil {
		return IssueRepeatDisagreement{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) DeleteRepeat(ctx context.Context, repeatID string) error {
	var env Envelope[map[string]string]
	if err := c.delete(ctx, "/api/v1/repeats/"+repeatID, &env); err != nil {
		return err
	}
	if env.Error != nil {
		return fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return nil
}

func (c *Client) StartConversation(ctx context.Context, issueID, pairID string, goal, questions, startState *string, ruleViolationLimit int) (ConversationSession, error) {
	var env Envelope[ConversationSession]
	body := map[string]any{"pair_id": pairID, "goal": goal, "questions": questions, "start_state": startState, "rule_violation_limit": ruleViolationLimit}
	if err := c.post(ctx, "/api/v1/issues/"+issueID+"/conversations", body, &env); err != nil {
		return ConversationSession{}, err
	}
	if env.Error != nil {
		return ConversationSession{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) PauseConversation(ctx context.Context, conversationID string) (ConversationSession, error) {
	var env Envelope[ConversationSession]
	if err := c.patch(ctx, "/api/v1/conversations/"+conversationID+"/pause", map[string]any{}, &env); err != nil {
		return ConversationSession{}, err
	}
	if env.Error != nil {
		return ConversationSession{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ResumeConversation(ctx context.Context, conversationID string) (ConversationSession, error) {
	var env Envelope[ConversationSession]
	if err := c.patch(ctx, "/api/v1/conversations/"+conversationID+"/resume", map[string]any{}, &env); err != nil {
		return ConversationSession{}, err
	}
	if env.Error != nil {
		return ConversationSession{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

type FinishConversationRequest struct {
	ResultStatus    string  `json:"result_status"`
	ResultText      *string `json:"result_text"`
	EndState        *string `json:"end_state"`
	EndedEarly      bool    `json:"ended_early,omitempty"`
	EndedByUserID   *string `json:"ended_by_user_id,omitempty"`
	EndedInitiative *string `json:"ended_initiative,omitempty"`
	EndReason       *string `json:"end_reason,omitempty"`
}

func (c *Client) FinishConversation(ctx context.Context, conversationID string, req FinishConversationRequest) (ConversationSession, error) {
	var env Envelope[ConversationSession]
	if err := c.patch(ctx, "/api/v1/conversations/"+conversationID+"/finish", req, &env); err != nil {
		return ConversationSession{}, err
	}
	if env.Error != nil {
		return ConversationSession{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) AddConversationNote(ctx context.Context, conversationID, userID, text string) error {
	var env Envelope[map[string]string]
	body := map[string]any{"user_id": userID, "text": text}
	if err := c.post(ctx, "/api/v1/conversations/"+conversationID+"/notes", body, &env); err != nil {
		return err
	}
	if env.Error != nil {
		return fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return nil
}

func (c *Client) DeleteConversationNote(ctx context.Context, noteID string) error {
	var env Envelope[map[string]string]
	if err := c.delete(ctx, "/api/v1/notes/"+noteID, &env); err != nil {
		return err
	}
	if env.Error != nil {
		return fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return nil
}

func (c *Client) AddConversationRuleViolation(ctx context.Context, conversationID, userID, ruleCode, note string) error {
	var env Envelope[map[string]string]
	body := map[string]any{"user_id": userID, "rule_code": ruleCode, "note": note}
	if err := c.post(ctx, "/api/v1/conversations/"+conversationID+"/rule-violations", body, &env); err != nil {
		return err
	}
	if env.Error != nil {
		return fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return nil
}

func (c *Client) ListConversationRuleViolations(ctx context.Context, conversationID string, limit int) ([]ConversationRuleViolation, error) {
	path := fmt.Sprintf("/api/v1/conversations/%s/rule-violations?limit=%d", conversationID, limit)
	var env Envelope[[]ConversationRuleViolation]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

type AddSideIssueRequest struct {
	CreatedByUserID string `json:"created_by_user_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
}

func (c *Client) AddSideIssue(ctx context.Context, conversationID string, req AddSideIssueRequest) (string, error) {
	var env Envelope[map[string]string]
	if err := c.post(ctx, "/api/v1/conversations/"+conversationID+"/side-issues", req, &env); err != nil {
		return "", err
	}
	if env.Error != nil {
		return "", fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data["issue_id"], nil
}

func (c *Client) ListConversations(ctx context.Context, pairID string, status string) ([]ConversationSession, error) {
	path := "/api/v1/pairs/" + pairID + "/conversations"
	if strings.TrimSpace(status) != "" {
		path += "?status=" + status
	}
	var env Envelope[[]ConversationSession]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ListPairNotes(ctx context.Context, pairID string, limit int) ([]ConversationPairNote, error) {
	path := "/api/v1/pairs/" + pairID + "/notes"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var env Envelope[[]ConversationPairNote]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) GetConversation(ctx context.Context, conversationID string) (ConversationSession, error) {
	var env Envelope[ConversationSession]
	if err := c.get(ctx, "/api/v1/conversations/"+conversationID, &env); err != nil {
		return ConversationSession{}, err
	}
	if env.Error != nil {
		return ConversationSession{}, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ListConversationNotes(ctx context.Context, conversationID string, limit int) ([]ConversationNote, error) {
	path := fmt.Sprintf("/api/v1/conversations/%s/notes?limit=%d", conversationID, limit)
	var env Envelope[[]ConversationNote]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) ListConversationSideIssues(ctx context.Context, conversationID string) ([]Issue, error) {
	var env Envelope[[]Issue]
	if err := c.get(ctx, "/api/v1/conversations/"+conversationID+"/side-issues", &env); err != nil {
		return nil, err
	}
	if env.Error != nil {
		return nil, fmt.Errorf("api error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env.Data, nil
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api http %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) patch(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api http %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api http %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) delete(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api http %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}
