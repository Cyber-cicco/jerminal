package utils

import "time"

func FormatEuropeanDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format("02/01/2006 15:04")
}

