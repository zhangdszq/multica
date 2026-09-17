ALTER TABLE project
    ADD COLUMN created_by UUID,
    ADD COLUMN access_restricted BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN allowed_user_ids UUID[] NOT NULL DEFAULT '{}';

-- Historical projects did not record their creator. Leave it unknown instead
-- of granting creator permissions based on workspace ownership.
