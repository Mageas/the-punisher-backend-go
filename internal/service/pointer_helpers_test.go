package service

import "time"

func doubleTimePtr(value time.Time) *time.Time {
	timeValue := value
	return &timeValue
}
