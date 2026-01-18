ALTER TABLE students
  RENAME COLUMN photo_object_key TO photo_url;

DROP INDEX IF EXISTS imagekit_outbox_available_at_idx;
DROP TABLE IF EXISTS imagekit_outbox;
