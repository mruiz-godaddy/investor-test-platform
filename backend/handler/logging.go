package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ANSI colors
const (
	cReset = "\033[0m"
	cBold  = "\033[1m"
	cDim   = "\033[2m"
	cRed   = "\033[31m"
	cGreen = "\033[32m"
	cYell  = "\033[33m"
	cCyan  = "\033[36m"
)

// SetupSessionLog creates a timestamped log file in logs/ and configures the
// global logger to write to both stdout and the file.  Returns a cleanup func.
func SetupSessionLog(logsDir string) func() {
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		log.Printf("WARN: could not create logs dir: %v", err)
		return func() {}
	}

	name := time.Now().Format("2006-01-02_15-04-05") + ".log"
	path := filepath.Join(logsDir, name)

	f, err := os.Create(path)
	if err != nil {
		log.Printf("WARN: could not create log file: %v", err)
		return func() {}
	}

	multi := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multi)
	log.SetFlags(0)

	log.Printf("Session log → %s", path)
	return func() { f.Close() }
}

// responseCapture wraps http.ResponseWriter to record status and full body.
type responseCapture struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.status = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

func statusColor(code int) string {
	switch {
	case code < 300:
		return cGreen
	case code < 400:
		return cYell
	default:
		return cRed
	}
}

func prettyJSON(raw string, indent string) string {
	if raw == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(raw), indent, "  "); err == nil {
		return buf.String()
	}
	return raw
}

// dedup tracks the last response hash per route to suppress identical repeated responses.
type dedup struct {
	mu    sync.Mutex
	cache map[string][32]byte // key = "METHOD path" → sha256 of response body
}

func (d *dedup) isDuplicate(key string, body []byte) bool {
	h := sha256.Sum256(body)
	d.mu.Lock()
	defer d.mu.Unlock()
	if prev, ok := d.cache[key]; ok && prev == h {
		return true
	}
	d.cache[key] = h
	return false
}

// RequestResponseLogger is a mux middleware that logs the incoming request
// immediately, then logs the outgoing response when the handler completes.
func RequestResponseLogger(next http.Handler) http.Handler {
	dd := &dedup{cache: make(map[string][32]byte)}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RequestURI()

		// Skip admin endpoints — only log app traffic
		if strings.HasPrefix(path, "/admin") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		ts := start.Format("15:04:05.000")
		ind := "      "

		// ── Log incoming request ──
		var reqSB strings.Builder
		fmt.Fprintf(&reqSB, "%s%s%s %s▶ %s %s%s",
			cDim, ts, cReset,
			cCyan, r.Method, path, cReset)

		if auth := r.Header.Get("Authorization"); auth != "" {
			if len(auth) > 80 {
				auth = auth[:80] + "…"
			}
			fmt.Fprintf(&reqSB, "\n%sAuth:%s %s", cDim, cReset, auth)
		}

		// Buffer request body for POST/PUT/PATCH
		var reqBody string
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				reqBody = string(bodyBytes)
			}
			if reqBody != "" {
				fmt.Fprintf(&reqSB, "\n%s", prettyJSON(reqBody, ind))
			}
		}

		log.Print(reqSB.String())

		// ── Execute handler + capture response ──
		rc := &responseCapture{ResponseWriter: w, status: 200}
		next.ServeHTTP(rc, r)

		dur := time.Since(start)
		respTs := time.Now().Format("15:04:05.000")
		sc := statusColor(rc.status)
		respBody := rc.body.String()
		routeKey := r.Method + " " + path

		// For GET 2xx: suppress body on repeated identical responses
		if r.Method == "GET" && rc.status < 300 && dd.isDuplicate(routeKey, rc.body.Bytes()) {
			log.Printf("%s%s%s %s◀ %s%d%s %s %s%dms%s %s(unchanged)%s",
				cDim, respTs, cReset,
				sc, cBold, rc.status, cReset,
				path,
				cDim, dur.Milliseconds(), cReset,
				cDim, cReset)
			return
		}

		// ── Log outgoing response ──
		var resSB strings.Builder
		fmt.Fprintf(&resSB, "%s%s%s %s%s◀ %d%s %s %s%dms%s",
			cDim, respTs, cReset,
			sc, cBold, rc.status, cReset,
			path,
			cDim, dur.Milliseconds(), cReset)

		if respBody != "" {
			fmt.Fprintf(&resSB, "\n%s", prettyJSON(respBody, ind))
		}

		log.Print(resSB.String())
	})
}
