package service

import "testing"

func TestWithdrawalFee(t *testing.T) {
	if got := int64(1000000) * 200 / 10000; got != 20000 {
		t.Fatalf("fee=%d", got)
	}
}
