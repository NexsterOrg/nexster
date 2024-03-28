package time

import (
	"time"

	umath "github.com/NamalSanjaya/nexster/pkgs/utill/math"
)

func CurrentUTCTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func CurrentUTCTimeTillMinutes() string {
	return time.Now().UTC().Format("2006-01-02T15:04Z")
}

func AddMinToCurrentTime(amountInMin int) int64 {
	return time.Now().Add(time.Duration(amountInMin) * time.Minute).Unix()
}

func HasUnixTimeExceeded(unixTime int64) bool {
	return unixTime < time.Now().Unix()
}

func GetRandomDateBetweenDays(n, m int) string {
	return time.Now().AddDate(0, 0, umath.GetRandValBetweenNumbers(n, m)).Format(time.RFC3339)
}

func SleepInSecond(secs int) {
	time.Sleep(time.Duration(secs) * time.Second)
}
