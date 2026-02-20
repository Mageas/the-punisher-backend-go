package service

import "time"

func doubleTimePtr(value time.Time) **time.Time {
	ptr := &value
	return &ptr
}
