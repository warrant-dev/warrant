BEGIN;

ALTER TABLE object
ADD COLUMN meta json DEFAULT NULL AFTER objectId;

# backfill tenants
UPDATE object
INNER JOIN tenant ON object.objectId = tenant.tenantId
SET meta = JSON_OBJECT("name", tenant.name)
WHERE
    object.objectType = "tenant" AND
    tenant.name IS NOT NULL;

# backfill users
UPDATE object
INNER JOIN user ON object.objectId = user.userId
SET meta = JSON_OBJECT("email", user.email)
WHERE
    object.objectType = "user" AND
    user.email IS NOT NULL;

# backfill roles
UPDATE object
INNER JOIN role ON object.objectId = role.roleId
SET meta = JSON_OBJECT("name", role.name)
WHERE
    object.objectType = "role" AND
    role.name IS NOT NULL AND
    role.description IS NULL;

UPDATE object
INNER JOIN role ON object.objectId = role.roleId
SET meta = JSON_OBJECT("description", role.description)
WHERE
    object.objectType = "role" AND
    role.name IS NULL AND
    role.description IS NOT NULL;

UPDATE object
INNER JOIN role ON object.objectId = role.roleId
SET meta = JSON_OBJECT("name", role.name, "description", role.description)
WHERE
    object.objectType = "role" AND
    role.name IS NOT NULL AND
    role.description IS NOT NULL;

# backfill permissions
UPDATE object
INNER JOIN permission ON object.objectId = permission.permissionId
SET meta = JSON_OBJECT("name", permission.name)
WHERE
    object.objectType = "permission" AND
    permission.name IS NOT NULL AND
    permission.description IS NULL;

UPDATE object
INNER JOIN permission ON object.objectId = permission.permissionId
SET meta = JSON_OBJECT("description", permission.description)
WHERE
    object.objectType = "permission" AND
    permission.name IS NULL AND
    permission.description IS NOT NULL;

UPDATE object
INNER JOIN permission ON object.objectId = permission.permissionId
SET meta = JSON_OBJECT("name", permission.name, "description", permission.description)
WHERE
    object.objectType = "permission" AND
    permission.name IS NOT NULL AND
    permission.description IS NOT NULL;

# backfill pricing-tiers
UPDATE object
INNER JOIN pricingTier ON object.objectId = pricingTier.pricingTierId
SET meta = JSON_OBJECT("name", pricingTier.name)
WHERE
    object.objectType = "pricing-tier" AND
    pricingTier.name IS NOT NULL AND
    pricingTier.description IS NULL;

UPDATE object
INNER JOIN pricingTier ON object.objectId = pricingTier.pricingTierId
SET meta = JSON_OBJECT("description", pricingTier.description)
WHERE
    object.objectType = "pricing-tier" AND
    pricingTier.name IS NULL AND
    pricingTier.description IS NOT NULL;

UPDATE object
INNER JOIN pricingTier ON object.objectId = pricingTier.pricingTierId
SET meta = JSON_OBJECT("name", pricingTier.name, "description", pricingTier.description)
WHERE
    object.objectType = "pricing-tier" AND
    pricingTier.name IS NOT NULL AND
    pricingTier.description IS NOT NULL;

# backfill features
UPDATE object
INNER JOIN feature ON object.objectId = feature.featureId
SET meta = JSON_OBJECT("name", feature.name)
WHERE
    object.objectType = "feature" AND
    feature.name IS NOT NULL AND
    feature.description IS NULL;

UPDATE object
INNER JOIN feature ON object.objectId = feature.featureId
SET meta = JSON_OBJECT("description", feature.description)
WHERE
    object.objectType = "feature" AND
    feature.name IS NULL AND
    feature.description IS NOT NULL;

UPDATE object
INNER JOIN feature ON object.objectId = feature.featureId
SET meta = JSON_OBJECT("name", feature.name, "description", feature.description)
WHERE
    object.objectType = "feature" AND
    feature.name IS NOT NULL AND
    feature.description IS NOT NULL;

# drop unused tables
DROP TABLE tenant;
DROP TABLE user;
DROP TABLE role;
DROP TABLE permission;
DROP TABLE pricingTier;
DROP TABLE feature;

COMMIT;
