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

const (
	EducationBelowDiploma Education     = "below_diploma"
	EducationDiploma      Education     = "diploma"
	EducationAssociate    Education     = "associate"
	EducationBachelor     Education     = "bachelor"
	EducationMaster       Education     = "master"
	EducationDoctorate    Education     = "doctorate"
	GenderMale            Gender        = "male"
	GenderFemale          Gender        = "female"
	GenderOther           Gender        = "other"
	MaritalUnknown        MaritalStatus = "unknown"
	MaritalSingle         MaritalStatus = "single"
	MaritalMarried        MaritalStatus = "married"
)

type User struct {
	ID                       string        `json:"id"`
	Role                     Role          `json:"role"`
	FirstName                string        `json:"first_name"`
	LastName                 string        `json:"last_name"`
	FatherName               *string       `json:"father_name,omitempty"`
	NationalID               string        `json:"national_id"`
	BankIBAN                 *string       `json:"bank_iban,omitempty"`
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
	Active                   bool          `json:"active"`
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
	DocumentProfile   DocumentKind = "profile"
)

type UserDocument struct {
	ID         string       `json:"id"`
	UserID     string       `json:"user_id"`
	Kind       DocumentKind `json:"kind"`
	StorageKey string       `json:"storage_key"`
	MIMEType   string       `json:"mime_type"`
	Size       int64        `json:"size_bytes"`
	CreatedAt  time.Time    `json:"created_at"`
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
	Capacity       int64       `json:"capacity"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type GymImage struct {
	ID         string    `json:"id"`
	GymID      string    `json:"gym_id"`
	StorageKey string    `json:"storage_key"`
	MIMEType   string    `json:"mime_type"`
	Size       int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
}
type GymAthlete struct {
	GymID             string    `json:"gym_id"`
	AthleteID         string    `json:"athlete_id"`
	BillingStartYear  int       `json:"billing_start_year"`
	BillingStartMonth int       `json:"billing_start_month"`
	RegisteredAt      time.Time `json:"registered_at"`
}
type GymAthleteProfile struct {
	ID                string    `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Phone             *string   `json:"phone,omitempty"`
	Age               int       `json:"age"`
	TuitionStatus     string    `json:"tuition_status"`
	UnpaidMonths      int       `json:"unpaid_months"`
	OutstandingAmount int64     `json:"outstanding_amount"`
	RegisteredAt      time.Time `json:"registered_at"`
	LastWeightKG      *float64  `json:"last_weight_kg,omitempty"`
}

type GymAttendanceRecord struct {
	AttendanceDate time.Time `json:"attendance_date"`
	AthleteID      string    `json:"athlete_id"`
	Present        bool      `json:"present"`
}

// GymAttendanceAthlete is a gym roster row annotated with attendance for one date.
type GymAttendanceAthlete struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Present   bool   `json:"present"`
}

type GymJoinRequestStatus string

const (
	GymJoinPending  GymJoinRequestStatus = "pending"
	GymJoinApproved GymJoinRequestStatus = "approved"
	GymJoinRejected GymJoinRequestStatus = "rejected"
)

type GymJoinRequest struct {
	ID              string               `json:"id"`
	GymID           string               `json:"gym_id"`
	GymName         string               `json:"gym_name"`
	CoachID         string               `json:"coach_id"`
	AthleteID       string               `json:"athlete_id"`
	AthleteName     string               `json:"athlete_name"`
	AthletePhone    *string              `json:"athlete_phone,omitempty"`
	Status          GymJoinRequestStatus `json:"status"`
	RejectionReason *string              `json:"rejection_reason,omitempty"`
	ReviewedAt      *time.Time           `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type BodyMeasurement struct {
	ID         string    `json:"id"`
	AthleteID  string    `json:"athlete_id"`
	HeightCM   float64   `json:"height_cm"`
	WeightKG   float64   `json:"weight_kg"`
	BMI        float64   `json:"bmi"`
	AgeYears   int       `json:"age_years"`
	Percentile *float64  `json:"bmi_percentile,omitempty"`
	HealthyMin float64   `json:"healthy_bmi_min,omitempty"`
	HealthyMax float64   `json:"healthy_bmi_max,omitempty"`
	Reference  string    `json:"reference"`
	Status     string    `json:"status"`
	MeasuredAt time.Time `json:"measured_at"`
}

type BMIOverview struct {
	Measurements []BodyMeasurement `json:"measurements"`
	Latest       *BodyMeasurement  `json:"latest,omitempty"`
	CanRecord    bool              `json:"can_record"`
	NextAllowed  *time.Time        `json:"next_allowed_at,omitempty"`
}

type GymCatalogItem struct {
	Gym
	CoachName                  string                `json:"coach_name"`
	Joined                     bool                  `json:"joined"`
	RegistrationPaid           bool                  `json:"registration_paid"`
	JoinRequestStatus          *GymJoinRequestStatus `json:"join_request_status,omitempty"`
	JoinRequestRejectionReason *string               `json:"join_request_rejection_reason,omitempty"`
	CapacityRemaining          *int64                `json:"capacity_remaining"`
	UnpaidMonths               int                   `json:"unpaid_months"`
	OutstandingAmount          int64                 `json:"outstanding_amount"`
	MonthlyPaymentDue          bool                  `json:"monthly_payment_due"`
}

// GymShowcaseItem contains only the public fields used by the home-page banner.
type GymShowcaseItem struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	CoachName         string  `json:"coach_name"`
	Sport             string  `json:"sport"`
	SportStyle        string  `json:"sport_style"`
	StartTime         string  `json:"start_time"`
	EndTime           string  `json:"end_time"`
	ImageStorageKey   *string `json:"image_storage_key,omitempty"`
	CapacityRemaining *int64  `json:"capacity_remaining"`
}

type RegistrationFeePolicy struct {
	FeeAmount int64     `json:"fee_amount"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TuitionPayment struct {
	ID           string    `json:"id"`
	AthleteID    string    `json:"athlete_id"`
	GymID        string    `json:"gym_id"`
	CoachID      string    `json:"coach_id"`
	Tuition      int64     `json:"tuition_amount"`
	PlatformFee  int64     `json:"platform_fee"`
	Total        int64     `json:"total_amount"`
	Status       string    `json:"status"`
	TrackingID   string    `json:"tracking_id"`
	BillingYear  int       `json:"billing_year,omitempty"`
	BillingMonth int       `json:"billing_month,omitempty"`
	PaymentKind  string    `json:"payment_kind"`
	CreatedAt    time.Time `json:"created_at"`
}

type MonthlyTuitionPayment struct {
	TrackingID string `json:"tracking_id"`
	PaidMonths int    `json:"paid_months"`
	PaidAmount int64  `json:"paid_amount"`
	FromYear   int    `json:"from_year"`
	FromMonth  int    `json:"from_month"`
	ToYear     int    `json:"to_year"`
	ToMonth    int    `json:"to_month"`
}

type PaymentOrderStatus string

const (
	PaymentOrderCreated   PaymentOrderStatus = "created"
	PaymentOrderPending   PaymentOrderStatus = "pending"
	PaymentOrderSucceeded PaymentOrderStatus = "succeeded"
	PaymentOrderFailed    PaymentOrderStatus = "failed"
)

// PaymentOrder is the immutable invoice sent to the payment provider. Amounts
// are stored in rials, matching both the existing ClubMax ledger and Zibal.
type PaymentOrder struct {
	ID           string             `json:"id"`
	AthleteID    string             `json:"-"`
	GymID        string             `json:"gym_id"`
	Kind         string             `json:"kind"`
	Tuition      int64              `json:"tuition_amount"`
	PlatformFee  int64              `json:"platform_fee"`
	Total        int64              `json:"total_amount"`
	BillingYear  int                `json:"billing_year,omitempty"`
	BillingMonth int                `json:"billing_month,omitempty"`
	Status       PaymentOrderStatus `json:"status"`
	TrackID      *int64             `json:"-"`
	RefNumber    *string            `json:"ref_number,omitempty"`
	Failure      *string            `json:"-"`
	ExpiresAt    time.Time          `json:"-"`
	CompletedAt  *time.Time         `json:"completed_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type PaymentOrderMonth struct {
	BillingYear  int   `json:"billing_year"`
	BillingMonth int   `json:"billing_month"`
	Amount       int64 `json:"amount"`
}

type PaymentStart struct {
	OrderID    string `json:"order_id"`
	PaymentURL string `json:"payment_url"`
}

type AthletePayment struct {
	TuitionPayment
	GymName   string `json:"gym_name"`
	CoachName string `json:"coach_name"`
}

type OTPPurpose string

const (
	OTPRegister      OTPPurpose = "register"
	OTPChangePhone   OTPPurpose = "change_phone"
	OTPLogin         OTPPurpose = "login"
	OTPDelete        OTPPurpose = "delete_account"
	OTPPasswordReset OTPPurpose = "password_reset"
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

type MessageAudience string

const (
	AudienceAll        MessageAudience = "all"
	AudienceGymMembers MessageAudience = "gym_members"
	AudienceDirect     MessageAudience = "direct"
)

type Message struct {
	ID          string          `json:"id"`
	SenderID    string          `json:"sender_id"`
	GymID       *string         `json:"gym_id,omitempty"`
	RecipientID *string         `json:"recipient_id,omitempty"`
	Audience    MessageAudience `json:"audience"`
	Subject     string          `json:"subject"`
	Body        string          `json:"body"`
	CreatedAt   time.Time       `json:"created_at"`
}
type MessageRecipient struct {
	MessageID, UserID string
	ReadAt            *time.Time
}

type TicketCategory string
type TicketStatus string

const (
	TicketCategoryPayment   TicketCategory = "payment"
	TicketCategoryAccount   TicketCategory = "account"
	TicketCategoryGym       TicketCategory = "gym"
	TicketCategoryTechnical TicketCategory = "technical"
	TicketCategoryOther     TicketCategory = "other"
	TicketWaitingAdmin      TicketStatus   = "waiting_admin"
	TicketWaitingMember     TicketStatus   = "waiting_member"
	TicketClosed            TicketStatus   = "closed"
)

type SupportTicket struct {
	ID              string         `json:"id"`
	Number          int64          `json:"number"`
	OwnerID         string         `json:"owner_id"`
	OwnerName       string         `json:"owner_name,omitempty"`
	OwnerRole       Role           `json:"owner_role,omitempty"`
	OwnerPhone      *string        `json:"owner_phone,omitempty"`
	Category        TicketCategory `json:"category"`
	Subject         string         `json:"subject"`
	Status          TicketStatus   `json:"status"`
	AssignedAdminID *string        `json:"assigned_admin_id,omitempty"`
	Unread          bool           `json:"unread"`
	LastMessageAt   time.Time      `json:"last_message_at"`
	ClosedAt        *time.Time     `json:"closed_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type SupportTicketMessage struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	SenderRole Role      `json:"sender_role"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

type SupportTicketDetail struct {
	Ticket   SupportTicket          `json:"ticket"`
	Messages []SupportTicketMessage `json:"messages"`
}

type LedgerEntryType string

const (
	LedgerTuition           LedgerEntryType = "tuition_credit"
	LedgerWithdrawalHold    LedgerEntryType = "withdrawal_hold"
	LedgerWithdrawalRelease LedgerEntryType = "withdrawal_release"
	LedgerWithdrawalDebit   LedgerEntryType = "withdrawal_debit"
	LedgerWithdrawalFee     LedgerEntryType = "withdrawal_fee"
)

type WalletEntry struct {
	ID            string          `json:"id"`
	CoachID       string          `json:"coach_id"`
	Type          LedgerEntryType `json:"type"`
	Amount        int64           `json:"amount"`
	ReferenceType string          `json:"reference_type"`
	ReferenceID   string          `json:"reference_id"`
	Description   string          `json:"description"`
	BalanceAfter  int64           `json:"balance_after"`
	CreatedAt     time.Time       `json:"created_at"`
}
type WithdrawalStatus string

const (
	WithdrawalUnderReview WithdrawalStatus = "under_review"
	WithdrawalQueued      WithdrawalStatus = "queued"
	WithdrawalCompleted   WithdrawalStatus = "completed"
	WithdrawalRejected    WithdrawalStatus = "rejected"
)

type WithdrawalRequest struct {
	ID               string           `json:"id"`
	CoachID          string           `json:"coach_id"`
	GrossAmount      int64            `json:"gross_amount"`
	FeeAmount        int64            `json:"fee_amount"`
	NetAmount        int64            `json:"net_amount"`
	IBAN             string           `json:"iban"`
	Status           WithdrawalStatus `json:"status"`
	BankTrackingCode *string          `json:"bank_tracking_code,omitempty"`
	RejectionReason  *string          `json:"rejection_reason,omitempty"`
	ReviewedBy       *string          `json:"reviewed_by,omitempty"`
	CoachName        string           `json:"coach_name,omitempty"`
	ReviewedAt       *time.Time       `json:"reviewed_at,omitempty"`
	QueuedAt         *time.Time       `json:"queued_at,omitempty"`
	ScheduledAt      *time.Time       `json:"scheduled_completion_at,omitempty"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}
type WithdrawalPolicy struct {
	MinAmount int64     `json:"min_amount"`
	MaxAmount int64     `json:"max_amount"`
	FeeBPS    int64     `json:"fee_bps"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BackupPolicy struct {
	ScheduleMinutes []int32   `json:"schedule_minutes"`
	Retention       int       `json:"retention"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminOverview struct {
	Users              int64 `json:"users"`
	Coaches            int64 `json:"coaches"`
	Athletes           int64 `json:"athletes"`
	Gyms               int64 `json:"gyms"`
	ActiveAthletes     int64 `json:"active_athletes"`
	PaymentCount       int64 `json:"payment_count"`
	GrossPayments      int64 `json:"gross_payments"`
	PlatformRevenue    int64 `json:"platform_revenue"`
	WithdrawalFees     int64 `json:"withdrawal_fees"`
	PendingWithdrawals int64 `json:"pending_withdrawals"`
}

type AdminGym struct {
	Gym
	CoachName      string     `json:"coach_name"`
	CoachPhone     *string    `json:"coach_phone,omitempty"`
	ActiveAthletes int64      `json:"active_athletes"`
	PaymentCount   int64      `json:"payment_count"`
	PaidTuition    int64      `json:"paid_tuition"`
	LastPaymentAt  *time.Time `json:"last_payment_at,omitempty"`
}

type AdminPayment struct {
	TuitionPayment
	AthleteName  string  `json:"athlete_name"`
	AthletePhone *string `json:"athlete_phone,omitempty"`
	GymName      string  `json:"gym_name"`
	CoachName    string  `json:"coach_name"`
}
