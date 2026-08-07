package repository

import (
	"context"
	"github.com/gymax/gymax/internal/domain"
	"time"
)

type UserFilter struct {
	Role          *domain.Role
	Search        string
	Limit, Offset int
}
type UserRepository interface {
	Create(context.Context, *domain.User) error
	Update(context.Context, *domain.User) error
	UpdatePhone(context.Context, string, string) error
	GetByID(context.Context, string) (*domain.User, error)
	GetByPhone(context.Context, string) (*domain.User, error)
	GetByUsername(context.Context, string) (*domain.User, error)
	List(context.Context, UserFilter) ([]domain.User, error)
	Delete(context.Context, string) error
}
type OTPRepository interface {
	Create(context.Context, *domain.OTPChallenge) error
	Latest(context.Context, string, domain.OTPPurpose) (*domain.OTPChallenge, error)
	Consume(context.Context, string, time.Time) error
	IncrementAttempts(context.Context, string) error
}
type GymRepository interface {
	Create(context.Context, *domain.Gym) error
	GetByID(context.Context, string) (*domain.Gym, error)
	ListByCoach(context.Context, string) ([]domain.Gym, error)
	CountByCoach(context.Context, string) (int, error)
	Delete(context.Context, string) error
	AddAthlete(context.Context, *domain.GymAthlete) error
	RemoveAthlete(context.Context, string, string) error
}
type DocumentRepository interface {
	AddUserDocument(context.Context, *domain.UserDocument) error
	CountUserDocuments(context.Context, string, domain.DocumentKind) (int, error)
	AddGymImage(context.Context, *domain.GymImage) error
	CountGymImages(context.Context, string) (int, error)
}
