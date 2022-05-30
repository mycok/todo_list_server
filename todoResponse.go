package main

import (
	"encoding/json"
	"time"

	"github.com/mycok/todo_list_cli/todo"
)

type todoResponse struct {
	Results todo.List `json:"results"`
}

func (td *todoResponse) MarshalJSON() ([]byte, error) {
	resp := struct {
		Results      todo.List `json:"results"`
		Date         int64     `json:"date"`
		TotalResults int       `json:"total_results"`
	}{
		Results:      td.Results,
		Date:         time.Now().Unix(),
		TotalResults: len(td.Results),
	}

	return json.Marshal(resp)
}
