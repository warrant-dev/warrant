BEGIN;

UPDATE access_event
SET type = CONCAT(object_type, '.', type);

UPDATE resource_event
SET type = CONCAT(resource_type, '.', type);

COMMIT;
