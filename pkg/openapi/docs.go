// Package openapi serves interactive OpenAPI documentation for API servers.
//
// The docs UI is Scalar API Reference loaded from a pinned jsDelivr CDN URL.
// That keeps service binaries small and is acceptable for local/dev and typical
// cluster environments with egress. Air-gapped deployments that need a fully
// self-contained UI should vendor/embed the Scalar assets instead.
package openapi

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// Pin Scalar for reproducible docs UI; bump intentionally when upgrading.
const scalarCDN = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.43.5"

const scalarHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    body { margin: 0; }
  </style>
</head>
<body>
  <div id="app"></div>
  <script src="{{.CDN}}" crossorigin="anonymous"></script>
  <script>
    Scalar.createApiReference("#app", {
      url: {{.SpecURL}},
      hideClientButton: true,
    });
  </script>
</body>
</html>`

var docsTmpl = template.Must(template.New("scalar").Parse(scalarHTML))

// DocsHandler returns an http.Handler that serves Scalar API Reference
// configured to load the OpenAPI/Swagger spec from specURL (e.g. "/swagger.json").
func DocsHandler(specURL, title string) http.Handler {
	if title == "" {
		title = "API documentation"
	}
	if specURL == "" {
		specURL = "/swagger.json"
	}
	specJSON, err := json.Marshal(specURL)
	if err != nil {
		// Impossible for a string, but keep the handler usable.
		specJSON = []byte(`"/swagger.json"`)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := docsTmpl.Execute(w, map[string]any{
			"Title":   title,
			"CDN":     scalarCDN,
			"SpecURL": template.JS(specJSON),
		}); err != nil {
			http.Error(w, fmt.Sprintf("render docs: %v", err), http.StatusInternalServerError)
		}
	})
}

// SpecHandler returns an http.Handler that serves the OpenAPI/Swagger JSON
// document with the correct Content-Type.
func SpecHandler(spec []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(spec)
	})
}

// Mount registers the OpenAPI spec at /swagger.json and Scalar docs at /docs/.
// Requests to /docs (no trailing slash) redirect to /docs/.
func Mount(mux *http.ServeMux, spec []byte, title string) {
	docs := DocsHandler("/swagger.json", title)
	mux.Handle("/swagger.json", SpecHandler(spec))
	mux.Handle("/docs", http.RedirectHandler("/docs/", http.StatusFound))
	mux.Handle("/docs/", docs)
}
