ALTER TABLE todos ADD COLUMN favorite BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX todos_favorites_idx ON todos(user_id) WHERE favorite;
