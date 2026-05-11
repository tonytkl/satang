package utils

import "time"

func ParseDate(dateText string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, dateText); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", dateText)
}
