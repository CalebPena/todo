WITH positions AS (
    SELECT id,
        row_number() OVER (PARTITION BY user_id, parent_id ORDER BY value_rank, id) AS value_position,
        row_number() OVER (PARTITION BY user_id, parent_id ORDER BY time_rank, id) AS time_position
    FROM todos
    WHERE value_rank IS NOT NULL
)
UPDATE todos
SET value_rank = positions.value_position, time_rank = positions.time_position
FROM positions
WHERE todos.id = positions.id;
