package service

import (
	"testing"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
)

func TestEnrichBodyMeasurementAdultCategories(t *testing.T) {
	birth := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	measured := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		weight float64
		status string
	}{
		{50, "underweight"},
		{65, "healthy"},
		{80, "overweight"},
		{95, "obesity_1"},
		{110, "obesity_2"},
		{125, "obesity_3"},
	}
	for _, test := range tests {
		item := domain.BodyMeasurement{HeightCM: 170, WeightKG: test.weight, MeasuredAt: measured}
		enrichBodyMeasurement(&item, birth, domain.GenderMale)
		if item.Status != test.status {
			t.Errorf("weight %.1f: status=%s, want %s", test.weight, item.Status, test.status)
		}
	}
}

func TestEnrichBodyMeasurementUsesCDCReferenceBelowTwenty(t *testing.T) {
	item := domain.BodyMeasurement{
		HeightCM:   170,
		WeightKG:   65,
		MeasuredAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	enrichBodyMeasurement(&item, time.Date(2010, time.January, 2, 0, 0, 0, 0, time.UTC), domain.GenderFemale)
	if item.Reference != "cdc_bmi_for_age" || item.Percentile == nil {
		t.Fatalf("reference=%s percentile=%v, want CDC age/sex result", item.Reference, item.Percentile)
	}
	if item.HealthyMin == 0 || item.HealthyMax == 0 {
		t.Fatal("age-specific healthy BMI range was not calculated")
	}
}
