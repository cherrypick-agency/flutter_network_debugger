package domain

import (
	"encoding/json"
	"testing"
)

func TestDirection_ConstantsAndJSON(t *testing.T) {
	t.Parallel()

	// Check stability of string values
	if string(DirectionClientToUpstream) != "client->upstream" {
		t.Fatalf("DirectionClientToUpstream changed: %q", DirectionClientToUpstream)
	}
	if string(DirectionUpstreamToClient) != "upstream->client" {
		t.Fatalf("DirectionUpstreamToClient changed: %q", DirectionUpstreamToClient)
	}

	// JSON roundtrip in wrapper
	type wrap struct {
		D Direction `json:"d"`
	}
	for _, d := range []Direction{DirectionClientToUpstream, DirectionUpstreamToClient, Direction("custom")} {
		b, err := json.Marshal(wrap{D: d})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var w wrap
		if err := json.Unmarshal(b, &w); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if w.D != d {
			t.Fatalf("roundtrip mismatch: %q != %q", w.D, d)
		}
	}
}
