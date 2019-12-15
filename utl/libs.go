// Package utl provides helpers to extend the xelf language and some common library function specs.
package utl

import (
	"strings"
	"time"
)

var StrLib = Lazy(Fmap{
	"str_contains":     strings.Contains,
	"str_contains_any": strings.ContainsAny,
	"str_prefix":       strings.HasPrefix,
	"str_suffix":       strings.HasSuffix,
	"str_upper":        strings.ToUpper,
	"str_lower":        strings.ToLower,
	"str_trim":         strings.TrimSpace,
})

var TimeLib = Lazy(Fmap{
	"time_sub":       time.Time.Sub,
	"time_add_date":  time.Time.AddDate,
	"time_add_days":  timeAddDays,
	"time_year":      time.Time.Year,
	"time_month":     time.Time.Month,
	"time_weekday":   time.Time.Weekday,
	"time_yearday":   time.Time.YearDay,
	"time_format":    time.Time.Format,
	"time_date_long": timeDateLong,
	"time_day_start": DayStart,
	"time_day_end":   DayEnd,
})

func timeAddDays(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }
func timeDateLong(t time.Time) string             { return t.Format("2006-01-02") }
func DayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func DayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, -1, 0, t.Location())
}
