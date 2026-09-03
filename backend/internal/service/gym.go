package service

import (
	"context"
	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/reference"
	"github.com/clubmax/clubmax/internal/repository"
	"github.com/google/uuid"
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
func (s *GymService) Update(ctx context.Context, id string, input *domain.Gym, actor *domain.User) (*domain.Gym, error) {
	current, err := s.gyms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin && current.CoachID != actor.ID {
		return nil, domain.ErrForbidden
	}
	if err = validateGym(input); err != nil {
		return nil, err
	}
	input.ID, input.CoachID, input.CreatedAt, input.UpdatedAt = current.ID, current.CoachID, current.CreatedAt, s.now()
	if err = s.gyms.Update(ctx, input); err != nil {
		return nil, err
	}
	return s.gyms.GetByID(ctx, id)
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
func (s *GymService) List(ctx context.Context, actor *domain.User) ([]domain.Gym, error) {
	switch actor.Role {
	case domain.RoleCoach:
		return s.gyms.ListByCoach(ctx, actor.ID)
	case domain.RoleAthlete:
		return s.gyms.ListForAthlete(ctx, actor.ID)
	default:
		return nil, domain.ErrForbidden
	}
}
func (s *GymService) Catalog(ctx context.Context, actor *domain.User) ([]domain.GymCatalogItem, error) {
	if actor.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	year, month := persianYearMonth(s.now())
	return s.gyms.ListCatalog(ctx, actor.ID, year, month)
}
func (s *GymService) CatalogItem(ctx context.Context, id string, actor *domain.User) (*domain.GymCatalogItem, error) {
	if actor.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	year, month := persianYearMonth(s.now())
	return s.gyms.GetCatalogItem(ctx, id, actor.ID, year, month)
}
func (s *GymService) Showcase(ctx context.Context) ([]domain.GymShowcaseItem, error) {
	return s.gyms.ListShowcase(ctx)
}
func (s *GymService) Get(ctx context.Context, id string, actor *domain.User) (*domain.Gym, error) {
	gym, err := s.gyms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Role == domain.RoleAdmin || gym.CoachID == actor.ID {
		return gym, nil
	}
	if actor.Role == domain.RoleAthlete {
		gyms, listErr := s.gyms.ListForAthlete(ctx, actor.ID)
		if listErr != nil {
			return nil, listErr
		}
		for _, item := range gyms {
			if item.ID == id {
				return gym, nil
			}
		}
	}
	return nil, domain.ErrForbidden
}
func (s *GymService) AddImage(ctx context.Context, gymID string, actor *domain.User, storageKey, mimeType string, size int64) (*domain.GymImage, error) {
	gym, err := s.gyms.GetByID(ctx, gymID)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin && gym.CoachID != actor.ID {
		return nil, domain.ErrForbidden
	}
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return nil, domain.ErrInvalid
	}
	if size < 1 || size > 500*1024 {
		return nil, domain.ErrInvalid
	}
	count, err := s.gyms.CountGymImages(ctx, gymID)
	if err != nil {
		return nil, err
	}
	if count >= 3 {
		return nil, domain.ErrGymImageLimit
	}
	image := &domain.GymImage{ID: uuid.NewString(), GymID: gymID, StorageKey: storageKey, MIMEType: mimeType, Size: size, CreatedAt: s.now()}
	return image, s.gyms.AddGymImage(ctx, image)
}
func (s *GymService) Images(ctx context.Context, gymID string, actor *domain.User) ([]domain.GymImage, error) {
	gym, err := s.gyms.GetByID(ctx, gymID)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAthlete && actor.Role != domain.RoleAdmin && gym.CoachID != actor.ID {
		return nil, domain.ErrForbidden
	}
	return s.gyms.ListGymImages(ctx, gymID)
}
func (s *GymService) DeleteImage(ctx context.Context, gymID, imageID string, actor *domain.User) (*domain.GymImage, error) {
	gym, err := s.gyms.GetByID(ctx, gymID)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin && gym.CoachID != actor.ID {
		return nil, domain.ErrForbidden
	}
	image, err := s.gyms.GetGymImage(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.GymID != gymID {
		return nil, domain.ErrNotFound
	}
	if err = s.gyms.DeleteGymImage(ctx, imageID); err != nil {
		return nil, err
	}
	return image, nil
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
	now := s.now()
	year, month := persianYearMonth(now)
	return s.gyms.AddAthlete(ctx, &domain.GymAthlete{GymID: gymID, AthleteID: athleteID, BillingStartYear: year, BillingStartMonth: month, RegisteredAt: now})
}
func (s *GymService) Athletes(ctx context.Context, gymID string, actor *domain.User) ([]domain.GymAthleteProfile, error) {
	gym, e := s.gyms.GetByID(ctx, gymID)
	if e != nil {
		return nil, e
	}
	if actor.Role != domain.RoleAdmin && gym.CoachID != actor.ID {
		return nil, domain.ErrForbidden
	}
	year, month := persianYearMonth(s.now())
	return s.gyms.ListGymAthletes(ctx, gymID, year, month)
}
func validateGym(c *domain.Gym) error {
	if strings.TrimSpace(c.Name) == "" || c.Province == "" || c.City == "" || c.Address == "" || c.Sport == "" || c.SportStyle == "" || len(c.ActiveWeekdays) == 0 || c.StartTime == "" || c.EndTime == "" || c.TuitionAmount < 0 || c.Capacity < 0 {
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
	if !reference.ValidAddress(c.Province, c.City) {
		return domain.ErrInvalid
	}
	start, startErr := time.Parse("15:04", c.StartTime)
	end, endErr := time.Parse("15:04", c.EndTime)
	if startErr != nil || endErr != nil || !start.Before(end) {
		return domain.ErrInvalid
	}
	c.Name, c.Province, c.City, c.Address = strings.TrimSpace(c.Name), strings.TrimSpace(c.Province), strings.TrimSpace(c.City), strings.TrimSpace(c.Address)
	c.Sport, c.SportStyle = strings.TrimSpace(c.Sport), strings.TrimSpace(c.SportStyle)
	return nil
}
