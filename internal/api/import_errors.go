package api

type ImportErrorResponse struct {
	Error        string              `json:"error"`
	ErrorDetails []ImportErrorDetail `json:"error_details,omitempty"`
	ErrorCode    int                 `json:"error_code"`
}

type ImportErrorDetail struct {
	Row          *int     `json:"row,omitempty"`
	Field        string   `json:"field"`
	Error        string   `json:"error"`
	Value        string   `json:"value,omitempty"`
	ErrorDetails []string `json:"error_details,omitempty"`
}

type ImportValidationError struct {
	Message    string
	StatusCode int
	Details    []ImportErrorDetail
}

func (e *ImportValidationError) Error() string { return e.Message }

func NewImportValidationError(details ...ImportErrorDetail) *ImportValidationError {
	return &ImportValidationError{
		Message:    ErrImportValidationFailed.Message,
		StatusCode: ErrImportValidationFailed.StatusCode,
		Details:    details,
	}
}
