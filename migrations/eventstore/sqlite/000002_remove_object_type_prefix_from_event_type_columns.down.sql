UPDATE accessEvent
SET type = objectType || "." || type;

UPDATE resourceEvent
SET type = resourceType || "." || type;
