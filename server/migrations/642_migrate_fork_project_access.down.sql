-- Translate the current official policy back before rolling back application
-- code. Keep official rows/schema for recovery; do not reverse later DDL.
BEGIN;
UPDATE project p
SET access_restricted = (p.visibility = 'private'),
    created_by = CASE WHEN p.creator_type = 'member' AND (
        p.visibility <> 'private' OR EXISTS (
            SELECT 1 FROM project_member pm
            WHERE pm.workspace_id = p.workspace_id AND pm.project_id = p.id
                AND pm.user_id = p.creator_id
        ) OR EXISTS (
            SELECT 1 FROM member m WHERE m.workspace_id = p.workspace_id
                AND m.user_id = p.creator_id AND m.role IN ('owner', 'admin')
        )
    ) THEN p.creator_id ELSE NULL END,
    allowed_user_ids = ARRAY(
        SELECT pm.user_id FROM project_member pm
        WHERE pm.workspace_id = p.workspace_id AND pm.project_id = p.id
        UNION
        SELECT m.user_id FROM member m
        WHERE m.workspace_id = p.workspace_id AND m.role IN ('owner', 'admin')
    );
COMMIT;
