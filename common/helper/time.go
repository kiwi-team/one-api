package helper

import (
	"fmt"
	"time"
)

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetDayString() string {
	return time.Now().Add(time.Hour * 8).Format("20060102")
}
func GetYesterDayString() string {
	return time.Now().Add(-time.Hour * 24).Add(time.Hour * 8).Format("20060102")
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}
