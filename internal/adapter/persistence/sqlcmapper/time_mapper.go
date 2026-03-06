package sqlcmapper

import "time"

func normalizeAPITime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func normalizeOptionalAPITime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	timeValue := normalizeAPITime(*value)
	return &timeValue
}
