package service

import (
	"testing"
	"time"
)

func TestPersianYearMonth(t *testing.T) {
	tests := []struct {
		date        string
		year, month int
	}{
		{"2024-03-20", 1403, 1},
		{"2025-03-20", 1403, 12},
		{"2025-03-21", 1404, 1},
		{"2026-08-14", 1405, 5},
	}
	for _, test := range tests {
		value, _ := time.Parse("2006-01-02", test.date)
		year, month := persianYearMonth(value)
		if year != test.year || month != test.month {
			t.Fatalf("%s: got %d/%d, want %d/%d", test.date, year, month, test.year, test.month)
		}
	}
}
