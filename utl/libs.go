package utl

import (
	"strings"
	"time"
)

var StrLib = Lazy(Fmap{
	"str.contains":     strings.Contains,
	"str.contains_any": strings.ContainsAny,
	"str.prefix":       strings.HasPrefix,
	"str.suffix":       strings.HasSuffix,
	"str.upper":        strings.ToUpper,
	"str.lower":        strings.ToLower,
	"str.trim":         strings.TrimSpace,
})

var TimeLib = Lazy(Fmap{
	"time.sub":       time.Time.Sub,
	"time.add_date":  time.Time.AddDate,
	"time.add_days":  timeAddDays,
	"time.year":      time.Time.Year,
	"time.month":     time.Time.Month,
	"time.weekday":   time.Time.Weekday,
	"time.yearday":   time.Time.YearDay,
	"time.format":    time.Time.Format,
	"time.date_long": timeDateLong,
})

func timeAddDays(t time.Time, days int) time.Time { return t.AddDate(0, 0, days) }
func timeDateLong(t time.Time) string             { return t.Format("2006-01-02") }
