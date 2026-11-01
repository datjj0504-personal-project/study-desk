package study

import (
	"testing"
	"time"
)

func TestNewStatisticsWindowUsesLocalCalendarAndMondayWeek(t *testing.T) {
	location := time.FixedZone("UTC-7", -7*60*60)
	now := time.Date(2026, time.October, 25, 12, 30, 0, 0, location)

	window := NewStatisticsWindow(now)

	wantTodayStart := time.Date(2026, time.October, 25, 0, 0, 0, 0, location)
	wantWeekStart := time.Date(2026, time.October, 19, 0, 0, 0, 0, location)
	if !window.TodayStart.Equal(wantTodayStart) || !window.TodayEnd.Equal(now) {
		t.Errorf("today window = [%s, %s), want [%s, %s)", window.TodayStart, window.TodayEnd, wantTodayStart, now)
	}
	if !window.WeekStart.Equal(wantWeekStart) || !window.WeekEnd.Equal(now) {
		t.Errorf("week window = [%s, %s), want [%s, %s)", window.WeekStart, window.WeekEnd, wantWeekStart, now)
	}
	if window.TodayStart.Location() != location || window.WeekStart.Location() != location {
		t.Errorf("statistics windows lost local timezone: %+v", window)
	}
}
