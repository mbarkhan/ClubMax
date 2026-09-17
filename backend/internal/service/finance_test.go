package service

import (
	"testing"
	"time"
)

func TestWithdrawalFee(t *testing.T) {
	if got := int64(1000000) * 200 / 10000; got != 20000 {
		t.Fatalf("fee=%d", got)
	}
}

func TestNextTehranSettlement(t *testing.T) {
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Fatal(err)
	}
	before := time.Date(2026, 8, 21, 22, 0, 0, 0, location)
	if got := nextTehranSettlement(before); !got.Equal(time.Date(2026, 8, 21, 23, 30, 0, 0, location)) {
		t.Fatalf("before cutoff: got %s", got)
	}
	after := time.Date(2026, 8, 21, 23, 30, 0, 0, location)
	if got := nextTehranSettlement(after); !got.Equal(time.Date(2026, 8, 22, 23, 30, 0, 0, location)) {
		t.Fatalf("at cutoff: got %s", got)
	}
}
