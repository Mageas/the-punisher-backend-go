package handler_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/testutil"
)

// ============================================================
// CreateStudent
// ============================================================

var endpoint = "/v1/students/"

func TestCreateStudent_Success(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "Alice",
		"last_name":  "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusCreated)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	if result["first_name"] != "Alice" {
		t.Errorf("expected first_name 'Alice', got %v", result["first_name"])
	}
	if result["last_name"] != "Martin" {
		t.Errorf("expected last_name 'Martin', got %v", result["last_name"])
	}
	if result["id"] == nil || result["id"] == "" {
		t.Error("expected a non-empty id")
	}
}

func TestCreateStudent_NoAuth(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	body := map[string]string{
		"first_name": "Alice",
		"last_name":  "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, "")
	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

func TestCreateStudent_EmptyBody(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequestRaw(t, http.MethodPost, env.Server.URL+endpoint, "", token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

func TestCreateStudent_MissingFirstName(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"last_name": "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "validation_failed" {
		t.Errorf("expected error 'validation_failed', got '%s'", result.Error)
	}
	assertErrorDetailContainsField(t, result.ErrorDetails, "first_name")
}

func TestCreateStudent_MissingLastName(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "Alice",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "validation_failed" {
		t.Errorf("expected error 'validation_failed', got '%s'", result.Error)
	}
	assertErrorDetailContainsField(t, result.ErrorDetails, "last_name")
}

func TestCreateStudent_FirstNameTooShort(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "A",
		"last_name":  "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "validation_failed" {
		t.Errorf("expected error 'validation_failed', got '%s'", result.Error)
	}
	assertErrorDetailContainsField(t, result.ErrorDetails, "first_name")
}

func TestCreateStudent_FirstNameTooLong(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": strings.Repeat("A", 71),
		"last_name":  "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "validation_failed" {
		t.Errorf("expected error 'validation_failed', got '%s'", result.Error)
	}
	assertErrorDetailContainsField(t, result.ErrorDetails, "first_name")
}

func TestCreateStudent_LastNameTooShort(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "Alice",
		"last_name":  "M",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "last_name")
}

func TestCreateStudent_LastNameTooLong(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "Alice",
		"last_name":  strings.Repeat("M", 71),
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "last_name")
}

func TestCreateStudent_UnknownField(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{
		"first_name": "Alice",
		"last_name":  "Martin",
		"unknown":    "field",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "invalid_request_body" {
		t.Errorf("expected error 'invalid_request_body', got '%s'", result.Error)
	}
}

func TestCreateStudent_InvalidJSONType(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	// first_name should be a string, not a number
	body := map[string]any{
		"first_name": 42,
		"last_name":  "Martin",
	}

	resp := testutil.DoRequest(t, http.MethodPost, env.Server.URL+endpoint, body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "malformed_parameter" {
		t.Errorf("expected error 'malformed_parameter', got '%s'", result.Error)
	}
}

func TestCreateStudent_MalformedJSON(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequestRaw(t, http.MethodPost, env.Server.URL+endpoint, `{invalid json`, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

// ============================================================
// GetStudent
// ============================================================

func TestGetStudent_Success(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Bob", "Dupont")

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+student.ID.String(), nil, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	if result["first_name"] != "Bob" {
		t.Errorf("expected first_name 'Bob', got %v", result["first_name"])
	}
	if result["last_name"] != "Dupont" {
		t.Errorf("expected last_name 'Dupont', got %v", result["last_name"])
	}
}

func TestGetStudent_NoAuth(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+uuid.New().String(), nil, "")
	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

func TestGetStudent_MalformedID(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+"not-a-uuid", nil, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "malformed_parameter" {
		t.Errorf("expected error 'malformed_parameter', got '%s'", result.Error)
	}
}

func TestGetStudent_NotFound(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+uuid.New().String(), nil, token)
	testutil.AssertStatus(t, resp, http.StatusNotFound)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "student_not_found" {
		t.Errorf("expected error 'student_not_found', got '%s'", result.Error)
	}
}

func TestGetStudent_BelongsToAnotherUser(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	ownerID := uuid.New()
	otherUserID := uuid.New()

	student := env.Mock.SeedStudent(ownerID, "Bob", "Dupont")

	token := testutil.GenerateTestJWT(t, otherUserID)
	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+student.ID.String(), nil, token)
	testutil.AssertStatus(t, resp, http.StatusNotFound)
}

// ============================================================
// ListStudents
// ============================================================

func TestListStudents_Empty(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint, nil, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	data, ok := result["data"].([]any)
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) != 0 {
		t.Errorf("expected empty data array, got %d items", len(data))
	}

	totalCount := result["total_count"].(float64)
	if totalCount != 0 {
		t.Errorf("expected total_count 0, got %v", totalCount)
	}
}

func TestListStudents_WithData(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	env.Mock.SeedStudent(userID, "Alice", "Martin")
	env.Mock.SeedStudent(userID, "Bob", "Dupont")

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint, nil, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	data := result["data"].([]any)
	if len(data) != 2 {
		t.Errorf("expected 2 items, got %d", len(data))
	}

	totalCount := result["total_count"].(float64)
	if totalCount != 2 {
		t.Errorf("expected total_count 2, got %v", totalCount)
	}
}

func TestListStudents_DoesNotShowOtherUsersStudents(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	env.Mock.SeedStudent(userID, "Alice", "Martin")
	env.Mock.SeedStudent(otherUserID, "Bob", "Dupont")

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint, nil, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 item (only own students), got %d", len(data))
	}
}

func TestListStudents_Pagination(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	env.Mock.SeedStudent(userID, "Alice", "Martin")
	env.Mock.SeedStudent(userID, "Bob", "Dupont")

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+"?page=1", nil, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	page := result["page"].(float64)
	if page != 1 {
		t.Errorf("expected page 1, got %v", page)
	}

	itemPerPage := result["item_per_page"].(float64)
	if itemPerPage != 20 {
		t.Errorf("expected item_per_page 20, got %v", itemPerPage)
	}
}

func TestListStudents_NoAuth(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	resp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint, nil, "")
	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

// ============================================================
// UpdateStudent
// ============================================================

func TestUpdateStudent_FullUpdate(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{
		"first_name": "Alicia",
		"last_name":  "Martinez",
	}

	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	if result["first_name"] != "Alicia" {
		t.Errorf("expected first_name 'Alicia', got %v", result["first_name"])
	}
	if result["last_name"] != "Martinez" {
		t.Errorf("expected last_name 'Martinez', got %v", result["last_name"])
	}
}

func TestUpdateStudent_PartialFirstName(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{
		"first_name": "Alicia",
	}

	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	if result["first_name"] != "Alicia" {
		t.Errorf("expected first_name 'Alicia', got %v", result["first_name"])
	}
	if result["last_name"] != "Martin" {
		t.Errorf("expected last_name to remain 'Martin', got %v", result["last_name"])
	}
}

func TestUpdateStudent_PartialLastName(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{
		"last_name": "Martinez",
	}

	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusOK)

	var result map[string]any
	testutil.ParseResponseBody(t, resp, &result)

	if result["first_name"] != "Alice" {
		t.Errorf("expected first_name to remain 'Alice', got %v", result["first_name"])
	}
	if result["last_name"] != "Martinez" {
		t.Errorf("expected last_name 'Martinez', got %v", result["last_name"])
	}
}

func TestUpdateStudent_EmptyBody(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	// Empty JSON body — all fields optional, so no validation error
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), map[string]string{}, token)
	testutil.AssertStatus(t, resp, http.StatusOK)
}

func TestUpdateStudent_NoAuth(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+uuid.New().String(), map[string]string{"first_name": "X"}, "")
	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

func TestUpdateStudent_MalformedID(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{"first_name": "Alicia"}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+"not-a-uuid", body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "malformed_parameter" {
		t.Errorf("expected error 'malformed_parameter', got '%s'", result.Error)
	}
}

func TestUpdateStudent_NotFound(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	body := map[string]string{"first_name": "Alicia"}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+uuid.New().String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusNotFound)
}

func TestUpdateStudent_BelongsToAnotherUser(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	ownerID := uuid.New()
	otherUserID := uuid.New()

	student := env.Mock.SeedStudent(ownerID, "Alice", "Martin")

	token := testutil.GenerateTestJWT(t, otherUserID)
	body := map[string]string{"first_name": "Hacked"}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusNotFound)
}

func TestUpdateStudent_FirstNameTooShort(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{"first_name": "A"}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "first_name")
}

func TestUpdateStudent_FirstNameTooLong(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{"first_name": strings.Repeat("A", 71)}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "first_name")
}

func TestUpdateStudent_LastNameTooShort(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{"last_name": "M"}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "last_name")
}

func TestUpdateStudent_LastNameTooLong(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{"last_name": strings.Repeat("M", 71)}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)
	assertErrorDetailContainsField(t, result.ErrorDetails, "last_name")
}

func TestUpdateStudent_InvalidJSONType(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]any{"first_name": 42}
	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

func TestUpdateStudent_UnknownField(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	body := map[string]string{
		"first_name": "Alicia",
		"unknown":    "field",
	}

	resp := testutil.DoRequest(t, http.MethodPut, env.Server.URL+endpoint+student.ID.String(), body, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "invalid_request_body" {
		t.Errorf("expected error 'invalid_request_body', got '%s'", result.Error)
	}
}

// ============================================================
// DeleteStudent
// ============================================================

func TestDeleteStudent_Success(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	student := env.Mock.SeedStudent(userID, "Alice", "Martin")

	resp := testutil.DoRequest(t, http.MethodDelete, env.Server.URL+endpoint+student.ID.String(), nil, token)
	testutil.AssertStatus(t, resp, http.StatusNoContent)

	// Verify the student is actually deleted
	getResp := testutil.DoRequest(t, http.MethodGet, env.Server.URL+endpoint+student.ID.String(), nil, token)
	testutil.AssertStatus(t, getResp, http.StatusNotFound)
}

func TestDeleteStudent_NoAuth(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	resp := testutil.DoRequest(t, http.MethodDelete, env.Server.URL+endpoint+uuid.New().String(), nil, "")
	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

func TestDeleteStudent_MalformedID(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	resp := testutil.DoRequest(t, http.MethodDelete, env.Server.URL+endpoint+"not-a-uuid", nil, token)
	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var result api.Error
	testutil.ParseResponseBody(t, resp, &result)

	if result.Error != "malformed_parameter" {
		t.Errorf("expected error 'malformed_parameter', got '%s'", result.Error)
	}
}

func TestDeleteStudent_NonExistentDoesNotError(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	userID := uuid.New()
	token := testutil.GenerateTestJWT(t, userID)

	// DELETE on non-existent student — SQL DELETE is :exec, no error
	resp := testutil.DoRequest(t, http.MethodDelete, env.Server.URL+endpoint+uuid.New().String(), nil, token)
	testutil.AssertStatus(t, resp, http.StatusNoContent)
}

// ============================================================
// Invalid JWT Tests
// ============================================================

func TestAllEndpoints_InvalidJWT(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.token"

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, endpoint},
		{http.MethodGet, endpoint},
		{http.MethodGet, fmt.Sprintf(endpoint+"%s", uuid.New().String())},
		{http.MethodPut, fmt.Sprintf(endpoint+"%s", uuid.New().String())},
		{http.MethodDelete, fmt.Sprintf(endpoint+"%s", uuid.New().String())},
	}

	for _, e := range endpoints {
		t.Run(fmt.Sprintf("%s %s", e.method, e.path), func(t *testing.T) {
			resp := testutil.DoRequest(t, e.method, env.Server.URL+e.path, nil, invalidToken)
			testutil.AssertStatus(t, resp, http.StatusUnauthorized)
			resp.Body.Close()
		})
	}
}

// ============================================================
// Helpers
// ============================================================

func assertErrorDetailContainsField(t *testing.T, details []api.ErrorDetail, field string) {
	t.Helper()
	for _, d := range details {
		if d.Field == field {
			return
		}
	}
	t.Errorf("expected error_details to contain field '%s', got: %+v", field, details)
}
