package util

import (
	"strconv"
	"time"

	"gopkg.in/errgo.v2/fmt/errors"
)

var dateFormats = []string{
	"2006-01-02T15:04:05.123Z",
	"2006-01-02T15:04:05.123",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02",
}

func ParseTimestamp(timestamp string) (time.Time, error) {
	if timeMs, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		return time.UnixMilli(timeMs), nil
	}

	for _, format := range dateFormats {
		if val, e := time.Parse(format, timestamp); e == nil {
			return val, nil
		}
	}
	return time.Time{}, errors.Newf("unable to parse timestamp: %s", timestamp)
}

func ContainsString(list []string, element string) bool {
	for _, it := range list {
		if it == element {
			return true
		}
	}
	return false
}

func ContainsInt32(list []int32, element int32) bool {
	for _, it := range list {
		if it == element {
			return true
		}
	}
	return false
}

func StringArraysEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
