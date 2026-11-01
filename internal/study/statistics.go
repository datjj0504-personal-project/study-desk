package study

import "time"

type StatisticsWindow struct {
	TodayStart time.Time
	TodayEnd   time.Time
	WeekStart  time.Time
	WeekEnd    time.Time
}

type Statistics struct {
	TodaySeconds   int64               `json:"todaySeconds"`
	WeekSeconds    int64               `json:"weekSeconds"`
	AllTimeSeconds int64               `json:"allTimeSeconds"`
	BySubject      []SubjectStatistics `json:"bySubject"`
}

type SubjectStatistics struct {
	SubjectID       *int64 `json:"subjectId"`
	SubjectName     string `json:"subjectName"`
	DurationSeconds int64  `json:"durationSeconds"`
}

func NewStatisticsWindow(now time.Time) StatisticsWindow {
	localNow := now.In(now.Location())
	todayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	daysSinceMonday := (int(localNow.Weekday()) + 6) % 7
	weekStart := todayStart.AddDate(0, 0, -daysSinceMonday)
	return StatisticsWindow{
		TodayStart: todayStart,
		TodayEnd:   localNow,
		WeekStart:  weekStart,
		WeekEnd:    localNow,
	}
}
