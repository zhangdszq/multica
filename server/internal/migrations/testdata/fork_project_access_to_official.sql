\set ON_ERROR_STOP on
-- Run in one disposable psql session. All fixtures are temporary and shadow
-- application tables; the real migration SQL is exercised without live rows.
CREATE TEMP TABLE project (
    id UUID PRIMARY KEY, workspace_id UUID NOT NULL,
    created_by UUID, access_restricted BOOLEAN NOT NULL,
    allowed_user_ids UUID[] NOT NULL DEFAULT '{}',
    creator_type TEXT, creator_id UUID,
    visibility TEXT NOT NULL DEFAULT 'workspace'
);
CREATE TEMP TABLE project_member (
    workspace_id UUID NOT NULL, project_id UUID NOT NULL, user_id UUID NOT NULL,
    added_by UUID, PRIMARY KEY(project_id,user_id)
);
CREATE TEMP TABLE member (workspace_id UUID NOT NULL, user_id UUID NOT NULL, role TEXT NOT NULL);

INSERT INTO project (id,workspace_id,created_by,access_restricted,allowed_user_ids) VALUES
('00000000-0000-0000-0000-000000000101','00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000011',true,
 ARRAY['00000000-0000-0000-0000-000000000012'::uuid,'00000000-0000-0000-0000-000000000012'::uuid]),
('00000000-0000-0000-0000-000000000102','00000000-0000-0000-0000-000000000001',NULL,false,'{}'),
('00000000-0000-0000-0000-000000000103','00000000-0000-0000-0000-000000000002','00000000-0000-0000-0000-000000000021',true,'{}'),
('00000000-0000-0000-0000-000000000104','00000000-0000-0000-0000-000000000002',NULL,false,
 ARRAY['00000000-0000-0000-0000-000000000022'::uuid]);

\ir ../../../migrations/642_migrate_fork_project_access.up.sql
\ir ../../../migrations/642_migrate_fork_project_access.up.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM project_member) <> 4 THEN
        RAISE EXCEPTION 'Duplicate grants, creator grants or dormant public grants were not preserved';
    END IF;
    IF (SELECT count(*) FROM project WHERE visibility='private') <> 2 THEN
        RAISE EXCEPTION 'Public/private policy changed';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM project WHERE id='00000000-0000-0000-0000-000000000101'
        AND creator_type='member' AND creator_id=created_by AND access_restricted) THEN
        RAISE EXCEPTION 'Creator identity or rollback source lost';
    END IF;
    IF EXISTS (SELECT 1 FROM project WHERE id='00000000-0000-0000-0000-000000000102'
        AND (creator_type IS NOT NULL OR creator_id IS NOT NULL)) THEN
        RAISE EXCEPTION 'Unknown historical creator was invented';
    END IF;
    IF EXISTS (SELECT 1 FROM project_member WHERE project_id='00000000-0000-0000-0000-000000000101'
        AND user_id NOT IN ('00000000-0000-0000-0000-000000000011','00000000-0000-0000-0000-000000000012')) THEN
        RAISE EXCEPTION 'Unexpected member gained a private project grant';
    END IF;
    IF EXISTS (SELECT 1 FROM project_member pm JOIN project p ON p.id=pm.project_id
        WHERE pm.workspace_id<>p.workspace_id) THEN
        RAISE EXCEPTION 'Grant escaped its project workspace';
    END IF;
END $$;

-- Simulate access changes after cutover. A rollback must translate the latest
-- official rows, rather than restoring a stale pre-cutover grant snapshot.
DELETE FROM project_member WHERE project_id='00000000-0000-0000-0000-000000000101'
    AND user_id='00000000-0000-0000-0000-000000000012';
DELETE FROM project_member WHERE project_id='00000000-0000-0000-0000-000000000103';
INSERT INTO project_member VALUES ('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000101','00000000-0000-0000-0000-000000000013',NULL);
INSERT INTO member VALUES ('00000000-0000-0000-0000-000000000001','00000000-0000-0000-0000-000000000014','admin');
\ir ../../../migrations/642_migrate_fork_project_access.down.sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM project WHERE id='00000000-0000-0000-0000-000000000101'
        AND access_restricted
        AND '00000000-0000-0000-0000-000000000013'::uuid=ANY(allowed_user_ids)
        AND '00000000-0000-0000-0000-000000000014'::uuid=ANY(allowed_user_ids)
        AND NOT ('00000000-0000-0000-0000-000000000012'::uuid=ANY(allowed_user_ids))) THEN
        RAISE EXCEPTION 'Rollback lost a new grant, revived a removed grant or dropped admin access';
    END IF;
    IF EXISTS (SELECT 1 FROM project WHERE id='00000000-0000-0000-0000-000000000103'
        AND created_by IS NOT NULL) THEN
        RAISE EXCEPTION 'Rollback revived a revoked creator grant';
    END IF;
END $$;
