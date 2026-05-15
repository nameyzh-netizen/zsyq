-- Remove unsafe built-in Anthropic Claude Code impersonation request template.
-- Existing monitors retain copied request snapshot fields; template_id is cleared by ON DELETE SET NULL.
DELETE FROM channel_monitor_request_templates
WHERE provider = 'anthropic'
  AND name = 'Claude Code 伪装';
