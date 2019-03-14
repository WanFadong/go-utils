package go_utils

import (
	"time"
)

func GetTimesFromUnix(unixs []int64) (tfs []string) {
	tfs = make([]string, len(unixs))
	var tf string
	for i, unix := range unixs {
		tf = GetTimeFromUnix(unix)
		tfs[i] = tf
	}
	return
}

func GetTimeFromUnix(unix int64) (tf string) {
	t := time.Unix(unix/1e7, 0)
	tf = t.Format("2006-01-02 15:04:05")
	return
}

func FormatTime(t time.Time) (tf string) {
	tf = t.Format("2006-01-02 15:04:05")
	return
}
