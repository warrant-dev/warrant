BEGIN;

ALTER TABLE object
ADD COLUMN meta jsonb DEFAULT NULL;

-- backfill tenants
UPDATE object
SET meta = jsonb_build_object('name', tenant.name)
FROM tenant
WHERE
    object.object_id = tenant.tenant_id AND
    object.object_type = 'tenant' AND
    tenant.name IS NOT NULL;

-- backfill users
UPDATE object
SET meta = jsonb_build_object('email', "user".email)
FROM "user"
WHERE
    object.object_id = "user".user_id AND
    object.object_type = 'user' AND
    "user".email IS NOT NULL;

-- backfill roles
UPDATE object
SET meta = jsonb_build_object('name', role.name)
FROM role
WHERE
    object.object_id = role.role_id AND
    object.object_type = 'role' AND
    role.name IS NOT NULL AND
    role.description IS NULL;

UPDATE object
SET meta = jsonb_build_object('description', role.description)
FROM role
WHERE
    object.object_id = role.role_id AND
    object.object_type = 'role' AND
    role.name IS NULL AND
    role.description IS NOT NULL;

UPDATE object
SET meta = jsonb_build_object('name', role.name, 'description', role.description)
FROM role
WHERE
    object.object_id = role.role_id AND
    object.object_type = 'role' AND
    role.name IS NOT NULL AND
    role.description IS NOT NULL;

-- backfill permissions
UPDATE object
SET meta = jsonb_build_object('name', permission.name)
FROM permission
WHERE
    object.object_id = permission.permission_id AND
    object.object_type = 'permission' AND
    permission.name IS NOT NULL AND
    permission.description IS NULL;

UPDATE object
SET meta = jsonb_build_object('description', permission.description)
FROM permission
WHERE
    object.object_id = permission.permission_id AND
    object.object_type = 'permission' AND
    permission.name IS NULL AND
    permission.description IS NOT NULL;

UPDATE object
SET meta = jsonb_build_object('name', permission.name, 'description', permission.description)
FROM permission
WHERE
    object.object_id = permission.permission_id AND
    object.object_type = 'permission' AND
    permission.name IS NOT NULL AND
    permission.description IS NOT NULL;

-- backfill pricing-tiers
UPDATE object
SET meta = jsonb_build_object('name', pricing_tier.name)
FROM pricing_tier
WHERE
    object.object_id = pricing_tier.pricing_tier_id AND
    object.object_type = 'pricing-tier' AND
    pricing_tier.name IS NOT NULL AND
    pricing_tier.description IS NULL;

UPDATE object
SET meta = jsonb_build_object('description', pricing_tier.description)
FROM pricing_tier
WHERE
    object.object_id = pricing_tier.pricing_tier_id AND
    object.object_type = 'pricing-tier' AND
    pricing_tier.name IS NULL AND
    pricing_tier.description IS NOT NULL;

UPDATE object
SET meta = jsonb_build_object('name', pricing_tier.name, 'description', pricing_tier.description)
FROM pricing_tier
WHERE
    object.object_id = pricing_tier.pricing_tier_id AND
    object.object_type = 'pricing-tier' AND
    pricing_tier.name IS NOT NULL AND
    pricing_tier.description IS NOT NULL;

-- backfill features
UPDATE object
SET meta = jsonb_build_object('name', feature.name)
FROM feature
WHERE
    object.object_id = feature.feature_id AND
    object.object_type = 'feature' AND
    feature.name IS NOT NULL AND
    feature.description IS NULL;

UPDATE object
SET meta = jsonb_build_object('description', feature.description)
FROM feature
WHERE
    object.object_id = feature.feature_id AND
    object.object_type = 'feature' AND
    feature.name IS NULL AND
    feature.description IS NOT NULL;

UPDATE object
SET meta = jsonb_build_object('name', feature.name, 'description', feature.description)
FROM feature
WHERE
    object.object_id = feature.feature_id AND
    object.object_type = 'feature' AND
    feature.name IS NOT NULL AND
    feature.description IS NOT NULL;

-- drop unused tables
DROP TABLE tenant;
DROP TABLE "user";
DROP TABLE role;
DROP TABLE permission;
DROP TABLE pricing_tier;
DROP TABLE feature;

COMMIT;
