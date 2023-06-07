BEGIN;

ALTER TABLE accessEvent
ADD COLUMN context json DEFAULT NULL AFTER subjectRelation;

COMMIT;
