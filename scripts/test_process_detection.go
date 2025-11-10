//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Session struct {
	ID          string       `json:"id"`
	Target      string       `json:"target"`
	ClientAddr  string       `json:"clientAddr"`
	StartedAt   time.Time    `json:"startedAt"`
	ProcessInfo *ProcessInfo `json:"processInfo,omitempty"`
}

type ProcessInfo struct {
	PID            int32     `json:"pid"`
	Name           string    `json:"name"`
	ExecutablePath string    `json:"executablePath"`
	BundleID       *string   `json:"bundleId,omitempty"`
	DetectedAt     time.Time `json:"detectedAt"`
}

type SessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Total    int       `json:"total"`
}

func main() {
	fmt.Println("Testing process detection...")

	// 1. Configure proxy
	proxyURL, _ := url.Parse("http://localhost:9091")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	// 2. Make a request through the proxy
	fmt.Println("\n1. Making HTTP request through proxy...")
	resp, err := client.Get("http://httpbin.org/get")
	if err != nil {
		fmt.Printf("   ❌ Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	fmt.Printf("   ✅ Request completed: %s\n", resp.Status)

	// 3. Wait a bit for session to be created
	time.Sleep(1 * time.Second)

	// 4. Fetch sessions from API
	fmt.Println("\n2. Fetching sessions from API...")
	apiResp, err := http.Get("http://localhost:9092/_api/v1/sessions?limit=10")
	if err != nil {
		fmt.Printf("   ❌ Failed to fetch sessions: %v\n", err)
		return
	}
	defer apiResp.Body.Close()

	var sessionsResp SessionsResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&sessionsResp); err != nil {
		fmt.Printf("   ❌ Failed to decode response: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Found %d sessions\n", len(sessionsResp.Sessions))

	// 5. Check process detection results
	fmt.Println("\n3. Process detection results:")
	foundProcessInfo := false
	for i, session := range sessionsResp.Sessions {
		fmt.Printf("\n   Session #%d: %s\n", i+1, session.ID)
		fmt.Printf("   Target:     %s\n", session.Target)
		fmt.Printf("   Client:     %s\n", session.ClientAddr)

		if session.ProcessInfo != nil {
			foundProcessInfo = true
			fmt.Printf("   ✅ Process detected!\n")
			fmt.Printf("      PID:        %d\n", session.ProcessInfo.PID)
			fmt.Printf("      Name:       %s\n", session.ProcessInfo.Name)
			fmt.Printf("      Path:       %s\n", session.ProcessInfo.ExecutablePath)
			if session.ProcessInfo.BundleID != nil {
				fmt.Printf("      Bundle ID:  %s\n", *session.ProcessInfo.BundleID)
			}
		} else {
			fmt.Printf("   ⚠️  No process info detected\n")
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	if foundProcessInfo {
		fmt.Println("✅ SUCCESS: Process detection is working!")
	} else {
		fmt.Println("⚠️  Process info not detected in sessions")
		fmt.Println("   This could be expected if:")
		fmt.Println("   - Detection is disabled in config")
		fmt.Println("   - No suitable process found for the connection")
		fmt.Println("   - Insufficient permissions")
	}
	fmt.Println(strings.Repeat("=", 60))
}
