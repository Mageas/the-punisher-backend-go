package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api/handler"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/jwt"
	"github.com/mageas/the-punisher-backend/internal/service"
)

const TestAccessSecret = "test-access-secret-for-e2e-tests"

type TestEnv struct {
	Server *httptest.Server
	Mock   *MockQuerier
}

// SetupTestEnv creates a fully wired test server with in-memory mock.
func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	mock := NewMockQuerier()

	studentService := service.NewStudentService(mock)
	studentHandler := handler.NewStudentHandler(studentService)

	r := chi.NewRouter()
	r.Route("/v1/students", func(r chi.Router) {
		r.Use(auth.AuthMiddleware(TestAccessSecret))
		r.Post("/", studentHandler.CreateStudent)
		r.Get("/", studentHandler.ListStudents)
		r.Get("/{id}", studentHandler.GetStudent)
		r.Put("/{id}", studentHandler.UpdateStudent)
		r.Delete("/{id}", studentHandler.DeleteStudent)
	})

	server := httptest.NewServer(r)
	t.Cleanup(server.Close)

	return &TestEnv{
		Server: server,
		Mock:   mock,
	}
}

// GenerateTestJWT creates a valid JWT for the given userID.
func GenerateTestJWT(t *testing.T, userID uuid.UUID) string {
	t.Helper()

	token, err := jwt.Generate(jwt.Config{
		Secret:     TestAccessSecret,
		Expiration: 15 * time.Minute,
		Issuer:     "test",
		Audience:   "test",
	}, userID.String())

	if err != nil {
		t.Fatalf("failed to generate test JWT: %v", err)
	}
	return token
}

// DoRequest performs an HTTP request against the test server.
func DoRequest(t *testing.T, method, url string, body any, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}

	return resp
}

// DoRequestRaw performs an HTTP request with a raw string body (for malformed JSON testing).
func DoRequestRaw(t *testing.T, method, url string, rawBody string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(rawBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to do request: %v", err)
	}

	return resp
}

// ParseResponseBody decodes a JSON response body into the given target.
func ParseResponseBody(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// AssertStatus checks the HTTP status code and fails the test if it doesn't match.
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", expected, resp.StatusCode, string(body))
	}
}
