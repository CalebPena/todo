UPDATE todos SET value_rank = NULL, time_rank = NULL WHERE completed;

WITH active_siblings AS (
    SELECT id, value_rank, time_rank,
        count(*) OVER (PARTITION BY user_id, parent_id) AS sibling_count,
        user_id, parent_id
    FROM todos WHERE NOT completed
), positions AS (
    SELECT id,
        row_number() OVER (PARTITION BY user_id, parent_id ORDER BY value_rank, id) AS value_position,
        row_number() OVER (PARTITION BY user_id, parent_id ORDER BY time_rank, id) AS time_position
    FROM active_siblings WHERE value_rank IS NOT NULL OR sibling_count = 1
)
UPDATE todos SET value_rank = positions.value_position, time_rank = positions.time_position
FROM positions WHERE todos.id = positions.id;

ALTER TABLE todos ADD CONSTRAINT todos_completed_unranked_check
    CHECK (NOT completed OR (value_rank IS NULL AND time_rank IS NULL));
