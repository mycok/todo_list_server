package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func newMux(todoFile string) http.Handler {
	mu := &sync.Mutex{}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	r := todoRouter(todoFile, mu)

	mux.Handle("/todo", http.StripPrefix("/todo", r))
	mux.Handle("/todo/", http.StripPrefix("/todo/", r))

	return mux
}

func replyWithTextContent(w http.ResponseWriter, r *http.Request, statusCode int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(content))
}

func replyWithJSONContent(w http.ResponseWriter, r *http.Request, statusCode int, resp *todoResponse) {
	body, err := json.Marshal(resp)
	if err != nil {
		replyWithErr(w, r, http.StatusInternalServerError, err.Error())

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func replyWithErr(w http.ResponseWriter, r *http.Request, statusCode int, errMsg string) {
	log.Printf("%s: %s: %d: %s", r.URL, r.Method, statusCode, errMsg)

	http.Error(w, http.StatusText(statusCode), statusCode)
}
