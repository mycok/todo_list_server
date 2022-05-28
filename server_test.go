package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
			name:               "NotFoundUrlPath",
			urlPath:            "/todo",
			expectedStatusCode: http.StatusNotFound,
			expectedItems:      0,
			expectedContent:    "404 page not found",
		},
	}

	tServer := setupAPI(t)
	defer t.Cleanup(func() {
		tServer.Close()
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				body []byte
				err  error
			)

			res, err := http.Get(tServer.URL + tc.urlPath)
			if err != nil {
				t.Fatal(err)
			}

			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatusCode {
				t.Errorf("Expected status code: %d, but got: %d instead", tc.expectedStatusCode, res.StatusCode)
			}

			switch {
			case strings.Contains(res.Header.Get("Content-Type"), "text/plain"):
				if body, err = io.ReadAll(res.Body); err != nil {
					t.Errorf("Error reading response body: %q", err)
				}

				if !strings.Contains(string(body), tc.expectedContent) {
					t.Errorf("Expected %q, but got: %q instead", tc.expectedContent, string(body))
				}
			default:
				t.Errorf("Unsupported content-type: %q", res.Header.Get("Content-Type"))
			}
		})
	}

}

func setupAPI(t *testing.T) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(newMux(""))

	return ts
}
