package service

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mageas/the-punisher-backend/internal/api"
	"github.com/mageas/the-punisher-backend/internal/dto"
	"github.com/xuri/excelize/v2"
)

func assertAPIError(t *testing.T, err error, wantMessage string) *api.APIError {
	t.Helper()

	var apiErr *api.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.Message != wantMessage {
		t.Fatalf("expected APIError message %q, got %q", wantMessage, apiErr.Message)
	}
	return apiErr
}

func makeXLSXBuffer(t *testing.T, rows [][]string) *bytes.Buffer {
	t.Helper()

	f := excelize.NewFile()
	sheetName := f.GetSheetName(0)

	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				t.Fatalf("failed to build cell name: %v", err)
			}
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				t.Fatalf("failed to set cell %s: %v", cell, err)
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to write xlsx buffer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close workbook: %v", err)
	}

	return buf
}

func TestResolveStudentImportColumnIndexes(t *testing.T) {
	indexes, err := resolveStudentImportColumnIndexes([]string{"\uFEFFEleves", "Classes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if indexes.students != 0 || indexes.classes != 1 {
		t.Fatalf("unexpected indexes: %+v", indexes)
	}
}

func TestResolveStudentImportColumnIndexesMissing(t *testing.T) {
	_, err := resolveStudentImportColumnIndexes([]string{"students"})
	apiErr := assertAPIError(t, err, api.ErrImportTemplateInvalid.Message)
	if len(apiErr.Details) == 0 {
		t.Fatalf("expected details")
	}
}

func TestParseCSVRows(t *testing.T) {
	rows, err := parseCSVRows(strings.NewReader("eleves,classes\nDUPONT Jean,6A\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[1][0] != "DUPONT Jean" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestParseCSVRowsInvalid(t *testing.T) {
	_, err := parseCSVRows(strings.NewReader("\"unterminated"))
	assertAPIError(t, err, api.ErrImportFileInvalid.Message)
}

func TestParseCSVRowsEmpty(t *testing.T) {
	_, err := parseCSVRows(strings.NewReader(""))
	assertAPIError(t, err, api.ErrImportTemplateInvalid.Message)
}

func TestParseXLSXRows(t *testing.T) {
	buf := makeXLSXBuffer(t, [][]string{{"eleves", "classes"}, {"DUPONT Jean", "6A"}})
	rows, err := parseXLSXRows(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[1][0] != "DUPONT Jean" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestParseXLSXRowsInvalidFile(t *testing.T) {
	_, err := parseXLSXRows(strings.NewReader("not-an-xlsx"))
	assertAPIError(t, err, api.ErrImportFileInvalid.Message)
}

func TestParseXLSXRowsMissingHeaders(t *testing.T) {
	buf := makeXLSXBuffer(t, nil)
	_, err := parseXLSXRows(bytes.NewReader(buf.Bytes()))
	assertAPIError(t, err, api.ErrImportTemplateInvalid.Message)
}

func TestParseStudentImportFile(t *testing.T) {
	csvRows, err := parseStudentImportFile(strings.NewReader("eleves,classes\nDUPONT Jean,6A\n"), "import.CSV")
	if err != nil {
		t.Fatalf("unexpected csv error: %v", err)
	}
	if len(csvRows) != 2 {
		t.Fatalf("unexpected csv rows: %#v", csvRows)
	}

	xlsxBuf := makeXLSXBuffer(t, [][]string{{"eleves", "classes"}, {"DUPONT Jean", "6A"}})
	xlsxRows, err := parseStudentImportFile(bytes.NewReader(xlsxBuf.Bytes()), "import.xlsx")
	if err != nil {
		t.Fatalf("unexpected xlsx error: %v", err)
	}
	if len(xlsxRows) != 2 {
		t.Fatalf("unexpected xlsx rows: %#v", xlsxRows)
	}
}

func TestParseImportClassNames(t *testing.T) {
	classes, errs := parseImportClassNames("6A; 6B,6A")
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %+v", errs)
	}
	if len(classes) != 2 || classes[0] != "6A" || classes[1] != "6B" {
		t.Fatalf("unexpected classes: %+v", classes)
	}
}

func TestParseImportClassNamesInvalid(t *testing.T) {
	classes, errs := parseImportClassNames("A")
	if len(classes) != 0 {
		t.Fatalf("expected no classes")
	}
	if len(errs) == 0 {
		t.Fatalf("expected at least one error")
	}
}

func TestParseImportStudentName(t *testing.T) {
	first, last, errMsg := parseImportStudentName("DUPONT Jean Claude")
	if errMsg != "" {
		t.Fatalf("unexpected validation error: %s", errMsg)
	}
	if first != "Jean Claude" || last != "DUPONT" {
		t.Fatalf("unexpected parsed names: first=%q last=%q", first, last)
	}
}

func TestParseImportStudentNameInvalid(t *testing.T) {
	_, _, errMsg := parseImportStudentName("Dupont Jean")
	if errMsg == "" {
		t.Fatalf("expected validation error")
	}
}

func TestParseAndValidateStudentImportRows(t *testing.T) {
	raw := [][]string{
		{"eleves", "classes"},
		{"DUPONT Jean", "6A,6B"},
		{"", ""},
		{"INVALID", "6A"},
	}

	parsed, rowErrors, err := parseAndValidateStudentImportRows(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected one valid row, got %d", len(parsed))
	}
	if len(rowErrors) == 0 {
		t.Fatalf("expected row errors")
	}
}

func TestParseStudentImportFileUnsupportedExtension(t *testing.T) {
	_, err := parseStudentImportFile(strings.NewReader("irrelevant"), "import.txt")
	assertAPIError(t, err, api.ErrImportFileInvalid.Message)
}

func TestShouldUppercaseLastNamePart(t *testing.T) {
	if !isUppercaseStudentLastNamePart("DUPONT") {
		t.Fatalf("expected uppercase name part")
	}
	if isUppercaseStudentLastNamePart("Dupont") {
		t.Fatalf("expected lowercase part to be invalid")
	}
}

func TestReadStudentImportCell(t *testing.T) {
	row := []string{"  DUPONT Jean  "}
	if got := readStudentImportCell(row, 0); got != "DUPONT Jean" {
		t.Fatalf("unexpected trimmed cell: %q", got)
	}
	if got := readStudentImportCell(row, 1); got != "" {
		t.Fatalf("expected empty for out-of-range, got %q", got)
	}
	if got := readStudentImportCell(row, -1); got != "" {
		t.Fatalf("expected empty for negative index, got %q", got)
	}
}

func TestCollectDistinctClassNames(t *testing.T) {
	rows := []parsedStudentImportRow{
		{ClassNames: []string{"6A", "6B"}},
		{ClassNames: []string{"6B", "6C"}},
	}

	got := collectDistinctClassNames(rows)
	want := []string{"6A", "6B", "6C"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected class names: got=%v want=%v", got, want)
	}
}

func TestMakeStudentImportKey(t *testing.T) {
	key := makeStudentImportKey("  Jean  ", "  DUPONT ")
	if key.FirstName != "Jean" || key.LastName != "DUPONT" {
		t.Fatalf("unexpected key: %+v", key)
	}
}

func TestNewImportValidationError(t *testing.T) {
	err := newImportValidationError([]dto.StudentImportRowErrorDto{{
		Row:     4,
		Field:   "classes",
		Message: "invalid",
		Value:   "x",
	}})

	apiErr := assertAPIError(t, err, api.ErrImportValidationFailed.Message)
	if len(apiErr.Details) != 1 {
		t.Fatalf("expected one detail, got %+v", apiErr.Details)
	}
	if apiErr.Details[0].Row == nil || *apiErr.Details[0].Row != 4 {
		t.Fatalf("unexpected row pointer: %+v", apiErr.Details[0])
	}
}
