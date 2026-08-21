ALTER TABLE users DROP CONSTRAINT athlete_fields;
ALTER TABLE users ADD CONSTRAINT athlete_fields CHECK (
 role <> 'athlete' OR (
  has_medical_history IS NOT NULL AND
  emergency_phone IS NOT NULL AND
  emergency_relation IS NOT NULL
 )
);

CREATE OR REPLACE FUNCTION protect_default_admin() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.username = 'admin' THEN
  RAISE EXCEPTION 'PROTECTED_DEFAULT_ADMIN';
 END IF;
 RETURN OLD;
END $$;

CREATE TRIGGER protect_default_admin_delete
BEFORE DELETE ON users
FOR EACH ROW EXECUTE FUNCTION protect_default_admin();
