package domain

import "time"

// InterceptItemState -- interception item state
type InterceptItemState string

const (
	StatePending  InterceptItemState = "PENDING"
	StateApplied  InterceptItemState = "APPLIED"
	StateCanceled InterceptItemState = "CANCELED"
	StateTimedOut InterceptItemState = "TIMED_OUT"
)

// InterceptItem -- unit awaiting user decision
type InterceptItem struct {
	ID        string             `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	Deadline  time.Time          `json:"deadline"`
	Direction InterceptDirection `json:"direction"`
	SessionID string             `json:"sessionId"`
	RuleID    string             `json:"ruleId,omitempty"`

	Req *HTTPRequestSnapshot  `json:"req,omitempty"`
	Res *HTTPResponseSnapshot `json:"res,omitempty"`

	State InterceptItemState `json:"state"`
}

// HTTPRequestSnapshot -- part that can be shown/edited
type HTTPRequestSnapshot struct {
	Method        string              `json:"method"`
	URL           string              `json:"url"`
	Headers       map[string][]string `json:"headers"`
	BodyBase64    string              `json:"bodyBase64,omitempty"`
	BodyTruncated bool                `json:"bodyTruncated,omitempty"`
	ContentType   string              `json:"contentType,omitempty"`
}

// HTTPResponseSnapshot -- similarly for response
type HTTPResponseSnapshot struct {
	Status        int                 `json:"status"`
	Headers       map[string][]string `json:"headers"`
	BodyBase64    string              `json:"bodyBase64,omitempty"`
	BodyTruncated bool                `json:"bodyTruncated,omitempty"`
	ContentType   string              `json:"contentType,omitempty"`
}

// HTTPRequestDecision -- user decision for request
type HTTPRequestDecision struct {
	Action  string              `json:"action"` // continue|drop
	Method  string              `json:"method,omitempty"`
	URL     string              `json:"url,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    []byte              `json:"-"`
}

// HTTPResponseDecision -- user decision for response
type HTTPResponseDecision struct {
	Action  string              `json:"action"` // continue
	Status  int                 `json:"status,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    []byte              `json:"-"`
}
