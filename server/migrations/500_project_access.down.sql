ALTER TABLE project
    DROP COLUMN IF EXISTS allowed_user_ids,
    DROP COLUMN IF EXISTS access_restricted,
    DROP COLUMN IF EXISTS created_by;
