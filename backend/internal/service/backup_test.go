package service

import (
	"testing"
	"time"
)

func TestBackupNextRunUsesTehranSchedule(t *testing.T) {
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Fatal(err)
	}
	service := &BackupService{location: location}
	schedule := []int32{6 * 60, 14 * 60, 23 * 60}
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{"before first run", time.Date(2026, 8, 31, 5, 10, 0, 0, location), time.Date(2026, 8, 31, 6, 0, 0, 0, location)},
		{"between runs", time.Date(2026, 8, 31, 14, 1, 0, 0, location), time.Date(2026, 8, 31, 23, 0, 0, 0, location)},
		{"after last run", time.Date(2026, 8, 31, 23, 1, 0, 0, location), time.Date(2026, 9, 1, 6, 0, 0, 0, location)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := service.nextRun(test.now, schedule); !got.Equal(test.want) {
				t.Fatalf("nextRun() = %s, want %s", got, test.want)
			}
		})
	}
}

func TestBackupNextRunSupportsMinutes(t *testing.T) {
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Fatal(err)
	}
	service := &BackupService{location: location}
	now := time.Date(2026, 8, 31, 14, 5, 0, 0, location)
	want := time.Date(2026, 8, 31, 14, 15, 0, 0, location)
	if got := service.nextRun(now, []int32{6 * 60, 14*60 + 15, 23 * 60}); !got.Equal(want) {
		t.Fatalf("nextRun() = %s, want %s", got, want)
	}
}

func TestBackupScheduleValidation(t *testing.T) {
	if validSchedule([]int32{-1, 360}) || validSchedule([]int32{1440}) || validSchedule(nil) {
		t.Fatal("invalid backup schedule was accepted")
	}
	if !validSchedule([]int32{360, 840, 1380}) {
		t.Fatal("valid backup schedule was rejected")
	}
}
