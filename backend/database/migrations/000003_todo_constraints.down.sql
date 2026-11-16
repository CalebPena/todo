DROP INDEX todos_user_parent_idx;
ALTER TABLE todos DROP CONSTRAINT todos_ranks_check;
ALTER TABLE todos ADD CONSTRAINT todos_name_key UNIQUE (name);
