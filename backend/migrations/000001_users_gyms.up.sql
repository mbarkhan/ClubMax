CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('admin','coach','athlete','operator');
CREATE TYPE document_kind AS ENUM ('coaching','insurance','other','profile');
CREATE TYPE otp_purpose AS ENUM ('register','change_phone','login','delete_account');
CREATE TYPE tuition_mode AS ENUM ('monthly','per_session','course');

CREATE TABLE users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), role user_role NOT NULL,
 first_name varchar(100) NOT NULL, last_name varchar(100) NOT NULL DEFAULT '', father_name varchar(100),
 national_id char(10), birth_date date, province varchar(100), city varchar(100), address text,
 education varchar(30), field_of_study varchar(150), gender varchar(20), marital_status varchar(20) NOT NULL DEFAULT 'unknown', description text,
 phone varchar(15) UNIQUE, phone_verified boolean NOT NULL DEFAULT false, active boolean NOT NULL DEFAULT true, bank_iban varchar(26),
 username varchar(100) UNIQUE, password_hash text,
 has_medical_history boolean, medical_history text, blood_type varchar(3), emergency_phone varchar(15), emergency_relation varchar(30),
 sports_insurance_expires_at date, face_image_id text,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CONSTRAINT admin_credentials CHECK (role <> 'admin' OR (username IS NOT NULL AND password_hash IS NOT NULL)),
 CONSTRAINT member_credentials CHECK (role NOT IN ('coach','athlete') OR (phone IS NOT NULL AND password_hash IS NOT NULL)),
 CONSTRAINT athlete_fields CHECK (role <> 'athlete' OR (has_medical_history IS NOT NULL AND blood_type IS NOT NULL AND emergency_phone IS NOT NULL AND emergency_relation IS NOT NULL))
 ,CONSTRAINT users_bank_iban_format CHECK (bank_iban IS NULL OR (role IN ('coach','athlete') AND bank_iban ~ '^IR[0-9]{24}$'))
);
CREATE UNIQUE INDEX users_national_id_unique ON users(national_id) WHERE national_id IS NOT NULL;
CREATE UNIQUE INDEX users_bank_iban_unique ON users(bank_iban) WHERE bank_iban IS NOT NULL;

CREATE TABLE user_documents (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 kind document_kind NOT NULL, storage_key text NOT NULL UNIQUE, mime_type varchar(100) NOT NULL, size_bytes integer NOT NULL CHECK(size_bytes BETWEEN 1 AND 512000), created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX user_insurance_document ON user_documents(user_id) WHERE kind='insurance';
CREATE UNIQUE INDEX user_profile_document ON user_documents(user_id) WHERE kind='profile';

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
 start_time time NOT NULL, end_time time NOT NULL, tuition_mode tuition_mode NOT NULL, tuition_amount bigint NOT NULL CHECK(tuition_amount >= 0), capacity bigint NOT NULL DEFAULT 0 CHECK(capacity >= 0),
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
 billing_start_year smallint NOT NULL CHECK(billing_start_year BETWEEN 1300 AND 1600), billing_start_month smallint NOT NULL CHECK(billing_start_month BETWEEN 1 AND 12),
 registered_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(gym_id,athlete_id)
);

CREATE TABLE messages (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), sender_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 gym_id uuid REFERENCES gyms(id) ON DELETE CASCADE, recipient_id uuid REFERENCES users(id) ON DELETE CASCADE,
 audience varchar(20) NOT NULL CHECK (audience IN ('all','gym_members','direct')),
 subject varchar(200) NOT NULL, body text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
 CHECK ((audience='all' AND gym_id IS NULL AND recipient_id IS NULL) OR (audience='gym_members' AND gym_id IS NOT NULL AND recipient_id IS NULL) OR (audience='direct' AND recipient_id IS NOT NULL AND gym_id IS NULL))
);
CREATE TABLE message_recipients (
 message_id uuid NOT NULL REFERENCES messages(id) ON DELETE CASCADE, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 read_at timestamptz, PRIMARY KEY(message_id,user_id)
);

CREATE TABLE withdrawal_policy (
 id boolean PRIMARY KEY DEFAULT true CHECK (id=true), min_amount bigint NOT NULL CHECK(min_amount>0), max_amount bigint NOT NULL CHECK(max_amount>=min_amount), fee_bps integer NOT NULL CHECK(fee_bps BETWEEN 0 AND 10000), updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO withdrawal_policy(min_amount,max_amount,fee_bps) VALUES (100000, 1000000000, 200);
CREATE TABLE registration_fee_policy (
 id boolean PRIMARY KEY DEFAULT true CHECK (id=true), fee_amount bigint NOT NULL CHECK(fee_amount>=0), updated_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO registration_fee_policy(fee_amount) VALUES (0);
CREATE TABLE tuition_payments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), athlete_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, gym_id uuid NOT NULL REFERENCES gyms(id) ON DELETE CASCADE, coach_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 tuition_amount bigint NOT NULL CHECK(tuition_amount>=0), platform_fee bigint NOT NULL CHECK(platform_fee>=0), total_amount bigint NOT NULL CHECK(total_amount=tuition_amount+platform_fee),
 status varchar(20) NOT NULL CHECK(status IN ('succeeded')), tracking_id varchar(100) NOT NULL UNIQUE,
 billing_year smallint CHECK(billing_year BETWEEN 1300 AND 1600), billing_month smallint CHECK(billing_month BETWEEN 1 AND 12), payment_kind varchar(20) NOT NULL CHECK(payment_kind IN ('registration','monthly')),
 created_at timestamptz NOT NULL DEFAULT now(), CHECK((billing_year IS NULL)=(billing_month IS NULL))
);
CREATE UNIQUE INDEX tuition_payment_month_unique ON tuition_payments(athlete_id,gym_id,billing_year,billing_month) WHERE billing_year IS NOT NULL;
CREATE TABLE wallet_entries (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), coach_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, entry_type varchar(30) NOT NULL,
 amount bigint NOT NULL CHECK(amount<>0), reference_type varchar(50) NOT NULL, reference_id uuid NOT NULL, description text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(reference_type,reference_id,entry_type)
);
CREATE INDEX wallet_entries_coach_created ON wallet_entries(coach_id,created_at DESC);
CREATE TABLE withdrawal_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), coach_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, gross_amount bigint NOT NULL CHECK(gross_amount>0), fee_amount bigint NOT NULL CHECK(fee_amount>=0), net_amount bigint NOT NULL CHECK(net_amount>0),
 iban varchar(26) NOT NULL CHECK(iban ~ '^IR[0-9]{24}$'), status varchar(20) NOT NULL CHECK(status IN ('under_review','queued','completed','rejected')),
 bank_tracking_code varchar(100), rejection_reason text, reviewed_by uuid REFERENCES users(id) ON DELETE SET NULL,
 reviewed_at timestamptz, queued_at timestamptz, scheduled_completion_at timestamptz, completed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK ((status IN ('queued','completed')) = (bank_tracking_code IS NOT NULL)),
 CHECK (status <> 'rejected' OR rejection_reason IS NOT NULL)
);
CREATE INDEX withdrawal_requests_status_created ON withdrawal_requests(status,created_at DESC);

CREATE OR REPLACE FUNCTION enforce_gym_coach_role() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NOT EXISTS(SELECT 1 FROM users WHERE id=NEW.coach_id AND role='coach') THEN RAISE EXCEPTION 'gym owner must be coach'; END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE FUNCTION enforce_gym_athlete_role() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NOT EXISTS(SELECT 1 FROM users WHERE id=NEW.athlete_id AND role='athlete') THEN RAISE EXCEPTION 'member must be athlete'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER gyms_role_guard BEFORE INSERT OR UPDATE ON gyms FOR EACH ROW EXECUTE FUNCTION enforce_gym_coach_role();
CREATE TRIGGER athletes_role_guard BEFORE INSERT OR UPDATE ON gym_athletes FOR EACH ROW EXECUTE FUNCTION enforce_gym_athlete_role();

CREATE OR REPLACE FUNCTION enforce_gym_limits() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='gyms' THEN
  PERFORM pg_advisory_xact_lock(hashtext('gym-limit:' || NEW.coach_id::text));
  IF (SELECT count(*) FROM gyms WHERE coach_id=NEW.coach_id) >= 36 THEN RAISE EXCEPTION 'GYM_LIMIT_REACHED'; END IF;
 END IF;
 IF TG_TABLE_NAME='gym_images' THEN
  PERFORM pg_advisory_xact_lock(hashtext('gym-image-limit:' || NEW.gym_id::text));
  IF (SELECT count(*) FROM gym_images WHERE gym_id=NEW.gym_id) >= 3 THEN RAISE EXCEPTION 'GYM_IMAGE_LIMIT_REACHED'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER gyms_limit_guard BEFORE INSERT ON gyms FOR EACH ROW EXECUTE FUNCTION enforce_gym_limits();
CREATE TRIGGER gym_images_limit_guard BEFORE INSERT ON gym_images FOR EACH ROW EXECUTE FUNCTION enforce_gym_limits();
