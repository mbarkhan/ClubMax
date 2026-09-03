package repository

import (
	"context"
	"github.com/clubmax/clubmax/internal/domain"
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
	PromoteToAdmin(context.Context, string, string, string) error
	GetByID(context.Context, string) (*domain.User, error)
	GetByPhone(context.Context, string) (*domain.User, error)
	GetByNationalID(context.Context, string) (*domain.User, error)
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
	Update(context.Context, *domain.Gym) error
	GetByID(context.Context, string) (*domain.Gym, error)
	ListByCoach(context.Context, string) ([]domain.Gym, error)
	ListForAthlete(context.Context, string) ([]domain.Gym, error)
	ListCatalog(context.Context, string, int, int) ([]domain.GymCatalogItem, error)
	GetCatalogItem(context.Context, string, string, int, int) (*domain.GymCatalogItem, error)
	ListShowcase(context.Context) ([]domain.GymShowcaseItem, error)
	CountByCoach(context.Context, string) (int, error)
	Delete(context.Context, string) error
	AddAthlete(context.Context, *domain.GymAthlete) error
	ListGymAthletes(context.Context, string, int, int) ([]domain.GymAthleteProfile, error)
	RemoveAthlete(context.Context, string, string) error
	AddGymImage(context.Context, *domain.GymImage) error
	GetGymImage(context.Context, string) (*domain.GymImage, error)
	DeleteGymImage(context.Context, string) error
	CountGymImages(context.Context, string) (int, error)
	ListGymImages(context.Context, string) ([]domain.GymImage, error)
}
type DocumentRepository interface {
	AddUserDocument(context.Context, *domain.UserDocument) error
	CountUserDocuments(context.Context, string, domain.DocumentKind) (int, error)
}

type MessageRepository interface {
	Create(context.Context, *domain.Message, []domain.MessageRecipient) error
	ListForUser(context.Context, string, int, int) ([]domain.Message, error)
	MarkRead(context.Context, string, string) error
	AllUserIDs(context.Context, string) ([]string, error)
	GymRecipientIDs(context.Context, string, string) ([]string, error)
	CanAccessGym(context.Context, string, string) (bool, error)
	ShareGym(context.Context, string, string) (bool, error)
}
type WalletRepository interface {
	GetPolicy(context.Context) (domain.WithdrawalPolicy, error)
	SetPolicy(context.Context, domain.WithdrawalPolicy) error
	Balance(context.Context, string) (int64, error)
	CreateWithdrawal(context.Context, *domain.WithdrawalRequest) error
	ListWithdrawals(context.Context, *domain.WithdrawalStatus, int, int) ([]domain.WithdrawalRequest, error)
	ListCoachWithdrawals(context.Context, string, int, int) ([]domain.WithdrawalRequest, error)
	ReviewWithdrawal(context.Context, string, string, domain.WithdrawalStatus, *string, *string, *time.Time) error
	CompleteDueWithdrawals(context.Context, time.Time) (int64, error)
	ListLedger(context.Context, string, int, int) ([]domain.WalletEntry, error)
	GetRegistrationFee(context.Context) (domain.RegistrationFeePolicy, error)
	PayTuition(context.Context, *domain.TuitionPayment) error
	PayMonthlyTuition(context.Context, string, string, int, int, string, time.Time) (*domain.MonthlyTuitionPayment, error)
	ListAthletePayments(context.Context, string) ([]domain.AthletePayment, error)
	ListGymAthletePayments(context.Context, string, string) ([]domain.AthletePayment, error)
}

type AdminRepository interface {
	AdminOverview(context.Context) (domain.AdminOverview, error)
	ListAdminGyms(context.Context, string, int, int) ([]domain.AdminGym, error)
	ListAdminPayments(context.Context, string, int, int) ([]domain.AdminPayment, error)
}
