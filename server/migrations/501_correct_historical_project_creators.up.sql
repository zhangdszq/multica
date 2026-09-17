-- The initial SHUZ-152 test migration incorrectly assigned historical
-- projects to the earliest workspace owner. New projects created after that
-- migration record their actual creator and must retain that attribution.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM project p
        JOIN schema_migrations m ON m.version = '500_project_access'
        WHERE p.created_at < m.applied_at AND p.created_by IS NOT NULL
          AND p.access_restricted
    ) THEN
        RAISE EXCEPTION 'Historical restricted projects require verified creator attribution before correcting migration 500';
    END IF;
END $$;

UPDATE project p SET created_by = NULL
FROM schema_migrations m
WHERE m.version = '500_project_access'
  AND p.created_at < m.applied_at AND p.created_by IS NOT NULL;
