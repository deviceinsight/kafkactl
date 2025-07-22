package util

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/errgo.v2/fmt/errors"
)

var dateFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	time.DateOnly,
}

func ParseTimestamp(timestamp string) (time.Time, error) {
	if timeMs, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
		return time.UnixMilli(timeMs), nil
	}

	loc, _ := time.LoadLocation("Local")

	for _, format := range dateFormats {
		if val, e := time.ParseInLocation(format, timestamp, loc); e == nil {
			return val, nil
		}
	}
	return time.Time{}, errors.Newf("unable to parse timestamp: %s", timestamp)
}

func ConvertControlChars(value string) string {
	value = strings.Replace(value, "\\n", "\n", -1)
	value = strings.Replace(value, "\\r", "\r", -1)
	value = strings.Replace(value, "\\t", "\t", -1)
	return value
}

func ContainsString(list []string, element string) bool {
	return slices.Contains(list, element)
}

func ContainsInt32(list []int32, element int32) bool {
	return slices.Contains(list, element)
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
