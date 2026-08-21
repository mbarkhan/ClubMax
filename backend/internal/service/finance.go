package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"strings"
	"time"
)

const defaultWithdrawalFeeBPS int64 = 200 // 2.00%

type FinanceService struct {
	wallet repository.WalletRepository
	users  repository.UserRepository
	gyms   repository.GymRepository
	now    func() time.Time
}

func NewFinanceService(w repository.WalletRepository, u repository.UserRepository, g repository.GymRepository) *FinanceService {
	return &FinanceService{wallet: w, users: u, gyms: g, now: time.Now}
}

func (s *FinanceService) RequestWithdrawal(ctx context.Context, coachID string, gross int64) (*domain.WithdrawalRequest, error) {
	if gross <= 0 {
		return nil, domain.ErrInvalid
	}
	coach, e := s.users.GetByID(ctx, coachID)
	if e != nil {
		return nil, e
	}
	if coach.Role != domain.RoleCoach || coach.BankIBAN == nil {
		return nil, domain.ErrInvalid
	}
	iban, ok := NormalizeIranianIBAN(*coach.BankIBAN)
	if !ok {
		return nil, domain.ErrInvalid
	}
	policy, e := s.wallet.GetPolicy(ctx)
	if e != nil {
		return nil, e
	}
	if policy.FeeBPS == 0 {
		policy.FeeBPS = defaultWithdrawalFeeBPS
	}
	if gross < policy.MinAmount || gross > policy.MaxAmount {
		return nil, domain.ErrWithdrawalLimit
	}
	balance, e := s.wallet.Balance(ctx, coachID)
	if e != nil {
		return nil, e
	}
	if gross > balance {
		return nil, domain.ErrInsufficientFunds
	}
	fee := (gross * policy.FeeBPS) / 10000
	r := &domain.WithdrawalRequest{ID: uuid.NewString(), CoachID: coachID, GrossAmount: gross, FeeAmount: fee, NetAmount: gross - fee, IBAN: iban, Status: domain.WithdrawalUnderReview, CreatedAt: s.now(), UpdatedAt: s.now()}
	if r.NetAmount <= 0 {
		return nil, domain.ErrInvalid
	}
	if e = s.wallet.CreateWithdrawal(ctx, r); e != nil {
		return nil, e
	}
	return r, nil
}
func (s *FinanceService) ReviewWithdrawal(ctx context.Context, adminID, id string, status domain.WithdrawalStatus, tracking, reason *string) error {
	admin, e := s.users.GetByID(ctx, adminID)
	if e != nil {
		return e
	}
	if admin.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if status != domain.WithdrawalQueued && status != domain.WithdrawalRejected {
		return domain.ErrInvalidState
	}
	if status == domain.WithdrawalQueued && (tracking == nil || strings.TrimSpace(*tracking) == "") {
		return domain.ErrInvalid
	}
	if status == domain.WithdrawalRejected && (reason == nil || strings.TrimSpace(*reason) == "") {
		return domain.ErrInvalid
	}
	var scheduledAt *time.Time
	if status == domain.WithdrawalQueued {
		clean := strings.TrimSpace(*tracking)
		tracking = &clean
		next := nextTehranSettlement(s.now())
		scheduledAt = &next
	} else {
		clean := strings.TrimSpace(*reason)
		reason = &clean
	}
	return s.wallet.ReviewWithdrawal(ctx, id, adminID, status, tracking, reason, scheduledAt)
}

func nextTehranSettlement(now time.Time) time.Time {
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		location = time.FixedZone("Asia/Tehran", 3*60*60+30*60)
	}
	local := now.In(location)
	cutoff := time.Date(local.Year(), local.Month(), local.Day(), 23, 30, 0, 0, location)
	if !local.Before(cutoff) {
		cutoff = cutoff.AddDate(0, 0, 1)
	}
	return cutoff
}

func (s *FinanceService) CompleteDueWithdrawals(ctx context.Context) (int64, error) {
	return s.wallet.CompleteDueWithdrawals(ctx, s.now())
}
func (s *FinanceService) SetPolicy(ctx context.Context, adminID string, p domain.WithdrawalPolicy) error {
	u, e := s.users.GetByID(ctx, adminID)
	if e != nil {
		return e
	}
	if u.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if p.MinAmount <= 0 || p.MaxAmount < p.MinAmount || p.FeeBPS < 0 || p.FeeBPS > 10000 {
		return domain.ErrInvalid
	}
	return s.wallet.SetPolicy(ctx, p)
}
func (s *FinanceService) Policy(ctx context.Context) (domain.WithdrawalPolicy, error) {
	p, e := s.wallet.GetPolicy(ctx)
	if errors.Is(e, domain.ErrNotFound) {
		return domain.WithdrawalPolicy{MinAmount: 1, MaxAmount: 1 << 62, FeeBPS: defaultWithdrawalFeeBPS}, nil
	}
	return p, e
}
func (s *FinanceService) Balance(ctx context.Context, coach string) (int64, error) {
	return s.wallet.Balance(ctx, coach)
}
func (s *FinanceService) Ledger(ctx context.Context, coach string, limit, offset int) ([]domain.WalletEntry, error) {
	return s.wallet.ListLedger(ctx, coach, limit, offset)
}
func (s *FinanceService) RegistrationFee(ctx context.Context) (domain.RegistrationFeePolicy, error) {
	return s.wallet.GetRegistrationFee(ctx)
}
func (s *FinanceService) PayTuition(ctx context.Context, athleteID, gymID string) (*domain.TuitionPayment, error) {
	u, e := s.users.GetByID(ctx, athleteID)
	if e != nil {
		return nil, e
	}
	if u.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	now := s.now()
	year, month := persianYearMonth(now)
	p := &domain.TuitionPayment{ID: uuid.NewString(), AthleteID: athleteID, GymID: gymID, Status: "succeeded", TrackingID: "TEST-" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:16]), BillingYear: year, BillingMonth: month, PaymentKind: "registration", CreatedAt: now}
	if e = s.wallet.PayTuition(ctx, p); e != nil {
		return nil, e
	}
	return p, nil
}
func (s *FinanceService) PayMonthlyTuition(ctx context.Context, athleteID, gymID string) (*domain.MonthlyTuitionPayment, error) {
	u, e := s.users.GetByID(ctx, athleteID)
	if e != nil {
		return nil, e
	}
	if u.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	now := s.now()
	year, month := persianYearMonth(now)
	tracking := "TEST-" + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:16])
	return s.wallet.PayMonthlyTuition(ctx, athleteID, gymID, year, month, tracking, now)
}
func (s *FinanceService) AthletePayments(ctx context.Context, athleteID string) ([]domain.AthletePayment, error) {
	u, e := s.users.GetByID(ctx, athleteID)
	if e != nil {
		return nil, e
	}
	if u.Role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	return s.wallet.ListAthletePayments(ctx, athleteID)
}
func (s *FinanceService) GymAthletePayments(ctx context.Context, coachID, gymID, athleteID string) ([]domain.AthletePayment, error) {
	gym, e := s.gyms.GetByID(ctx, gymID)
	if e != nil {
		return nil, e
	}
	if gym.CoachID != coachID {
		return nil, domain.ErrForbidden
	}
	athlete, e := s.users.GetByID(ctx, athleteID)
	if e != nil {
		return nil, e
	}
	if athlete.Role != domain.RoleAthlete {
		return nil, domain.ErrInvalid
	}
	return s.wallet.ListGymAthletePayments(ctx, gymID, athleteID)
}
func (s *FinanceService) Withdrawals(ctx context.Context, actor *domain.User, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if actor.Role == domain.RoleAdmin {
		return s.wallet.ListWithdrawals(ctx, nil, limit, offset)
	}
	if actor.Role == domain.RoleCoach {
		return s.wallet.ListCoachWithdrawals(ctx, actor.ID, limit, offset)
	}
	return nil, domain.ErrForbidden
}
