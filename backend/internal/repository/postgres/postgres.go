package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Store struct{ DB *pgxpool.Pool }

func New(ctx context.Context, url string) (*Store, error) {
	p, e := pgxpool.New(ctx, url)
	if e != nil {
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		p.Close()
		return nil, e
	}
	return &Store{p}, nil
}
func dbError(e error) error {
	if errors.Is(e, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pe *pgconn.PgError
	if errors.As(e, &pe) && pe.Code == "23505" {
		return domain.ErrConflict
	}
	if errors.As(e, &pe) && strings.Contains(pe.Message, "GYM_IMAGE_LIMIT_REACHED") {
		return domain.ErrGymImageLimit
	}
	if errors.As(e, &pe) && strings.Contains(pe.Message, "GYM_LIMIT_REACHED") {
		return domain.ErrGymLimit
	}
	return e
}

const userCols = `id,role,first_name,last_name,father_name,COALESCE(national_id,''),bank_iban,COALESCE(birth_date,DATE '0001-01-01'),province,city,address,COALESCE(education,''),field_of_study,COALESCE(gender,''),marital_status,description,phone,phone_verified,active,username,password_hash,has_medical_history,medical_history,blood_type,emergency_phone,emergency_relation,sports_insurance_expires_at,COALESCE(NULLIF(face_image_id,''),(SELECT storage_key FROM user_documents WHERE user_id=users.id AND kind='profile' ORDER BY created_at DESC LIMIT 1),''),created_at,updated_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	u := new(domain.User)
	e := row.Scan(&u.ID, &u.Role, &u.FirstName, &u.LastName, &u.FatherName, &u.NationalID, &u.BankIBAN, &u.BirthDate, &u.Province, &u.City, &u.Address, &u.Education, &u.FieldOfStudy, &u.Gender, &u.MaritalStatus, &u.Description, &u.Phone, &u.PhoneVerified, &u.Active, &u.Username, &u.PasswordHash, &u.HasMedicalHistory, &u.MedicalHistory, &u.BloodType, &u.EmergencyPhone, &u.EmergencyRelation, &u.SportsInsuranceExpiresAt, &u.FaceImageID, &u.CreatedAt, &u.UpdatedAt)
	return u, dbError(e)
}
func (s *Store) Create(ctx context.Context, u *domain.User) error {
	_, e := s.DB.Exec(ctx, `INSERT INTO users(id,role,first_name,last_name,father_name,national_id,bank_iban,birth_date,province,city,address,education,field_of_study,gender,marital_status,description,phone,phone_verified,active,username,password_hash,has_medical_history,medical_history,blood_type,emergency_phone,emergency_relation,sports_insurance_expires_at,face_image_id) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7,NULLIF($8::date,DATE '0001-01-01'),$9,$10,$11,NULLIF($12,''),$13,NULLIF($14,''),$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,NULLIF($28,''))`, u.ID, u.Role, u.FirstName, u.LastName, u.FatherName, u.NationalID, u.BankIBAN, u.BirthDate.Format("2006-01-02"), u.Province, u.City, u.Address, u.Education, u.FieldOfStudy, u.Gender, u.MaritalStatus, u.Description, u.Phone, u.PhoneVerified, u.Active, u.Username, u.PasswordHash, u.HasMedicalHistory, u.MedicalHistory, u.BloodType, u.EmergencyPhone, u.EmergencyRelation, u.SportsInsuranceExpiresAt, u.FaceImageID)
	return dbError(e)
}
func (s *Store) Update(ctx context.Context, u *domain.User) error {
	tag, e := s.DB.Exec(ctx, `UPDATE users SET first_name=$2,last_name=$3,father_name=$4,province=$5,city=$6,address=$7,education=$8,field_of_study=$9,gender=$10,marital_status=$11,description=$12,has_medical_history=$13,medical_history=$14,blood_type=$15,emergency_phone=$16,emergency_relation=$17,sports_insurance_expires_at=$18,face_image_id=$19,bank_iban=$20,updated_at=now() WHERE id=$1`, u.ID, u.FirstName, u.LastName, u.FatherName, u.Province, u.City, u.Address, u.Education, u.FieldOfStudy, u.Gender, u.MaritalStatus, u.Description, u.HasMedicalHistory, u.MedicalHistory, u.BloodType, u.EmergencyPhone, u.EmergencyRelation, u.SportsInsuranceExpiresAt, u.FaceImageID, u.BankIBAN)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) UpdatePhone(c context.Context, id, phone string) error {
	tag, e := s.DB.Exec(c, `UPDATE users SET phone=$2,phone_verified=true,updated_at=now() WHERE id=$1`, id, phone)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) PromoteToAdmin(c context.Context, id, username, passwordHash string) error {
	tag, e := s.DB.Exec(c, `UPDATE users SET role='admin',username=$2,password_hash=$3,phone=NULL,updated_at=now() WHERE id=$1`, id, username, passwordHash)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) GetByID(c context.Context, id string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE id=$1`, id))
}
func (s *Store) GetByPhone(c context.Context, v string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE phone=$1`, v))
}
func (s *Store) GetByNationalID(c context.Context, v string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE national_id=$1`, v))
}
func (s *Store) GetByUsername(c context.Context, v string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE username=$1`, v))
}
func (s *Store) List(c context.Context, f repository.UserFilter) ([]domain.User, error) {
	rows, e := s.DB.Query(c, `SELECT `+userCols+` FROM users WHERE ($1::user_role IS NULL OR role=$1) AND ($2='' OR first_name ILIKE '%'||$2||'%' OR last_name ILIKE '%'||$2||'%' OR phone LIKE '%'||$2||'%') ORDER BY created_at DESC LIMIT $3 OFFSET $4`, f.Role, f.Search, f.Limit, f.Offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.User{}
	for rows.Next() {
		u, e := scanUser(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}
func (s *Store) Delete(c context.Context, id string) error {
	tag, e := s.DB.Exec(c, `DELETE FROM users WHERE id=$1`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) Latest(c context.Context, p string, k domain.OTPPurpose) (*domain.OTPChallenge, error) {
	o := new(domain.OTPChallenge)
	e := s.DB.QueryRow(c, `SELECT id,phone,code_hash,purpose,role,user_id,expires_at,attempts,consumed_at,created_at FROM otp_challenges WHERE phone=$1 AND purpose=$2 ORDER BY created_at DESC LIMIT 1`, p, k).Scan(&o.ID, &o.Phone, &o.CodeHash, &o.Purpose, &o.Role, &o.UserID, &o.ExpiresAt, &o.Attempts, &o.ConsumedAt, &o.CreatedAt)
	return o, dbError(e)
}
func (s *Store) IncrementAttempts(c context.Context, id string) error {
	_, e := s.DB.Exec(c, `UPDATE otp_challenges SET attempts=attempts+1 WHERE id=$1`, id)
	return e
}
func (s *Store) CreateChallenge(c context.Context, o *domain.OTPChallenge) error {
	_, e := s.DB.Exec(c, `INSERT INTO otp_challenges(id,phone,code_hash,purpose,role,user_id,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, o.ID, o.Phone, o.CodeHash, o.Purpose, o.Role, o.UserID, o.ExpiresAt)
	return dbError(e)
}

func scanGym(r pgx.Row) (*domain.Gym, error) {
	v := new(domain.Gym)
	e := r.Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(e)
}

const gymCols = `id,coach_id,name,province,city,address,phone,sport,sport_style,active_weekdays,start_time::text,end_time::text,tuition_mode,tuition_amount,capacity,created_at,updated_at`
const gymCatalogCols = `g.id,g.coach_id,g.name,g.province,g.city,g.address,g.phone,g.sport,g.sport_style,g.active_weekdays,g.start_time::text,g.end_time::text,g.tuition_mode,g.tuition_amount,g.capacity,g.created_at,g.updated_at`

func (s *Store) CreateGym(c context.Context, v *domain.Gym) error {
	_, e := s.DB.Exec(c, `INSERT INTO gyms(id,coach_id,name,province,city,address,phone,sport,sport_style,active_weekdays,start_time,end_time,tuition_mode,tuition_amount,capacity) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, v.ID, v.CoachID, v.Name, v.Province, v.City, v.Address, v.Phone, v.Sport, v.SportStyle, v.ActiveWeekdays, v.StartTime, v.EndTime, v.TuitionMode, v.TuitionAmount, v.Capacity)
	return dbError(e)
}
func (s *Store) UpdateGym(c context.Context, v *domain.Gym) error {
	tag, e := s.DB.Exec(c, `UPDATE gyms SET name=$2,province=$3,city=$4,address=$5,phone=$6,sport=$7,sport_style=$8,active_weekdays=$9,start_time=$10,end_time=$11,tuition_mode=$12,tuition_amount=$13,capacity=$14,updated_at=now() WHERE id=$1`, v.ID, v.Name, v.Province, v.City, v.Address, v.Phone, v.Sport, v.SportStyle, v.ActiveWeekdays, v.StartTime, v.EndTime, v.TuitionMode, v.TuitionAmount, v.Capacity)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) GetGymByID(c context.Context, id string) (*domain.Gym, error) {
	return scanGym(s.DB.QueryRow(c, `SELECT `+gymCols+` FROM gyms WHERE id=$1`, id))
}
func (s *Store) ListByCoach(c context.Context, id string) ([]domain.Gym, error) {
	rows, e := s.DB.Query(c, `SELECT `+gymCols+` FROM gyms WHERE coach_id=$1 ORDER BY created_at DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []domain.Gym
	for rows.Next() {
		v, e := scanGym(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}
func (s *Store) ListForAthlete(c context.Context, id string) ([]domain.Gym, error) {
	rows, e := s.DB.Query(c, `SELECT `+gymCols+` FROM gyms WHERE id IN (SELECT gym_id FROM gym_athletes WHERE athlete_id=$1) ORDER BY created_at DESC`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Gym{}
	for rows.Next() {
		v, err := scanGym(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}
func (s *Store) ListCatalog(c context.Context, athleteID string, year, month int) ([]domain.GymCatalogItem, error) {
	rows, e := s.DB.Query(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),(ga.athlete_id IS NOT NULL),CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,CASE WHEN g.tuition_mode='monthly' AND ga.athlete_id IS NOT NULL THEN GREATEST((($2*12+$3)-(ga.billing_start_year*12+ga.billing_start_month)+1)-(SELECT count(*) FROM tuition_payments tp WHERE tp.gym_id=g.id AND tp.athlete_id=$1 AND tp.billing_year IS NOT NULL AND (tp.billing_year*12+tp.billing_month) BETWEEN (ga.billing_start_year*12+ga.billing_start_month) AND ($2*12+$3)),0) ELSE 0 END FROM gyms g JOIN users u ON u.id=g.coach_id LEFT JOIN gym_athletes ga ON ga.gym_id=g.id AND ga.athlete_id=$1 ORDER BY g.created_at DESC`, athleteID, year, month)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.GymCatalogItem{}
	for rows.Next() {
		var v domain.GymCatalogItem
		if e = rows.Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CoachName, &v.Joined, &v.CapacityRemaining, &v.UnpaidMonths); e != nil {
			return nil, e
		}
		v.OutstandingAmount = int64(v.UnpaidMonths) * v.TuitionAmount
		v.MonthlyPaymentDue = v.UnpaidMonths > 0
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetCatalogItem(c context.Context, gymID, athleteID string, year, month int) (*domain.GymCatalogItem, error) {
	v := new(domain.GymCatalogItem)
	e := s.DB.QueryRow(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),(ga.athlete_id IS NOT NULL),CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,CASE WHEN g.tuition_mode='monthly' AND ga.athlete_id IS NOT NULL THEN GREATEST((($3*12+$4)-(ga.billing_start_year*12+ga.billing_start_month)+1)-(SELECT count(*) FROM tuition_payments tp WHERE tp.gym_id=g.id AND tp.athlete_id=$2 AND tp.billing_year IS NOT NULL AND (tp.billing_year*12+tp.billing_month) BETWEEN (ga.billing_start_year*12+ga.billing_start_month) AND ($3*12+$4)),0) ELSE 0 END FROM gyms g JOIN users u ON u.id=g.coach_id LEFT JOIN gym_athletes ga ON ga.gym_id=g.id AND ga.athlete_id=$2 WHERE g.id=$1`, gymID, athleteID, year, month).Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CoachName, &v.Joined, &v.CapacityRemaining, &v.UnpaidMonths)
	v.OutstandingAmount = int64(v.UnpaidMonths) * v.TuitionAmount
	v.MonthlyPaymentDue = v.UnpaidMonths > 0
	return v, dbError(e)
}
func (s *Store) CountByCoach(c context.Context, id string) (int, error) {
	var n int
	e := s.DB.QueryRow(c, `SELECT count(*) FROM gyms WHERE coach_id=$1`, id).Scan(&n)
	return n, e
}
func (s *Store) DeleteGym(c context.Context, id string) error {
	tag, e := s.DB.Exec(c, `DELETE FROM gyms WHERE id=$1`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) AddAthlete(c context.Context, v *domain.GymAthlete) error {
	tx, e := s.DB.Begin(c)
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	var capacity, registered int64
	if e = tx.QueryRow(c, `SELECT capacity FROM gyms WHERE id=$1 FOR UPDATE`, v.GymID).Scan(&capacity); e != nil {
		return dbError(e)
	}
	if capacity > 0 {
		if e = tx.QueryRow(c, `SELECT count(*) FROM gym_athletes WHERE gym_id=$1`, v.GymID).Scan(&registered); e != nil {
			return e
		}
		if registered >= capacity {
			return domain.ErrGymFull
		}
	}
	if _, e = tx.Exec(c, `INSERT INTO gym_athletes(gym_id,athlete_id,billing_start_year,billing_start_month,registered_at) VALUES($1,$2,$3,$4,$5)`, v.GymID, v.AthleteID, v.BillingStartYear, v.BillingStartMonth, v.RegisteredAt); e != nil {
		return dbError(e)
	}
	return tx.Commit(c)
}
func (s *Store) ListGymAthletes(c context.Context, gymID string) ([]domain.GymAthleteProfile, error) {
	rows, e := s.DB.Query(c, `SELECT u.id,u.first_name,u.last_name,u.phone,ga.registered_at FROM gym_athletes ga JOIN users u ON u.id=ga.athlete_id WHERE ga.gym_id=$1 ORDER BY ga.registered_at DESC`, gymID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.GymAthleteProfile{}
	for rows.Next() {
		var v domain.GymAthleteProfile
		if e = rows.Scan(&v.ID, &v.FirstName, &v.LastName, &v.Phone, &v.RegisteredAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) RemoveAthlete(c context.Context, gym, user string) error {
	_, e := s.DB.Exec(c, `DELETE FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2`, gym, user)
	return e
}
func (s *Store) AddUserDocument(c context.Context, v *domain.UserDocument) error {
	_, e := s.DB.Exec(c, `INSERT INTO user_documents(id,user_id,kind,storage_key,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5,$6)`, v.ID, v.UserID, v.Kind, v.StorageKey, v.MIMEType, v.Size)
	return dbError(e)
}
func (s *Store) CountUserDocuments(c context.Context, user string, k domain.DocumentKind) (int, error) {
	var n int
	e := s.DB.QueryRow(c, `SELECT count(*) FROM user_documents WHERE user_id=$1 AND kind=$2`, user, k).Scan(&n)
	return n, e
}
func (s *Store) AddGymImage(c context.Context, v *domain.GymImage) error {
	_, e := s.DB.Exec(c, `INSERT INTO gym_images(id,gym_id,storage_key,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5)`, v.ID, v.GymID, v.StorageKey, v.MIMEType, v.Size)
	return dbError(e)
}
func (s *Store) GetGymImage(c context.Context, id string) (*domain.GymImage, error) {
	v := new(domain.GymImage)
	e := s.DB.QueryRow(c, `SELECT id,gym_id,storage_key,mime_type,size_bytes,created_at FROM gym_images WHERE id=$1`, id).Scan(&v.ID, &v.GymID, &v.StorageKey, &v.MIMEType, &v.Size, &v.CreatedAt)
	return v, dbError(e)
}
func (s *Store) DeleteGymImage(c context.Context, id string) error {
	tag, e := s.DB.Exec(c, `DELETE FROM gym_images WHERE id=$1`, id)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) CountGymImages(c context.Context, gym string) (int, error) {
	var n int
	e := s.DB.QueryRow(c, `SELECT count(*) FROM gym_images WHERE gym_id=$1`, gym).Scan(&n)
	return n, e
}
func (s *Store) ListGymImages(c context.Context, gym string) ([]domain.GymImage, error) {
	rows, e := s.DB.Query(c, `SELECT id,gym_id,storage_key,mime_type,size_bytes,created_at FROM gym_images WHERE gym_id=$1 ORDER BY created_at`, gym)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.GymImage{}
	for rows.Next() {
		var v domain.GymImage
		if e = rows.Scan(&v.ID, &v.GymID, &v.StorageKey, &v.MIMEType, &v.Size, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Adapters keep the single Store usable through small domain-specific interfaces.
type OTPStore struct{ *Store }

func (o OTPStore) Create(c context.Context, v *domain.OTPChallenge) error {
	return o.CreateChallenge(c, v)
}
func (o OTPStore) Consume(c context.Context, id string, t time.Time) error {
	_, e := o.DB.Exec(c, `UPDATE otp_challenges SET consumed_at=$2 WHERE id=$1 AND consumed_at IS NULL`, id, t)
	return e
}

type GymStore struct{ *Store }

func (o GymStore) Create(c context.Context, v *domain.Gym) error { return o.CreateGym(c, v) }
func (o GymStore) Update(c context.Context, v *domain.Gym) error { return o.UpdateGym(c, v) }
func (o GymStore) GetByID(c context.Context, id string) (*domain.Gym, error) {
	return o.GetGymByID(c, id)
}
func (o GymStore) Delete(c context.Context, id string) error { return o.DeleteGym(c, id) }

func (s *Store) GetPolicy(c context.Context) (domain.WithdrawalPolicy, error) {
	var p domain.WithdrawalPolicy
	e := s.DB.QueryRow(c, `SELECT min_amount,max_amount,fee_bps,updated_at FROM withdrawal_policy WHERE id=true`).Scan(&p.MinAmount, &p.MaxAmount, &p.FeeBPS, &p.UpdatedAt)
	return p, dbError(e)
}
func (s *Store) SetPolicy(c context.Context, p domain.WithdrawalPolicy) error {
	_, e := s.DB.Exec(c, `INSERT INTO withdrawal_policy(id,min_amount,max_amount,fee_bps,updated_at) VALUES(true,$1,$2,$3,now()) ON CONFLICT(id) DO UPDATE SET min_amount=EXCLUDED.min_amount,max_amount=EXCLUDED.max_amount,fee_bps=EXCLUDED.fee_bps,updated_at=now()`, p.MinAmount, p.MaxAmount, p.FeeBPS)
	return dbError(e)
}
func (s *Store) Balance(c context.Context, coach string) (int64, error) {
	var v int64
	e := s.DB.QueryRow(c, `SELECT COALESCE(sum(amount),0) FROM wallet_entries WHERE coach_id=$1`, coach).Scan(&v)
	return v, e
}
func (s *Store) GetRegistrationFee(c context.Context) (domain.RegistrationFeePolicy, error) {
	var p domain.RegistrationFeePolicy
	e := s.DB.QueryRow(c, `SELECT fee_amount,updated_at FROM registration_fee_policy WHERE id=true`).Scan(&p.FeeAmount, &p.UpdatedAt)
	return p, dbError(e)
}
func (s *Store) PayTuition(c context.Context, p *domain.TuitionPayment) error {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	var role domain.Role
	if e = tx.QueryRow(c, `SELECT role FROM users WHERE id=$1 FOR SHARE`, p.AthleteID).Scan(&role); e != nil {
		return dbError(e)
	}
	if role != domain.RoleAthlete {
		return domain.ErrForbidden
	}
	var capacity, registered int64
	var tuitionMode domain.TuitionMode
	if e = tx.QueryRow(c, `SELECT coach_id,tuition_amount,capacity,tuition_mode FROM gyms WHERE id=$1 FOR UPDATE`, p.GymID).Scan(&p.CoachID, &p.Tuition, &capacity, &tuitionMode); e != nil {
		return dbError(e)
	}
	if e = tx.QueryRow(c, `SELECT count(*) FROM gym_athletes WHERE gym_id=$1`, p.GymID).Scan(&registered); e != nil {
		return e
	}
	if capacity > 0 && registered >= capacity {
		return domain.ErrGymFull
	}
	if e = tx.QueryRow(c, `SELECT fee_amount FROM registration_fee_policy WHERE id=true FOR SHARE`).Scan(&p.PlatformFee); e != nil {
		return dbError(e)
	}
	p.Total = p.Tuition + p.PlatformFee
	var billingYear, billingMonth any
	if tuitionMode == domain.TuitionMonthly {
		billingYear, billingMonth = p.BillingYear, p.BillingMonth
	}
	if _, e = tx.Exec(c, `INSERT INTO tuition_payments(id,athlete_id,gym_id,coach_id,tuition_amount,platform_fee,total_amount,status,tracking_id,billing_year,billing_month,payment_kind,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, p.ID, p.AthleteID, p.GymID, p.CoachID, p.Tuition, p.PlatformFee, p.Total, p.Status, p.TrackingID, billingYear, billingMonth, p.PaymentKind, p.CreatedAt); e != nil {
		return dbError(e)
	}
	if _, e = tx.Exec(c, `INSERT INTO gym_athletes(gym_id,athlete_id,billing_start_year,billing_start_month,registered_at) VALUES($1,$2,$3,$4,$5)`, p.GymID, p.AthleteID, p.BillingYear, p.BillingMonth, p.CreatedAt); e != nil {
		return dbError(e)
	}
	if p.Tuition > 0 {
		if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'tuition_payment',$4,'athlete tuition payment')`, p.CoachID, domain.LedgerTuition, p.Tuition, p.ID); e != nil {
			return dbError(e)
		}
	}
	return tx.Commit(c)
}
func (s *Store) PayMonthlyTuition(c context.Context, athleteID, gymID string, currentYear, currentMonth int, tracking string, createdAt time.Time) (*domain.MonthlyTuitionPayment, error) {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(c)
	var coachID string
	var tuition int64
	var mode domain.TuitionMode
	if e = tx.QueryRow(c, `SELECT coach_id,tuition_amount,tuition_mode FROM gyms WHERE id=$1 FOR UPDATE`, gymID).Scan(&coachID, &tuition, &mode); e != nil {
		return nil, dbError(e)
	}
	if mode != domain.TuitionMonthly {
		return nil, domain.ErrInvalid
	}
	var startYear, startMonth int
	if e = tx.QueryRow(c, `SELECT billing_start_year,billing_start_month FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2 FOR UPDATE`, gymID, athleteID).Scan(&startYear, &startMonth); e != nil {
		return nil, dbError(e)
	}
	rows, e := tx.Query(c, `SELECT billing_year,billing_month FROM tuition_payments WHERE gym_id=$1 AND athlete_id=$2 AND billing_year IS NOT NULL`, gymID, athleteID)
	if e != nil {
		return nil, e
	}
	paid := map[int]bool{}
	for rows.Next() {
		var year, month int
		if e = rows.Scan(&year, &month); e != nil {
			rows.Close()
			return nil, e
		}
		paid[year*12+month] = true
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	start, end := startYear*12+startMonth, currentYear*12+currentMonth
	result := &domain.MonthlyTuitionPayment{}
	for serial := start; serial <= end; serial++ {
		if paid[serial] {
			continue
		}
		year, month := (serial-1)/12, (serial-1)%12+1
		paymentID := uuid.NewString()
		monthTracking := fmt.Sprintf("%s-%04d%02d", tracking, year, month)
		if _, e = tx.Exec(c, `INSERT INTO tuition_payments(id,athlete_id,gym_id,coach_id,tuition_amount,platform_fee,total_amount,status,tracking_id,billing_year,billing_month,payment_kind,created_at) VALUES($1,$2,$3,$4,$5,0,$5,'succeeded',$6,$7,$8,'monthly',$9)`, paymentID, athleteID, gymID, coachID, tuition, monthTracking, year, month, createdAt); e != nil {
			return nil, dbError(e)
		}
		if tuition > 0 {
			description := fmt.Sprintf("monthly tuition %04d/%02d", year, month)
			if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'tuition_payment',$4,$5)`, coachID, domain.LedgerTuition, tuition, paymentID, description); e != nil {
				return nil, dbError(e)
			}
		}
		if result.PaidMonths == 0 {
			result.FromYear, result.FromMonth = year, month
			result.TrackingID = monthTracking
		}
		result.ToYear, result.ToMonth = year, month
		result.PaidMonths++
		result.PaidAmount += tuition
	}
	if result.PaidMonths == 0 {
		return nil, domain.ErrConflict
	}
	if e = tx.Commit(c); e != nil {
		return nil, dbError(e)
	}
	return result, nil
}
func (s *Store) ListAthletePayments(c context.Context, athleteID string) ([]domain.AthletePayment, error) {
	rows, e := s.DB.Query(c, `SELECT tp.id,tp.athlete_id,tp.gym_id,tp.coach_id,tp.tuition_amount,tp.platform_fee,tp.total_amount,tp.status,tp.tracking_id,COALESCE(tp.billing_year,0),COALESCE(tp.billing_month,0),tp.payment_kind,tp.created_at,g.name,concat_ws(' ',u.first_name,u.last_name) FROM tuition_payments tp JOIN gyms g ON g.id=tp.gym_id JOIN users u ON u.id=tp.coach_id WHERE tp.athlete_id=$1 ORDER BY tp.created_at DESC,tp.billing_year DESC NULLS LAST,tp.billing_month DESC NULLS LAST`, athleteID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.AthletePayment{}
	for rows.Next() {
		var payment domain.AthletePayment
		if e = rows.Scan(&payment.ID, &payment.AthleteID, &payment.GymID, &payment.CoachID, &payment.Tuition, &payment.PlatformFee, &payment.Total, &payment.Status, &payment.TrackingID, &payment.BillingYear, &payment.BillingMonth, &payment.PaymentKind, &payment.CreatedAt, &payment.GymName, &payment.CoachName); e != nil {
			return nil, e
		}
		out = append(out, payment)
	}
	return out, rows.Err()
}
func (s *Store) CreateWithdrawal(c context.Context, w *domain.WithdrawalRequest) error {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	if _, e = tx.Exec(c, `SELECT pg_advisory_xact_lock(hashtext($1))`, w.CoachID); e != nil {
		return e
	}
	var balance int64
	if e = tx.QueryRow(c, `SELECT COALESCE(sum(amount),0) FROM wallet_entries WHERE coach_id=$1`, w.CoachID).Scan(&balance); e != nil {
		return e
	}
	if balance < w.GrossAmount {
		return domain.ErrInsufficientFunds
	}
	_, e = tx.Exec(c, `INSERT INTO withdrawal_requests(id,coach_id,gross_amount,fee_amount,net_amount,iban,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$8)`, w.ID, w.CoachID, w.GrossAmount, w.FeeAmount, w.NetAmount, w.IBAN, w.Status, w.CreatedAt)
	if e != nil {
		return dbError(e)
	}
	_, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'withdrawal',$4,'withdrawal amount reserved')`, w.CoachID, domain.LedgerWithdrawalHold, -w.GrossAmount, w.ID)
	if e != nil {
		return dbError(e)
	}
	return tx.Commit(c)
}
func scanWithdrawal(r pgx.Row) (*domain.WithdrawalRequest, error) {
	v := new(domain.WithdrawalRequest)
	e := r.Scan(&v.ID, &v.CoachID, &v.GrossAmount, &v.FeeAmount, &v.NetAmount, &v.IBAN, &v.Status, &v.BankTrackingCode, &v.RejectionReason, &v.ReviewedBy, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(e)
}

const withdrawalCols = `id,coach_id,gross_amount,fee_amount,net_amount,iban,status,bank_tracking_code,rejection_reason,reviewed_by,created_at,updated_at`

func (s *Store) ListWithdrawals(c context.Context, status *domain.WithdrawalStatus, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `SELECT `+withdrawalCols+` FROM withdrawal_requests WHERE ($1::varchar IS NULL OR status=$1) ORDER BY created_at DESC LIMIT $2 OFFSET $3`, status, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.WithdrawalRequest{}
	for rows.Next() {
		v, e := scanWithdrawal(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}
func (s *Store) ListCoachWithdrawals(c context.Context, coach string, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `SELECT `+withdrawalCols+` FROM withdrawal_requests WHERE coach_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, coach, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.WithdrawalRequest{}
	for rows.Next() {
		v, e := scanWithdrawal(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *v)
	}
	return out, rows.Err()
}
func (s *Store) ReviewWithdrawal(c context.Context, id, admin string, status domain.WithdrawalStatus, tracking, reason *string) error {
	tx, e := s.DB.Begin(c)
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	var coach string
	var gross int64
	var current domain.WithdrawalStatus
	e = tx.QueryRow(c, `SELECT coach_id,gross_amount,status FROM withdrawal_requests WHERE id=$1 FOR UPDATE`, id).Scan(&coach, &gross, &current)
	if e != nil {
		return dbError(e)
	}
	if current != domain.WithdrawalPending {
		return domain.ErrInvalidState
	}
	tag, e := tx.Exec(c, `UPDATE withdrawal_requests SET status=$2,bank_tracking_code=$3,rejection_reason=$4,reviewed_by=$5,updated_at=now() WHERE id=$1`, id, status, tracking, reason, admin)
	if e != nil || tag.RowsAffected() != 1 {
		return dbError(e)
	}
	if status == domain.WithdrawalRejected {
		_, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'withdrawal',$4,'rejected withdrawal released')`, coach, domain.LedgerWithdrawalRelease, gross, id)
		if e != nil {
			return dbError(e)
		}
	}
	return tx.Commit(c)
}
func (s *Store) ListLedger(c context.Context, coach string, limit, offset int) ([]domain.WalletEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `SELECT id,coach_id,entry_type,amount,reference_type,reference_id,description,created_at FROM wallet_entries WHERE coach_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, coach, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.WalletEntry{}
	for rows.Next() {
		var v domain.WalletEntry
		if e = rows.Scan(&v.ID, &v.CoachID, &v.Type, &v.Amount, &v.ReferenceType, &v.ReferenceID, &v.Description, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) CreateMessage(c context.Context, m *domain.Message, recipients []domain.MessageRecipient) error {
	tx, e := s.DB.Begin(c)
	if e != nil {
		return e
	}
	defer tx.Rollback(c)
	_, e = tx.Exec(c, `INSERT INTO messages(id,sender_id,gym_id,recipient_id,audience,subject,body,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, m.ID, m.SenderID, m.GymID, m.RecipientID, m.Audience, m.Subject, m.Body, m.CreatedAt)
	if e != nil {
		return dbError(e)
	}
	for _, r := range recipients {
		if _, e = tx.Exec(c, `INSERT INTO message_recipients(message_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, m.ID, r.UserID); e != nil {
			return dbError(e)
		}
	}
	return tx.Commit(c)
}
func (s *Store) ListForUser(c context.Context, user string, limit, offset int) ([]domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `SELECT m.id,m.sender_id,m.gym_id,m.recipient_id,m.audience,m.subject,m.body,m.created_at FROM messages m JOIN message_recipients r ON r.message_id=m.id WHERE r.user_id=$1 ORDER BY m.created_at DESC LIMIT $2 OFFSET $3`, user, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Message{}
	for rows.Next() {
		var m domain.Message
		if e = rows.Scan(&m.ID, &m.SenderID, &m.GymID, &m.RecipientID, &m.Audience, &m.Subject, &m.Body, &m.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Store) MarkRead(c context.Context, message, user string) error {
	tag, e := s.DB.Exec(c, `UPDATE message_recipients SET read_at=COALESCE(read_at,now()) WHERE message_id=$1 AND user_id=$2`, message, user)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) AllUserIDs(c context.Context, exclude string) ([]string, error) {
	rows, e := s.DB.Query(c, `SELECT id FROM users WHERE id<>$1`, exclude)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			return nil, e
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
func (s *Store) GymRecipientIDs(c context.Context, gym, exclude string) ([]string, error) {
	rows, e := s.DB.Query(c, `SELECT coach_id FROM gyms WHERE id=$1 AND coach_id<>$2 UNION SELECT athlete_id FROM gym_athletes WHERE gym_id=$1 AND athlete_id<>$2`, gym, exclude)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			return nil, e
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
func (s *Store) CanAccessGym(c context.Context, user, gym string) (bool, error) {
	var ok bool
	e := s.DB.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM gyms WHERE id=$2 AND coach_id=$1 UNION SELECT 1 FROM gym_athletes WHERE gym_id=$2 AND athlete_id=$1)`, user, gym).Scan(&ok)
	return ok, e
}
func (s *Store) ShareGym(c context.Context, a, b string) (bool, error) {
	var ok bool
	e := s.DB.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM (SELECT id FROM gyms WHERE coach_id=$1 UNION SELECT gym_id FROM gym_athletes WHERE athlete_id=$1) x JOIN (SELECT id FROM gyms WHERE coach_id=$2 UNION SELECT gym_id FROM gym_athletes WHERE athlete_id=$2) y ON y.id=x.id)`, a, b).Scan(&ok)
	return ok, e
}

type MessageStore struct{ *Store }

func (m MessageStore) Create(c context.Context, v *domain.Message, r []domain.MessageRecipient) error {
	return m.CreateMessage(c, v, r)
}
