package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mycok/todo_list_cli/todo"
)

func TestMain(m *testing.M) {
	// Get rid of the server log output from the test output.
	log.SetOutput(io.Discard)

	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name               string
		urlPath            string
		expectedStatusCode int
		expectedItems      int
		expectedContent    string
	}{
		{
			name:               "RootUrlPath",
			urlPath:            "/",
			expectedStatusCode: http.StatusOK,
			expectedItems:      0,
			expectedContent:    "Our API is live",
		},
		{
			name:               "GetAll",
			urlPath:            "/todo",
			expectedStatusCode: http.StatusOK,
			expectedItems:      3,
			expectedContent:    "task 1",
		},
		{
			name:               "GetOne",
			urlPath:            "/todo/3",
			expectedStatusCode: http.StatusOK,
			expectedItems:      1,
			expectedContent:    "task 3",
		},
		{
			name:               "NotFoundUrlPath",
			urlPath:            "/task",
			expectedStatusCode: http.StatusNotFound,
			expectedItems:      0,
			expectedContent:    "Not Found\n",
		},
	}

	// Setup a test server.
	tServer := setupAPI(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				body []byte
				err  error
				resp struct {
					Results      todo.List `json:"results"`
					Date         int64     `json:"date"`
					TotalResults int       `json:"total_results"`
				}
			)

			// Perform a GET request using the default client and the test server url path.
			res, err := http.Get(tServer.URL + tc.urlPath)
			if err != nil {
				t.Fatalf("Error sending a GET request: %q", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatusCode {
				t.Errorf("Expected status code: %d, but got: %d instead", tc.expectedStatusCode, res.StatusCode)
			}

			switch {
			case strings.Contains(res.Header.Get("Content-Type"), "text/plain"):
				if body, err = io.ReadAll(res.Body); err != nil {
					t.Fatalf("Error reading response body: %q", err)
				}

				if !strings.Contains(string(body), tc.expectedContent) {
					t.Errorf("Expected %q, but got: %q instead", tc.expectedContent, string(body))
				}

			case res.Header.Get("Content-Type") == "application/json":
				if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
					t.Fatalf("Error decoding JSON response body: %q", err)
				}

				if tc.expectedItems != resp.TotalResults {
					t.Errorf("Expected: %d todo items, but got: %d instead", tc.expectedItems, resp.TotalResults)
				}

				if tc.expectedContent != resp.Results[0].Task {
					t.Errorf("Expected: %s todo item, but got: %s instead", tc.expectedContent, resp.Results[0].Task)
				}

			default:
				t.Errorf("Unsupported content-type: %q", res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestAdd(t *testing.T) {
	// Setup a test server.
	tServer := setupAPI(t)
	
	taskName := "test add todo task"

	t.Run("AddNew", func(t *testing.T) {
		var body bytes.Buffer

		task := struct {
			Task string `json:"task"`
		}{
			Task: taskName,
		}

		if err := json.NewEncoder(&body).Encode(task); err != nil {
			t.Fatalf("Error encoding JSON body: %q", err)
		}

		// Perform a POST request using the default client and the test server url path.
		resp, err := http.Post(tServer.URL+"/todo", "application/json", &body)
		if err != nil {
			t.Fatalf("Error sending a POST request: %q", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf(
				"Expected status: %q, but got: %q instead",
				http.StatusText(http.StatusCreated), http.StatusText(resp.StatusCode))
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Error reading response body: %q", err)
		}
		defer resp.Body.Close()

		msg := "Todo added successfully"

		if string(respBody) != msg {
			t.Errorf("Expected: %s, but got: %s instead", msg, string(respBody))
		}

		// Check for the newly added todo task.
		t.Run("CheckNew", func(t *testing.T) {
			// Perform a GET request using the default client and the test server url path.
			resp, err := http.Get(tServer.URL + "/todo/4")
			if err != nil {
				t.Fatalf("Error sending a POST request: %q", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf(
					"Expected status: %q, but got: %q instead",
					http.StatusText(http.StatusCreated), http.StatusText(resp.StatusCode))
			}

			var todoResp todoResponse

			if err := json.NewDecoder(resp.Body).Decode(&todoResp); err != nil {
				t.Fatalf("Error decoding JSON body: %q", err)
			}
			defer resp.Body.Close()

			if len(todoResp.Results) != 1 {
				t.Errorf("Expected only one todo item, but got: %d instead", len(todoResp.Results))
			}

			if todoResp.Results[0].Task != taskName {
				t.Errorf("Expected todo task: %s, but got: %s instead", taskName, todoResp.Results[0].Task)
			}
		})
	})
}

func TestDelete(t *testing.T) {
	// Setup a test server.
	tServer := setupAPI(t)

	t.Run("Delete", func(t *testing.T) {
		url := fmt.Sprintf("%s/todo/1", tServer.URL)

		// Create a new custom DELETE request.
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Fatalf("Error creating a DELETE request: %q", err)
		}

		// Send a DELETE request to the test server.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error sending a DELETE request: %q", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf(
				"Expected status: %q, but got: %q instead",
				http.StatusText(http.StatusNoContent), http.StatusText(resp.StatusCode))
		}

		// respBody, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	t.Fatalf("Error reading response body: %q", err)
		// }
		// defer resp.Body.Close()

		// fmt.Printf("*****response body*****: %+v", resp)

		// msg := "Todo deleted successfully"

		// if string(respBody) != msg {
		// 	t.Errorf("Expected: %s, but got: %s instead", msg, string(respBody))
		// }

		// Check for successful deletion.
		t.Run("CheckDelete", func(t *testing.T) {
			// Perform a GET request using the default client and the test server url path.
			resp, err := http.Get(tServer.URL + "/todo")
			if err != nil {
				t.Fatalf("Error sending a GET request: %q", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf(
					"Expected status: %q, but got: %q instead",
					http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
			}

			var todoResp todoResponse

			if err := json.NewDecoder(resp.Body).Decode(&todoResp); err != nil {
				t.Fatalf("Error decoding JSON body: %q", err)
			}
			defer resp.Body.Close()

			if len(todoResp.Results) != 2 {
				t.Errorf("Expected only two todo items, but got: %d instead", len(todoResp.Results))
			}

			expectedTodoTask := "task 2"

			if todoResp.Results[0].Task != expectedTodoTask {
				t.Errorf("Expected todo task: %s, but got: %s instead", expectedTodoTask, todoResp.Results[0].Task)
			}
		})
	})
}

func TestPatch(t *testing.T) {
	// Setup a test server.
	tServer := setupAPI(t)
	
	t.Run("Complete", func(t *testing.T) {
		url := fmt.Sprintf("%s/todo/1?complete", tServer.URL)

		// Create a new custom PATCH request.
		req, err := http.NewRequest(http.MethodPatch, url, nil)
		if err != nil {
			t.Fatalf("Error creating a PATCH request: %q", err)
		}

		// Send a PATCH request to the test server.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error sending a PATCH request: %q", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf(
				"Expected status: %q, but got: %q instead",
				http.StatusText(http.StatusNoContent), http.StatusText(resp.StatusCode))
		}

		// Check for successful [mark as complete] operation.
		t.Run("CheckComplete", func(t *testing.T) {
			resp, err := http.Get(tServer.URL + "/todo/1")
			if err != nil {
				t.Fatalf("Error sending a GET request: %q", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf(
					"Expected status: %q, but got: %q instead",
					http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
			}

			var todoResp todoResponse

			if err := json.NewDecoder(resp.Body).Decode(&todoResp); err != nil {
				t.Fatalf("Error decoding JSON body: %q", err)
			}
			defer resp.Body.Close()

			if len(todoResp.Results) != 1 {
				t.Errorf("Expected only one todo item, but got: %d instead", len(todoResp.Results))
			}

			if !todoResp.Results[0].Done {
				t.Errorf("Expected todo item %s to be marked complete, but got an incomplete todo item instead", todoResp.Results[0].Task)
			}
		})
	})

}

func setupAPI(t *testing.T) *httptest.Server {
	t.Helper()

	tempTodoFile, err := os.CreateTemp("", "todo")
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(newMux(tempTodoFile.Name()))

	// Add some todo items for initial testing.
	for i := 1; i < 4; i++ {
		var body bytes.Buffer

		taskName := fmt.Sprintf("task %d", i)

		item := struct {
			Task string `json:"task"`
		}{Task: taskName}

		if err := json.NewEncoder(&body).Encode(item); err != nil {
			t.Fatal(err)
		}

		res, err := http.Post(ts.URL+"/todo", "application/json", &body)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Failed to add initial todo items: Status: %d", res.StatusCode)
		}
	}

	t.Cleanup(func() {
		os.Remove(tempTodoFile.Name())
		ts.Close()
	})

	return ts
}
