package handlers

import (
	"net/http"
	"strings"
)

func requestEmbed(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.FormValue("embed") == "1" || r.URL.Query().Get("embed") == "1" {
		return true
	}
	referer := strings.TrimSpace(r.Header.Get("Referer"))
	return strings.Contains(referer, "embed=1")
}
