package main

import (
	"net/http"
)

func newMux(todoFile string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	return mux
}

// TODO: remove the request argument if not used.
func replyTextContent(w http.ResponseWriter, r *http.Request, statusCode int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(content))
}
