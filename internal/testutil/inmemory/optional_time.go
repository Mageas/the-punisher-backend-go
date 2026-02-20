package inmemory

import "time"

func hasTime(value **time.Time) bool {
	return value != nil && *value != nil
}

func doubleTimePtr(value time.Time) **time.Time {
	timeValue := value
	ptr := &timeValue
	return &ptr
}
