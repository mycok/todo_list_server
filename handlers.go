package main

import (
	"net/http"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	content := "Our API is live"

	replyTextContent(w, r, http.StatusOK, content)
}
