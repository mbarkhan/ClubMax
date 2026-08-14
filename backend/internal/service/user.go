package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/reference"
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
	docs   repository.DocumentRepository
	notify OTPNotifier
	pepper string
	now    func() time.Time
}

func NewUserService(u repository.UserRepository, o repository.OTPRepository, d repository.DocumentRepository, n OTPNotifier, pepper string) *UserService {
	return &UserService{users: u, otps: o, docs: d, notify: n, pepper: pepper, now: time.Now}
}

func (s *UserService) AddDocument(ctx context.Context, userID string, kind domain.DocumentKind, storageKey, mimeType string, size int64) (*domain.UserDocument, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	limits := map[domain.DocumentKind]int{domain.DocumentCoaching: 5, domain.DocumentInsurance: 1, domain.DocumentOther: 10, domain.DocumentProfile: 1}
	limit, ok := limits[kind]
	if !ok || size < 1 || size > 500*1024 {
		return nil, domain.ErrInvalid
	}
	if kind == domain.DocumentCoaching && u.Role != domain.RoleCoach {
		return nil, domain.ErrForbidden
	}
	count, err := s.docs.CountUserDocuments(ctx, userID, kind)
	if err != nil {
		return nil, err
	}
	if count >= limit {
		return nil, domain.ErrConflict
	}
	doc := &domain.UserDocument{ID: uuid.NewString(), UserID: userID, Kind: kind, StorageKey: storageKey, MIMEType: mimeType, Size: size, CreatedAt: s.now()}
	return doc, s.docs.AddUserDocument(ctx, doc)
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

type RegisterUser struct {
	Phone    string      `json:"phone"`
	Password string      `json:"password"`
	Role     domain.Role `json:"role"`
	domain.User
}

func (s *UserService) Register(ctx context.Context, in RegisterUser) (*domain.User, error) {
	in.Phone = normalizePhone(in.Phone)
	if !validPhone(in.Phone) {
		return nil, domain.InvalidField("phone")
	}
	if !in.Role.ValidForRegistration() {
		return nil, domain.InvalidField("role")
	}
	if len(in.Password) < 8 {
		return nil, domain.InvalidField("password")
	}
	if _, err := s.users.GetByPhone(ctx, in.Phone); err == nil {
		return nil, domain.ErrConflict
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u := in.User
	u.ID, u.Role, u.Phone, u.PhoneVerified, u.Active = uuid.NewString(), in.Role, &in.Phone, true, true
	u.PasswordHash, u.CreatedAt, u.UpdatedAt = &hash, s.now(), s.now()
	if err = normalizeUserIBAN(&u); err != nil {
		return nil, err
	}
	if err = validateUser(&u); err != nil {
		return nil, err
	}
	if err = s.users.Create(ctx, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserService) AuthenticateMember(ctx context.Context, phone, password string) (*domain.User, error) {
	phone = normalizePhone(phone)
	u, err := s.users.GetByPhone(ctx, phone)
	if err != nil || !u.Active || !u.Role.ValidForRegistration() || u.PasswordHash == nil || !comparePassword(*u.PasswordHash, password) {
		return nil, domain.ErrForbidden
	}
	return u, nil
}
func (s *UserService) Authenticate(ctx context.Context, identity, password string) (*domain.User, error) {
	identity = strings.TrimSpace(identity)
	if u, e := s.users.GetByNationalID(ctx, identity); e == nil {
		if u.Active && u.Role.ValidForRegistration() && u.PasswordHash != nil && comparePassword(*u.PasswordHash, password) {
			return u, nil
		}
		return nil, domain.ErrForbidden
	}
	u, e := s.users.GetByUsername(ctx, identity)
	if e != nil || !u.Active || u.Role != domain.RoleAdmin || u.PasswordHash == nil || !comparePassword(*u.PasswordHash, password) {
		return nil, domain.ErrForbidden
	}
	return u, nil
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
	in.User.Active = true
	in.User.Role = *ch.Role
	in.User.CreatedAt = s.now()
	in.User.UpdatedAt = s.now()
	if e = normalizeUserIBAN(&in.User); e != nil {
		return nil, e
	}
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
func (s *UserService) StartLogin(ctx context.Context, phone string) error {
	phone = normalizePhone(phone)
	u, e := s.users.GetByPhone(ctx, phone)
	if e != nil {
		return domain.ErrNotFound
	}
	if !u.PhoneVerified {
		return domain.ErrForbidden
	}
	code, e := generateOTP()
	if e != nil {
		return e
	}
	id := uuid.NewString()
	ch := &domain.OTPChallenge{ID: id, Phone: phone, CodeHash: otpHash(id, code, s.pepper), Purpose: domain.OTPLogin, UserID: &u.ID, ExpiresAt: s.now().Add(2 * time.Minute), CreatedAt: s.now()}
	if e = s.otps.Create(ctx, ch); e != nil {
		return e
	}
	return s.notify.Send(ctx, phone, code)
}
func (s *UserService) CompleteLogin(ctx context.Context, phone, code string) (*domain.User, error) {
	phone = normalizePhone(phone)
	ch, e := s.verifyOTP(ctx, phone, code, domain.OTPLogin)
	if e != nil {
		return nil, e
	}
	if ch.UserID == nil {
		return nil, domain.ErrInvalidOTP
	}
	u, e := s.users.GetByID(ctx, *ch.UserID)
	if e != nil {
		return nil, e
	}
	if e = s.otps.Consume(ctx, ch.ID, s.now()); e != nil {
		return nil, e
	}
	return u, nil
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
func (s *UserService) StartAccountDeletion(ctx context.Context, userID string) error {
	u, e := s.users.GetByID(ctx, userID)
	if e != nil {
		return e
	}
	if u.Role == domain.RoleAdmin || u.Phone == nil || !u.PhoneVerified {
		return domain.ErrForbidden
	}
	code, e := generateOTP()
	if e != nil {
		return e
	}
	id := uuid.NewString()
	ch := &domain.OTPChallenge{ID: id, Phone: *u.Phone, CodeHash: otpHash(id, code, s.pepper), Purpose: domain.OTPDelete, UserID: &userID, ExpiresAt: s.now().Add(2 * time.Minute), CreatedAt: s.now()}
	if e = s.otps.Create(ctx, ch); e != nil {
		return e
	}
	return s.notify.Send(ctx, *u.Phone, code)
}
func (s *UserService) CompleteAccountDeletion(ctx context.Context, userID, code string) error {
	u, e := s.users.GetByID(ctx, userID)
	if e != nil {
		return e
	}
	if u.Role == domain.RoleAdmin || u.Phone == nil {
		return domain.ErrForbidden
	}
	ch, e := s.verifyOTP(ctx, *u.Phone, code, domain.OTPDelete)
	if e != nil {
		return e
	}
	if ch.UserID == nil || *ch.UserID != userID {
		return domain.ErrInvalidOTP
	}
	return s.users.Delete(ctx, userID)
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
	if e = normalizeUserIBAN(in); e != nil {
		return nil, e
	}
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
	u := &domain.User{ID: uuid.NewString(), Role: domain.RoleAdmin, FirstName: parts[0], Active: true, Username: &username, PasswordHash: &h, CreatedAt: s.now(), UpdatedAt: s.now()}
	if len(parts) > 1 {
		u.LastName = strings.Join(parts[1:], " ")
	}
	return u, s.users.Create(ctx, u)
}
func (s *UserService) AuthenticateAdmin(ctx context.Context, username, password string) (*domain.User, error) {
	u, e := s.users.GetByUsername(ctx, username)
	if e != nil || !u.Active || u.Role != domain.RoleAdmin || u.PasswordHash == nil || !comparePassword(*u.PasswordHash, password) {
		return nil, domain.ErrForbidden
	}
	return u, nil
}
func (s *UserService) PromoteToAdmin(ctx context.Context, id, username, password string) error {
	if strings.TrimSpace(username) == "" || len(password) < 12 {
		return domain.ErrInvalid
	}
	if _, err := s.users.GetByID(ctx, id); err != nil {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	return s.users.PromoteToAdmin(ctx, id, strings.TrimSpace(username), hash)
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
	if strings.TrimSpace(u.FirstName) == "" {
		return domain.InvalidField("first_name")
	}
	if strings.TrimSpace(u.LastName) == "" {
		return domain.InvalidField("last_name")
	}
	if !validNationalID(u.NationalID) {
		return domain.InvalidField("national_id")
	}
	if u.BirthDate.IsZero() {
		return domain.InvalidField("birth_date")
	}
	if !validEducation(u.Education) {
		return domain.InvalidField("education")
	}
	if !validGender(u.Gender) {
		return domain.InvalidField("gender")
	}
	if !validMaritalStatus(u.MaritalStatus) {
		return domain.InvalidField("marital_status")
	}
	if (u.Province == nil) != (u.City == nil) || (u.Province != nil && !reference.ValidAddress(*u.Province, *u.City)) {
		return domain.InvalidField("city")
	}
	if u.Role == domain.RoleAthlete {
		if u.HasMedicalHistory == nil {
			return domain.InvalidField("has_medical_history")
		}
		if u.BloodType == nil {
			return domain.InvalidField("blood_type")
		}
		if u.EmergencyPhone == nil {
			return domain.InvalidField("emergency_phone")
		}
		if u.EmergencyRelation == nil {
			return domain.InvalidField("emergency_relation")
		}
	}
	return nil
}

func validEducation(v domain.Education) bool {
	return v == domain.EducationBelowDiploma || v == domain.EducationDiploma || v == domain.EducationAssociate || v == domain.EducationBachelor || v == domain.EducationMaster || v == domain.EducationDoctorate
}
func validGender(v domain.Gender) bool { return v == domain.GenderMale || v == domain.GenderFemale }
func validMaritalStatus(v domain.MaritalStatus) bool {
	return v == domain.MaritalUnknown || v == domain.MaritalSingle || v == domain.MaritalMarried
}
func validNationalID(v string) bool {
	if len(v) != 10 {
		return false
	}
	allSame := true
	for i := 0; i < 10; i++ {
		if v[i] < '0' || v[i] > '9' {
			return false
		}
		if i > 0 && v[i] != v[0] {
			allSame = false
		}
	}
	if allSame {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(v[i]-'0') * (10 - i)
	}
	r := sum % 11
	if r >= 2 {
		r = 11 - r
	}
	return r == int(v[9]-'0')
}

func normalizeUserIBAN(u *domain.User) error {
	if u.Role != domain.RoleCoach && u.Role != domain.RoleAthlete {
		if u.BankIBAN != nil {
			return domain.ErrInvalid
		}
		return nil
	}
	if u.BankIBAN == nil {
		return nil
	}
	iban, ok := NormalizeIranianIBAN(*u.BankIBAN)
	if !ok {
		return domain.InvalidField("bank_iban")
	}
	u.BankIBAN = &iban
	return nil
}

// NormalizeIranianIBAN validates the Iranian Sheba/IBAN checksum (ISO 13616).
// It accepts optional spaces and hyphens but always returns IR plus 24 digits.
func NormalizeIranianIBAN(value string) (string, bool) {
	value = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), " ", ""), "-", ""))
	if len(value) != 26 || !strings.HasPrefix(value, "IR") {
		return "", false
	}
	for _, r := range value[2:] {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	rearranged := value[4:] + value[:4]
	remainder := 0
	for _, r := range rearranged {
		if r >= '0' && r <= '9' {
			remainder = (remainder*10 + int(r-'0')) % 97
			continue
		}
		for _, digit := range []byte(fmt.Sprintf("%d", int(r-'A')+10)) {
			remainder = (remainder*10 + int(digit-'0')) % 97
		}
	}
	if remainder != 1 {
		return "", false
	}
	return value, true
}
