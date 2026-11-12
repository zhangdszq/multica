-- Move the fork's existing policy into MUL-8472's official model before the
-- new application starts. Retain the legacy columns only as rollback data.
BEGIN;

UPDATE project
SET creator_type = 'member', creator_id = created_by
WHERE created_by IS NOT NULL AND creator_type IS NULL AND creator_id IS NULL;

INSERT INTO project_member (workspace_id, project_id, user_id, added_by)
SELECT p.workspace_id, p.id, grants.user_id, p.created_by
FROM project p
CROSS JOIN LATERAL (
    SELECT p.created_by AS user_id
    UNION
    SELECT unnest(p.allowed_user_ids)
) grants
WHERE grants.user_id IS NOT NULL
ON CONFLICT (project_id, user_id) DO NOTHING;

UPDATE project
SET visibility = CASE WHEN access_restricted THEN 'private' ELSE 'workspace' END;

-- Abort the whole transaction if any formerly authorized principal was lost.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM project p
        CROSS JOIN LATERAL (
            SELECT p.created_by AS user_id
            UNION SELECT unnest(p.allowed_user_ids)
        ) grants
        WHERE p.access_restricted AND grants.user_id IS NOT NULL
          AND NOT EXISTS (
              SELECT 1 FROM project_member pm
              WHERE pm.workspace_id = p.workspace_id AND pm.project_id = p.id
                AND pm.user_id = grants.user_id
          )
    ) THEN
        RAISE EXCEPTION 'Fork project access migration lost an existing grant';
    END IF;
END $$;

COMMIT;
