BEGIN;

DELETE FROM object_type
WHERE type_id IN ('role', 'permission', 'tenant', 'user', 'pricing-tier', 'feature');

COMMIT;
