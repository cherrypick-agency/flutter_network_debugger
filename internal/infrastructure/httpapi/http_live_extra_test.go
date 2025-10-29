package httpapi

import "testing"

func TestLiveSessions_SendText_InvalidDirection(t *testing.T) {
    ls := NewLiveSessions()
    // зарегистрируем с пустыми ws, чтобы дошло до ветки invalid direction
    ls.Register("s1", nil, nil)
    if err := ls.SendText("s1", "wrong", "x"); err == nil {
        t.Fatalf("expected error for invalid direction")
    }
}

func TestLiveSessions_SendText_NotFound(t *testing.T) {
    ls := NewLiveSessions()
    if err := ls.SendText("unknown", "client->upstream", "x"); err == nil {
        t.Fatalf("expected not found error")
    }
}

func TestLiveSessions_SendText_UpstreamOrClientUnavailable(t *testing.T) {
    ls := NewLiveSessions()
    ls.Register("s1", nil, nil)
    if err := ls.SendText("s1", "client->upstream", "x"); err == nil {
        t.Fatalf("expected upstream not available error")
    }
    if err := ls.SendText("s1", "upstream->client", "x"); err == nil {
        t.Fatalf("expected client not available error")
    }
}


