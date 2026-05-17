CREATE TABLE IF NOT EXISTS ops_alert_silences (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    platform VARCHAR(64) NOT NULL,
    group_id BIGINT,
    region VARCHAR(64),
    until TIMESTAMPTZ NOT NULL,
    reason TEXT,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_alert_silences_lookup
    ON ops_alert_silences (rule_id, platform, group_id, region, until);

UPDATE accounts
SET credentials = jsonb_set(
    credentials,
    '{tier_id}',
    '"LEGACY"',
    true
)
WHERE platform = 'gemini'
  AND type = 'oauth'
  AND jsonb_typeof(credentials) = 'object'
  AND credentials->>'tier_id' IS NULL
  AND (
    credentials->>'oauth_type' = 'code_assist'
    OR (credentials->>'oauth_type' IS NULL AND credentials->>'project_id' IS NOT NULL)
  );

WITH candidate AS (
    SELECT id
    FROM user_attribute_definitions
    WHERE key = 'wechat'
      AND deleted_at IS NOT NULL
      AND NOT EXISTS (
          SELECT 1
          FROM user_attribute_definitions active
          WHERE active.key = 'wechat'
            AND active.deleted_at IS NULL
      )
    ORDER BY id
    LIMIT 1
)
UPDATE user_attribute_definitions d
SET name = '微信',
    description = '用户微信号',
    type = 'text',
    options = '[]'::jsonb,
    required = false,
    validation = '{}'::jsonb,
    placeholder = '请输入微信号',
    display_order = -1,
    enabled = true,
    deleted_at = NULL,
    updated_at = NOW()
FROM candidate
WHERE d.id = candidate.id;

INSERT INTO user_attribute_definitions (key, name, description, type, options, required, validation, placeholder, display_order, enabled, created_at, updated_at)
SELECT 'wechat', '微信', '用户微信号', 'text', '[]'::jsonb, false, '{}'::jsonb, '请输入微信号', -1, true, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM user_attribute_definitions WHERE key = 'wechat' AND deleted_at IS NULL
);

DO $$
DECLARE
    wechat_attribute_id BIGINT;
BEGIN
    SELECT id INTO wechat_attribute_id
    FROM user_attribute_definitions
    WHERE key = 'wechat'
      AND deleted_at IS NULL
    ORDER BY id
    LIMIT 1;

    IF wechat_attribute_id IS NOT NULL AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'users'
          AND column_name = 'wechat'
    ) THEN
        EXECUTE format($sql$
            INSERT INTO user_attribute_values (user_id, attribute_id, value, created_at, updated_at)
            SELECT u.id, %s, u.wechat, NOW(), NOW()
            FROM users u
            WHERE u.wechat IS NOT NULL
              AND u.wechat != ''
              AND u.deleted_at IS NULL
            ON CONFLICT (user_id, attribute_id) DO UPDATE
            SET value = EXCLUDED.value,
                updated_at = NOW()
            WHERE COALESCE(user_attribute_values.value, '') = ''
        $sql$, wechat_attribute_id);
    END IF;
END $$;

UPDATE user_attribute_definitions
SET display_order = -1,
    updated_at = NOW()
WHERE key = 'wechat'
  AND deleted_at IS NULL;

WITH ordered AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY display_order, id) - 1 as new_order
    FROM user_attribute_definitions
    WHERE deleted_at IS NULL
)
UPDATE user_attribute_definitions
SET display_order = ordered.new_order,
    updated_at = NOW()
FROM ordered
WHERE user_attribute_definitions.id = ordered.id;

ALTER TABLE users DROP COLUMN IF EXISTS wechat;
