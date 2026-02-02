package services

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

const parcaSidecarURL = "http://localhost:7070"

// ProfilingProxyHandler proxies requests to the Parca sidecar for profile visualization.
func ProfilingProxyHandler() gin.HandlerFunc {
	target, _ := url.Parse(parcaSidecarURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Profiling service unavailable: " + err.Error()))
	}

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
