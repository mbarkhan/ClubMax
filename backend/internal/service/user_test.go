package service

import (
	"testing"

	"github.com/clubmax/clubmax/internal/domain"
)

func TestValidGender(t *testing.T) {
	for _, gender := range []domain.Gender{domain.GenderMale, domain.GenderFemale, domain.GenderOther} {
		if !validGender(gender) {
			t.Errorf("expected gender %q to be valid", gender)
		}
	}
	if validGender(domain.Gender("invalid")) {
		t.Error("expected unknown gender to be invalid")
	}
}

func TestNormalizeIranianIBAN(t *testing.T) {
	tests := []struct {
		input, want string
		valid       bool
	}{
		{"IR820540102680020817909002", "IR820540102680020817909002", true},
		{"ir82 0540-1026 8002 0817 9090 02", "IR820540102680020817909002", true},
		{"IR820540102680020817909003", "", false},
		{"DE89370400440532013000", "", false},
		{"IR123", "", false},
	}
	for _, tt := range tests {
		got, valid := NormalizeIranianIBAN(tt.input)
		if got != tt.want || valid != tt.valid {
			t.Errorf("NormalizeIranianIBAN(%q) = (%q,%v), want (%q,%v)", tt.input, got, valid, tt.want, tt.valid)
		}
	}
}

func TestValidNationalID(t *testing.T) {
	for _, value := range []string{"0013547429", "0499370899"} {
		if !validNationalID(value) {
			t.Errorf("expected %s to be valid", value)
		}
	}
	for _, value := range []string{"1111111111", "0013547428", "123"} {
		if validNationalID(value) {
			t.Errorf("expected %s to be invalid", value)
		}
	}
}
