package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/mageas/the-punisher-backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

const (
	studentImportMaxRows = 1000

	studentNameMinLen = 2
	studentNameMaxLen = 70

	classroomNameMinLen = 2
	classroomNameMaxLen = 100
)

type transactionalStudentRepo interface {
	repository.Querier
	WithinTransaction(ctx context.Context, fn func(repository.Querier) error) error
}

type studentImportColumnIndexes struct {
	students int
	classes  int
}

type parsedStudentImportRow struct {
	RowNumber  int
	FirstName  string
	LastName   string
	ClassNames []string
}

type studentImportKey struct {
	FirstName string
	LastName  string
}

type studentClassroomImportKey struct {
	StudentID   uuid.UUID
	ClassroomID uuid.UUID
}

func (s *studentService) ImportStudents(ctx context.Context, userID uuid.UUID, file io.Reader, filename string) (*dto.StudentImportResultDto, error) {
	rawRows, err := parseStudentImportFile(file, filename)
	if err != nil {
		return nil, err
	}

	parsedRows, rowErrors, err := parseAndValidateStudentImportRows(rawRows)
	if err != nil {
		return nil, err
	}

	if len(parsedRows) == 0 {
		rowErrors = append(rowErrors, dto.StudentImportRowErrorDto{
			Row:     1,
			Field:   "file",
			Message: "at least one non-empty data row is required",
		})
	}
	if len(parsedRows) > studentImportMaxRows {
		rowErrors = append(rowErrors, dto.StudentImportRowErrorDto{
			Row:     1,
			Field:   "rows",
			Message: fmt.Sprintf("maximum %d rows are allowed", studentImportMaxRows),
			Value:   fmt.Sprintf("%d", len(parsedRows)),
		})
	}
	if len(rowErrors) > 0 {
		return nil, newImportValidationError(rowErrors)
	}

	txRepo, ok := s.repo.(transactionalStudentRepo)
	if !ok {
		return nil, fmt.Errorf("student repository does not support transactions")
	}

	result := &dto.StudentImportResultDto{
		Summary: dto.StudentImportSummaryDto{
			RowsTotal:     len(parsedRows),
			RowsProcessed: 0,
			RowsFailed:    0,
		},
		Errors: []dto.StudentImportRowErrorDto{},
	}

	err = txRepo.WithinTransaction(ctx, func(txQuerier repository.Querier) error {
		classroomIDsByName, err := ensureImportClassrooms(ctx, txQuerier, userID, parsedRows, &result.Summary)
		if err != nil {
			return err
		}

		studentIDsByKey, err := listStudentsForImport(ctx, txQuerier, userID)
		if err != nil {
			return err
		}
		existingLinks, err := listStudentClassroomLinksForImport(ctx, txQuerier, userID, studentIDsByKey)
		if err != nil {
			return err
		}

		for _, row := range parsedRows {
			studentKey := makeStudentImportKey(row.FirstName, row.LastName)

			studentID, exists := studentIDsByKey[studentKey]
			if exists {
				result.Summary.StudentsExisting++
			} else {
				createdStudent, err := txQuerier.CreateStudent(ctx, repository.CreateStudentParams{
					UserID:    userID,
					FirstName: row.FirstName,
					LastName:  row.LastName,
				})
				if err != nil {
					return fmt.Errorf("failed to create student at row %d: %w", row.RowNumber, err)
				}
				studentID = createdStudent.ID
				studentIDsByKey[studentKey] = studentID
				result.Summary.StudentsCreated++
			}

			for _, className := range row.ClassNames {
				classroomID, found := classroomIDsByName[className]
				if !found {
					return fmt.Errorf("missing classroom mapping for %q at row %d", className, row.RowNumber)
				}
				linkKey := studentClassroomImportKey{
					StudentID:   studentID,
					ClassroomID: classroomID,
				}
				if _, exists := existingLinks[linkKey]; exists {
					result.Summary.LinksExisting++
					continue
				}

				rowsAffected, err := txQuerier.AddStudentToClassroom(ctx, repository.AddStudentToClassroomParams{
					StudentID:   studentID,
					ClassroomID: classroomID,
					UserID:      userID,
				})
				if err != nil {
					return fmt.Errorf("failed to create student-classroom link at row %d: %w", row.RowNumber, err)
				}
				if rowsAffected == 0 {
					return fmt.Errorf("failed to create student-classroom link at row %d: student or classroom not found", row.RowNumber)
				}

				result.Summary.LinksCreated++
				existingLinks[linkKey] = struct{}{}
			}

			result.Summary.RowsProcessed++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseStudentImportFile(file io.Reader, filename string) ([][]string, error) {
	extension := strings.ToLower(strings.TrimSpace(filepath.Ext(filename)))

	switch extension {
	case ".xlsx":
		return parseXLSXRows(file)
	case ".csv":
		return parseCSVRows(file)
	default:
		return nil, api.NewAPIError(api.ErrImportFileInvalid.StatusCode, api.ErrImportFileInvalid.Message, api.ErrorDetail{
			Field: "file",
			Error: "unsupported_file_type",
			Value: extension,
		})
	}
}

func parseXLSXRows(file io.Reader) ([][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, api.NewAPIError(api.ErrImportFileInvalid.StatusCode, api.ErrImportFileInvalid.Message, api.ErrorDetail{
			Field: "file",
			Error: "failed_to_read_xlsx",
		})
	}
	defer workbook.Close()

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, api.NewAPIError(api.ErrImportTemplateInvalid.StatusCode, api.ErrImportTemplateInvalid.Message, api.ErrorDetail{
			Field: "headers",
			Error: "missing_headers_row",
		})
	}

	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return nil, api.NewAPIError(api.ErrImportFileInvalid.StatusCode, api.ErrImportFileInvalid.Message, api.ErrorDetail{
			Field: "file",
			Error: "failed_to_read_xlsx_rows",
		})
	}
	if len(rows) == 0 {
		return nil, api.NewAPIError(api.ErrImportTemplateInvalid.StatusCode, api.ErrImportTemplateInvalid.Message, api.ErrorDetail{
			Field: "headers",
			Error: "missing_headers_row",
		})
	}

	return rows, nil
}

func parseCSVRows(file io.Reader) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, api.NewAPIError(api.ErrImportFileInvalid.StatusCode, api.ErrImportFileInvalid.Message, api.ErrorDetail{
			Field: "file",
			Error: "failed_to_read_csv",
		})
	}
	if len(rows) == 0 {
		return nil, api.NewAPIError(api.ErrImportTemplateInvalid.StatusCode, api.ErrImportTemplateInvalid.Message, api.ErrorDetail{
			Field: "headers",
			Error: "missing_headers_row",
		})
	}

	return rows, nil
}

func parseAndValidateStudentImportRows(rawRows [][]string) ([]parsedStudentImportRow, []dto.StudentImportRowErrorDto, error) {
	columnIndexes, err := resolveStudentImportColumnIndexes(rawRows[0])
	if err != nil {
		return nil, nil, err
	}

	parsedRows := make([]parsedStudentImportRow, 0, len(rawRows)-1)
	rowErrors := make([]dto.StudentImportRowErrorDto, 0)

	for rowIndex, row := range rawRows[1:] {
		excelRowNumber := rowIndex + 2
		rawStudent := readStudentImportCell(row, columnIndexes.students)
		rawClasses := readStudentImportCell(row, columnIndexes.classes)

		if strings.TrimSpace(rawStudent) == "" && strings.TrimSpace(rawClasses) == "" {
			continue
		}

		firstName, lastName, studentErr := parseImportStudentName(rawStudent)
		if studentErr != "" {
			rowErrors = append(rowErrors, dto.StudentImportRowErrorDto{
				Row:     excelRowNumber,
				Field:   "eleves",
				Message: studentErr,
				Value:   strings.TrimSpace(rawStudent),
			})
		}

		classNames, classErrors := parseImportClassNames(rawClasses)
		for _, classError := range classErrors {
			rowErrors = append(rowErrors, dto.StudentImportRowErrorDto{
				Row:     excelRowNumber,
				Field:   "classes",
				Message: classError.Message,
				Value:   classError.Value,
			})
		}

		if studentErr != "" || len(classErrors) > 0 {
			continue
		}

		parsedRows = append(parsedRows, parsedStudentImportRow{
			RowNumber:  excelRowNumber,
			FirstName:  firstName,
			LastName:   lastName,
			ClassNames: classNames,
		})
	}

	return parsedRows, rowErrors, nil
}

func resolveStudentImportColumnIndexes(headerRow []string) (studentImportColumnIndexes, error) {
	indexes := studentImportColumnIndexes{
		students: -1,
		classes:  -1,
	}

	for index, rawHeader := range headerRow {
		normalizedHeader := normalizeImportHeader(rawHeader)
		switch normalizedHeader {
		case "eleves":
			if indexes.students == -1 {
				indexes.students = index
			}
		case "classes":
			if indexes.classes == -1 {
				indexes.classes = index
			}
		}
	}

	if indexes.students == -1 || indexes.classes == -1 {
		details := make([]api.ErrorDetail, 0, 2)
		if indexes.students == -1 {
			details = append(details, api.ErrorDetail{
				Field: "headers",
				Error: "missing_header_eleves",
			})
		}
		if indexes.classes == -1 {
			details = append(details, api.ErrorDetail{
				Field: "headers",
				Error: "missing_header_classes",
			})
		}
		return studentImportColumnIndexes{}, api.NewAPIError(api.ErrImportTemplateInvalid.StatusCode, api.ErrImportTemplateInvalid.Message, details...)
	}

	return indexes, nil
}

type classImportError struct {
	Message string
	Value   string
}

func parseImportClassNames(rawClasses string) ([]string, []classImportError) {
	trimmedClasses := strings.TrimSpace(rawClasses)
	if trimmedClasses == "" {
		return nil, []classImportError{{
			Message: "at least one classroom is required",
			Value:   "",
		}}
	}

	replacedSeparators := strings.ReplaceAll(trimmedClasses, ";", ",")
	parts := strings.Split(replacedSeparators, ",")

	seenClassNames := make(map[string]struct{}, len(parts))
	classNames := make([]string, 0, len(parts))
	classErrors := make([]classImportError, 0)

	for _, part := range parts {
		className := strings.TrimSpace(part)
		if className == "" {
			classErrors = append(classErrors, classImportError{
				Message: "classroom name is required",
				Value:   part,
			})
			continue
		}

		if !isLengthValid(className, classroomNameMinLen, classroomNameMaxLen) {
			classErrors = append(classErrors, classImportError{
				Message: fmt.Sprintf("classroom name must be between %d and %d characters", classroomNameMinLen, classroomNameMaxLen),
				Value:   className,
			})
			continue
		}

		if _, exists := seenClassNames[className]; exists {
			continue
		}
		seenClassNames[className] = struct{}{}

		classNames = append(classNames, className)
	}

	if len(classNames) == 0 && len(classErrors) == 0 {
		classErrors = append(classErrors, classImportError{
			Message: "at least one classroom is required",
			Value:   strings.TrimSpace(rawClasses),
		})
	}

	return classNames, classErrors
}

func parseImportStudentName(rawStudent string) (firstName string, lastName string, validationError string) {
	trimmedStudent := strings.TrimSpace(rawStudent)
	if trimmedStudent == "" {
		return "", "", "student name is required"
	}

	nameParts := strings.Fields(trimmedStudent)
	if len(nameParts) < 2 {
		return "", "", "student name format must be 'NOM Prenom'"
	}

	lastNamePartCount := 0
	for lastNamePartCount < len(nameParts) && isUppercaseStudentLastNamePart(nameParts[lastNamePartCount]) {
		lastNamePartCount++
	}
	if lastNamePartCount == 0 || lastNamePartCount == len(nameParts) {
		return "", "", "student name format must be 'NOM Prenom' with uppercase last name"
	}

	lastName = strings.Join(nameParts[:lastNamePartCount], " ")
	firstName = strings.Join(nameParts[lastNamePartCount:], " ")

	if !isLengthValid(lastName, studentNameMinLen, studentNameMaxLen) {
		return "", "", fmt.Sprintf("last name must be between %d and %d characters", studentNameMinLen, studentNameMaxLen)
	}
	if !isLengthValid(firstName, studentNameMinLen, studentNameMaxLen) {
		return "", "", fmt.Sprintf("first name must be between %d and %d characters", studentNameMinLen, studentNameMaxLen)
	}

	return firstName, lastName, ""
}

func isUppercaseStudentLastNamePart(part string) bool {
	hasLetter := false
	for _, character := range part {
		if !unicode.IsLetter(character) {
			continue
		}

		hasLetter = true
		if unicode.IsLower(character) {
			return false
		}
	}

	return hasLetter
}

func isLengthValid(value string, min int, max int) bool {
	length := utf8.RuneCountInString(value)
	return length >= min && length <= max
}

func readStudentImportCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func normalizeImportHeader(header string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(header, "\uFEFF")))
}

func collectDistinctClassNames(rows []parsedStudentImportRow) []string {
	seenClassNames := make(map[string]struct{})
	distinctClassNames := make([]string, 0)

	for _, row := range rows {
		for _, className := range row.ClassNames {
			if _, exists := seenClassNames[className]; exists {
				continue
			}
			seenClassNames[className] = struct{}{}
			distinctClassNames = append(distinctClassNames, className)
		}
	}

	return distinctClassNames
}

func ensureImportClassrooms(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	rows []parsedStudentImportRow,
	summary *dto.StudentImportSummaryDto,
) (map[string]uuid.UUID, error) {
	existingClassrooms, err := repo.ListClassroomsByUserForImport(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list classrooms for import: %w", err)
	}

	classroomIDsByName := make(map[string]uuid.UUID, len(existingClassrooms))
	for _, classroom := range existingClassrooms {
		if _, exists := classroomIDsByName[classroom.Name]; exists {
			continue
		}
		classroomIDsByName[classroom.Name] = classroom.ID
	}

	for _, className := range collectDistinctClassNames(rows) {
		if _, exists := classroomIDsByName[className]; exists {
			summary.ClassroomsExisting++
			continue
		}

		createdClassroom, err := repo.CreateClassroom(ctx, repository.CreateClassroomParams{
			UserID: userID,
			Name:   className,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create classroom %q during import: %w", className, err)
		}

		classroomIDsByName[className] = createdClassroom.ID
		summary.ClassroomsCreated++
	}

	return classroomIDsByName, nil
}

func listStudentsForImport(ctx context.Context, repo repository.Querier, userID uuid.UUID) (map[studentImportKey]uuid.UUID, error) {
	existingStudents, err := repo.ListStudentsByUserForImport(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list students for import: %w", err)
	}

	studentIDsByKey := make(map[studentImportKey]uuid.UUID, len(existingStudents))
	for _, student := range existingStudents {
		key := makeStudentImportKey(student.FirstName, student.LastName)
		if _, exists := studentIDsByKey[key]; exists {
			continue
		}
		studentIDsByKey[key] = student.ID
	}

	return studentIDsByKey, nil
}

func listStudentClassroomLinksForImport(
	ctx context.Context,
	repo repository.Querier,
	userID uuid.UUID,
	studentIDsByKey map[studentImportKey]uuid.UUID,
) (map[studentClassroomImportKey]struct{}, error) {
	if len(studentIDsByKey) == 0 {
		return map[studentClassroomImportKey]struct{}{}, nil
	}

	studentIDs := make([]uuid.UUID, 0, len(studentIDsByKey))
	seenStudentIDs := make(map[uuid.UUID]struct{}, len(studentIDsByKey))
	for _, studentID := range studentIDsByKey {
		if _, exists := seenStudentIDs[studentID]; exists {
			continue
		}
		seenStudentIDs[studentID] = struct{}{}
		studentIDs = append(studentIDs, studentID)
	}

	rows, err := repo.ListClassroomRefsByStudentIDs(ctx, repository.ListClassroomRefsByStudentIDsParams{
		UserID:     userID,
		StudentIds: studentIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list student-classroom links for import: %w", err)
	}

	links := make(map[studentClassroomImportKey]struct{}, len(rows))
	for _, row := range rows {
		links[studentClassroomImportKey{
			StudentID:   row.StudentID,
			ClassroomID: row.ClassroomID,
		}] = struct{}{}
	}

	return links, nil
}

func makeStudentImportKey(firstName string, lastName string) studentImportKey {
	return studentImportKey{
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
	}
}

func newImportValidationError(rowErrors []dto.StudentImportRowErrorDto) error {
	errorDetails := make([]api.ErrorDetail, 0, len(rowErrors))
	for _, rowError := range rowErrors {
		row := rowError.Row
		errorDetails = append(errorDetails, api.ErrorDetail{
			Row:   &row,
			Field: rowError.Field,
			Error: rowError.Message,
			Value: rowError.Value,
		})
	}

	return api.NewAPIError(api.ErrImportValidationFailed.StatusCode, api.ErrImportValidationFailed.Message, errorDetails...)
}
