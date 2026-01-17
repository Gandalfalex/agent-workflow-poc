WITH upserted AS (
  INSERT INTO group_memberships (group_id, user_id)
  VALUES ($1, $2)
  ON CONFLICT (group_id, user_id) DO UPDATE SET user_id = EXCLUDED.user_id
  RETURNING group_id, user_id
)
SELECT upserted.group_id, upserted.user_id, u.name
FROM upserted
LEFT JOIN users u ON u.id = upserted.user_id
