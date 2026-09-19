package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	if e = runMigrations(ctx, p); e != nil {
		p.Close()
		return nil, fmt.Errorf("initialize database: %w", e)
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
	if errors.As(e, &pe) && strings.Contains(pe.Message, "PROTECTED_DEFAULT_ADMIN") {
		return domain.ErrProtectedAdmin
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
func (s *Store) UpdatePassword(c context.Context, id, passwordHash string) error {
	tag, e := s.DB.Exec(c, `UPDATE users SET password_hash=$2,updated_at=now() WHERE id=$1`, id, passwordHash)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(e)
}
func (s *Store) PromoteToAdmin(c context.Context, id, username, passwordHash string) error {
	tag, e := s.DB.Exec(c, `UPDATE users SET role='admin',username=$2,password_hash=$3,phone=NULL,bank_iban=NULL,active=true,updated_at=now() WHERE id=$1`, id, username, passwordHash)
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
	rows, e := s.DB.Query(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),(ga.athlete_id IS NOT NULL),EXISTS(SELECT 1 FROM tuition_payments rp WHERE rp.gym_id=g.id AND rp.athlete_id=$1 AND rp.payment_kind='registration' AND rp.status='succeeded'),jr.status,jr.rejection_reason,CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,CASE WHEN g.tuition_mode='monthly' AND ga.athlete_id IS NOT NULL THEN GREATEST((($2*12+$3)-(ga.billing_start_year*12+ga.billing_start_month)+1)-(SELECT count(*) FROM tuition_payments tp WHERE tp.gym_id=g.id AND tp.athlete_id=$1 AND tp.billing_year IS NOT NULL AND (tp.billing_year*12+tp.billing_month) BETWEEN (ga.billing_start_year*12+ga.billing_start_month) AND ($2*12+$3)),0) ELSE 0 END FROM gyms g JOIN users u ON u.id=g.coach_id LEFT JOIN gym_athletes ga ON ga.gym_id=g.id AND ga.athlete_id=$1 LEFT JOIN LATERAL (SELECT status,rejection_reason FROM gym_join_requests WHERE gym_id=g.id AND athlete_id=$1 ORDER BY created_at DESC LIMIT 1) jr ON true ORDER BY g.created_at DESC`, athleteID, year, month)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.GymCatalogItem{}
	for rows.Next() {
		var v domain.GymCatalogItem
		var requestStatus *string
		if e = rows.Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CoachName, &v.Joined, &v.RegistrationPaid, &requestStatus, &v.JoinRequestRejectionReason, &v.CapacityRemaining, &v.UnpaidMonths); e != nil {
			return nil, e
		}
		if requestStatus != nil {
			status := domain.GymJoinRequestStatus(*requestStatus)
			v.JoinRequestStatus = &status
		}
		v.OutstandingAmount = int64(v.UnpaidMonths) * v.TuitionAmount
		v.MonthlyPaymentDue = v.UnpaidMonths > 0
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) GetCatalogItem(c context.Context, gymID, athleteID string, year, month int) (*domain.GymCatalogItem, error) {
	v := new(domain.GymCatalogItem)
	var requestStatus *string
	e := s.DB.QueryRow(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),(ga.athlete_id IS NOT NULL),EXISTS(SELECT 1 FROM tuition_payments rp WHERE rp.gym_id=g.id AND rp.athlete_id=$2 AND rp.payment_kind='registration' AND rp.status='succeeded'),jr.status,jr.rejection_reason,CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,CASE WHEN g.tuition_mode='monthly' AND ga.athlete_id IS NOT NULL THEN GREATEST((($3*12+$4)-(ga.billing_start_year*12+ga.billing_start_month)+1)-(SELECT count(*) FROM tuition_payments tp WHERE tp.gym_id=g.id AND tp.athlete_id=$2 AND tp.billing_year IS NOT NULL AND (tp.billing_year*12+tp.billing_month) BETWEEN (ga.billing_start_year*12+ga.billing_start_month) AND ($3*12+$4)),0) ELSE 0 END FROM gyms g JOIN users u ON u.id=g.coach_id LEFT JOIN gym_athletes ga ON ga.gym_id=g.id AND ga.athlete_id=$2 LEFT JOIN LATERAL (SELECT status,rejection_reason FROM gym_join_requests WHERE gym_id=g.id AND athlete_id=$2 ORDER BY created_at DESC LIMIT 1) jr ON true WHERE g.id=$1`, gymID, athleteID, year, month).Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CoachName, &v.Joined, &v.RegistrationPaid, &requestStatus, &v.JoinRequestRejectionReason, &v.CapacityRemaining, &v.UnpaidMonths)
	if requestStatus != nil {
		status := domain.GymJoinRequestStatus(*requestStatus)
		v.JoinRequestStatus = &status
	}
	v.OutstandingAmount = int64(v.UnpaidMonths) * v.TuitionAmount
	v.MonthlyPaymentDue = v.UnpaidMonths > 0
	return v, dbError(e)
}
func (s *Store) ListPublicCatalog(c context.Context) ([]domain.GymCatalogItem, error) {
	rows, err := s.DB.Query(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),false,false,NULL::text,NULL::text,CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,0 FROM gyms g JOIN users u ON u.id=g.coach_id WHERE u.active=true ORDER BY g.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymCatalogItem{}
	for rows.Next() {
		var item domain.GymCatalogItem
		var requestStatus *string
		if err = rows.Scan(&item.ID, &item.CoachID, &item.Name, &item.Province, &item.City, &item.Address, &item.Phone, &item.Sport, &item.SportStyle, &item.ActiveWeekdays, &item.StartTime, &item.EndTime, &item.TuitionMode, &item.TuitionAmount, &item.Capacity, &item.CreatedAt, &item.UpdatedAt, &item.CoachName, &item.Joined, &item.RegistrationPaid, &requestStatus, &item.JoinRequestRejectionReason, &item.CapacityRemaining, &item.UnpaidMonths); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) GetPublicCatalogItem(c context.Context, gymID string) (*domain.GymCatalogItem, error) {
	item := new(domain.GymCatalogItem)
	var requestStatus *string
	err := s.DB.QueryRow(c, `SELECT `+gymCatalogCols+`,concat_ws(' ',u.first_name,u.last_name),false,false,NULL::text,NULL::text,CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes gc WHERE gc.gym_id=g.id),0) END,0 FROM gyms g JOIN users u ON u.id=g.coach_id WHERE g.id=$1 AND u.active=true`, gymID).Scan(&item.ID, &item.CoachID, &item.Name, &item.Province, &item.City, &item.Address, &item.Phone, &item.Sport, &item.SportStyle, &item.ActiveWeekdays, &item.StartTime, &item.EndTime, &item.TuitionMode, &item.TuitionAmount, &item.Capacity, &item.CreatedAt, &item.UpdatedAt, &item.CoachName, &item.Joined, &item.RegistrationPaid, &requestStatus, &item.JoinRequestRejectionReason, &item.CapacityRemaining, &item.UnpaidMonths)
	return item, dbError(err)
}

func (s *Store) ListShowcase(c context.Context) ([]domain.GymShowcaseItem, error) {
	rows, err := s.DB.Query(c, `
		SELECT g.id, g.name, concat_ws(' ', u.first_name, u.last_name),
		       g.sport, g.sport_style, g.start_time::text, g.end_time::text,
		       image.storage_key,
		       CASE WHEN g.capacity=0 THEN NULL ELSE GREATEST(g.capacity-(SELECT count(*) FROM gym_athletes ga WHERE ga.gym_id=g.id),0) END
		FROM gyms g
		JOIN users u ON u.id = g.coach_id
		LEFT JOIN LATERAL (
			SELECT gi.storage_key
			FROM gym_images gi
			WHERE gi.gym_id = g.id
			ORDER BY gi.created_at, gi.id
			LIMIT 1
		) image ON true
		WHERE u.active = true
		ORDER BY g.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.GymShowcaseItem{}
	for rows.Next() {
		var item domain.GymShowcaseItem
		if err = rows.Scan(&item.ID, &item.Name, &item.CoachName, &item.Sport, &item.SportStyle, &item.StartTime, &item.EndTime, &item.ImageStorageKey, &item.CapacityRemaining); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

const joinRequestCols = `r.id,r.gym_id,g.name,g.coach_id,r.athlete_id,concat_ws(' ',u.first_name,u.last_name),u.phone,r.status,r.rejection_reason,r.reviewed_at,r.created_at,r.updated_at`

func scanJoinRequest(row pgx.Row) (*domain.GymJoinRequest, error) {
	v := new(domain.GymJoinRequest)
	err := row.Scan(&v.ID, &v.GymID, &v.GymName, &v.CoachID, &v.AthleteID, &v.AthleteName, &v.AthletePhone, &v.Status, &v.RejectionReason, &v.ReviewedAt, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(err)
}

func (s *Store) CreateJoinRequest(c context.Context, v *domain.GymJoinRequest) error {
	tx, err := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(c)
	var capacity, members int64
	if err = tx.QueryRow(c, `SELECT g.coach_id,g.name,g.capacity FROM gyms g JOIN users coach ON coach.id=g.coach_id AND coach.active=true WHERE g.id=$1 FOR UPDATE OF g`, v.GymID).Scan(&v.CoachID, &v.GymName, &capacity); err != nil {
		return dbError(err)
	}
	var role domain.Role
	if err = tx.QueryRow(c, `SELECT role,concat_ws(' ',first_name,last_name),phone FROM users WHERE id=$1 AND active=true FOR SHARE`, v.AthleteID).Scan(&role, &v.AthleteName, &v.AthletePhone); err != nil {
		return dbError(err)
	}
	if role != domain.RoleAthlete {
		return domain.ErrForbidden
	}
	var joined, pending bool
	if err = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2),EXISTS(SELECT 1 FROM gym_join_requests WHERE gym_id=$1 AND athlete_id=$2 AND status='pending')`, v.GymID, v.AthleteID).Scan(&joined, &pending); err != nil {
		return err
	}
	if joined {
		return domain.ErrConflict
	}
	if pending {
		return domain.ErrJoinRequestPending
	}
	if capacity > 0 {
		if err = tx.QueryRow(c, `SELECT count(*) FROM gym_athletes WHERE gym_id=$1`, v.GymID).Scan(&members); err != nil {
			return err
		}
		if members >= capacity {
			return domain.ErrGymFull
		}
	}
	if _, err = tx.Exec(c, `INSERT INTO gym_join_requests(id,gym_id,athlete_id,status,created_at,updated_at) VALUES($1,$2,$3,'pending',$4,$4)`, v.ID, v.GymID, v.AthleteID, v.CreatedAt); err != nil {
		return dbError(err)
	}
	return tx.Commit(c)
}

func (s *Store) ListCoachJoinRequests(c context.Context, coachID string, status *domain.GymJoinRequestStatus) ([]domain.GymJoinRequest, error) {
	filter := ""
	if status != nil {
		filter = string(*status)
	}
	rows, err := s.DB.Query(c, `SELECT `+joinRequestCols+` FROM gym_join_requests r JOIN gyms g ON g.id=r.gym_id JOIN users u ON u.id=r.athlete_id WHERE g.coach_id=$1 AND ($2='' OR r.status=$2) ORDER BY CASE r.status WHEN 'pending' THEN 0 ELSE 1 END,r.created_at DESC`, coachID, filter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymJoinRequest{}
	for rows.Next() {
		item, scanErr := scanJoinRequest(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (s *Store) ListAthleteJoinRequests(c context.Context, athleteID string) ([]domain.GymJoinRequest, error) {
	rows, err := s.DB.Query(c, `SELECT `+joinRequestCols+` FROM gym_join_requests r JOIN gyms g ON g.id=r.gym_id JOIN users u ON u.id=r.athlete_id WHERE r.athlete_id=$1 ORDER BY r.created_at DESC`, athleteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymJoinRequest{}
	for rows.Next() {
		item, scanErr := scanJoinRequest(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (s *Store) ReviewJoinRequest(c context.Context, requestID, coachID string, status domain.GymJoinRequestStatus, reason *string, year, month int, now time.Time) (*domain.GymJoinRequest, error) {
	tx, err := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(c)
	var gymID, athleteID, ownerID string
	var current domain.GymJoinRequestStatus
	var capacity int64
	if err = tx.QueryRow(c, `SELECT r.gym_id,r.athlete_id,r.status,g.coach_id,g.capacity FROM gym_join_requests r JOIN gyms g ON g.id=r.gym_id WHERE r.id=$1 FOR UPDATE OF r,g`, requestID).Scan(&gymID, &athleteID, &current, &ownerID, &capacity); err != nil {
		return nil, dbError(err)
	}
	if ownerID != coachID {
		return nil, domain.ErrForbidden
	}
	if current != domain.GymJoinPending {
		return nil, domain.ErrInvalidState
	}
	if status == domain.GymJoinApproved {
		var members int64
		if err = tx.QueryRow(c, `SELECT count(*) FROM gym_athletes WHERE gym_id=$1`, gymID).Scan(&members); err != nil {
			return nil, err
		}
		if capacity > 0 && members >= capacity {
			return nil, domain.ErrGymFull
		}
		if _, err = tx.Exec(c, `INSERT INTO gym_athletes(gym_id,athlete_id,billing_start_year,billing_start_month,registered_at) VALUES($1,$2,$3,$4,$5)`, gymID, athleteID, year, month, now); err != nil {
			return nil, dbError(err)
		}
		reason = nil
	}
	tag, err := tx.Exec(c, `UPDATE gym_join_requests SET status=$2,rejection_reason=$3,reviewed_at=$4,updated_at=$4 WHERE id=$1 AND status='pending'`, requestID, status, reason, now)
	if err != nil {
		return nil, dbError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrInvalidState
	}
	item, err := scanJoinRequest(tx.QueryRow(c, `SELECT `+joinRequestCols+` FROM gym_join_requests r JOIN gyms g ON g.id=r.gym_id JOIN users u ON u.id=r.athlete_id WHERE r.id=$1`, requestID))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(c); err != nil {
		return nil, dbError(err)
	}
	return item, nil
}
func (s *Store) ListGymAthletes(c context.Context, gymID string, year, month int) ([]domain.GymAthleteProfile, error) {
	rows, e := s.DB.Query(c, `SELECT u.id,u.first_name,u.last_name,u.phone,COALESCE(EXTRACT(year FROM age(current_date,u.birth_date))::int,0),CASE WHEN g.tuition_mode='monthly' THEN GREATEST((($2*12+$3)-(ga.billing_start_year*12+ga.billing_start_month)+1)-(SELECT count(*) FROM tuition_payments tp WHERE tp.gym_id=ga.gym_id AND tp.athlete_id=ga.athlete_id AND tp.billing_year IS NOT NULL AND (tp.billing_year*12+tp.billing_month) BETWEEN (ga.billing_start_year*12+ga.billing_start_month) AND ($2*12+$3)),0) ELSE 0 END,g.tuition_amount,ga.registered_at,latest.weight_kg::float8 FROM gym_athletes ga JOIN users u ON u.id=ga.athlete_id JOIN gyms g ON g.id=ga.gym_id LEFT JOIN LATERAL (SELECT weight_kg FROM athlete_body_measurements WHERE athlete_id=ga.athlete_id ORDER BY measured_at DESC,id DESC LIMIT 1) latest ON true WHERE ga.gym_id=$1 ORDER BY ga.registered_at DESC`, gymID, year, month)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.GymAthleteProfile{}
	for rows.Next() {
		var v domain.GymAthleteProfile
		var tuitionAmount int64
		if e = rows.Scan(&v.ID, &v.FirstName, &v.LastName, &v.Phone, &v.Age, &v.UnpaidMonths, &tuitionAmount, &v.RegisteredAt, &v.LastWeightKG); e != nil {
			return nil, e
		}
		v.OutstandingAmount = int64(v.UnpaidMonths) * tuitionAmount
		v.TuitionStatus = "settled"
		if v.UnpaidMonths > 0 {
			v.TuitionStatus = "debtor"
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
func (s *Store) ListUserDocuments(c context.Context, userID string) ([]domain.UserDocument, error) {
	rows, err := s.DB.Query(c, `SELECT id,user_id,kind,storage_key,mime_type,size_bytes,created_at FROM user_documents WHERE user_id=$1 ORDER BY kind,created_at,id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.UserDocument{}
	for rows.Next() {
		var item domain.UserDocument
		if err = rows.Scan(&item.ID, &item.UserID, &item.Kind, &item.StorageKey, &item.MIMEType, &item.Size, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) ReplaceUserDocuments(c context.Context, userID string, kind domain.DocumentKind, replacements []domain.UserDocument) ([]domain.UserDocument, error) {
	tx, err := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(c)
	if _, err = tx.Exec(c, `SELECT 1 FROM users WHERE id=$1 FOR UPDATE`, userID); err != nil {
		return nil, dbError(err)
	}
	rows, err := tx.Query(c, `SELECT id,user_id,kind,storage_key,mime_type,size_bytes,created_at FROM user_documents WHERE user_id=$1 AND kind=$2 FOR UPDATE`, userID, kind)
	if err != nil {
		return nil, err
	}
	old := []domain.UserDocument{}
	for rows.Next() {
		var item domain.UserDocument
		if err = rows.Scan(&item.ID, &item.UserID, &item.Kind, &item.StorageKey, &item.MIMEType, &item.Size, &item.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		old = append(old, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if _, err = tx.Exec(c, `DELETE FROM user_documents WHERE user_id=$1 AND kind=$2`, userID, kind); err != nil {
		return nil, dbError(err)
	}
	for index := range replacements {
		document := &replacements[index]
		if _, err = tx.Exec(c, `INSERT INTO user_documents(id,user_id,kind,storage_key,mime_type,size_bytes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, document.ID, document.UserID, document.Kind, document.StorageKey, document.MIMEType, document.Size, document.CreatedAt); err != nil {
			return nil, dbError(err)
		}
	}
	if kind == domain.DocumentProfile {
		if _, err = tx.Exec(c, `UPDATE users SET face_image_id=$2,updated_at=now() WHERE id=$1`, userID, replacements[0].StorageKey); err != nil {
			return nil, dbError(err)
		}
	}
	if err = tx.Commit(c); err != nil {
		return nil, dbError(err)
	}
	return old, nil
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

func (s *Store) CreateBodyMeasurement(c context.Context, v *domain.BodyMeasurement, minimumTime time.Time) error {
	tx, err := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(c)
	var role domain.Role
	if err = tx.QueryRow(c, `SELECT role FROM users WHERE id=$1 AND active=true FOR UPDATE`, v.AthleteID).Scan(&role); err != nil {
		return dbError(err)
	}
	if role != domain.RoleAthlete {
		return domain.ErrForbidden
	}
	var allowed bool
	if err = tx.QueryRow(c, `SELECT NOT EXISTS(SELECT 1 FROM athlete_body_measurements WHERE athlete_id=$1 AND measured_at>$2)`, v.AthleteID, minimumTime).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return domain.ErrMeasurementTooSoon
	}
	if _, err = tx.Exec(c, `INSERT INTO athlete_body_measurements(id,athlete_id,height_cm,weight_kg,measured_at) VALUES($1,$2,$3,$4,$5)`, v.ID, v.AthleteID, v.HeightCM, v.WeightKG, v.MeasuredAt); err != nil {
		return dbError(err)
	}
	return tx.Commit(c)
}

func (s *Store) ListBodyMeasurements(c context.Context, athleteID string, limit int) ([]domain.BodyMeasurement, error) {
	rows, err := s.DB.Query(c, `SELECT id,athlete_id,height_cm::float8,weight_kg::float8,measured_at FROM athlete_body_measurements WHERE athlete_id=$1 ORDER BY measured_at DESC LIMIT $2`, athleteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.BodyMeasurement{}
	for rows.Next() {
		var item domain.BodyMeasurement
		if err = rows.Scan(&item.ID, &item.AthleteID, &item.HeightCM, &item.WeightKG, &item.MeasuredAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

func (o GymStore) ListGymAttendance(c context.Context, gymID string, date time.Time) ([]domain.GymAttendanceAthlete, error) {
	rows, err := o.DB.Query(c, `SELECT u.id,u.first_name,u.last_name,(ga.athlete_id IS NOT NULL)
		FROM gym_athletes members JOIN users u ON u.id=members.athlete_id
		LEFT JOIN gym_attendance ga ON ga.gym_id=members.gym_id AND ga.athlete_id=members.athlete_id AND ga.attendance_date=$2
		WHERE members.gym_id=$1 ORDER BY u.first_name,u.last_name,u.id`, gymID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymAttendanceAthlete{}
	for rows.Next() {
		var item domain.GymAttendanceAthlete
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.Present); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (o GymStore) ListGymAttendanceHistory(c context.Context, gymID string, from, until time.Time) ([]domain.GymAttendanceRecord, error) {
	rows, err := o.DB.Query(c, `SELECT attendance_date,athlete_id FROM gym_attendance WHERE gym_id=$1 AND attendance_date >= $2 AND attendance_date < $3 ORDER BY attendance_date,athlete_id`, gymID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymAttendanceRecord{}
	for rows.Next() {
		var item domain.GymAttendanceRecord
		if err := rows.Scan(&item.AttendanceDate, &item.AthleteID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (o GymStore) ListAthleteGymAttendance(c context.Context, gymID, athleteID string) ([]domain.GymAttendanceRecord, error) {
	rows, err := o.DB.Query(c, `SELECT dates.attendance_date,$2,EXISTS(SELECT 1 FROM gym_attendance own WHERE own.gym_id=$1 AND own.athlete_id=$2 AND own.attendance_date=dates.attendance_date)
		FROM (SELECT DISTINCT attendance_date FROM gym_attendance WHERE gym_id=$1) dates ORDER BY dates.attendance_date DESC`, gymID, athleteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.GymAttendanceRecord{}
	for rows.Next() {
		var item domain.GymAttendanceRecord
		if err := rows.Scan(&item.AttendanceDate, &item.AthleteID, &item.Present); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (o GymStore) SetGymAttendance(c context.Context, gymID, athleteID string, date time.Time, present bool, recordedBy string) error {
	if present {
		tag, err := o.DB.Exec(c, `INSERT INTO gym_attendance(gym_id,athlete_id,attendance_date,recorded_by)
			SELECT $1,$2,$3,$4 WHERE EXISTS(SELECT 1 FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2)
			ON CONFLICT(gym_id,athlete_id,attendance_date) DO UPDATE SET recorded_by=EXCLUDED.recorded_by,recorded_at=now()`, gymID, athleteID, date, recordedBy)
		if err == nil && tag.RowsAffected() == 0 {
			return domain.ErrNotFound
		}
		return dbError(err)
	}
	tag, err := o.DB.Exec(c, `DELETE FROM gym_attendance WHERE gym_id=$1 AND athlete_id=$2 AND attendance_date=$3`, gymID, athleteID, date)
	if err == nil && tag.RowsAffected() == 0 {
		var member bool
		if membershipErr := o.DB.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2)`, gymID, athleteID).Scan(&member); membershipErr != nil {
			return dbError(membershipErr)
		}
		if !member {
			return domain.ErrNotFound
		}
	}
	return dbError(err)
}

func (s *Store) GetPolicy(c context.Context) (domain.WithdrawalPolicy, error) {
	var p domain.WithdrawalPolicy
	e := s.DB.QueryRow(c, `SELECT min_amount,max_amount,fee_bps,updated_at FROM withdrawal_policy WHERE id=true`).Scan(&p.MinAmount, &p.MaxAmount, &p.FeeBPS, &p.UpdatedAt)
	return p, dbError(e)
}
func (s *Store) SetPolicy(c context.Context, p domain.WithdrawalPolicy) error {
	_, e := s.DB.Exec(c, `INSERT INTO withdrawal_policy(id,min_amount,max_amount,fee_bps,updated_at) VALUES(true,$1,$2,$3,now()) ON CONFLICT(id) DO UPDATE SET min_amount=EXCLUDED.min_amount,max_amount=EXCLUDED.max_amount,fee_bps=EXCLUDED.fee_bps,updated_at=now()`, p.MinAmount, p.MaxAmount, p.FeeBPS)
	return dbError(e)
}
func (s *Store) GetBackupPolicy(c context.Context) (domain.BackupPolicy, error) {
	var p domain.BackupPolicy
	e := s.DB.QueryRow(c, `SELECT schedule_minutes,retention,updated_at FROM backup_policy WHERE id=true`).Scan(&p.ScheduleMinutes, &p.Retention, &p.UpdatedAt)
	return p, dbError(e)
}
func (s *Store) SetBackupPolicy(c context.Context, p domain.BackupPolicy) error {
	_, e := s.DB.Exec(c, `INSERT INTO backup_policy(id,schedule_minutes,retention,updated_at) VALUES(true,$1,$2,now()) ON CONFLICT(id) DO UPDATE SET schedule_minutes=EXCLUDED.schedule_minutes,retention=EXCLUDED.retention,updated_at=now()`, p.ScheduleMinutes, p.Retention)
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

const paymentOrderCols = `id,athlete_id,gym_id,kind,tuition_amount,platform_fee,total_amount,billing_year,billing_month,status,zibal_track_id,zibal_ref_number,failure_reason,expires_at,completed_at,created_at,updated_at`

func scanPaymentOrder(row pgx.Row) (*domain.PaymentOrder, error) {
	v := new(domain.PaymentOrder)
	e := row.Scan(&v.ID, &v.AthleteID, &v.GymID, &v.Kind, &v.Tuition, &v.PlatformFee, &v.Total, &v.BillingYear, &v.BillingMonth, &v.Status, &v.TrackID, &v.RefNumber, &v.Failure, &v.ExpiresAt, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(e)
}

func pendingPaymentOrder(c context.Context, tx pgx.Tx, athleteID, gymID string) (*domain.PaymentOrder, error) {
	v, err := scanPaymentOrder(tx.QueryRow(c, `SELECT `+paymentOrderCols+` FROM payment_orders WHERE athlete_id=$1 AND gym_id=$2 AND status='pending' ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, athleteID, gymID))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, nil
	}
	return v, err
}

func clearStaleCreatedPaymentOrders(c context.Context, tx pgx.Tx, athleteID, gymID string, now time.Time) error {
	_, err := tx.Exec(c, `UPDATE payment_orders SET status='failed',failure_reason='payment initialization timed out',updated_at=$3 WHERE athlete_id=$1 AND gym_id=$2 AND status='created' AND updated_at < ($3::timestamptz-interval '2 minutes')`, athleteID, gymID, now)
	return dbError(err)
}

func (s *Store) CreateRegistrationPaymentOrder(c context.Context, athleteID, gymID string, year, month int, now, expires time.Time) (*domain.PaymentOrder, error) {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(c)
	var role domain.Role
	if e = tx.QueryRow(c, `SELECT role FROM users WHERE id=$1 FOR SHARE`, athleteID).Scan(&role); e != nil {
		return nil, dbError(e)
	}
	if role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	var tuition int64
	if e = tx.QueryRow(c, `SELECT g.tuition_amount FROM gyms g JOIN users coach ON coach.id=g.coach_id AND coach.active=true WHERE g.id=$1 FOR UPDATE OF g`, gymID).Scan(&tuition); e != nil {
		return nil, dbError(e)
	}
	if e = clearStaleCreatedPaymentOrders(c, tx, athleteID, gymID, now); e != nil {
		return nil, e
	}
	if e = tx.QueryRow(c, `SELECT billing_start_year,billing_start_month FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2 FOR UPDATE`, gymID, athleteID).Scan(&year, &month); e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			return nil, domain.ErrJoinApprovalNeeded
		}
		return nil, dbError(e)
	}
	var alreadyPaid bool
	if e = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM tuition_payments WHERE gym_id=$1 AND athlete_id=$2 AND payment_kind='registration' AND status='succeeded')`, gymID, athleteID).Scan(&alreadyPaid); e != nil {
		return nil, e
	}
	if alreadyPaid {
		return nil, domain.ErrConflict
	}
	if existing, err := pendingPaymentOrder(c, tx, athleteID, gymID); err != nil {
		return nil, err
	} else if existing != nil {
		if err = tx.Commit(c); err != nil {
			return nil, dbError(err)
		}
		return existing, nil
	}
	var creating bool
	if e = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM payment_orders WHERE athlete_id=$1 AND gym_id=$2 AND status='created')`, athleteID, gymID).Scan(&creating); e != nil {
		return nil, e
	}
	if creating {
		return nil, domain.ErrConflict
	}
	var fee int64
	if e = tx.QueryRow(c, `SELECT fee_amount FROM registration_fee_policy WHERE id=true FOR SHARE`).Scan(&fee); e != nil {
		return nil, dbError(e)
	}
	if tuition > math.MaxInt64-fee || tuition+fee <= 1000 {
		return nil, domain.ErrPaymentAmount
	}
	v := &domain.PaymentOrder{ID: uuid.NewString(), AthleteID: athleteID, GymID: gymID, Kind: "registration", Tuition: tuition, PlatformFee: fee, Total: tuition + fee, BillingYear: year, BillingMonth: month, Status: domain.PaymentOrderCreated, ExpiresAt: expires, CreatedAt: now, UpdatedAt: now}
	if _, e = tx.Exec(c, `INSERT INTO payment_orders(id,athlete_id,gym_id,kind,tuition_amount,platform_fee,total_amount,billing_year,billing_month,status,expires_at,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, v.ID, v.AthleteID, v.GymID, v.Kind, v.Tuition, v.PlatformFee, v.Total, v.BillingYear, v.BillingMonth, v.Status, v.ExpiresAt, v.CreatedAt, v.UpdatedAt); e != nil {
		return nil, dbError(e)
	}
	if e = tx.Commit(c); e != nil {
		return nil, dbError(e)
	}
	return v, nil
}

func (s *Store) CreateMonthlyPaymentOrder(c context.Context, athleteID, gymID string, currentYear, currentMonth int, now, expires time.Time) (*domain.PaymentOrder, error) {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(c)
	var role domain.Role
	if e = tx.QueryRow(c, `SELECT role FROM users WHERE id=$1 FOR SHARE`, athleteID).Scan(&role); e != nil {
		return nil, dbError(e)
	}
	if role != domain.RoleAthlete {
		return nil, domain.ErrForbidden
	}
	var tuition int64
	var mode domain.TuitionMode
	if e = tx.QueryRow(c, `SELECT g.tuition_amount,g.tuition_mode FROM gyms g JOIN users coach ON coach.id=g.coach_id AND coach.active=true WHERE g.id=$1 FOR UPDATE OF g`, gymID).Scan(&tuition, &mode); e != nil {
		return nil, dbError(e)
	}
	if mode != domain.TuitionMonthly {
		return nil, domain.ErrInvalid
	}
	var startYear, startMonth int
	if e = tx.QueryRow(c, `SELECT billing_start_year,billing_start_month FROM gym_athletes WHERE gym_id=$1 AND athlete_id=$2 FOR UPDATE`, gymID, athleteID).Scan(&startYear, &startMonth); e != nil {
		return nil, dbError(e)
	}
	if e = clearStaleCreatedPaymentOrders(c, tx, athleteID, gymID, now); e != nil {
		return nil, e
	}
	if existing, err := pendingPaymentOrder(c, tx, athleteID, gymID); err != nil {
		return nil, err
	} else if existing != nil {
		if err = tx.Commit(c); err != nil {
			return nil, dbError(err)
		}
		return existing, nil
	}
	var creating bool
	if e = tx.QueryRow(c, `SELECT EXISTS(SELECT 1 FROM payment_orders WHERE athlete_id=$1 AND gym_id=$2 AND status='created')`, athleteID, gymID).Scan(&creating); e != nil {
		return nil, e
	}
	if creating {
		return nil, domain.ErrConflict
	}
	rows, e := tx.Query(c, `SELECT billing_year,billing_month FROM tuition_payments WHERE gym_id=$1 AND athlete_id=$2 AND status='succeeded' AND billing_year IS NOT NULL`, gymID, athleteID)
	if e != nil {
		return nil, e
	}
	paid := map[int]bool{}
	for rows.Next() {
		var y, m int
		if e = rows.Scan(&y, &m); e != nil {
			rows.Close()
			return nil, e
		}
		paid[y*12+m] = true
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	v := &domain.PaymentOrder{ID: uuid.NewString(), AthleteID: athleteID, GymID: gymID, Kind: "monthly", BillingYear: currentYear, BillingMonth: currentMonth, Status: domain.PaymentOrderCreated, ExpiresAt: expires, CreatedAt: now, UpdatedAt: now}
	months := []domain.PaymentOrderMonth{}
	for serial := startYear*12 + startMonth; serial <= currentYear*12+currentMonth; serial++ {
		if paid[serial] {
			continue
		}
		y, m := (serial-1)/12, (serial-1)%12+1
		if tuition > 0 && v.Tuition > math.MaxInt64-tuition {
			return nil, domain.ErrPaymentAmount
		}
		months = append(months, domain.PaymentOrderMonth{BillingYear: y, BillingMonth: m, Amount: tuition})
		v.Tuition += tuition
	}
	if len(months) == 0 {
		return nil, domain.ErrConflict
	}
	if v.Tuition <= 1000 {
		return nil, domain.ErrPaymentAmount
	}
	v.Total = v.Tuition
	if _, e = tx.Exec(c, `INSERT INTO payment_orders(id,athlete_id,gym_id,kind,tuition_amount,platform_fee,total_amount,billing_year,billing_month,status,expires_at,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, v.ID, v.AthleteID, v.GymID, v.Kind, v.Tuition, 0, v.Total, v.BillingYear, v.BillingMonth, v.Status, v.ExpiresAt, v.CreatedAt, v.UpdatedAt); e != nil {
		return nil, dbError(e)
	}
	for _, month := range months {
		if _, e = tx.Exec(c, `INSERT INTO payment_order_months(order_id,billing_year,billing_month,amount) VALUES($1,$2,$3,$4)`, v.ID, month.BillingYear, month.BillingMonth, month.Amount); e != nil {
			return nil, dbError(e)
		}
	}
	if e = tx.Commit(c); e != nil {
		return nil, dbError(e)
	}
	return v, nil
}

func (s *Store) SetPaymentOrderTrack(c context.Context, id string, trackID int64, now time.Time) error {
	tag, e := s.DB.Exec(c, `UPDATE payment_orders SET zibal_track_id=$2,status='pending',updated_at=$3 WHERE id=$1 AND status='created'`, id, trackID, now)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrInvalidState
	}
	return dbError(e)
}
func (s *Store) ReopenPaymentOrder(c context.Context, id string, now time.Time) error {
	tag, e := s.DB.Exec(c, `UPDATE payment_orders SET status='pending',failure_reason=NULL,updated_at=$2 WHERE id=$1 AND status='failed' AND zibal_track_id IS NOT NULL`, id, now)
	if e == nil && tag.RowsAffected() == 0 {
		return domain.ErrInvalidState
	}
	return dbError(e)
}
func (s *Store) FailPaymentOrder(c context.Context, id, reason string, now time.Time) error {
	_, e := s.DB.Exec(c, `UPDATE payment_orders SET status='failed',failure_reason=$2,updated_at=$3 WHERE id=$1 AND status<>'succeeded'`, id, reason, now)
	return dbError(e)
}
func (s *Store) GetPaymentOrder(c context.Context, id string) (*domain.PaymentOrder, error) {
	return scanPaymentOrder(s.DB.QueryRow(c, `SELECT `+paymentOrderCols+` FROM payment_orders WHERE id=$1`, id))
}
func (s *Store) GetPaymentOrderByTrack(c context.Context, trackID int64) (*domain.PaymentOrder, error) {
	return scanPaymentOrder(s.DB.QueryRow(c, `SELECT `+paymentOrderCols+` FROM payment_orders WHERE zibal_track_id=$1`, trackID))
}
func (s *Store) CompletePaymentOrder(c context.Context, id string, trackID, amount int64, refNumber string, completedAt time.Time) (*domain.PaymentOrder, error) {
	tx, e := s.DB.BeginTx(c, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(c)
	v, e := scanPaymentOrder(tx.QueryRow(c, `SELECT `+paymentOrderCols+` FROM payment_orders WHERE id=$1 FOR UPDATE`, id))
	if e != nil {
		return nil, e
	}
	if v.Status == domain.PaymentOrderSucceeded {
		return v, tx.Commit(c)
	}
	if (v.Status != domain.PaymentOrderPending && v.Status != domain.PaymentOrderFailed) || v.TrackID == nil || *v.TrackID != trackID || v.Total != amount || refNumber == "" {
		return nil, domain.ErrPaymentFailed
	}
	var coachID string
	var tuitionMode domain.TuitionMode
	if e = tx.QueryRow(c, `SELECT coach_id,tuition_mode FROM gyms WHERE id=$1 FOR UPDATE`, v.GymID).Scan(&coachID, &tuitionMode); e != nil {
		return nil, dbError(e)
	}
	if v.Kind == "registration" {
		billingYear, billingMonth := any(nil), any(nil)
		if tuitionMode == domain.TuitionMonthly {
			billingYear, billingMonth = v.BillingYear, v.BillingMonth
		}
		tracking := fmt.Sprintf("ZIBAL-%d", trackID)
		if _, e = tx.Exec(c, `INSERT INTO tuition_payments(id,athlete_id,gym_id,coach_id,tuition_amount,platform_fee,total_amount,status,tracking_id,billing_year,billing_month,payment_kind,created_at,payment_order_id) VALUES($1,$2,$3,$4,$5,$6,$7,'succeeded',$8,$9,$10,'registration',$11,$1)`, v.ID, v.AthleteID, v.GymID, coachID, v.Tuition, v.PlatformFee, v.Total, tracking, billingYear, billingMonth, completedAt); e != nil {
			return nil, dbError(e)
		}
		if v.Tuition > 0 {
			if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'tuition_payment',$4,'athlete tuition payment')`, coachID, domain.LedgerTuition, v.Tuition, v.ID); e != nil {
				return nil, dbError(e)
			}
		}
	} else {
		rows, e := tx.Query(c, `SELECT billing_year,billing_month,amount FROM payment_order_months WHERE order_id=$1 ORDER BY billing_year,billing_month FOR UPDATE`, v.ID)
		if e != nil {
			return nil, e
		}
		months := []domain.PaymentOrderMonth{}
		for rows.Next() {
			var m domain.PaymentOrderMonth
			if e = rows.Scan(&m.BillingYear, &m.BillingMonth, &m.Amount); e != nil {
				rows.Close()
				return nil, e
			}
			months = append(months, m)
		}
		if e = rows.Err(); e != nil {
			rows.Close()
			return nil, e
		}
		rows.Close()
		if len(months) == 0 {
			return nil, domain.ErrInvalidState
		}
		for _, m := range months {
			paymentID := uuid.NewString()
			tracking := fmt.Sprintf("ZIBAL-%d-%04d%02d", trackID, m.BillingYear, m.BillingMonth)
			if _, e = tx.Exec(c, `INSERT INTO tuition_payments(id,athlete_id,gym_id,coach_id,tuition_amount,platform_fee,total_amount,status,tracking_id,billing_year,billing_month,payment_kind,created_at,payment_order_id) VALUES($1,$2,$3,$4,$5,0,$5,'succeeded',$6,$7,$8,'monthly',$9,$10)`, paymentID, v.AthleteID, v.GymID, coachID, m.Amount, tracking, m.BillingYear, m.BillingMonth, completedAt, v.ID); e != nil {
				return nil, dbError(e)
			}
			if m.Amount > 0 {
				desc := fmt.Sprintf("monthly tuition %04d/%02d", m.BillingYear, m.BillingMonth)
				if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'tuition_payment',$4,$5)`, coachID, domain.LedgerTuition, m.Amount, paymentID, desc); e != nil {
					return nil, dbError(e)
				}
			}
		}
	}
	if _, e = tx.Exec(c, `UPDATE payment_orders SET status='succeeded',zibal_ref_number=$2,failure_reason=NULL,completed_at=$3,updated_at=$3 WHERE id=$1`, v.ID, refNumber, completedAt); e != nil {
		return nil, dbError(e)
	}
	v.Status = domain.PaymentOrderSucceeded
	v.RefNumber = nil
	if refNumber != "" {
		v.RefNumber = &refNumber
	}
	v.Failure = nil
	v.CompletedAt = &completedAt
	v.UpdatedAt = completedAt
	if e = tx.Commit(c); e != nil {
		return nil, dbError(e)
	}
	return v, nil
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
func (s *Store) ListGymAthletePayments(c context.Context, gymID, athleteID string) ([]domain.AthletePayment, error) {
	rows, e := s.DB.Query(c, `SELECT tp.id,tp.athlete_id,tp.gym_id,tp.coach_id,tp.tuition_amount,tp.platform_fee,tp.total_amount,tp.status,tp.tracking_id,COALESCE(tp.billing_year,0),COALESCE(tp.billing_month,0),tp.payment_kind,tp.created_at,g.name,concat_ws(' ',u.first_name,u.last_name) FROM tuition_payments tp JOIN gyms g ON g.id=tp.gym_id JOIN users u ON u.id=tp.coach_id WHERE tp.gym_id=$1 AND tp.athlete_id=$2 ORDER BY tp.created_at DESC,tp.billing_year DESC NULLS LAST,tp.billing_month DESC NULLS LAST`, gymID, athleteID)
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
	e := r.Scan(&v.ID, &v.CoachID, &v.GrossAmount, &v.FeeAmount, &v.NetAmount, &v.IBAN, &v.Status, &v.BankTrackingCode, &v.RejectionReason, &v.ReviewedBy, &v.CoachName, &v.ReviewedAt, &v.QueuedAt, &v.ScheduledAt, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(e)
}

const withdrawalCols = `wr.id,wr.coach_id,wr.gross_amount,wr.fee_amount,wr.net_amount,wr.iban,wr.status,wr.bank_tracking_code,wr.rejection_reason,wr.reviewed_by,concat(coach.first_name,' ',coach.last_name),wr.reviewed_at,wr.queued_at,wr.scheduled_completion_at,wr.completed_at,wr.created_at,wr.updated_at`

func (s *Store) ListWithdrawals(c context.Context, status *domain.WithdrawalStatus, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `SELECT `+withdrawalCols+` FROM withdrawal_requests wr JOIN users coach ON coach.id=wr.coach_id WHERE ($1::varchar IS NULL OR wr.status=$1) ORDER BY wr.created_at DESC LIMIT $2 OFFSET $3`, status, limit, offset)
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
	rows, e := s.DB.Query(c, `SELECT `+withdrawalCols+` FROM withdrawal_requests wr JOIN users coach ON coach.id=wr.coach_id WHERE wr.coach_id=$1 ORDER BY wr.created_at DESC LIMIT $2 OFFSET $3`, coach, limit, offset)
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
func (s *Store) ReviewWithdrawal(c context.Context, id, admin string, status domain.WithdrawalStatus, tracking, reason *string, scheduledAt *time.Time) error {
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
	if current != domain.WithdrawalUnderReview {
		return domain.ErrInvalidState
	}
	tag, e := tx.Exec(c, `UPDATE withdrawal_requests SET status=$2::varchar,bank_tracking_code=$3,rejection_reason=$4,reviewed_by=$5,reviewed_at=now(),queued_at=CASE WHEN $2::varchar='queued' THEN now() END,scheduled_completion_at=$6,updated_at=now() WHERE id=$1`, id, status, tracking, reason, admin, scheduledAt)
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

func (s *Store) CompleteDueWithdrawals(c context.Context, now time.Time) (int64, error) {
	tx, e := s.DB.Begin(c)
	if e != nil {
		return 0, e
	}
	defer tx.Rollback(c)
	rows, e := tx.Query(c, `SELECT id,coach_id,gross_amount,fee_amount,net_amount,COALESCE(reviewed_by,coach_id),bank_tracking_code FROM withdrawal_requests WHERE status='queued' AND scheduled_completion_at<=$1 ORDER BY scheduled_completion_at FOR UPDATE SKIP LOCKED`, now)
	if e != nil {
		return 0, e
	}
	type due struct {
		id, coach, admin, tracking string
		gross, fee, net            int64
	}
	items := []due{}
	for rows.Next() {
		var v due
		if e = rows.Scan(&v.id, &v.coach, &v.gross, &v.fee, &v.net, &v.admin, &v.tracking); e != nil {
			rows.Close()
			return 0, e
		}
		items = append(items, v)
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return 0, e
	}
	rows.Close()
	for _, v := range items {
		if _, e = tx.Exec(c, `UPDATE withdrawal_requests SET status='completed',completed_at=$2,updated_at=$2 WHERE id=$1`, v.id, now); e != nil {
			return 0, e
		}
		if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'withdrawal',$4,'release reserved amount after settlement'),(gen_random_uuid(),$1,$5,$6,'withdrawal',$4,'bank payout')`, v.coach, domain.LedgerWithdrawalRelease, v.gross, v.id, domain.LedgerWithdrawalDebit, -v.net); e != nil {
			return 0, dbError(e)
		}
		if v.fee > 0 {
			if _, e = tx.Exec(c, `INSERT INTO wallet_entries(id,coach_id,entry_type,amount,reference_type,reference_id,description) VALUES(gen_random_uuid(),$1,$2,$3,'withdrawal',$4,'ClubMax withdrawal fee')`, v.coach, domain.LedgerWithdrawalFee, -v.fee, v.id); e != nil {
				return 0, dbError(e)
			}
		}
		var messageID string
		if e = tx.QueryRow(c, `INSERT INTO messages(id,sender_id,recipient_id,audience,subject,body,created_at) VALUES(gen_random_uuid(),$1,$2,'direct','تکمیل واریز برداشت','درخواست برداشت شما تکمیل شد. کد پیگیری بانکی: ' || $3::text,$4) RETURNING id`, v.admin, v.coach, v.tracking, now).Scan(&messageID); e != nil {
			return 0, e
		}
		if _, e = tx.Exec(c, `INSERT INTO message_recipients(message_id,user_id) VALUES($1,$2)`, messageID, v.coach); e != nil {
			return 0, e
		}
	}
	if e = tx.Commit(c); e != nil {
		return 0, e
	}
	return int64(len(items)), nil
}
func (s *Store) ListLedger(c context.Context, coach string, limit, offset int) ([]domain.WalletEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, e := s.DB.Query(c, `WITH ledger AS (SELECT we.id,we.coach_id,we.entry_type,we.amount,we.reference_type,we.reference_id,CASE WHEN we.entry_type='tuition_credit' THEN concat('واریز شهریه توسط ',athlete.first_name,' ',athlete.last_name,' - ',g.name) WHEN we.entry_type='withdrawal_hold' THEN 'رزرو مبلغ درخواست برداشت' WHEN we.entry_type='withdrawal_release' THEN 'آزادسازی مبلغ رزروشده' WHEN we.entry_type='withdrawal_debit' THEN 'واریز بانکی به مربی' WHEN we.entry_type='withdrawal_fee' THEN 'کارمزد خدمات جی‌مکس' ELSE we.description END AS description,we.created_at,SUM(we.amount) OVER(ORDER BY we.created_at,we.id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS balance_after FROM wallet_entries we LEFT JOIN tuition_payments tp ON we.reference_type='tuition_payment' AND tp.id=we.reference_id LEFT JOIN users athlete ON athlete.id=tp.athlete_id LEFT JOIN gyms g ON g.id=tp.gym_id WHERE we.coach_id=$1) SELECT id,coach_id,entry_type,amount,reference_type,reference_id,description,balance_after,created_at FROM ledger ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, coach, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.WalletEntry{}
	for rows.Next() {
		var v domain.WalletEntry
		if e = rows.Scan(&v.ID, &v.CoachID, &v.Type, &v.Amount, &v.ReferenceType, &v.ReferenceID, &v.Description, &v.BalanceAfter, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) AdminOverview(c context.Context) (domain.AdminOverview, error) {
	var v domain.AdminOverview
	e := s.DB.QueryRow(c, `SELECT
		(SELECT count(*) FROM users),
		(SELECT count(*) FROM users WHERE role='coach'),
		(SELECT count(*) FROM users WHERE role='athlete'),
		(SELECT count(*) FROM gyms),
		(SELECT count(DISTINCT athlete_id) FROM gym_athletes),
		(SELECT count(*) FROM tuition_payments WHERE status='succeeded'),
		(SELECT COALESCE(sum(total_amount),0) FROM tuition_payments WHERE status='succeeded'),
		(SELECT COALESCE(sum(platform_fee),0) FROM tuition_payments WHERE status='succeeded'),
		(SELECT COALESCE(sum(fee_amount),0) FROM withdrawal_requests WHERE status='completed'),
		(SELECT count(*) FROM withdrawal_requests WHERE status IN ('under_review','queued'))`).Scan(
		&v.Users, &v.Coaches, &v.Athletes, &v.Gyms, &v.ActiveAthletes, &v.PaymentCount,
		&v.GrossPayments, &v.PlatformRevenue, &v.WithdrawalFees, &v.PendingWithdrawals)
	return v, e
}

func (s *Store) ListAdminGyms(c context.Context, search string, limit, offset int) ([]domain.AdminGym, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, e := s.DB.Query(c, `SELECT g.id,g.coach_id,g.name,g.province,g.city,g.address,g.phone,g.sport,g.sport_style,g.active_weekdays,g.start_time::text,g.end_time::text,g.tuition_mode,g.tuition_amount,g.capacity,g.created_at,g.updated_at,concat(coach.first_name,' ',coach.last_name),coach.phone,member_stats.active_athletes,payment_stats.payment_count,payment_stats.paid_tuition,payment_stats.last_payment_at FROM gyms g JOIN users coach ON coach.id=g.coach_id LEFT JOIN LATERAL (SELECT count(*) AS active_athletes FROM gym_athletes ga JOIN users athlete ON athlete.id=ga.athlete_id AND athlete.active WHERE ga.gym_id=g.id) member_stats ON true LEFT JOIN LATERAL (SELECT count(*) AS payment_count,COALESCE(sum(tp.tuition_amount),0) AS paid_tuition,max(tp.created_at) AS last_payment_at FROM tuition_payments tp WHERE tp.gym_id=g.id AND tp.status='succeeded') payment_stats ON true WHERE ($1='' OR g.name ILIKE '%'||$1||'%' OR concat(coach.first_name,' ',coach.last_name) ILIKE '%'||$1||'%' OR g.sport ILIKE '%'||$1||'%') ORDER BY g.created_at DESC LIMIT $2 OFFSET $3`, search, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.AdminGym{}
	for rows.Next() {
		var v domain.AdminGym
		if e = rows.Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CoachName, &v.CoachPhone, &v.ActiveAthletes, &v.PaymentCount, &v.PaidTuition, &v.LastPaymentAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) ListAdminPayments(c context.Context, search string, limit, offset int) ([]domain.AdminPayment, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, e := s.DB.Query(c, `SELECT tp.id,tp.athlete_id,tp.gym_id,tp.coach_id,tp.tuition_amount,tp.platform_fee,tp.total_amount,tp.status,tp.tracking_id,COALESCE(tp.billing_year,0),COALESCE(tp.billing_month,0),tp.payment_kind,tp.created_at,concat(athlete.first_name,' ',athlete.last_name),athlete.phone,g.name,concat(coach.first_name,' ',coach.last_name) FROM tuition_payments tp JOIN users athlete ON athlete.id=tp.athlete_id JOIN users coach ON coach.id=tp.coach_id JOIN gyms g ON g.id=tp.gym_id WHERE ($1='' OR tp.tracking_id ILIKE '%'||$1||'%' OR g.name ILIKE '%'||$1||'%' OR concat(athlete.first_name,' ',athlete.last_name) ILIKE '%'||$1||'%') ORDER BY tp.created_at DESC LIMIT $2 OFFSET $3`, search, limit, offset)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.AdminPayment{}
	for rows.Next() {
		var v domain.AdminPayment
		if e = rows.Scan(&v.ID, &v.AthleteID, &v.GymID, &v.CoachID, &v.Tuition, &v.PlatformFee, &v.Total, &v.Status, &v.TrackingID, &v.BillingYear, &v.BillingMonth, &v.PaymentKind, &v.CreatedAt, &v.AthleteName, &v.AthletePhone, &v.GymName, &v.CoachName); e != nil {
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

func (s *Store) CreateTicket(c context.Context, t *domain.SupportTicket, first *domain.SupportTicketMessage) error {
	tx, err := s.DB.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)
	err = tx.QueryRow(c, `INSERT INTO support_tickets(id,owner_id,category,subject,status,last_message_at,member_last_read_at,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$6,$6,$6) RETURNING number`, t.ID, t.OwnerID, t.Category, t.Subject, t.Status, t.CreatedAt).Scan(&t.Number)
	if err != nil {
		return dbError(err)
	}
	_, err = tx.Exec(c, `INSERT INTO support_ticket_messages(id,ticket_id,sender_id,body,created_at) VALUES($1,$2,$3,$4,$5)`, first.ID, first.TicketID, first.SenderID, first.Body, first.CreatedAt)
	if err != nil {
		return dbError(err)
	}
	return tx.Commit(c)
}

func (s *Store) ListTickets(c context.Context, f repository.TicketFilter) ([]domain.SupportTicket, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 30
	}
	rows, err := s.DB.Query(c, `SELECT t.id,t.number,t.owner_id,concat(u.first_name,' ',u.last_name),u.role,u.phone,t.category,t.subject,t.status,t.assigned_admin_id,
	CASE WHEN $2='admin' THEN t.admin_last_read_at IS NULL OR t.last_message_at>t.admin_last_read_at ELSE t.last_message_at>t.member_last_read_at END,
	t.last_message_at,t.closed_at,t.created_at,t.updated_at
	FROM support_tickets t JOIN users u ON u.id=t.owner_id
	WHERE ($2='admin' OR t.owner_id=$1) AND ($3='' OR t.status::text=$3) AND ($4='' OR t.category::text=$4)
	AND ($5='' OR t.subject ILIKE '%'||$5||'%' OR t.number::text=$5 OR concat(u.first_name,' ',u.last_name) ILIKE '%'||$5||'%')
	ORDER BY t.last_message_at DESC LIMIT $6 OFFSET $7`, f.ActorID, f.ActorRole, optionalTicketStatus(f.Status), optionalTicketCategory(f.Category), f.Search, f.Limit, f.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SupportTicket{}
	for rows.Next() {
		var t domain.SupportTicket
		if err = rows.Scan(&t.ID, &t.Number, &t.OwnerID, &t.OwnerName, &t.OwnerRole, &t.OwnerPhone, &t.Category, &t.Subject, &t.Status, &t.AssignedAdminID, &t.Unread, &t.LastMessageAt, &t.ClosedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func optionalTicketStatus(v *domain.TicketStatus) string {
	if v == nil {
		return ""
	}
	return string(*v)
}
func optionalTicketCategory(v *domain.TicketCategory) string {
	if v == nil {
		return ""
	}
	return string(*v)
}

func (s *Store) GetTicket(c context.Context, id string) (*domain.SupportTicket, error) {
	var t domain.SupportTicket
	err := s.DB.QueryRow(c, `SELECT t.id,t.number,t.owner_id,concat(u.first_name,' ',u.last_name),u.role,u.phone,t.category,t.subject,t.status,t.assigned_admin_id,false,t.last_message_at,t.closed_at,t.created_at,t.updated_at FROM support_tickets t JOIN users u ON u.id=t.owner_id WHERE t.id=$1`, id).Scan(&t.ID, &t.Number, &t.OwnerID, &t.OwnerName, &t.OwnerRole, &t.OwnerPhone, &t.Category, &t.Subject, &t.Status, &t.AssignedAdminID, &t.Unread, &t.LastMessageAt, &t.ClosedAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, dbError(err)
	}
	return &t, nil
}

func (s *Store) ListTicketMessages(c context.Context, ticket string) ([]domain.SupportTicketMessage, error) {
	rows, err := s.DB.Query(c, `SELECT m.id,m.ticket_id,m.sender_id,concat(u.first_name,' ',u.last_name),u.role,m.body,m.created_at FROM support_ticket_messages m JOIN users u ON u.id=m.sender_id WHERE m.ticket_id=$1 ORDER BY m.created_at,m.id`, ticket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.SupportTicketMessage{}
	for rows.Next() {
		var m domain.SupportTicketMessage
		if err = rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.SenderName, &m.SenderRole, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) AddTicketReply(c context.Context, m *domain.SupportTicketMessage, status domain.TicketStatus, assigned *string) error {
	tx, err := s.DB.Begin(c)
	if err != nil {
		return err
	}
	defer tx.Rollback(c)
	_, err = tx.Exec(c, `INSERT INTO support_ticket_messages(id,ticket_id,sender_id,body,created_at) VALUES($1,$2,$3,$4,$5)`, m.ID, m.TicketID, m.SenderID, m.Body, m.CreatedAt)
	if err != nil {
		return dbError(err)
	}
	tag, err := tx.Exec(c, `UPDATE support_tickets SET status=$2::varchar,assigned_admin_id=COALESCE($3,assigned_admin_id),last_message_at=$4,updated_at=$4,closed_at=NULL,
	member_last_read_at=CASE WHEN $2::varchar='waiting_admin' THEN $4 ELSE member_last_read_at END,
	admin_last_read_at=CASE WHEN $2::varchar='waiting_member' THEN $4 ELSE admin_last_read_at END WHERE id=$1`, m.TicketID, status, assigned, m.CreatedAt)
	if err != nil {
		return dbError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return tx.Commit(c)
}

func (s *Store) SetTicketStatus(c context.Context, id string, status domain.TicketStatus, admin *string, closed *time.Time) error {
	tag, err := s.DB.Exec(c, `UPDATE support_tickets SET status=$2,assigned_admin_id=COALESCE($3,assigned_admin_id),closed_at=$4,updated_at=now() WHERE id=$1`, id, status, admin, closed)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(err)
}

func (s *Store) MarkTicketRead(c context.Context, id string, role domain.Role) error {
	column := "member_last_read_at"
	if role == domain.RoleAdmin {
		column = "admin_last_read_at"
	}
	tag, err := s.DB.Exec(c, `UPDATE support_tickets SET `+column+`=now() WHERE id=$1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return dbError(err)
}

func (s *Store) CountUnreadTickets(c context.Context, actorID string, role domain.Role) (int64, error) {
	var count int64
	err := s.DB.QueryRow(c, `SELECT count(*) FROM support_tickets
		WHERE ($2='admin' AND (admin_last_read_at IS NULL OR last_message_at>admin_last_read_at))
		   OR ($2<>'admin' AND owner_id=$1 AND last_message_at>member_last_read_at)`, actorID, role).Scan(&count)
	return count, err
}

type MessageStore struct{ *Store }

func (m MessageStore) Create(c context.Context, v *domain.Message, r []domain.MessageRecipient) error {
	return m.CreateMessage(c, v, r)
}
