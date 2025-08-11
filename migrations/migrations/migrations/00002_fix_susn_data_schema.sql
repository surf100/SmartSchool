-- +goose Up
-- +goose StatementBegin
-- 1) типы: not null, дефолты
ALTER TABLE external_susn_data
  ALTER COLUMN iin SET NOT NULL,
  ALTER COLUMN school_bin SET NOT NULL,
  ALTER COLUMN social_payment SET NOT NULL,
  ALTER COLUMN created_at SET DEFAULT now(),
  ALTER COLUMN updated_at SET DEFAULT now();

-- 2) уникальность пары (iin, school_bin)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'ux_external_susn_data_iin_bin'
  ) THEN
    ALTER TABLE external_susn_data
      ADD CONSTRAINT ux_external_susn_data_iin_bin UNIQUE (iin, school_bin);
  END IF;
END $$;

-- 3) индексы
CREATE INDEX IF NOT EXISTS ix_external_susn_data_iin ON external_susn_data (iin);
CREATE INDEX IF NOT EXISTS ix_external_susn_data_school_bin ON external_susn_data (school_bin);

-- 4) тригер на updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'tr_external_susn_data_set_updated_at'
  ) THEN
    CREATE TRIGGER tr_external_susn_data_set_updated_at
    BEFORE UPDATE ON external_susn_data
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS tr_external_susn_data_set_updated_at ON external_susn_data;
DROP INDEX IF EXISTS ix_external_susn_data_iin;
DROP INDEX IF EXISTS ix_external_susn_data_school_bin;
ALTER TABLE external_susn_data DROP CONSTRAINT IF EXISTS ux_external_susn_data_iin_bin;
ALTER TABLE external_susn_data
  ALTER COLUMN iin DROP NOT NULL,
  ALTER COLUMN school_bin DROP NOT NULL,
  ALTER COLUMN social_payment DROP NOT NULL,
  ALTER COLUMN created_at DROP DEFAULT,
  ALTER COLUMN updated_at DROP DEFAULT;
-- +goose StatementEnd
