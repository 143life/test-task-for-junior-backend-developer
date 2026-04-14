ALTER TABLE tasks ADD COLUMN schedule JSONB DEFAULT NULL;

CREATE INDEX idx_tasks_schedule_not_null ON tasks (schedule) WHERE schedule IS NOT NULL;

CREATE OR REPLACE FUNCTION notify_schedule_change()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') AND NEW.schedule IS NOT NULL THEN
        PERFORM pg_notify('task_schedule_changed', NEW.id::text);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER task_schedule_notify
AFTER INSERT OR UPDATE OF schedule ON tasks
FOR EACH ROW EXECUTE FUNCTION notify_schedule_change();