BEGIN;

UPDATE accessEvent
SET type = SUBSTRING_INDEX(type, ".", -1);

UPDATE resourceEvent
SET type = SUBSTRING_INDEX(type, ".", -1);

COMMIT;
