package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"regexp"
	"strings"
	"time"
)

type OTPNotifier interface {
	Send(context.Context, string, string) error
}
type UserService struct {
	users  repository.UserRepository
	otps   repository.OTPRepository
	notify OTPNotifier
	pepper string
	now    func() time.Time
}

func NewUserService(u repository.UserRepository, o repository.OTPRepository, n OTPNotifier, pepper string) *UserService {
	return &UserService{u, o, n, pepper, time.Now}
}

type StartRegistration struct {
	Phone string
	Role  domain.Role
}

func (s *UserService) StartRegistration(ctx context.Context, in StartRegistration) error {
	in.Phone = normalizePhone(in.Phone)
	if !validPhone(in.Phone) || !in.Role.ValidForRegistration() {
		return domain.ErrInvalid
	}
	if _, e := s.users.GetByPhone(ctx, in.Phone); e == nil {
		return domain.ErrConflict
	} else if !errors.Is(e, domain.ErrNotFound) {
		return e
	}
	code, e := generateOTP()
	if e != nil {
		return e
	}
	id := uuid.NewString()
	ch := &domain.OTPChallenge{ID: id, Phone: in.Phone, CodeHash: otpHash(id, code, s.pepper), Purpose: domain.OTPRegister, Role: &in.Role, ExpiresAt: s.now().Add(2 * time.Minute), CreatedAt: s.now()}
	if e = s.otps.Create(ctx, ch); e != nil {
		return e
	}
	return s.notify.Send(ctx, in.Phone, code)
}

type CompleteRegistration struct {
	Phone, Code string
	User        domain.User
}

func (s *UserService) CompleteRegistration(ctx context.Context, in CompleteRegistration) (*domain.User, error) {
	in.Phone = normalizePhone(in.Phone)
	ch, e := s.verifyOTP(ctx, in.Phone, in.Code, domain.OTPRegister)
	if e != nil {
		return nil, e
	}
	in.User.ID = uuid.NewString()
	in.User.Phone = &in.Phone
	in.User.PhoneVerified = true
	in.User.Role = *ch.Role
	in.User.CreatedAt = s.now()
	in.User.UpdatedAt = s.now()
	if e = validateUser(&in.User); e != nil {
		return nil, e
	}
	if e = s.users.Create(ctx, &in.User); e != nil {
		return nil, e
	}
	if e = s.otps.Consume(ctx, ch.ID, s.now()); e != nil {
		return nil, e
	}
	return &in.User, nil
}
func (s *UserService) StartPhoneChange(ctx context.Context, userID, phone string) error {
	phone = normalizePhone(phone)
	if !validPhone(phone) {
		return domain.ErrInvalid
	}
	if _, e := s.users.GetByPhone(ctx, phone); e == nil {
		return domain.ErrConflict
	} else if !errors.Is(e, domain.ErrNotFound) {
		return e
	}
	code, e := generateOTP()
	if e != nil {
		return e
	}
	id := uuid.NewString()
	ch := &domain.OTPChallenge{ID: id, Phone: phone, CodeHash: otpHash(id, code, s.pepper), Purpose: domain.OTPChangePhone, UserID: &userID, ExpiresAt: s.now().Add(2 * time.Minute), CreatedAt: s.now()}
	if e = s.otps.Create(ctx, ch); e != nil {
		return e
	}
	return s.notify.Send(ctx, phone, code)
}
func (s *UserService) CompletePhoneChange(ctx context.Context, userID, phone, code string) error {
	phone = normalizePhone(phone)
	ch, e := s.verifyOTP(ctx, phone, code, domain.OTPChangePhone)
	if e != nil {
		return e
	}
	if ch.UserID == nil || *ch.UserID != userID {
		return domain.ErrInvalidOTP
	}
	if e = s.users.UpdatePhone(ctx, userID, phone); e != nil {
		return e
	}
	return s.otps.Consume(ctx, ch.ID, s.now())
}
func (s *UserService) verifyOTP(ctx context.Context, phone, code string, p domain.OTPPurpose) (*domain.OTPChallenge, error) {
	ch, e := s.otps.Latest(ctx, phone, p)
	if e != nil {
		return nil, domain.ErrInvalidOTP
	}
	if ch.ConsumedAt != nil || ch.Attempts >= 5 || s.now().After(ch.ExpiresAt) {
		return nil, domain.ErrInvalidOTP
	}
	if !secureEqual(ch.CodeHash, otpHash(ch.ID, code, s.pepper)) {
		_ = s.otps.IncrementAttempts(ctx, ch.ID)
		return nil, domain.ErrInvalidOTP
	}
	return ch, nil
}
func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}
func (s *UserService) UpdateProfile(ctx context.Context, id string, in *domain.User) (*domain.User, error) {
	current, e := s.users.GetByID(ctx, id)
	if e != nil {
		return nil, e
	}
	in.ID, in.Role, in.Phone, in.PhoneVerified = current.ID, current.Role, current.Phone, current.PhoneVerified
	in.Username, in.PasswordHash, in.NationalID, in.CreatedAt = current.Username, current.PasswordHash, current.NationalID, current.CreatedAt
	if in.Role != domain.RoleAdmin {
		if e = validateUser(in); e != nil {
			return nil, e
		}
	}
	if e = s.users.Update(ctx, in); e != nil {
		return nil, e
	}
	return s.users.GetByID(ctx, id)
}
func (s *UserService) List(ctx context.Context, f repository.UserFilter) ([]domain.User, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.users.List(ctx, f)
}
func (s *UserService) Delete(ctx context.Context, id, actorID string) error {
	if id == actorID {
		return fmt.Errorf("%w: cannot delete current admin", domain.ErrInvalid)
	}
	return s.users.Delete(ctx, id)
}
func (s *UserService) CreateAdmin(ctx context.Context, name, username, password string) (*domain.User, error) {
	if strings.TrimSpace(name) == "" || len(password) < 12 {
		return nil, domain.ErrInvalid
	}
	h, e := hashPassword(password)
	if e != nil {
		return nil, e
	}
	parts := strings.Fields(name)
	u := &domain.User{ID: uuid.NewString(), Role: domain.RoleAdmin, FirstName: parts[0], Username: &username, PasswordHash: &h, CreatedAt: s.now(), UpdatedAt: s.now()}
	if len(parts) > 1 {
		u.LastName = strings.Join(parts[1:], " ")
	}
	return u, s.users.Create(ctx, u)
}
func (s *UserService) AuthenticateAdmin(ctx context.Context, username, password string) (*domain.User, error) {
	u, e := s.users.GetByUsername(ctx, username)
	if e != nil || u.Role != domain.RoleAdmin || u.PasswordHash == nil || !comparePassword(*u.PasswordHash, password) {
		return nil, domain.ErrForbidden
	}
	return u, nil
}
func normalizePhone(p string) string {
	p = strings.ReplaceAll(strings.TrimSpace(p), " ", "")
	if strings.HasPrefix(p, "09") {
		return "+98" + p[1:]
	}
	return p
}

var phoneRE = regexp.MustCompile(`^\+989\d{9}$`)

func validPhone(p string) bool { return phoneRE.MatchString(p) }
func validateUser(u *domain.User) error {
	if strings.TrimSpace(u.FirstName) == "" || strings.TrimSpace(u.LastName) == "" || len(u.NationalID) != 10 || u.BirthDate.IsZero() || u.Education == "" || u.Gender == "" || u.FaceImageID == "" {
		return domain.ErrInvalid
	}
	if u.Role == domain.RoleAthlete && (u.HasMedicalHistory == nil || u.BloodType == nil || u.EmergencyPhone == nil || u.EmergencyRelation == nil) {
		return domain.ErrInvalid
	}
	return nil
}
