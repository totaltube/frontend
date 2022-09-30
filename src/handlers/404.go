package handlers

import (
	"net/http"
)

var Handle404 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	Output404(w, r, "page not found")
})