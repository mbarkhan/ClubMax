DROP TRIGGER IF EXISTS protect_default_admin_delete ON users;
DROP FUNCTION IF EXISTS protect_default_admin();
ALTER TABLE users DROP CONSTRAINT athlete_fields;
ALTER TABLE users ADD CONSTRAINT athlete_fields CHECK (
 role <> 'athlete' OR (
  has_medical_history IS NOT NULL AND
  blood_type IS NOT NULL AND
  emergency_phone IS NOT NULL AND
  emergency_relation IS NOT NULL
 )
);
