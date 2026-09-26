package config

import (
	"fmt"
	"time"
)

type TimeOfDay struct {
	Hour   int
	Minute int
}

func (t *TimeOfDay) UnmarshalText(text []byte) error {
	parsed, err := time.Parse("15:04", string(text))
	if err != nil {
		return fmt.Errorf("want HH:MM: %w", err)
	}
	t.Hour, t.Minute = parsed.Hour(), parsed.Minute()
	return nil
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

func (t TimeOfDay) On(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, t.Hour, t.Minute, 0, 0, date.Location())
}

func (t TimeOfDay) Before(u TimeOfDay) bool {
	return t.Hour*60+t.Minute < u.Hour*60+u.Minute
}
