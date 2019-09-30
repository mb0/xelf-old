// Package utl provides helpers to extend the xelf language and some common library function specs.
package utl

import (
	"strings"
	"time"
)

var StrLib = Lazy(Fmap{
	"str:contains":     strings.Contains,
	"str:contains_any": strings.ContainsAny,
	"str:prefix":       strings.HasPrefix,
	"str:suffix":       strings.HasSuffix,
	"str:upper":        strings.ToUpper,
	"str:lower":        strings.ToLower,
	"str:trim":         strings.TrimSpace,
})

var TimeLib = Lazy(Fmap{
	"time:sub":       time.Time.Sub,
	"time:add_date":  time.Time.AddDate,
	"time:add_days":  timeAddDays,
	"time:year":      time.Time.Year,
	"time:month":     time.Time.Month,
	"time:weekday":   time.Time.Weekday,
	"time:yearday":   time.Time.YearDay,
	"time:format":    time.Time.Format,
	"time:date_long": timeDateLong,
	"time:day_start": DayStart,
	"time:day_end":   DayEnd,
})

func timeAddDays(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }
func timeDateLong(t time.Time) string             { return t.Format("2006-01-02") }
func DayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func DayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, -1, 0, t.Location())
}
