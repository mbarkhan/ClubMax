CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('admin','coach','athlete','operator');
CREATE TYPE document_kind AS ENUM ('coaching','insurance','other');
CREATE TYPE otp_purpose AS ENUM ('register','change_phone');
CREATE TYPE tuition_mode AS ENUM ('monthly','per_session','course');

CREATE TABLE users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), role user_role NOT NULL,
 first_name varchar(100) NOT NULL, last_name varchar(100) NOT NULL DEFAULT '', father_name varchar(100),
 national_id char(10), birth_date date, province varchar(100), city varchar(100), address text,
 education varchar(30), field_of_study varchar(150), gender varchar(20), marital_status varchar(20) NOT NULL DEFAULT 'unknown', description text,
 phone varchar(15) UNIQUE, phone_verified boolean NOT NULL DEFAULT false,
 username varchar(100) UNIQUE, password_hash text,
 has_medical_history boolean, medical_history text, blood_type varchar(3), emergency_phone varchar(15), emergency_relation varchar(30),
 sports_insurance_expires_at date, face_image_id text,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CONSTRAINT admin_credentials CHECK (role <> 'admin' OR (username IS NOT NULL AND password_hash IS NOT NULL)),
 CONSTRAINT member_phone CHECK (role NOT IN ('coach','athlete') OR phone IS NOT NULL),
 CONSTRAINT athlete_fields CHECK (role <> 'athlete' OR (has_medical_history IS NOT NULL AND blood_type IS NOT NULL AND emergency_phone IS NOT NULL AND emergency_relation IS NOT NULL))
);
CREATE UNIQUE INDEX users_national_id_unique ON users(national_id) WHERE national_id IS NOT NULL;

CREATE TABLE user_documents (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 kind document_kind NOT NULL, storage_key text NOT NULL UNIQUE, mime_type varchar(100) NOT NULL, size_bytes integer NOT NULL CHECK(size_bytes BETWEEN 1 AND 512000), created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE otp_challenges (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), phone varchar(15) NOT NULL, code_hash text NOT NULL,
 purpose otp_purpose NOT NULL, role user_role, user_id uuid REFERENCES users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL, attempts smallint NOT NULL DEFAULT 0, consumed_at timestamptz, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX otp_lookup ON otp_challenges(phone,purpose,created_at DESC);

CREATE TABLE gyms (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), coach_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 name varchar(200) NOT NULL, province varchar(100) NOT NULL, city varchar(100) NOT NULL, address text NOT NULL, phone varchar(15),
 sport varchar(100) NOT NULL, sport_style varchar(150) NOT NULL, active_weekdays smallint[] NOT NULL,
 start_time time NOT NULL, end_time time NOT NULL, tuition_mode tuition_mode NOT NULL, tuition_amount bigint NOT NULL CHECK(tuition_amount >= 0),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK(cardinality(active_weekdays) BETWEEN 1 AND 7), CHECK(start_time < end_time)
);
CREATE INDEX gyms_coach_id ON gyms(coach_id);
CREATE TABLE gym_images (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), gym_id uuid NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
 storage_key text NOT NULL UNIQUE, mime_type varchar(100) NOT NULL, size_bytes integer NOT NULL CHECK(size_bytes > 0), created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE gym_athletes (
 gym_id uuid NOT NULL REFERENCES gyms(id) ON DELETE CASCADE, athlete_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 registered_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(gym_id,athlete_id)
);

CREATE OR REPLACE FUNCTION enforce_gym_roles() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='gyms' AND NOT EXISTS(SELECT 1 FROM users WHERE id=NEW.coach_id AND role='coach') THEN RAISE EXCEPTION 'gym owner must be coach'; END IF;
 IF TG_TABLE_NAME='gym_athletes' AND NOT EXISTS(SELECT 1 FROM users WHERE id=NEW.athlete_id AND role='athlete') THEN RAISE EXCEPTION 'member must be athlete'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER gyms_role_guard BEFORE INSERT OR UPDATE ON gyms FOR EACH ROW EXECUTE FUNCTION enforce_gym_roles();
CREATE TRIGGER athletes_role_guard BEFORE INSERT OR UPDATE ON gym_athletes FOR EACH ROW EXECUTE FUNCTION enforce_gym_roles();
