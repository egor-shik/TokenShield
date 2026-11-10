package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/egor-shik/tokenshield/internal/cache"
	"github.com/egor-shik/tokenshield/internal/sanitizer"
)

type ProxyServer struct {
	handler http.Handler
}

func NewProxy(targetURL string, c *cache.Cache) (*ProxyServer, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ModifyResponse = func(resp *http.Response) error {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		if resp.StatusCode == http.StatusOK && resp.Request != nil && resp.Request.Body != nil {
			reqBody, err := io.ReadAll(resp.Request.Body)
			if err == nil {
				_ = c.Set(resp.Request.Context(), reqBody, string(bodyBytes))
			}
		}

		return nil
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				sanitizedBytes := sanitizer.Clean(bodyBytes)
				cachedResp, found, err := c.Get(r.Context(), sanitizedBytes)
				if err == nil && found {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(cachedResp))
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(sanitizedBytes))
				r.ContentLength = int64(len(sanitizedBytes))
			}
		}

		proxy.ServeHTTP(w, r)
	})

	return &ProxyServer{handler: handler}, nil
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.handler.ServeHTTP(w, r)
}
