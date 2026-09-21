-- The initial SHUZ-152 test migration incorrectly assigned historical
-- projects to the earliest workspace owner. New projects created after that
-- migration record their actual creator and must retain that attribution.
DO $$
DECLARE
    access_migration_applied_at TIMESTAMPTZ;
BEGIN
    -- Installations that already ran this fork migration under its original
    -- version have completed the correction. Do not reinterpret projects that
    -- became restricted later as unverified historical data.
    IF EXISTS (
        SELECT 1 FROM schema_migrations
        WHERE version = '501_correct_historical_project_creators'
    ) THEN
        RETURN;
    END IF;

    SELECT MIN(applied_at) INTO access_migration_applied_at
    FROM schema_migrations
    WHERE version IN ('500_project_access', '535_project_access');

    IF EXISTS (
        SELECT 1 FROM project p
        WHERE p.created_at < access_migration_applied_at AND p.created_by IS NOT NULL
          AND p.access_restricted
    ) THEN
        RAISE EXCEPTION 'Historical restricted projects require verified creator attribution before correcting migration 500';
    END IF;
END $$;

WITH access_migration AS (
    SELECT MIN(applied_at) AS applied_at
    FROM schema_migrations
    WHERE version IN ('500_project_access', '535_project_access')
)
UPDATE project p SET created_by = NULL
FROM access_migration m
WHERE p.created_at < m.applied_at AND p.created_by IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM schema_migrations
      WHERE version = '501_correct_historical_project_creators'
  );
