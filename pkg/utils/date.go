package utils

import (
	"fmt"
	"time"
)

const datetimeFormat = "2006-01-02 15:04:05"

type DateTime struct {
	val time.Time
}

func NewDateTime(val time.Time) DateTime {
	return DateTime{
		val: val,
	}
}

func NewDateTimeFromString(val string) (DateTime, error) {
	timeParsed, err := time.Parse(datetimeFormat, val)
	if err != nil {
		return DateTime{}, fmt.Errorf("invalid datetime format: %w", err)
	}
	return NewDateTime(timeParsed), nil
}