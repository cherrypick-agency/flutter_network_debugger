package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHTTPTransaction_JSON_TagsAndZeroValue(t *testing.T) {
	t.Parallel()

	// zero-value безопасен
	var tx HTTPTransaction
	if _, err := json.Marshal(tx); err != nil {
		t.Fatalf("marshal zero: %v", err)
	}

	// заполненный
	tx = HTTPTransaction{
		ID:        "t1",
		SessionID: "s1",
		Method:    "GET",
		URL:       "http://example.com/x",
		Status:    200,
		ReqSize:   0,
		RespSize:  10,
		StartedAt: time.Unix(0, 0).UTC(),
		EndedAt:   time.Unix(1, 0).UTC(),
		Timings:   HTTPTimings{DNS: 1, Connect: 2, TLS: 3, TTFB: 4, Total: 5},
		// omitempty для ContentType/ReqBodyFile/RespBodyFile
	}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	// ключи с нужными тегами
	if !(containsAll(s, `"id"`, `"sessionId"`, `"method"`, `"url"`, `"status"`, `"reqSize"`, `"respSize"`, `"startedAt"`, `"endedAt"`, `"timings"`)) {
		t.Fatalf("missing keys: %s", s)
	}
	if contains(s, "contentType") || contains(s, "reqBodyFile") || contains(s, "respBodyFile") {
		t.Fatalf("omitempty fields must be omitted when empty: %s", s)
	}

	// Когда заданы — должны появляться
	tx.ContentType = "application/json"
	tx.ReqBodyFile = "req.bin"
	tx.RespBodyFile = "resp.bin"
	b2, _ := json.Marshal(tx)
	s2 := string(b2)
	if !(containsAll(s2, "contentType", "reqBodyFile", "respBodyFile")) {
		t.Fatalf("missing omitempty fields when set: %s", s2)
	}
}
