package utils

import (
	"fmt"
	"time"
)

func ParseStringToTime(timeString string) (time.Time, error) {
	// The layout "2006-01-02 15:04:05" matches "YYYY-MM-DD HH:MM:SS"
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeString)
	if err != nil {
		return time.Time{}, 
		fmt.Errorf("failed to parse transaction time '%s': %w", timeString, err)
	}
	return t, nil
}