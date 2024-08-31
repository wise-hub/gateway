package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"fibank.bg/fis-gateway-ws/internal/configuration"
	"github.com/go-chi/chi/v5"
)

func proxyHandler(proxy *httputil.ReverseProxy, d *configuration.Dependencies, method, context string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestPath := chi.URLParam(r, "proxyPath")

		if r.URL.RawQuery != "" {
			requestPath += "?" + r.URL.RawQuery
		}

		key := method + "," + context + "," + requestPath

		mu.RLock()
		_, allowed := allowedEndpoints[key]
		mu.RUnlock()

		if !allowed {
			http.Error(w, "resource not found", http.StatusNotFound)
			return
		}

		r.Header.Add("fis-intws-auth", d.Cfg.InternalWsPwd)
		r.URL.Path = requestPath

		proxy.ServeHTTP(w, r)
	}
}

func setupProxyRoutes(r chi.Router, d *configuration.Dependencies, proxyType string) {
	var backendServiceURL string

	if proxyType == "public" {
		backendServiceURL = d.Cfg.PublicFQDN
	} else if proxyType == "protected" {
		backendServiceURL = d.Cfg.ProtectedFQDN
	} else {
		return
	}

	targetURL, _ := url.Parse(backendServiceURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	r.Get("/*", proxyHandler(proxy, d, "GET", proxyType))
	r.Post("/*", proxyHandler(proxy, d, "POST", proxyType))
}
