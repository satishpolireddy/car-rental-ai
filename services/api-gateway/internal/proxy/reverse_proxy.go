package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// NewReverseProxy creates a gin handler that reverse-proxies to the given target URL.
// It strips the provided prefixToStrip from the request path before forwarding.
func NewReverseProxy(targetURL string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("invalid proxy target URL %q: %v", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom error handler for when the upstream is unreachable
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.WithError(err).WithField("target", targetURL).Error("upstream proxy error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream service unavailable"}`))
	}

	// Modify request before forwarding: rewrite host to target
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
