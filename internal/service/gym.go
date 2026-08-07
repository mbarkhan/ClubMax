package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"strings"
	"time"
)

const maxGymsPerCoach = 36

type GymService struct {
	gyms  repository.GymRepository
	users repository.UserRepository
	now   func() time.Time
}

func NewGymService(c repository.GymRepository, u repository.UserRepository) *GymService {
	return &GymService{c, u, time.Now}
}
func (s *GymService) Create(ctx context.Context, c *domain.Gym, coachID string) (*domain.Gym, error) {
	u, e := s.users.GetByID(ctx, coachID)
	if e != nil {
		return nil, e
	}
	if u.Role != domain.RoleCoach {
		return nil, domain.ErrForbidden
	}
	n, e := s.gyms.CountByCoach(ctx, coachID)
	if e != nil {
		return nil, e
	}
	if n >= maxGymsPerCoach {
		return nil, domain.ErrGymLimit
	}
	if e = validateGym(c); e != nil {
		return nil, e
	}
	c.ID = uuid.NewString()
	c.CoachID = coachID
	c.CreatedAt = s.now()
	c.UpdatedAt = s.now()
	return c, s.gyms.Create(ctx, c)
}
func (s *GymService) Delete(ctx context.Context, id string, actor *domain.User) error {
	c, e := s.gyms.GetByID(ctx, id)
	if e != nil {
		return e
	}
	if actor.Role != domain.RoleAdmin && c.CoachID != actor.ID {
		return domain.ErrForbidden
	}
	return s.gyms.Delete(ctx, id)
}
func (s *GymService) AddAthlete(ctx context.Context, gymID, athleteID string, actor *domain.User) error {
	c, e := s.gyms.GetByID(ctx, gymID)
	if e != nil {
		return e
	}
	if actor.Role != domain.RoleAdmin && c.CoachID != actor.ID && actor.ID != athleteID {
		return domain.ErrForbidden
	}
	u, e := s.users.GetByID(ctx, athleteID)
	if e != nil {
		return e
	}
	if u.Role != domain.RoleAthlete {
		return domain.ErrInvalid
	}
	return s.gyms.AddAthlete(ctx, &domain.GymAthlete{GymID: gymID, AthleteID: athleteID, RegisteredAt: s.now()})
}
func validateGym(c *domain.Gym) error {
	if strings.TrimSpace(c.Name) == "" || c.Province == "" || c.City == "" || c.Address == "" || c.Sport == "" || c.SportStyle == "" || len(c.ActiveWeekdays) == 0 || c.StartTime == "" || c.EndTime == "" || c.TuitionAmount < 0 {
		return domain.ErrInvalid
	}
	seen := map[int]bool{}
	for _, d := range c.ActiveWeekdays {
		if d < 0 || d > 6 || seen[d] {
			return domain.ErrInvalid
		}
		seen[d] = true
	}
	if c.TuitionMode != domain.TuitionMonthly && c.TuitionMode != domain.TuitionPerSession && c.TuitionMode != domain.TuitionCourse {
		return domain.ErrInvalid
	}
	return nil
}
