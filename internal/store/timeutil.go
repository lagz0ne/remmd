package store

import "time"

const sqliteTimeFmt = "2006-01-02 15:04:05"

func formatTime(t time.Time) string {
	return t.UTC().Format(sqliteTimeFmt)
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(sqliteTimeFmt, s)
	return t
}
