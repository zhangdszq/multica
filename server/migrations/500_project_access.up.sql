ALTER TABLE project
    ADD COLUMN created_by UUID,
    ADD COLUMN access_restricted BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN allowed_user_ids UUID[] NOT NULL DEFAULT '{}';

-- Historical projects did not record their creator. Assign their access
-- management to the earliest workspace owner without changing visibility.
UPDATE project p SET created_by = (
    SELECT m.user_id FROM member m
    WHERE m.workspace_id = p.workspace_id AND m.role = 'owner'
    ORDER BY m.created_at, m.id LIMIT 1
);
