package service

import (
	"testing"

	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/reference"
)

func TestValidateGym(t *testing.T) {
	valid := domain.Gym{Name: "باشگاه آریا", Province: "تهران", City: "تهران", Address: "ولیعصر", Sport: "کاراته", SportStyle: "کیوکوشین", ActiveWeekdays: []int{0, 2, 4}, StartTime: "09:00", EndTime: "11:00", TuitionMode: domain.TuitionMonthly, TuitionAmount: 2_000_000}
	if !reference.ValidAddress(valid.Province, valid.City) {
		t.Fatal("reference address rejected")
	}
	if err := validateGym(&valid); err != nil {
		t.Fatalf("valid gym rejected: %v", err)
	}
	for name, change := range map[string]func(*domain.Gym){
		"duplicate weekday":  func(g *domain.Gym) { g.ActiveWeekdays = []int{1, 1} },
		"invalid weekday":    func(g *domain.Gym) { g.ActiveWeekdays = []int{7} },
		"invalid time range": func(g *domain.Gym) { g.StartTime = "12:00"; g.EndTime = "10:00" },
		"invalid tuition":    func(g *domain.Gym) { g.TuitionMode = "weekly" },
	} {
		g := valid
		change(&g)
		if validateGym(&g) == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
