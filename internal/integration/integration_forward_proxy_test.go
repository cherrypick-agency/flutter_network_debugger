package integration

import (
    "bufio"
    "encoding/json"
    "io"
    "net"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strconv"
    "strings"
    "testing"
    "time"
)

// Helper: write a raw absolute-URI GET via proxy connection and return status line and body
func rawProxyGET(t *testing.T, proxyHost string, absoluteURL string, extraHeaders []string) (string, string) {
    t.Helper()
    conn, err := net.DialTimeout("tcp", proxyHost, 3*time.Second)
    if err != nil {
        t.Fatalf("dial proxy: %v", err)
    }
    defer conn.Close()
    var b strings.Builder
    b.WriteString("GET ")
    b.WriteString(absoluteURL)
    b.WriteString(" HTTP/1.1\r\n")
    b.WriteString("Host: ")
    b.WriteString(proxyHost)
    b.WriteString("\r\n")
    for _, h := range extraHeaders {
        b.WriteString(h)
        if !strings.HasSuffix(h, "\r\n") {
            b.WriteString("\r\n")
        }
    }
    b.WriteString("\r\n")
    if _, err := conn.Write([]byte(b.String())); err != nil {
        t.Fatalf("write: %v", err)
    }
    br := bufio.NewReader(conn)
    status, _ := br.ReadString('\n')
    // Read and parse headers to determine body length
    headers := map[string]string{}
    for {
        line, _ := br.ReadString('\n')
        if line == "\r\n" || line == "\n" || line == "" {
            break
        }
        if i := strings.IndexByte(line, ':'); i > 0 {
            k := strings.TrimSpace(line[:i])
            v := strings.TrimSpace(strings.TrimRight(line[i+1:], "\r\n"))
            headers[strings.ToLower(k)] = v
        }
    }
    // Prefer Content-Length if present, otherwise read until close
    var body []byte
    if cl, ok := headers["content-length"]; ok {
        if n, err := strconv.Atoi(cl); err == nil && n >= 0 {
            buf := make([]byte, n)
            _, _ = io.ReadFull(br, buf)
            body = buf
        } else {
            // fallback
            body, _ = io.ReadAll(br)
        }
    } else {
        body, _ = io.ReadAll(br)
    }
    return status, string(body)
}

func TestForwardProxy_AbsoluteURI_SetsForwardHeaders(t *testing.T) {
    t.Parallel()
    // upstream echoes forward headers
    mux := http.NewServeMux()
    mux.HandleFunc("/hdr", func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode(map[string]string{
            "xff": r.Header.Get("X-Forwarded-For"),
            "xfp": r.Header.Get("X-Forwarded-Proto"),
            "via": r.Header.Get("Via"),
        })
    })
    upstream := httptest.NewServer(mux)
    defer upstream.Close()

    app, _ := startHTTPApp(t)
    defer app.Close()
    pURL, _ := url.Parse(app.URL)

    status, body := rawProxyGET(t, pURL.Host, upstream.URL+"/hdr", nil)
    if !strings.HasPrefix(status, "HTTP/1.1 200") {
        t.Fatalf("status: %s", status)
    }
    var got map[string]string
    // body is an HTTP stream; scan for last JSON line
    dec := json.NewDecoder(strings.NewReader(body))
    _ = dec.Decode(&got)
    if got["xfp"] == "" || got["via"] == "" {
        t.Fatalf("forward headers not set: %+v", got)
    }
}

func TestForwardProxy_HopByHopStripped(t *testing.T) {
    t.Parallel()
    mux := http.NewServeMux()
    mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode(map[string]string{
            "conn": r.Header.Get("Connection"),
            "pc":   r.Header.Get("Proxy-Connection"),
            "te":   r.Header.Get("Te"),
        })
    })
    upstream := httptest.NewServer(mux)
    defer upstream.Close()
    app, _ := startHTTPApp(t)
    defer app.Close()
    pURL, _ := url.Parse(app.URL)

    headers := []string{
        "Connection: keep-alive",
        "Proxy-Connection: keep-alive",
        "Te: trailers",
    }
    status, body := rawProxyGET(t, pURL.Host, upstream.URL+"/hop", headers)
    if !strings.HasPrefix(status, "HTTP/1.1 200") {
        t.Fatalf("status: %s", status)
    }
    var got map[string]string
    _ = json.NewDecoder(strings.NewReader(body)).Decode(&got)
    if got["conn"] != "" || got["pc"] != "" {
        t.Fatalf("hop-by-hop must be stripped: %+v", got)
    }
}

func TestForwardProxy_DNSError_Returns502(t *testing.T) {
    t.Parallel()
    app, _ := startHTTPApp(t)
    defer app.Close()
    pURL, _ := url.Parse(app.URL)
    status, _ := rawProxyGET(t, pURL.Host, "http://nonexistent.invalid/", nil)
    if !strings.HasPrefix(status, "HTTP/1.1 502") {
        t.Fatalf("expected 502 on DNS error, got %s", status)
    }
}

func TestForwardProxy_ConnectionRefused_Returns502(t *testing.T) {
    t.Parallel()
    app, _ := startHTTPApp(t)
    defer app.Close()
    pURL, _ := url.Parse(app.URL)
    status, _ := rawProxyGET(t, pURL.Host, "http://127.0.0.1:1/", nil)
    if !strings.HasPrefix(status, "HTTP/1.1 502") {
        t.Fatalf("expected 502 on refused, got %s", status)
    }
}


