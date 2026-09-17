\set ON_ERROR_STOP on
BEGIN;
CREATE TEMP TABLE project (id UUID PRIMARY KEY, workspace_id UUID, created_at TIMESTAMPTZ);
CREATE TEMP TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ);
INSERT INTO project VALUES
('00000000-0000-0000-0000-000000000001', NULL, '2026-01-01'),
('00000000-0000-0000-0000-000000000002', NULL, '2026-01-03');
\ir ../../../migrations/500_project_access.up.sql
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM project WHERE created_by IS NOT NULL) THEN
        RAISE EXCEPTION 'Fresh migration invented project creators';
    END IF;
END $$;
INSERT INTO schema_migrations VALUES ('500_project_access', '2026-01-02');
-- Simulate the original historical backfill and a genuinely new project.
UPDATE project SET created_by = '00000000-0000-0000-0000-000000000010';
UPDATE project SET access_restricted = true,
    allowed_user_ids = ARRAY['00000000-0000-0000-0000-000000000020'::uuid]
WHERE id = '00000000-0000-0000-0000-000000000002';
\ir ../../../migrations/501_correct_historical_project_creators.up.sql
\ir ../../../migrations/501_correct_historical_project_creators.up.sql
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM project WHERE id = '00000000-0000-0000-0000-000000000001'
        AND (created_by IS NOT NULL OR access_restricted)) THEN
        RAISE EXCEPTION 'Historical creator or access not corrected';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM project WHERE id = '00000000-0000-0000-0000-000000000002'
        AND created_by = '00000000-0000-0000-0000-000000000010'
        AND access_restricted AND allowed_user_ids = ARRAY['00000000-0000-0000-0000-000000000020'::uuid]) THEN
        RAISE EXCEPTION 'New project attribution or policy changed';
    END IF;
END $$;
ROLLBACK;
