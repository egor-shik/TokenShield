package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/egor-shik/tokenshield/internal/sanitizer" 
)

func NewProxy(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err == nil {
				sanitizedBytes := sanitizer.Clean(bodyBytes)
				req.Body = io.NopCloser(bytes.NewReader(sanitizedBytes))
				req.ContentLength = int64(len(sanitizedBytes))
			}
		}
	}

	return proxy
}