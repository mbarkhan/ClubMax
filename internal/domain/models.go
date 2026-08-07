package domain

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCoach    Role = "coach"
	RoleAthlete  Role = "athlete"
	RoleOperator Role = "operator"
)

func (r Role) ValidForRegistration() bool { return r == RoleCoach || r == RoleAthlete }

type Education string
type Gender string
type MaritalStatus string
type BloodType string

type User struct {
	ID                       string        `json:"id"`
	Role                     Role          `json:"role"`
	FirstName                string        `json:"first_name"`
	LastName                 string        `json:"last_name"`
	FatherName               *string       `json:"father_name,omitempty"`
	NationalID               string        `json:"national_id"`
	BirthDate                time.Time     `json:"birth_date"`
	Province                 *string       `json:"province,omitempty"`
	City                     *string       `json:"city,omitempty"`
	Address                  *string       `json:"address,omitempty"`
	Education                Education     `json:"education"`
	FieldOfStudy             *string       `json:"field_of_study,omitempty"`
	Gender                   Gender        `json:"gender"`
	MaritalStatus            MaritalStatus `json:"marital_status"`
	Description              *string       `json:"description,omitempty"`
	Phone                    *string       `json:"phone,omitempty"`
	PhoneVerified            bool          `json:"phone_verified"`
	Username                 *string       `json:"username,omitempty"`
	PasswordHash             *string       `json:"-"`
	HasMedicalHistory        *bool         `json:"has_medical_history,omitempty"`
	MedicalHistory           *string       `json:"medical_history,omitempty"`
	BloodType                *BloodType    `json:"blood_type,omitempty"`
	EmergencyPhone           *string       `json:"emergency_phone,omitempty"`
	EmergencyRelation        *string       `json:"emergency_relation,omitempty"`
	SportsInsuranceExpiresAt *time.Time    `json:"sports_insurance_expires_at,omitempty"`
	FaceImageID              string        `json:"face_image_id"`
	CreatedAt                time.Time     `json:"created_at"`
	UpdatedAt                time.Time     `json:"updated_at"`
}

type DocumentKind string

const (
	DocumentCoaching  DocumentKind = "coaching"
	DocumentInsurance DocumentKind = "insurance"
	DocumentOther     DocumentKind = "other"
)

type UserDocument struct {
	ID, UserID           string
	Kind                 DocumentKind
	StorageKey, MIMEType string
	Size                 int64
	CreatedAt            time.Time
}

type TuitionMode string

const (
	TuitionMonthly    TuitionMode = "monthly"
	TuitionPerSession TuitionMode = "per_session"
	TuitionCourse     TuitionMode = "course"
)

type Gym struct {
	ID             string      `json:"id"`
	CoachID        string      `json:"coach_id"`
	Name           string      `json:"name"`
	Province       string      `json:"province"`
	City           string      `json:"city"`
	Address        string      `json:"address"`
	Phone          *string     `json:"phone,omitempty"`
	Sport          string      `json:"sport"`
	SportStyle     string      `json:"sport_style"`
	ActiveWeekdays []int       `json:"active_weekdays"`
	StartTime      string      `json:"start_time"`
	EndTime        string      `json:"end_time"`
	TuitionMode    TuitionMode `json:"tuition_mode"`
	TuitionAmount  int64       `json:"tuition_amount"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type GymImage struct {
	ID, GymID, StorageKey, MIMEType string
	Size                            int64
	CreatedAt                       time.Time
}
type GymAthlete struct {
	GymID        string    `json:"gym_id"`
	AthleteID    string    `json:"athlete_id"`
	RegisteredAt time.Time `json:"registered_at"`
}

type OTPPurpose string

const (
	OTPRegister    OTPPurpose = "register"
	OTPChangePhone OTPPurpose = "change_phone"
)

type OTPChallenge struct {
	ID, Phone, CodeHash string
	Purpose             OTPPurpose
	Role                *Role
	UserID              *string
	ExpiresAt           time.Time
	Attempts            int
	ConsumedAt          *time.Time
	CreatedAt           time.Time
}
