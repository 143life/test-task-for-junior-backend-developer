DROP TRIGGER IF EXISTS task_schedule_notify ON tasks;
DROP FUNCTION IF EXISTS notify_schedule_change();
DROP INDEX IF EXISTS idx_tasks_schedule_not_null;
ALTER TABLE tasks DROP COLUMN IF EXISTS schedule;