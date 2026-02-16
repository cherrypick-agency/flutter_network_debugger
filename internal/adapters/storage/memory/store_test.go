package memory

import (
	"context"
	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
	"os"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	maxSessions := 100
	maxFrames := 1000
	ttl := 5 * time.Minute

	store := NewStore(maxSessions, maxFrames, ttl)

	if store == nil {
		t.Fatal("NewStore returned nil")
	}
	if store.maxSessions != maxSessions {
		t.Errorf("maxSessions = %d, want %d", store.maxSessions, maxSessions)
	}
	if store.maxFramesPerSession != maxFrames {
		t.Errorf("maxFramesPerSession = %d, want %d", store.maxFramesPerSession, maxFrames)
	}
	if store.ttl != ttl {
		t.Errorf("ttl = %v, want %v", store.ttl, ttl)
	}
	if !store.recording {
		t.Error("recording should be true by default")
	}
	if store.currentCapture != 0 {
		t.Errorf("currentCapture = %d, want 0", store.currentCapture)
	}
}

func TestRecordingState(t *testing.T) {
	store := NewStore(10, 100, 0)

	recording, captureID := store.RecordingState()
	if !recording {
		t.Error("recording should be true by default")
	}
	if captureID != 0 {
		t.Errorf("captureID = %d, want 0", captureID)
	}
}

func TestStartCapture(t *testing.T) {
	store := NewStore(10, 100, 0)

	captureID := store.StartCapture()
	if captureID != 1 {
		t.Errorf("first StartCapture() = %d, want 1", captureID)
	}

	recording, _ := store.RecordingState()
	if !recording {
		t.Error("recording should be true after StartCapture")
	}

	captureID2 := store.StartCapture()
	if captureID2 != 2 {
		t.Errorf("second StartCapture() = %d, want 2", captureID2)
	}
}

func TestStopCapture(t *testing.T) {
	store := NewStore(10, 100, 0)
	store.StartCapture()

	captureID := store.StopCapture()
	if captureID != 1 {
		t.Errorf("StopCapture() = %d, want 1", captureID)
	}

	recording, _ := store.RecordingState()
	if recording {
		t.Error("recording should be false after StopCapture")
	}
}

func TestCreateSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{
		ID:     "test-session",
		Target: "ws://example.com",
	}

	err := store.CreateSession(ctx, sess)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session was created
	got, ok, err := store.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if !ok {
		t.Fatal("session not found after creation")
	}
	if got.ID != sess.ID {
		t.Errorf("got session ID %s, want %s", got.ID, sess.ID)
	}
	if got.Target != sess.Target {
		t.Errorf("got Target %s, want %s", got.Target, sess.Target)
	}
}

func TestCreateSession_AssignsCaptureID(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.StartCapture()

	sess := domain.Session{
		ID:     "test-session",
		Target: "ws://example.com",
	}

	err := store.CreateSession(ctx, sess)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	got, ok, _ := store.GetSession(ctx, sess.ID)
	if !ok {
		t.Fatal("session not found")
	}
	if got.CaptureID == nil {
		t.Error("CaptureID should be assigned when recording")
	} else if *got.CaptureID != 1 {
		t.Errorf("CaptureID = %d, want 1", *got.CaptureID)
	}
}

func TestCreateSession_EvictsByCapacity(t *testing.T) {
	store := NewStore(2, 100, 0)
	ctx := context.Background()

	// Create 3 sessions
	sess1 := domain.Session{ID: "sess1", Target: "ws://1.com"}
	sess2 := domain.Session{ID: "sess2", Target: "ws://2.com"}
	sess3 := domain.Session{ID: "sess3", Target: "ws://3.com"}

	store.CreateSession(ctx, sess1)
	store.CreateSession(ctx, sess2)
	store.CreateSession(ctx, sess3) // Should evict sess1

	// sess1 should be evicted
	_, ok, _ := store.GetSession(ctx, sess1.ID)
	if ok {
		t.Error("sess1 should have been evicted")
	}

	// sess2 and sess3 should exist
	_, ok, _ = store.GetSession(ctx, sess2.ID)
	if !ok {
		t.Error("sess2 should exist")
	}
	_, ok, _ = store.GetSession(ctx, sess3.ID)
	if !ok {
		t.Error("sess3 should exist")
	}
}

func TestGetSession_NotFound(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	_, ok, err := store.GetSession(ctx, "non-existent")
	if err != nil {
		t.Errorf("GetSession returned error: %v", err)
	}
	if ok {
		t.Error("GetSession should return false for non-existent session")
	}
}

func TestDeleteSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test-session", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	err := store.DeleteSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, ok, _ := store.GetSession(ctx, sess.ID)
	if ok {
		t.Error("session should be deleted")
	}
}

func TestDeleteSession_NonExistent(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	err := store.DeleteSession(ctx, "non-existent")
	if err != nil {
		t.Errorf("DeleteSession should not error for non-existent session: %v", err)
	}
}

func TestDeleteSession_WithSpoolFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create session and add spool files
	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Create temporary spool files
	tempFile1, _ := os.CreateTemp("", "spool-test-1-*.bin")
	tempFile2, _ := os.CreateTemp("", "spool-test-2-*.bin")
	tempFile1.Close()
	tempFile2.Close()

	path1 := tempFile1.Name()
	path2 := tempFile2.Name()

	store.AddSpoolFile(ctx, sess.ID, path1)
	store.AddSpoolFile(ctx, sess.ID, path2)

	// Delete session - should cleanup spool files
	err := store.DeleteSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify spool files are deleted
	if _, err := os.Stat(path1); !os.IsNotExist(err) {
		t.Error("spool file 1 should be deleted")
	}
	if _, err := os.Stat(path2); !os.IsNotExist(err) {
		t.Error("spool file 2 should be deleted")
	}
}

func TestDeleteSession_WithFrameBodyFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Create temporary body file
	tempFile, _ := os.CreateTemp("", "frame-body-*.bin")
	tempFile.Close()
	bodyPath := tempFile.Name()

	// Add frame with body file
	frame := domain.Frame{ID: "frame1", BodyFile: bodyPath}
	store.AppendFrame(ctx, sess.ID, frame)

	// Delete session
	err := store.DeleteSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify body file is deleted
	if _, err := os.Stat(bodyPath); !os.IsNotExist(err) {
		t.Error("frame body file should be deleted")
	}
}

func TestDeleteSession_WithHTTPTxBodyFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Create temporary body files
	reqFile, _ := os.CreateTemp("", "req-body-*.bin")
	respFile, _ := os.CreateTemp("", "resp-body-*.bin")
	reqFile.Close()
	respFile.Close()
	reqPath := reqFile.Name()
	respPath := respFile.Name()

	// Add HTTP transaction with body files
	tx := domain.HTTPTransaction{
		ID:           "tx1",
		SessionID:    sess.ID,
		ReqBodyFile:  reqPath,
		RespBodyFile: respPath,
	}
	store.AppendHTTPTransaction(ctx, tx)

	// Delete session
	err := store.DeleteSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify body files are deleted
	if _, err := os.Stat(reqPath); !os.IsNotExist(err) {
		t.Error("request body file should be deleted")
	}
	if _, err := os.Stat(respPath); !os.IsNotExist(err) {
		t.Error("response body file should be deleted")
	}
}

func TestClearAllSessions(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create multiple sessions
	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://1.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://2.com"})

	err := store.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("ClearAllSessions failed: %v", err)
	}

	// Verify all sessions are cleared
	sessions, total, _ := store.ListSessions(ctx, usecase.SessionFilter{Limit: 100, IncludeUnassigned: true})
	if total != 0 {
		t.Errorf("total sessions = %d, want 0", total)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

func TestClearAllSessions_EmptyStore(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Clear when already empty - should not error
	err := store.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("ClearAllSessions on empty store failed: %v", err)
	}
}

func TestClearAllSessions_WithSpoolFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create sessions with spool files
	sess1 := domain.Session{ID: "sess1", Target: "ws://1.com"}
	sess2 := domain.Session{ID: "sess2", Target: "ws://2.com"}
	store.CreateSession(ctx, sess1)
	store.CreateSession(ctx, sess2)

	// Add spool files to both sessions
	tempFile1, _ := os.CreateTemp("", "clear-spool-1-*.bin")
	tempFile2, _ := os.CreateTemp("", "clear-spool-2-*.bin")
	tempFile1.Close()
	tempFile2.Close()
	path1 := tempFile1.Name()
	path2 := tempFile2.Name()

	store.AddSpoolFile(ctx, sess1.ID, path1)
	store.AddSpoolFile(ctx, sess2.ID, path2)

	// Clear all sessions
	err := store.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("ClearAllSessions failed: %v", err)
	}

	// Verify spool files are deleted
	if _, err := os.Stat(path1); !os.IsNotExist(err) {
		t.Error("spool file 1 should be deleted")
	}
	if _, err := os.Stat(path2); !os.IsNotExist(err) {
		t.Error("spool file 2 should be deleted")
	}

	// Verify sessions are cleared
	sessions, _, _ := store.ListSessions(ctx, usecase.SessionFilter{Limit: 100, IncludeUnassigned: true})
	if len(sessions) != 0 {
		t.Error("all sessions should be cleared")
	}
}

func TestClearAllSessions_WithFrameAndHTTPFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "sess1", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add frame with body file
	frameFile, _ := os.CreateTemp("", "clear-frame-*.bin")
	frameFile.Close()
	framePath := frameFile.Name()
	frame := domain.Frame{ID: "f1", BodyFile: framePath}
	store.AppendFrame(ctx, sess.ID, frame)

	// Add HTTP transaction with body files
	reqFile, _ := os.CreateTemp("", "clear-req-*.bin")
	respFile, _ := os.CreateTemp("", "clear-resp-*.bin")
	reqFile.Close()
	respFile.Close()
	reqPath := reqFile.Name()
	respPath := respFile.Name()

	tx := domain.HTTPTransaction{
		ID:           "tx1",
		SessionID:    sess.ID,
		ReqBodyFile:  reqPath,
		RespBodyFile: respPath,
	}
	store.AppendHTTPTransaction(ctx, tx)

	// Clear all sessions
	err := store.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("ClearAllSessions failed: %v", err)
	}

	// Verify files are deleted
	if _, err := os.Stat(framePath); !os.IsNotExist(err) {
		t.Error("frame body file should be deleted")
	}
	if _, err := os.Stat(reqPath); !os.IsNotExist(err) {
		t.Error("request body file should be deleted")
	}
	if _, err := os.Stat(respPath); !os.IsNotExist(err) {
		t.Error("response body file should be deleted")
	}
}

func TestListSessions_Empty(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

func TestListSessions_Basic(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.com"})

	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestListSessions_Pagination(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		store.CreateSession(ctx, domain.Session{ID: string(rune('a' + i)), Target: "ws://example.com"})
	}

	// Test pagination
	sessions, total, _ := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             2,
		Offset:            1,
		IncludeUnassigned: true,
	})
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestListSessions_CaptureIDExactMatch(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create sessions in different captures
	captureID1 := store.StartCapture()
	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})

	store.StartCapture()
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.com"})

	// Stop recording and create an unassigned session
	store.StopCapture()
	store.CreateSession(ctx, domain.Session{ID: "sess3", Target: "ws://other.com"})

	// Filter by exact captureID
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		CaptureID:         &captureID1,
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].ID != "sess1" {
		t.Errorf("session[0].ID = %s, want sess1", sessions[0].ID)
	}
	if sessions[1].ID != "sess3" {
		t.Errorf("session[1].ID = %s, want sess3 (unassigned)", sessions[1].ID)
	}
}

func TestListSessions_CaptureIDCurrentCapture(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create session in capture 1
	store.StartCapture()
	store.CreateSession(ctx, domain.Session{ID: "old", Target: "ws://old.com"})

	// Create session in capture 2 (current)
	store.StartCapture()
	store.CreateSession(ctx, domain.Session{ID: "current", Target: "ws://current.com"})

	// Stop recording and create an unassigned session (should be included when IncludeUnassigned=true)
	store.StopCapture()
	store.CreateSession(ctx, domain.Session{ID: "unassigned", Target: "ws://paused.com"})

	// Filter by current capture using -1
	minusOne := -1
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		CaptureID:         &minusOne,
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].ID != "current" || sessions[1].ID != "unassigned" {
		t.Errorf("unexpected sessions order/ids: %+v", []string{sessions[0].ID, sessions[1].ID})
	}
}

func TestListSessions_ExcludeUnassigned(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create an assigned session while recording
	store.StartCapture()
	store.CreateSession(ctx, domain.Session{ID: "assigned", Target: "ws://example.com"})

	// Stop recording and create an unassigned session
	store.StopCapture()
	store.CreateSession(ctx, domain.Session{ID: "unassigned", Target: "ws://test.com"})

	// IncludeUnassigned = false should exclude sessions without captureID
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             10,
		IncludeUnassigned: false,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(sessions) != 1 || sessions[0].ID != "assigned" {
		t.Error("expected only assigned session")
	}
}

func TestListSessions_TargetFilter(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.org"})
	store.CreateSession(ctx, domain.Session{ID: "sess3", Target: "wss://example.net"})

	// Filter by target substring (case-insensitive)
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Target:            "example",
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestListSessions_TextSearchFilter(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://api.example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://web.test.org"})
	store.CreateSession(ctx, domain.Session{ID: "sess3", Target: "ws://api.other.net"})

	// Search by 'Q' parameter
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Q:                 "api",
		Limit:             10,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestListSessions_OffsetExceedsTotal(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.com"})

	// Offset > total should return empty list
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             10,
		Offset:            100,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

func TestListSessions_ZeroLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.com"})

	// Limit = 0 should return all
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             0,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestListSessions_NegativeLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	store.CreateSession(ctx, domain.Session{ID: "sess1", Target: "ws://example.com"})
	store.CreateSession(ctx, domain.Session{ID: "sess2", Target: "ws://test.com"})

	// Negative limit should return all
	sessions, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             -1,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestIncrementCounters(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Increment with text frame
	frame := domain.Frame{Opcode: domain.OpcodeText}
	err := store.IncrementCounters(ctx, sess.ID, frame)
	if err != nil {
		t.Fatalf("IncrementCounters failed: %v", err)
	}

	// Verify counters
	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.Frames.Total != 1 {
		t.Errorf("Frames.Total = %d, want 1", got.Frames.Total)
	}
	if got.Frames.Text != 1 {
		t.Errorf("Frames.Text = %d, want 1", got.Frames.Text)
	}
}

func TestIncrementCounters_Binary(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	frame := domain.Frame{Opcode: domain.OpcodeBinary}
	store.IncrementCounters(ctx, sess.ID, frame)

	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.Frames.Binary != 1 {
		t.Errorf("Frames.Binary = %d, want 1", got.Frames.Binary)
	}
}

func TestIncrementCounters_Control(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	frame := domain.Frame{Opcode: domain.OpcodePing}
	store.IncrementCounters(ctx, sess.ID, frame)

	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.Frames.Control != 1 {
		t.Errorf("Frames.Control = %d, want 1", got.Frames.Control)
	}
}

func TestSetClosed(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	ts := time.Now()
	errMsg := "connection closed"
	err := store.SetClosed(ctx, sess.ID, ts, &errMsg)
	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}

	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.ClosedAt == nil {
		t.Error("ClosedAt should be set")
	} else if !got.ClosedAt.Equal(ts) {
		t.Errorf("ClosedAt = %v, want %v", got.ClosedAt, ts)
	}
	if got.Error == nil {
		t.Error("Error should be set")
	} else if *got.Error != errMsg {
		t.Errorf("Error = %s, want %s", *got.Error, errMsg)
	}
}

func TestSetClosed_WithSpoolFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Create and add spool files
	tempFile, _ := os.CreateTemp("", "spool-closed-*.bin")
	tempFile.Close()
	spoolPath := tempFile.Name()
	defer os.Remove(spoolPath) // cleanup after test
	store.AddSpoolFile(ctx, sess.ID, spoolPath)

	// Close session - should NOT cleanup spool files (session is still viewable)
	ts := time.Now()
	err := store.SetClosed(ctx, sess.ID, ts, nil)
	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}

	// Verify spool file is NOT deleted (body files needed for viewing closed sessions)
	if _, err := os.Stat(spoolPath); os.IsNotExist(err) {
		t.Error("spool file should NOT be deleted on close - it's needed for viewing")
	}
}

func TestSetClosed_WithFrameBodyFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add frame with body file
	tempFile, _ := os.CreateTemp("", "frame-closed-*.bin")
	tempFile.Close()
	bodyPath := tempFile.Name()
	defer os.Remove(bodyPath) // cleanup after test
	frame := domain.Frame{ID: "frame1", BodyFile: bodyPath}
	store.AppendFrame(ctx, sess.ID, frame)

	// Close session - should NOT cleanup body files (session is still viewable)
	ts := time.Now()
	err := store.SetClosed(ctx, sess.ID, ts, nil)
	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}

	// Verify body file is NOT deleted (needed for viewing closed sessions)
	if _, err := os.Stat(bodyPath); os.IsNotExist(err) {
		t.Error("frame body file should NOT be deleted on close - it's needed for viewing")
	}
}

func TestSetClosed_WithHTTPTxBodyFiles(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add HTTP transaction with body files
	reqFile, _ := os.CreateTemp("", "req-closed-*.bin")
	respFile, _ := os.CreateTemp("", "resp-closed-*.bin")
	reqFile.Close()
	respFile.Close()
	reqPath := reqFile.Name()
	respPath := respFile.Name()
	defer os.Remove(reqPath)  // cleanup after test
	defer os.Remove(respPath) // cleanup after test

	tx := domain.HTTPTransaction{
		ID:           "tx1",
		SessionID:    sess.ID,
		ReqBodyFile:  reqPath,
		RespBodyFile: respPath,
	}
	store.AppendHTTPTransaction(ctx, tx)

	// Close session - should NOT cleanup body files (session is still viewable)
	ts := time.Now()
	err := store.SetClosed(ctx, sess.ID, ts, nil)
	if err != nil {
		t.Fatalf("SetClosed failed: %v", err)
	}

	// Verify body files are NOT deleted (needed for viewing closed sessions)
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		t.Error("request body file should NOT be deleted on close - it's needed for viewing")
	}
	if _, err := os.Stat(respPath); os.IsNotExist(err) {
		t.Error("response body file should NOT be deleted on close - it's needed for viewing")
	}
}

func TestSetClosed_NonExistentSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	ts := time.Now()
	err := store.SetClosed(ctx, "non-existent", ts, nil)
	if err != nil {
		t.Errorf("SetClosed should not error for non-existent session: %v", err)
	}
}

func TestAppendFrame(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	frame := domain.Frame{ID: "frame1", Opcode: domain.OpcodeText}
	err := store.AppendFrame(ctx, sess.ID, frame)
	if err != nil {
		t.Fatalf("AppendFrame failed: %v", err)
	}

	frames, _, _ := store.ListFrames(ctx, sess.ID, "", 10)
	if len(frames) != 1 {
		t.Errorf("len(frames) = %d, want 1", len(frames))
	}
	if frames[0].ID != "frame1" {
		t.Errorf("frame ID = %s, want frame1", frames[0].ID)
	}
}

func TestCreateSession_IdempotentByID(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "same-id", Target: "ws://example.com"}
	if err := store.CreateSession(ctx, sess); err != nil {
		t.Fatalf("first CreateSession failed: %v", err)
	}
	if err := store.CreateSession(ctx, sess); err != nil {
		t.Fatalf("second CreateSession failed: %v", err)
	}

	items, total, err := store.ListSessions(ctx, usecase.SessionFilter{
		Limit:             100,
		IncludeUnassigned: true,
	})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("total=%d, want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("len(items)=%d, want 1", len(items))
	}
	if items[0].ID != "same-id" {
		t.Fatalf("id=%s, want same-id", items[0].ID)
	}
}

func TestAppendFrame_Eviction(t *testing.T) {
	store := NewStore(10, 2, 0) // maxFrames = 2
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add 3 frames
	store.AppendFrame(ctx, sess.ID, domain.Frame{ID: "frame1"})
	store.AppendFrame(ctx, sess.ID, domain.Frame{ID: "frame2"})
	store.AppendFrame(ctx, sess.ID, domain.Frame{ID: "frame3"})

	frames, _, _ := store.ListFrames(ctx, sess.ID, "", 10)
	if len(frames) != 2 {
		t.Errorf("len(frames) = %d, want 2", len(frames))
	}
	// frame1 should be evicted
	if frames[0].ID == "frame1" {
		t.Error("frame1 should have been evicted")
	}
}

func TestListFrames_Empty(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	frames, next, err := store.ListFrames(ctx, sess.ID, "", 10)
	if err != nil {
		t.Fatalf("ListFrames failed: %v", err)
	}
	if len(frames) != 0 {
		t.Errorf("len(frames) = %d, want 0", len(frames))
	}
	if next != "" {
		t.Errorf("next = %s, want empty", next)
	}
}

func TestListFrames_Pagination(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add 5 frames
	for i := 0; i < 5; i++ {
		store.AppendFrame(ctx, sess.ID, domain.Frame{ID: string(rune('a' + i))})
	}

	// Get first 2 frames
	frames, next, _ := store.ListFrames(ctx, sess.ID, "", 2)
	if len(frames) != 2 {
		t.Errorf("len(frames) = %d, want 2", len(frames))
	}
	if next != "b" {
		t.Errorf("next = %s, want b", next)
	}

	// Get next 2 frames
	frames, next, _ = store.ListFrames(ctx, sess.ID, next, 2)
	if len(frames) != 2 {
		t.Errorf("len(frames) = %d, want 2", len(frames))
	}
	if frames[0].ID != "c" {
		t.Errorf("first frame ID = %s, want c", frames[0].ID)
	}
}

func TestAppendEvent(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	event := domain.Event{ID: "event1", Namespace: "/chat"}
	err := store.AppendEvent(ctx, sess.ID, event)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	events, _, _ := store.ListEvents(ctx, sess.ID, "", 10)
	if len(events) != 1 {
		t.Errorf("len(events) = %d, want 1", len(events))
	}
	if events[0].ID != "event1" {
		t.Errorf("event ID = %s, want event1", events[0].ID)
	}

	// Verify counters updated
	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.Events.Total != 1 {
		t.Errorf("Events.Total = %d, want 1", got.Events.Total)
	}
	if got.Events.SIO != 1 {
		t.Errorf("Events.SIO = %d, want 1", got.Events.SIO)
	}
}

func TestAppendEvent_SystemEventNotCounted(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	event := domain.Event{ID: "event1", Namespace: "/_sys"}
	store.AppendEvent(ctx, sess.ID, event)

	got, _, _ := store.GetSession(ctx, sess.ID)
	if got.Events.Total != 0 {
		t.Errorf("Events.Total = %d, want 0 for system events", got.Events.Total)
	}
}

func TestAppendEvent_Eviction(t *testing.T) {
	store := NewStore(10, 2, 0) // maxEvents = 2
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add 3 events
	store.AppendEvent(ctx, sess.ID, domain.Event{ID: "event1", Namespace: "/chat"})
	store.AppendEvent(ctx, sess.ID, domain.Event{ID: "event2", Namespace: "/chat"})
	store.AppendEvent(ctx, sess.ID, domain.Event{ID: "event3", Namespace: "/chat"})

	events, _, _ := store.ListEvents(ctx, sess.ID, "", 10)
	if len(events) != 2 {
		t.Errorf("len(events) = %d, want 2", len(events))
	}
	if events[0].ID == "event1" {
		t.Error("event1 should have been evicted")
	}
	if store.droppedEventsTotal != 1 {
		t.Errorf("droppedEventsTotal = %d, want 1", store.droppedEventsTotal)
	}
}

func TestAppendHTTPTransaction(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	tx := domain.HTTPTransaction{ID: "tx1", SessionID: sess.ID}
	err := store.AppendHTTPTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("AppendHTTPTransaction failed: %v", err)
	}

	txs, _, _ := store.ListHTTPTransactions(ctx, sess.ID, "", 10)
	if len(txs) != 1 {
		t.Errorf("len(txs) = %d, want 1", len(txs))
	}
	if txs[0].ID != "tx1" {
		t.Errorf("tx ID = %s, want tx1", txs[0].ID)
	}
}

func TestListHTTPTransactions_Pagination(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add 5 transactions
	for i := 0; i < 5; i++ {
		tx := domain.HTTPTransaction{ID: string(rune('a' + i)), SessionID: sess.ID}
		store.AppendHTTPTransaction(ctx, tx)
	}

	// Get first 2 transactions
	txs, next, _ := store.ListHTTPTransactions(ctx, sess.ID, "", 2)
	if len(txs) != 2 {
		t.Errorf("len(txs) = %d, want 2", len(txs))
	}
	if next != "b" {
		t.Errorf("next = %s, want b", next)
	}
}

func TestListHTTPTransactions_NonExistentSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	txs, next, err := store.ListHTTPTransactions(ctx, "nonexistent", "", 10)
	if err != nil {
		t.Fatalf("ListHTTPTransactions should not error: %v", err)
	}
	if txs != nil {
		t.Error("txs should be nil for nonexistent session")
	}
	if next != "" {
		t.Error("next should be empty for nonexistent session")
	}
}

func TestListHTTPTransactions_ZeroLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		tx := domain.HTTPTransaction{ID: string(rune('a' + i)), SessionID: sess.ID}
		store.AppendHTTPTransaction(ctx, tx)
	}

	// Zero limit should return all
	txs, _, _ := store.ListHTTPTransactions(ctx, sess.ID, "", 0)
	if len(txs) != 3 {
		t.Errorf("len(txs) = %d, want 3", len(txs))
	}
}

func TestListHTTPTransactions_InvalidFrom(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		tx := domain.HTTPTransaction{ID: string(rune('a' + i)), SessionID: sess.ID}
		store.AppendHTTPTransaction(ctx, tx)
	}

	// "from" not found - should start from beginning
	txs, _, _ := store.ListHTTPTransactions(ctx, sess.ID, "nonexistent", 10)
	if len(txs) != 3 {
		t.Errorf("len(txs) = %d, want 3", len(txs))
	}
}

func TestListHTTPTransactions_NegativeLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		tx := domain.HTTPTransaction{ID: string(rune('a' + i)), SessionID: sess.ID}
		store.AppendHTTPTransaction(ctx, tx)
	}

	// Negative limit should return all
	txs, _, _ := store.ListHTTPTransactions(ctx, sess.ID, "", -1)
	if len(txs) != 3 {
		t.Errorf("len(txs) = %d, want 3", len(txs))
	}
}

func TestListHTTPTransactions_FromLastElement(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		tx := domain.HTTPTransaction{ID: string(rune('a' + i)), SessionID: sess.ID}
		store.AppendHTTPTransaction(ctx, tx)
	}

	// from="c" (last element) should return empty
	txs, next, _ := store.ListHTTPTransactions(ctx, sess.ID, "c", 10)
	if len(txs) != 0 {
		t.Errorf("len(txs) = %d, want 0", len(txs))
	}
	if next != "" {
		t.Errorf("next = %q, want empty", next)
	}
}

func TestListHTTPTransactions_EmptyList(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// No transactions added
	txs, next, err := store.ListHTTPTransactions(ctx, sess.ID, "", 10)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("len(txs) = %d, want 0", len(txs))
	}
	if next != "" {
		t.Errorf("next = %q, want empty", next)
	}
}

func TestListEvents_NonExistentSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	events, next, err := store.ListEvents(ctx, "nonexistent", "", 10)
	if err != nil {
		t.Fatalf("ListEvents should not error: %v", err)
	}
	if events != nil {
		t.Error("events should be nil for nonexistent session")
	}
	if next != "" {
		t.Error("next should be empty for nonexistent session")
	}
}

func TestListEvents_ZeroLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		evt := domain.Event{ID: string(rune('a' + i))}
		store.AppendEvent(ctx, sess.ID, evt)
	}

	// Zero limit should return all
	events, _, _ := store.ListEvents(ctx, sess.ID, "", 0)
	if len(events) != 3 {
		t.Errorf("len(events) = %d, want 3", len(events))
	}
}

func TestListFrames_NonExistentSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	frames, next, err := store.ListFrames(ctx, "nonexistent", "", 10)
	if err != nil {
		t.Fatalf("ListFrames should not error: %v", err)
	}
	if frames != nil {
		t.Error("frames should be nil for nonexistent session")
	}
	if next != "" {
		t.Error("next should be empty for nonexistent session")
	}
}

func TestListFrames_ZeroLimit(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	for i := 0; i < 3; i++ {
		frame := domain.Frame{ID: string(rune('a' + i))}
		store.AppendFrame(ctx, sess.ID, frame)
	}

	// Zero limit should return all
	frames, _, _ := store.ListFrames(ctx, sess.ID, "", 0)
	if len(frames) != 3 {
		t.Errorf("len(frames) = %d, want 3", len(frames))
	}
}

func TestContainsFold(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "hello", true},
		{"Hello World", "foo", false},
		{"", "", true},
		{"test", "", true},
	}

	for _, tt := range tests {
		got := containsFold(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsFold(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s    string
		sub  string
		want int
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "foo", -1},
		{"", "", 0},
		{"test", "", 0},
	}

	for _, tt := range tests {
		got := indexOf(tt.s, tt.sub)
		if got != tt.want {
			t.Errorf("indexOf(%q, %q) = %d, want %d", tt.s, tt.sub, got, tt.want)
		}
	}
}

func TestAddSpoolFile_ExistingSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add spool file
	store.AddSpoolFile(ctx, sess.ID, "/tmp/spool-file-1.bin")

	// Verify spool file was added
	store.mu.RLock()
	e, ok := store.items[sess.ID]
	store.mu.RUnlock()

	if !ok {
		t.Fatal("session should exist")
	}
	if len(e.spoolFiles) != 1 {
		t.Errorf("len(spoolFiles) = %d, want 1", len(e.spoolFiles))
	}
	if e.spoolFiles[0] != "/tmp/spool-file-1.bin" {
		t.Errorf("spoolFile = %s, want /tmp/spool-file-1.bin", e.spoolFiles[0])
	}
}

func TestAddSpoolFile_MultiplePaths(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	sess := domain.Session{ID: "test", Target: "ws://example.com"}
	store.CreateSession(ctx, sess)

	// Add multiple spool files
	store.AddSpoolFile(ctx, sess.ID, "/tmp/file-1.bin")
	store.AddSpoolFile(ctx, sess.ID, "/tmp/file-2.bin")
	store.AddSpoolFile(ctx, sess.ID, "/tmp/file-3.bin")

	store.mu.RLock()
	e := store.items[sess.ID]
	store.mu.RUnlock()

	if len(e.spoolFiles) != 3 {
		t.Errorf("len(spoolFiles) = %d, want 3", len(e.spoolFiles))
	}
}

func TestAddSpoolFile_NonExistentSession(t *testing.T) {
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Try to add spool file to non-existent session - should not panic
	store.AddSpoolFile(ctx, "non-existent", "/tmp/spool-file.bin")

	store.mu.RLock()
	e, ok := store.items["non-existent"]
	store.mu.RUnlock()

	if ok || e != nil {
		t.Error("non-existent session should not be created")
	}
}

func TestEvictExpiredSessions_WithTTL(t *testing.T) {
	// Create store with 100ms TTL
	store := NewStore(10, 100, 100*time.Millisecond)
	ctx := context.Background()

	// Create first session
	sess1 := domain.Session{ID: "sess1", Target: "ws://example.com"}
	store.CreateSession(ctx, sess1)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Create second session - should trigger eviction of sess1
	sess2 := domain.Session{ID: "sess2", Target: "ws://example.com"}
	store.CreateSession(ctx, sess2)

	// sess1 should be evicted
	_, ok, _ := store.GetSession(ctx, sess1.ID)
	if ok {
		t.Error("sess1 should have been evicted by TTL")
	}

	// sess2 should exist
	_, ok, _ = store.GetSession(ctx, sess2.ID)
	if !ok {
		t.Error("sess2 should exist")
	}
}

func TestEvictExpiredSessions_NoTTL(t *testing.T) {
	// Create store with zero TTL (disabled)
	store := NewStore(10, 100, 0)
	ctx := context.Background()

	// Create sessions
	sess1 := domain.Session{ID: "sess1", Target: "ws://example.com"}
	store.CreateSession(ctx, sess1)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Create another session - sess1 should NOT be evicted (TTL disabled)
	sess2 := domain.Session{ID: "sess2", Target: "ws://example.com"}
	store.CreateSession(ctx, sess2)

	// Both should exist (no TTL eviction)
	_, ok, _ := store.GetSession(ctx, sess1.ID)
	if !ok {
		t.Error("sess1 should still exist (TTL disabled)")
	}
	_, ok, _ = store.GetSession(ctx, sess2.ID)
	if !ok {
		t.Error("sess2 should exist")
	}
}

func TestEvictExpiredSessions_NegativeTTL(t *testing.T) {
	// Create store with negative TTL (should behave like disabled)
	store := NewStore(10, 100, -1*time.Second)
	ctx := context.Background()

	sess1 := domain.Session{ID: "sess1", Target: "ws://example.com"}
	store.CreateSession(ctx, sess1)

	sess2 := domain.Session{ID: "sess2", Target: "ws://example.com"}
	store.CreateSession(ctx, sess2)

	// Both should exist (negative TTL treated as disabled)
	_, ok, _ := store.GetSession(ctx, sess1.ID)
	if !ok {
		t.Error("sess1 should still exist (negative TTL)")
	}
}

func TestEvictExpiredSessions_MultipleExpired(t *testing.T) {
	// Create store with 100ms TTL
	store := NewStore(10, 100, 100*time.Millisecond)
	ctx := context.Background()

	// Create multiple sessions
	sess1 := domain.Session{ID: "sess1", Target: "ws://example.com"}
	sess2 := domain.Session{ID: "sess2", Target: "ws://example.com"}
	store.CreateSession(ctx, sess1)
	store.CreateSession(ctx, sess2)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Create new session - should evict both old ones
	sess3 := domain.Session{ID: "sess3", Target: "ws://example.com"}
	store.CreateSession(ctx, sess3)

	// sess1 and sess2 should be evicted
	_, ok, _ := store.GetSession(ctx, sess1.ID)
	if ok {
		t.Error("sess1 should have been evicted")
	}
	_, ok, _ = store.GetSession(ctx, sess2.ID)
	if ok {
		t.Error("sess2 should have been evicted")
	}

	// sess3 should exist
	_, ok, _ = store.GetSession(ctx, sess3.ID)
	if !ok {
		t.Error("sess3 should exist")
	}
}
