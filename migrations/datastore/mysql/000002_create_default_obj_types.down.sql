BEGIN;

DELETE FROM objectType
WHERE typeId IN ('role', 'permission', 'tenant', 'user', 'pricing-tier', 'feature');

COMMIT;
