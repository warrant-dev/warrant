BEGIN;

UPDATE access_event
SET type = SUBSTR(type, STRPOS(type, '.') + 1);

UPDATE resource_event
SET type = SUBSTR(type, STRPOS(type, '.') + 1);

COMMIT;
