DROP TABLE IF EXISTS withdrawal_requests, wallet_entries, tuition_payments, registration_fee_policy, withdrawal_policy, message_recipients, messages, gym_athletes, gym_images, gyms, otp_challenges, user_documents, users CASCADE;
DROP FUNCTION IF EXISTS enforce_gym_coach_role();
DROP FUNCTION IF EXISTS enforce_gym_athlete_role();
DROP FUNCTION IF EXISTS enforce_gym_limits();
DROP TYPE IF EXISTS tuition_mode, otp_purpose, document_kind, user_role;
