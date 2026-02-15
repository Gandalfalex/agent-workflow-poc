-- Make story_id mandatory on tickets (every ticket must belong to a story)
-- NOTE: existing rows with NULL story_id must be cleaned up before applying this migration.

ALTER TABLE tickets
  ALTER COLUMN story_id SET NOT NULL;

-- Replace ON DELETE SET NULL with ON DELETE CASCADE since story_id can no longer be null
ALTER TABLE tickets
  DROP CONSTRAINT IF EXISTS tickets_story_id_fkey,
  ADD CONSTRAINT tickets_story_id_fkey
    FOREIGN KEY (story_id) REFERENCES stories(id) ON DELETE CASCADE;
