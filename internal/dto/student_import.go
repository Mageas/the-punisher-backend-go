package dto

type StudentImportRowErrorDto struct {
	Row          int      `json:"row"`
	Field        string   `json:"field"`
	Key          string   `json:"key"`
	Value        string   `json:"value,omitempty"`
	ErrorDetails []string `json:"error_details,omitempty"`
}

type StudentImportSummaryDto struct {
	RowsTotal          int `json:"rows_total"`
	RowsProcessed      int `json:"rows_processed"`
	ClassroomsCreated  int `json:"classrooms_created"`
	ClassroomsExisting int `json:"classrooms_existing"`
	StudentsCreated    int `json:"students_created"`
	StudentsExisting   int `json:"students_existing"`
	LinksCreated       int `json:"links_created"`
	LinksExisting      int `json:"links_existing"`
	RowsFailed         int `json:"rows_failed"`
}

type StudentImportResultDto struct {
	Summary StudentImportSummaryDto    `json:"summary"`
	Errors  []StudentImportRowErrorDto `json:"errors"`
}
