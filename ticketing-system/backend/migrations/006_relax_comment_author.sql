ALTER TABLE ticket_comments
  DROP CONSTRAINT IF EXISTS ticket_comments_author_id_fkey;

ALTER TABLE ticket_comments
  ALTER COLUMN author_id DROP NOT NULL;
