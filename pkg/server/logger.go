package server

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type HTTPReqInof struct {
	method    string
	uri       string
	ipaddr    string
	code      int
	duration  time.Duration
	userAgent string
}

type wrRecorder struct {
	http.ResponseWriter
	status int
}

func (r *wrRecorder) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ri := &HTTPReqInof{
			method:    r.Method,
			uri:       r.URL.String(),
			userAgent: r.Header.Get("User-Agent"),
			ipaddr:    getRemoteAddress(r),
		}

		// Initialize the status to 200 in case WriteHeader is not called
		rec := &wrRecorder{w, http.StatusOK}

		start := time.Now()
		h.ServeHTTP(rec, r)
		ri.duration = time.Since(start)
		ri.code = rec.status

		logHTTPReq(ri)
	}

	return http.HandlerFunc(fn)
}

func logHTTPReq(ri *HTTPReqInof) {
	log.Printf("- %s: %s %d %s took %s\n", ri.ipaddr, ri.method, ri.code, ri.uri, ri.duration)
}

func getRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}

	return hdrRealIP
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
