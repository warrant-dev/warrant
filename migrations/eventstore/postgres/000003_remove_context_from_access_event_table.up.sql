BEGIN;

ALTER TABLE access_event
DROP COLUMN context;

COMMIT;
