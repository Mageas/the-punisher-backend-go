package dto

import (
	"bytes"
	"encoding/json"
)

// NullableStringField tracks field presence and allows explicit null values.
type NullableStringField struct {
	Set   bool
	Value *string
}

func (f *NullableStringField) UnmarshalJSON(data []byte) error {
	f.Set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.Value = &value
	return nil
}
