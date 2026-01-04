package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePOST(t *testing.T) {
	type want struct {
		status      int
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "post create",
			body: "http://yandex.ru",
			want: want{
				status:      http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "post empty",
			body: "",
			want: want{
				status:      http.StatusBadRequest,
				contentType: "",
			},
		},
	}
	s := NewURLShorter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			s.HandlePOST(rr, req)

			if rr.Code != tt.want.status {
				t.Errorf("status %d, got %d", tt.want.status, rr.Code)
			}

			if tt.want.contentType != "" {
				ct := rr.Header().Get("Content-Type")
				if ct != tt.want.contentType {
					t.Errorf("content type %s, got %s", tt.want.contentType, ct)
				}
			}
			if tt.body != "" && rr.Body.Len() == 0 {
				t.Errorf("non-empty body, length %d", rr.Body.Len())
			}
		})
	}
}

func TestHandleGET(t *testing.T) {
	s := NewURLShorter()
	longURL := "http://yandex.ru"
	postReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
	postRR := httptest.NewRecorder()
	s.HandlePOST(postRR, postReq)

	shortURL := postRR.Body.String()
	id := strings.TrimPrefix(shortURL, "http://localhost:8082/")

	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRR := httptest.NewRecorder()
	s.HandleGET(getRR, getReq)

	if getRR.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d",
			http.StatusTemporaryRedirect, getRR.Code)
	}

	location := getRR.Header().Get("Location")
	if location != longURL {
		t.Errorf("expected Location %s, got %s", longURL, location)
	}
}
