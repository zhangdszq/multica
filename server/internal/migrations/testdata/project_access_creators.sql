\set ON_ERROR_STOP on
BEGIN;
CREATE TEMP TABLE project (id UUID PRIMARY KEY, workspace_id UUID, created_at TIMESTAMPTZ);
CREATE TEMP TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ);
INSERT INTO project VALUES
('00000000-0000-0000-0000-000000000001', NULL, '2026-01-01'),
('00000000-0000-0000-0000-000000000002', NULL, '2026-01-03');
\ir ../../../migrations/535_project_access.up.sql
\ir ../../../migrations/535_project_access.up.sql
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
\ir ../../../migrations/536_correct_historical_project_creators.up.sql
\ir ../../../migrations/536_correct_historical_project_creators.up.sql
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

-- Confirmed SHUZ-152 attribution is scoped to the intended workspace and is
-- safe to apply and roll back more than once.
INSERT INTO project (id, workspace_id, created_at) VALUES
('7181cff2-8d1f-4cb4-9cae-c4959cf4c697', '5d54d458-153b-4612-a1a4-e3346de2bcb2', '2026-01-01'),
('0bf8cd8a-4604-4a1a-a754-03c1955e5cb8', '5d54d458-153b-4612-a1a4-e3346de2bcb2', '2026-01-01'),
('ef04ece8-8c18-4c34-b700-9f9cfff3124b', '00000000-0000-0000-0000-000000000099', '2026-01-01');
\ir ../../../migrations/537_backfill_confirmed_project_creators.up.sql
\ir ../../../migrations/537_backfill_confirmed_project_creators.up.sql
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM project
        WHERE id = '7181cff2-8d1f-4cb4-9cae-c4959cf4c697'
          AND created_by = 'b4a22a55-a3aa-499e-b33e-26de4b8fc65a'
    ) THEN
        RAISE EXCEPTION 'Confirmed Zhang Xinwei attribution was not applied';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM project
        WHERE id = '0bf8cd8a-4604-4a1a-a754-03c1955e5cb8'
          AND created_by = 'a01b5b0c-0491-4186-a8c4-f33f25f834c9'
    ) THEN
        RAISE EXCEPTION 'Confirmed Peng Bo attribution was not applied';
    END IF;
    IF EXISTS (
        SELECT 1 FROM project
        WHERE id = 'ef04ece8-8c18-4c34-b700-9f9cfff3124b'
          AND created_by IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Confirmed attribution escaped its workspace guard';
    END IF;
END $$;
\ir ../../../migrations/537_backfill_confirmed_project_creators.down.sql
\ir ../../../migrations/537_backfill_confirmed_project_creators.down.sql
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM project
        WHERE id IN (
            '7181cff2-8d1f-4cb4-9cae-c4959cf4c697',
            '0bf8cd8a-4604-4a1a-a754-03c1955e5cb8'
        ) AND created_by IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Confirmed attribution rollback was not idempotent';
    END IF;
END $$;
ROLLBACK;
