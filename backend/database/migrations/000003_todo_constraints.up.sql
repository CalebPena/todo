ALTER TABLE todos DROP CONSTRAINT todos_name_key;
ALTER TABLE todos ADD CONSTRAINT todos_ranks_check CHECK (
    (value_rank IS NULL AND time_rank IS NULL) OR
    (value_rank IS NOT NULL AND time_rank IS NOT NULL AND value_rank > 0 AND time_rank > 0)
);
CREATE INDEX todos_user_parent_idx ON todos (user_id, parent_id);
