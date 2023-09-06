ALTER TABLE object
ADD COLUMN meta json DEFAULT NULL;

-- backfill tenants
UPDATE object
SET meta = JSON('{"name":"' || tenant.name || '"}')
FROM tenant
WHERE
    object.objectId = tenant.tenantId AND
    object.objectType = "tenant" AND
    tenant.name IS NOT NULL;

-- backfill users
UPDATE object
SET meta = JSON('{"email":"' || user.email || '"}')
FROM user
WHERE
    object.objectId = user.userId AND
    object.objectType = "user" AND
    user.email IS NOT NULL;

-- backfill roles
UPDATE object
SET meta = JSON('{"name":"' || role.name || '"}')
FROM role
WHERE
    object.objectId = role.roleId AND
    object.objectType = "role" AND
    role.name IS NOT NULL AND
    role.description IS NULL;

UPDATE object
SET meta = JSON('{"desciption":"' || role.desciption || '"}')
FROM role
WHERE
    object.objectId = role.roleId AND
    object.objectType = "role" AND
    role.name IS NULL AND
    role.description IS NOT NULL;

UPDATE object
SET meta = JSON('{"name":"' || role.name || '", "description":"' || role.description || '"}')
FROM role
WHERE
    object.objectId = role.roleId AND
    object.objectType = "role" AND
    role.name IS NOT NULL AND
    role.description IS NOT NULL;

-- backfill permissions
UPDATE object
SET meta = JSON('{"name":"' || permission.name || '"}')
FROM permission
WHERE
    object.objectId = permission.permissionId AND
    object.objectType = "permission" AND
    permission.name IS NOT NULL AND
    permission.description IS NULL;

UPDATE object
SET meta = JSON('{"description":"' || permission.description || '"}')
FROM permission
WHERE
    object.objectId = permission.permissionId AND
    object.objectType = "permission" AND
    permission.name IS NULL AND
    permission.description IS NOT NULL;

UPDATE object
SET meta = JSON('{"name":"' || permission.name || '", "description":"' || permission.description || '"}')
FROM permission
WHERE
    object.objectId = permission.permissionId AND
    object.objectType = "permission" AND
    permission.name IS NOT NULL AND
    permission.description IS NOT NULL;

-- backfill pricing-tiers
UPDATE object
SET meta = JSON('{"name":"' || pricingTier.name || '"}')
FROM pricingTier
WHERE
    object.objectId = pricingTier.pricingTierId AND
    object.objectType = "pricingTier" AND
    pricingTier.name IS NOT NULL AND
    pricingTier.description IS NULL;

UPDATE object
SET meta = JSON('{"description":"' || pricingTier.description || '"}')
FROM pricingTier
WHERE
    object.objectId = pricingTier.pricingTierId AND
    object.objectType = "pricingTier" AND
    pricingTier.name IS NULL AND
    pricingTier.description IS NOT NULL;

UPDATE object
SET meta = JSON('{"name":"' || pricingTier.name || '", "description":"' || pricingTier.description || '"}')
FROM pricingTier
WHERE
    object.objectId = pricingTier.pricingTierId AND
    object.objectType = "pricingTier" AND
    pricingTier.name IS NOT NULL AND
    pricingTier.description IS NOT NULL;

-- backfill features
UPDATE object
SET meta = JSON('{"name":"' || feature.name || '"}')
FROM feature
WHERE
    object.objectId = feature.featureId AND
    object.objectType = "feature" AND
    feature.name IS NOT NULL AND
    feature.description IS NULL;

UPDATE object
SET meta = JSON('{"description":"' || feature.description || '"}')
FROM feature
WHERE
    object.objectId = feature.featureId AND
    object.objectType = "feature" AND
    feature.name IS NULL AND
    feature.description IS NOT NULL;

UPDATE object
SET meta = JSON('{"name":"' || feature.name || '", "description":"' || feature.description || '"}')
FROM feature
WHERE
    object.objectId = feature.featureId AND
    object.objectType = "feature" AND
    feature.name IS NOT NULL AND
    feature.description IS NOT NULL;

-- drop unused tables
DROP TABLE tenant;
DROP TABLE user;
DROP TABLE role;
DROP TABLE permission;
DROP TABLE pricingTier;
DROP TABLE feature;
