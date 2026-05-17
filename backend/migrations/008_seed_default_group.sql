-- Seed a default group for fresh installs.
-- Use WHERE NOT EXISTS to safely handle both fresh installs and existing deployments.
-- ON CONFLICT (name) cannot be used because migration 016 replaces the full unique constraint
-- with a partial unique index (WHERE deleted_at IS NULL), which PostgreSQL does not support
-- for ON CONFLICT resolution.
INSERT INTO groups (name, description, created_at, updated_at)
SELECT 'default', 'Default group', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM groups WHERE name = 'default' AND deleted_at IS NULL);