package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"network-debugger/internal/usecase"

	"github.com/zishang520/socket.io/v2/socket"
)

type sioFilters struct {
	Q                 string
	Target            string
	Types             []string
	StatusGroups      []string
	Tags              []string // filter by tags
	CaptureScope      string   // "current" | "all" | ""
	CaptureIDExplicit *int
	IncludeUnassigned bool
	Limit             int
	GroupBy           string
}

// NewSocketIOServer: Socket.IO server + sessions:subscribe subscription (initial + increments)
func NewSocketIOServer(d *Deps) http.Handler {
	// Create Socket.IO server v3/v4 (zishang520/socket.io/v2)
	// Library automatically configures WebSocket transport and CORS
	io := socket.NewServer(nil, nil)

	realtime := newSessionRealtimeService(d, &socketSessionRoomEmitter{io: io})
	newSessionMonitorAdapter(d.Monitor, realtime).bind()

	io.On("connection", func(clients ...any) {
		// Type assertion to get Socket from variadic parameter
		rawClient := clients[0].(*socket.Socket)
		client := &socketRealtimeClient{socket: rawClient}
		connID := client.ID()

		d.Logger.Info().Str("conn_id", connID).Msg("Socket.IO client connected")
		// Auto-subscribe with default filters so client immediately receives initial
		// Default Limit 1000 and current capture (passCapture treats empty scope as "current")
		// IncludeUnassigned: true allows seeing new sessions with CaptureID == nil
		f := sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true}
		realtime.addSubscription(client, f)
		realtime.sendInit(client, f)

		// Register event handlers - now variadic any instead of EventPayload
		rawClient.On("sessions:subscribe", func(datas ...any) {
			payload := socketPayload(datas)
			f := parseSioFilters(payload)
			if f.Limit <= 0 || f.Limit > 1000 {
				f.Limit = 1000
			}
			realtime.addSubscription(client, f)
			realtime.sendInit(client, f)
		})

		// Subscribe to specific session for detail page
		rawClient.On("session:subscribe", func(datas ...any) {
			sessionID := sessionIDFromPayload(socketPayload(datas))
			if sessionID == "" {
				d.Logger.Warn().Str("conn_id", connID).Msg("session:subscribe without sessionId")
				return
			}
			realtime.subscribeSession(client, sessionID)
			d.Logger.Info().Str("conn_id", connID).Str("session_id", sessionID).Str("room", string(sessionRoomName(sessionID))).Msg("Client joined session room")
		})

		// Unsubscribe from specific session
		rawClient.On("session:unsubscribe", func(datas ...any) {
			sessionID := sessionIDFromPayload(socketPayload(datas))
			if sessionID == "" {
				return
			}
			realtime.unsubscribeSession(client, sessionID)
			d.Logger.Info().Str("conn_id", connID).Str("session_id", sessionID).Str("room", string(sessionRoomName(sessionID))).Msg("Client left session room")
		})

		rawClient.On("error", func(datas ...any) {
			var err error
			if len(datas) > 0 {
				if e, ok := datas[0].(error); ok {
					err = e
				}
			}
			d.Logger.Error().Err(err).Str("conn_id", connID).Msg("Socket.IO connection error")
		})

		rawClient.On("disconnect", func(datas ...any) {
			var reason string
			if len(datas) > 0 {
				if r, ok := datas[0].(string); ok {
					reason = r
				}
			}
			d.Logger.Info().Str("conn_id", connID).Str("reason", reason).Msg("Socket.IO client disconnected")
			realtime.removeClient(client)
		})
	})

	return io.ServeHandler(nil)
}

func socketPayload(datas []any) map[string]any {
	if len(datas) == 0 {
		return nil
	}
	if payload, ok := datas[0].(map[string]any); ok {
		return payload
	}
	return nil
}

func sessionIDFromPayload(payload map[string]any) string {
	sessionID, _ := payload["sessionId"].(string)
	return strings.TrimSpace(sessionID)
}

func parseSioFilters(m map[string]any) sioFilters {
	f := sioFilters{Limit: 1000, GroupBy: "domain"}
	if v, ok := m["q"].(string); ok {
		f.Q = v
	}
	if v, ok := m["target"].(string); ok {
		f.Target = v
	}
	if v, ok := m["types"].([]any); ok {
		f.Types = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				f.Types = append(f.Types, strings.ToLower(strings.TrimSpace(s)))
			}
		}
	}
	if v, ok := m["status"].([]any); ok {
		f.StatusGroups = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				f.StatusGroups = append(f.StatusGroups, strings.ToLower(strings.TrimSpace(s)))
			}
		}
	}
	if v, ok := m["captureScope"].(string); ok {
		f.CaptureScope = v
	}
	if v, ok := m["includeUnassigned"].(bool); ok {
		f.IncludeUnassigned = v
	}
	if v, ok := m["limit"].(float64); ok {
		f.Limit = int(v)
	} else if s, ok2 := m["limit"].(string); ok2 {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			f.Limit = n
		}
	} else if iv, ok3 := m["limit"].(int); ok3 {
		f.Limit = iv
	}
	if v, ok := m["groupBy"].(string); ok && strings.TrimSpace(v) != "" {
		f.GroupBy = strings.ToLower(strings.TrimSpace(v))
	}
	if v, ok := m["captureId"].(float64); ok {
		n := int(v)
		f.CaptureIDExplicit = &n
	}
	if v, ok := m["tags"].([]any); ok {
		f.Tags = make([]string, 0, len(v))
		for _, x := range v {
			if s, ok2 := x.(string); ok2 {
				tagName := strings.TrimSpace(s)
				if tagName != "" {
					f.Tags = append(f.Tags, tagName)
				}
			}
		}
	}
	return f
}

// normalizeHost extracts and normalizes host from URL (removes default ports and userinfo)
func normalizeHost(targetURL string) string {
	return normalizeProjectionHost(targetURL)
}

// groupKeyFor calculates the grouping key based on groupBy setting.
func groupKeyFor(v sessionV1, groupBy string) string {
	return sessionGroupKey(v, groupBy)
}

func toUsecaseFilter(f sioFilters) usecase.SessionFilter {
	return projectionListFilterFromSIO(f)
}
