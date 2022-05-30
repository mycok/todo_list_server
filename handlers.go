package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/mycok/todo_list_cli/todo"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidData = errors.New("invalid data")
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		replyWithErr(w, r, http.StatusNotFound, "")

		return
	}

	content := "Our API is live"

	replyWithTextContent(w, r, http.StatusOK, content)
}

func todoRouter(todoFile string, l sync.Locker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := &todo.List{}

		// Lock to prevent concurrent access to the todoFile during the
		// file reading process.
		l.Lock()
		defer l.Unlock()
		// Load todo's from the todoFile and append them to the list.
		if err := list.Load(todoFile); err != nil {
			replyWithErr(w, r, http.StatusInternalServerError, err.Error())

			return
		}

		if r.URL.Path == "" {
			switch r.Method {
			case http.MethodGet:
				getAllHandler(w, r, list)
			case http.MethodPost:
				addHandler(w, r, list, todoFile)
			default:
				msg := "Method not supported"
				replyWithErr(w, r, http.StatusMethodNotAllowed, msg)
			}

			return
		}

		// Validate the id path parameter.
		id, err := validateID(r.URL.Path, list)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				replyWithErr(w, r, http.StatusNotFound, err.Error())

				return
			}

			replyWithErr(w, r, http.StatusBadRequest, err.Error())

			return
		}

		switch r.Method {
		case http.MethodGet:
			getOneHandler(w, r, list, id)
		case http.MethodPatch:
			patchHandler(w, r, list, id, todoFile)
		case http.MethodDelete:
			deleteHandler(w, r, list, id, todoFile)
		default:
			msg := "Method not supported"
			replyWithErr(w, r, http.StatusMethodNotAllowed, msg)
		}
	}
}

func validateID(path string, list *todo.List) (int, error) {
	id, err := strconv.Atoi(path)
	if err != nil {
		return 0, fmt.Errorf("%w, Invalid ID: %q", ErrInvalidData, err)
	}

	if id < 1 {
		return 0, fmt.Errorf("%w, Invalid ID: less than one", ErrInvalidData)
	}

	if id > len(*list) {
		return 0, fmt.Errorf("%w, ID %d not found", ErrNotFound, id)
	}

	return id, nil
}

func getAllHandler(w http.ResponseWriter, r *http.Request, l *todo.List) {
	resp := &todoResponse{
		Results: *l,
	}

	replyWithJSONContent(w, r, http.StatusOK, resp)
}

func getOneHandler(w http.ResponseWriter, r *http.Request, l *todo.List, id int) {
	resp := &todoResponse{
		Results: (*l)[id-1 : id],
	}

	replyWithJSONContent(w, r, http.StatusOK, resp)
}

func patchHandler(w http.ResponseWriter, r *http.Request, l *todo.List, id int, todoFile string) {
	q := r.URL.Query()

	if _, ok := q["complete"]; !ok {
		msg := "Missing query param 'complete'"
		replyWithErr(w, r, http.StatusBadRequest, msg)

		return
	}

	l.Complete(id)

	if err := l.Save(todoFile); err != nil {
		replyWithErr(w, r, http.StatusInternalServerError, err.Error())

		return
	}

	replyWithTextContent(w, r, http.StatusNoContent, "Todo completed successfully")
}

func deleteHandler(w http.ResponseWriter, r *http.Request, l *todo.List, id int, todoFile string) {

	l.Delete(id)

	if err := l.Save(todoFile); err != nil {
		replyWithErr(w, r, http.StatusInternalServerError, err.Error())

		return
	}

	replyWithTextContent(w, r, http.StatusNoContent, "Todo deleted successfully")
}

func addHandler(w http.ResponseWriter, r *http.Request, l *todo.List, todoFile string) {
	todo := struct {
		Task string `json:"task"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		msg := fmt.Sprintf("Invalid JSON: %s", err)
		replyWithErr(w, r, http.StatusBadRequest, msg)

		return
	}

	l.Add(todo.Task)

	if err := l.Save(todoFile); err != nil {
		replyWithErr(w, r, http.StatusInternalServerError, err.Error())

		return
	}

	replyWithTextContent(w, r, http.StatusCreated, "Todo added successfully")
}
