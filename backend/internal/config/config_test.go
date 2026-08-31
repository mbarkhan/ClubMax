package config

import (
	"reflect"
	"testing"
)

func TestParseBackupHours(t *testing.T) {
	got, err := parseHours("23, 6,14,6")
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{6, 14, 23}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parseHours() = %v, want %v", got, want)
	}
}

func TestParseBackupHoursRejectsInvalidHour(t *testing.T) {
	if _, err := parseHours("6,24"); err == nil {
		t.Fatal("parseHours() accepted hour 24")
	}
}
