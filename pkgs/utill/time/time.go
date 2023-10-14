package time

import (
	"time"
)

func CurrentUTCTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func CurrentUTCTimeTillMinutes() string {
	return time.Now().UTC().Format("2006-01-02T15:04Z")
}
