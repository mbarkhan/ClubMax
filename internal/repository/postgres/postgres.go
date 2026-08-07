package postgres

import (
	"context"
	"errors"
	"github.com/gymax/gymax/internal/domain"
	"github.com/gymax/gymax/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	return e
}

const userCols = `id,role,first_name,last_name,father_name,COALESCE(national_id,''),COALESCE(birth_date,DATE '0001-01-01'),province,city,address,COALESCE(education,''),field_of_study,COALESCE(gender,''),marital_status,description,phone,phone_verified,username,password_hash,has_medical_history,medical_history,blood_type,emergency_phone,emergency_relation,sports_insurance_expires_at,COALESCE(face_image_id,''),created_at,updated_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	u := new(domain.User)
	e := row.Scan(&u.ID, &u.Role, &u.FirstName, &u.LastName, &u.FatherName, &u.NationalID, &u.BirthDate, &u.Province, &u.City, &u.Address, &u.Education, &u.FieldOfStudy, &u.Gender, &u.MaritalStatus, &u.Description, &u.Phone, &u.PhoneVerified, &u.Username, &u.PasswordHash, &u.HasMedicalHistory, &u.MedicalHistory, &u.BloodType, &u.EmergencyPhone, &u.EmergencyRelation, &u.SportsInsuranceExpiresAt, &u.FaceImageID, &u.CreatedAt, &u.UpdatedAt)
	return u, dbError(e)
}
func (s *Store) Create(ctx context.Context, u *domain.User) error {
	_, e := s.DB.Exec(ctx, `INSERT INTO users(id,role,first_name,last_name,father_name,national_id,birth_date,province,city,address,education,field_of_study,gender,marital_status,description,phone,phone_verified,username,password_hash,has_medical_history,medical_history,blood_type,emergency_phone,emergency_relation,sports_insurance_expires_at,face_image_id) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,'0001-01-01'),$8,$9,$10,NULLIF($11,''),$12,NULLIF($13,''),$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,NULLIF($26,''))`, u.ID, u.Role, u.FirstName, u.LastName, u.FatherName, u.NationalID, u.BirthDate.Format("2006-01-02"), u.Province, u.City, u.Address, u.Education, u.FieldOfStudy, u.Gender, u.MaritalStatus, u.Description, u.Phone, u.PhoneVerified, u.Username, u.PasswordHash, u.HasMedicalHistory, u.MedicalHistory, u.BloodType, u.EmergencyPhone, u.EmergencyRelation, u.SportsInsuranceExpiresAt, u.FaceImageID)
	return dbError(e)
}
func (s *Store) Update(ctx context.Context, u *domain.User) error {
	tag, e := s.DB.Exec(ctx, `UPDATE users SET first_name=$2,last_name=$3,father_name=$4,province=$5,city=$6,address=$7,education=$8,field_of_study=$9,gender=$10,marital_status=$11,description=$12,has_medical_history=$13,medical_history=$14,blood_type=$15,emergency_phone=$16,emergency_relation=$17,sports_insurance_expires_at=$18,face_image_id=$19,updated_at=now() WHERE id=$1`, u.ID, u.FirstName, u.LastName, u.FatherName, u.Province, u.City, u.Address, u.Education, u.FieldOfStudy, u.Gender, u.MaritalStatus, u.Description, u.HasMedicalHistory, u.MedicalHistory, u.BloodType, u.EmergencyPhone, u.EmergencyRelation, u.SportsInsuranceExpiresAt, u.FaceImageID)
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
func (s *Store) GetByID(c context.Context, id string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE id=$1`, id))
}
func (s *Store) GetByPhone(c context.Context, v string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(c, `SELECT `+userCols+` FROM users WHERE phone=$1`, v))
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
	e := r.Scan(&v.ID, &v.CoachID, &v.Name, &v.Province, &v.City, &v.Address, &v.Phone, &v.Sport, &v.SportStyle, &v.ActiveWeekdays, &v.StartTime, &v.EndTime, &v.TuitionMode, &v.TuitionAmount, &v.CreatedAt, &v.UpdatedAt)
	return v, dbError(e)
}

const gymCols = `id,coach_id,name,province,city,address,phone,sport,sport_style,active_weekdays,start_time::text,end_time::text,tuition_mode,tuition_amount,created_at,updated_at`

func (s *Store) CreateGym(c context.Context, v *domain.Gym) error {
	_, e := s.DB.Exec(c, `INSERT INTO gyms(id,coach_id,name,province,city,address,phone,sport,sport_style,active_weekdays,start_time,end_time,tuition_mode,tuition_amount) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, v.ID, v.CoachID, v.Name, v.Province, v.City, v.Address, v.Phone, v.Sport, v.SportStyle, v.ActiveWeekdays, v.StartTime, v.EndTime, v.TuitionMode, v.TuitionAmount)
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
	_, e := s.DB.Exec(c, `INSERT INTO gym_athletes(gym_id,athlete_id,registered_at) VALUES($1,$2,$3)`, v.GymID, v.AthleteID, v.RegisteredAt)
	return dbError(e)
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
func (s *Store) CountGymImages(c context.Context, gym string) (int, error) {
	var n int
	e := s.DB.QueryRow(c, `SELECT count(*) FROM gym_images WHERE gym_id=$1`, gym).Scan(&n)
	return n, e
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
func (o GymStore) GetByID(c context.Context, id string) (*domain.Gym, error) {
	return o.GetGymByID(c, id)
}
func (o GymStore) Delete(c context.Context, id string) error { return o.DeleteGym(c, id) }
