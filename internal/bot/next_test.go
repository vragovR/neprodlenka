package bot

import (
	"testing"
	"time"

	"neprodlenka/internal/config"
)

func TestNextTick(t *testing.T) {
	b := &Bot{postTime: config.TimeOfDay{Hour: 7}, closeTime: config.TimeOfDay{Hour: 9}, summaryTime: config.TimeOfDay{Hour: 12}}
	at := func(day, hour, minute int) time.Time {
		return time.Date(2026, 9, day, hour, minute, 0, 0, time.Local)
	}

	tests := []struct {
		now, want time.Time
	}{
		{at(26, 0, 0), at(26, 7, 0)},
		{at(26, 7, 0), at(26, 9, 0)},
		{at(26, 8, 59), at(26, 9, 0)},
		{at(26, 9, 0), at(26, 12, 0)},
		{at(26, 12, 0), at(27, 7, 0)},
		{at(26, 23, 59), at(27, 7, 0)},
		{at(30, 12, 0), time.Date(2026, 10, 1, 7, 0, 0, 0, time.Local)},
	}
	for _, tt := range tests {
		if got := b.nextTick(tt.now); !got.Equal(tt.want) {
			t.Errorf("nextTick(%v) = %v, want %v", tt.now, got, tt.want)
		}
	}
}
