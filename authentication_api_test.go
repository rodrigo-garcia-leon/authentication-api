package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TestUserDao struct{}

func (u TestUserDao) Get(username string) (string, error) {
	return "test", nil
}

func (u TestUserDao) Set(username string, password string) error {
	return nil
}

func TestAuthenticationApi(t *testing.T) {
	serveMux := http.NewServeMux()
	u := TestUserDao{}
	now := func() time.Time { return time.Unix(int64(1645393928), 0) }
	setupMux(serveMux, u, now)

	t.Run("healthCheckHandler", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		serveMux.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
		}

		body := res.Body.String()
		if body != "" {
			t.Errorf("Expected body to be %q, got %q", "", body)
		}
	})

	t.Run("registerHandler", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/auth/register", strings.NewReader(`{"username":"test","password":"test"}`))

		serveMux.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
		}

		body := res.Body.String()
		if body != "" {
			t.Errorf("Expected body to be %q, got %q", "", body)
		}
	})

	t.Run("loginHandler", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/auth/login", strings.NewReader(`{"username":"test","password":"test"}`))

		serveMux.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.Code)
		}

		body := res.Body.String()
		if body != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDU0ODAzMjgsImlzcyI6ImF1dGhlbnRpY2F0aW9uLWFwaSIsInN1YiI6InRlc3QifQ.YZAcVBYWtMXo-WeXySKCFecHRMgdvQMaq8tuz5hcowY" {
			t.Errorf("Expected body to be %q, got %q", "", body)
		}
	})
}
