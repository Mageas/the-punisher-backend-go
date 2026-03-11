package handler

import (
	"net/http"
	"strings"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/platform/auth"
	"github.com/mageas/the-punisher-backend/internal/platform/validator"
	"github.com/mageas/the-punisher-backend/internal/platform/web"
	"github.com/mageas/the-punisher-backend/internal/service"
)

const studentImportMaxUploadSizeBytes int64 = 10 * 1024 * 1024

type StudentHandler struct {
	service service.StudentService
}

func NewStudentHandler(service service.StudentService) *StudentHandler {
	return &StudentHandler{service: service}
}

func (h *StudentHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	var req dto.RequestStudentDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	student, err := h.service.CreateStudent(r.Context(), userID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, student, nil)
}

func (h *StudentHandler) CreateStudentsInClassroom(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	classroomID, ok := parsePathUUID(w, r, "classroom_id")
	if !ok {
		return
	}

	var req dto.ClassroomStudentsBatchRequestDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}
	for _, studentReq := range req.Students {
		if err := validator.ValidateStruct(studentReq); err != nil {
			web.WriteValidationError(w, err)
			return
		}
	}

	students, err := h.service.CreateStudentsInClassroom(r.Context(), userID, classroomID, req.Students)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusCreated, students, nil)
}

func (h *StudentHandler) ImportStudents(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		web.WriteAPIError(w, api.ErrImportFileInvalid, []api.ErrorDetail{{
			Field: "content_type",
			Error: "expected_multipart_form_data",
			Value: contentType,
		}})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, studentImportMaxUploadSizeBytes)
	if err := r.ParseMultipartForm(studentImportMaxUploadSizeBytes); err != nil {
		web.WriteAPIError(w, api.ErrImportFileInvalid, []api.ErrorDetail{{
			Field: "file",
			Error: "failed_to_parse_multipart_form",
		}})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		web.WriteAPIError(w, api.ErrImportFileMissing, []api.ErrorDetail{{
			Field: "file",
			Error: "file_field_is_required",
		}})
		return
	}
	defer file.Close()

	importResult, err := h.service.ImportStudents(r.Context(), userID, file, fileHeader.Filename)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, importResult, nil)
}

func (h *StudentHandler) GetStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	student, err := h.service.GetStudent(r.Context(), userID, studentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, student, nil)
}

func (h *StudentHandler) GetStudentKpis(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	kpis, err := h.service.GetStudentKpis(r.Context(), userID, studentID)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, kpis, nil)
}

func (h *StudentHandler) GetStudentHistory(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	limit, offset, page := web.ParsePagination(r)

	history, totalCount, err := h.service.ListStudentHistory(r.Context(), userID, studentID, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(history, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *StudentHandler) ListStudents(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	limit, offset, page := web.ParsePagination(r)
	search := web.ParseSearchQueryParam(r, "search")

	students, totalCount, err := h.service.ListStudents(r.Context(), userID, search, limit, offset)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	response := web.NewPaginatedResponse(students, totalCount, page, int(limit))
	web.WriteJSON(w, http.StatusOK, response, nil)
}

func (h *StudentHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	var req dto.UpdateStudentDto
	if err := web.DecodeJSON(r, &req); err != nil {
		web.WriteJSONDecodeError(w, err)
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		web.WriteValidationError(w, err)
		return
	}

	student, err := h.service.UpdateStudent(r.Context(), userID, studentID, req)
	if err != nil {
		web.WriteFromError(w, err)
		return
	}

	web.WriteJSON(w, http.StatusOK, student, nil)
}

func (h *StudentHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	studentID, ok := parsePathUUID(w, r, "student_id")
	if !ok {
		return
	}

	if err := h.service.DeleteStudent(r.Context(), userID, studentID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StudentHandler) DeleteAllStudents(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserIDFromContext(r.Context())

	if err := h.service.DeleteAllStudents(r.Context(), userID); err != nil {
		web.WriteFromError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
