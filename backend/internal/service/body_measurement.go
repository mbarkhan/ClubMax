package service

import (
	"context"
	"math"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/reference"
	"github.com/clubmax/clubmax/internal/repository"
	"github.com/google/uuid"
)

const bodyMeasurementInterval = 7 * 24 * time.Hour

type BodyMeasurementService struct {
	measurements repository.BodyMeasurementRepository
	users        repository.UserRepository
	now          func() time.Time
}

func NewBodyMeasurementService(measurements repository.BodyMeasurementRepository, users repository.UserRepository) *BodyMeasurementService {
	return &BodyMeasurementService{measurements: measurements, users: users, now: time.Now}
}

func (s *BodyMeasurementService) Overview(ctx context.Context, athleteID string) (*domain.BMIOverview, error) {
	user, err := s.users.GetByID(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	if user.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	items, err := s.measurements.ListBodyMeasurements(ctx, athleteID, 104)
	if err != nil {
		return nil, err
	}
	for index := range items {
		enrichBodyMeasurement(&items[index], user.BirthDate, user.Gender)
	}
	overview := &domain.BMIOverview{Measurements: items, CanRecord: true}
	if len(items) > 0 {
		overview.Latest = &items[0]
		next := items[0].MeasuredAt.Add(bodyMeasurementInterval)
		overview.CanRecord = !s.now().Before(next)
		if !overview.CanRecord {
			overview.NextAllowed = &next
		}
	}
	return overview, nil
}

func (s *BodyMeasurementService) Record(ctx context.Context, athleteID string, heightCM, weightKG float64) (*domain.BodyMeasurement, error) {
	if math.IsNaN(heightCM) || math.IsInf(heightCM, 0) || heightCM < 40 || heightCM > 250 {
		return nil, domain.InvalidField("height_cm")
	}
	if math.IsNaN(weightKG) || math.IsInf(weightKG, 0) || weightKG < 5 || weightKG > 500 {
		return nil, domain.InvalidField("weight_kg")
	}
	user, err := s.users.GetByID(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	if user.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	now := s.now()
	item := &domain.BodyMeasurement{
		ID:         uuid.NewString(),
		AthleteID:  athleteID,
		HeightCM:   math.Round(heightCM*10) / 10,
		WeightKG:   math.Round(weightKG*10) / 10,
		MeasuredAt: now,
	}
	if err = s.measurements.CreateBodyMeasurement(ctx, item, now.Add(-bodyMeasurementInterval)); err != nil {
		return nil, err
	}
	enrichBodyMeasurement(item, user.BirthDate, user.Gender)
	return item, nil
}

func enrichBodyMeasurement(item *domain.BodyMeasurement, birthDate time.Time, gender domain.Gender) {
	heightM := item.HeightCM / 100
	item.BMI = math.Round(item.WeightKG/(heightM*heightM)*10) / 10
	months := ageMonthsAt(birthDate, item.MeasuredAt)
	item.AgeYears = months / 12
	if months < 240 {
		if enrichChildBodyMeasurement(item, birthDate, gender) {
			return
		}
		item.Status = "age_specific"
		item.Reference = "clinical_age_sex"
		return
	}
	item.Reference = "adult"
	item.HealthyMin, item.HealthyMax = 18.5, 25
	switch {
	case item.BMI < 18.5:
		item.Status = "underweight"
	case item.BMI < 25:
		item.Status = "healthy"
	case item.BMI < 30:
		item.Status = "overweight"
	case item.BMI < 35:
		item.Status = "obesity_1"
	case item.BMI < 40:
		item.Status = "obesity_2"
	default:
		item.Status = "obesity_3"
	}
}

func enrichChildBodyMeasurement(item *domain.BodyMeasurement, birthDate time.Time, gender domain.Gender) bool {
	months := ageMonthsAt(birthDate, item.MeasuredAt)
	sex := 0
	if gender == domain.GenderMale {
		sex = 1
	} else if gender == domain.GenderFemale {
		sex = 2
	}
	row, ok := reference.CDCBMIAgeRow(sex, months)
	if !ok {
		return false
	}
	z := math.Log(item.BMI/row.M) / row.S
	if row.L != 0 {
		z = (math.Pow(item.BMI/row.M, row.L) - 1) / (row.L * row.S)
	}
	percentile := math.Round(50*(1+math.Erf(z/math.Sqrt2))*10) / 10
	percentile = math.Max(0.1, math.Min(99.9, percentile))
	item.Percentile = &percentile
	item.HealthyMin = math.Round(row.P5*10) / 10
	item.HealthyMax = math.Round(row.P85*10) / 10
	item.Reference = "cdc_bmi_for_age"
	switch {
	case percentile < 5:
		item.Status = "underweight"
	case percentile < 85:
		item.Status = "healthy"
	case percentile < 95:
		item.Status = "overweight"
	case item.BMI >= math.Min(row.P95*1.2, 35):
		item.Status = "severe_obesity"
	default:
		item.Status = "obesity_1"
	}
	return true
}

func ageMonthsAt(birthDate, at time.Time) int {
	months := (at.Year()-birthDate.Year())*12 + int(at.Month()-birthDate.Month())
	if at.Day() < birthDate.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}
