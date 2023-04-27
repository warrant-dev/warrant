BEGIN;

UPDATE accessEvent
SET type = CONCAT(objectType, ".", type);

UPDATE resourceEvent
SET type = CONCAT(resourceType, ".", type);

COMMIT;
