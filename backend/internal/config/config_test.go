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

func TestValidZibalGatewayURL(t *testing.T) {
	for _, test := range []struct {
		url   string
		valid bool
	}{
		{url: "https://gateway.zibal.ir", valid: true},
		{url: "https://gateway.zibal.ir/", valid: true},
		{url: "http://gateway.zibal.ir", valid: false},
		{url: "https://gateway.zibal.ir.evil.example", valid: false},
		{url: "https://gateway.zibal.ir@evil.example", valid: false},
		{url: "https://gateway.zibal.ir/v1", valid: false},
		{url: "https://gateway.zibal.ir?redirect=evil", valid: false},
	} {
		if got := validZibalGatewayURL(test.url); got != test.valid {
			t.Errorf("validZibalGatewayURL(%q) = %v, want %v", test.url, got, test.valid)
		}
	}
}
