ALTER TABLE project
    ADD COLUMN IF NOT EXISTS created_by UUID,
    ADD COLUMN IF NOT EXISTS access_restricted BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS allowed_user_ids UUID[] NOT NULL DEFAULT '{}';

-- Historical projects did not record their creator. Leave it unknown instead
-- of granting creator permissions based on workspace ownership.
--
-- This migration was originally shipped as 500_project_access in the fork.
-- IF NOT EXISTS keeps the renumbered migration safe for upgraded installations.
