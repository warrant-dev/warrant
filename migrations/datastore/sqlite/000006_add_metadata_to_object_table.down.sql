CREATE TABLE IF NOT EXISTS permission (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectId INTEGER NOT NULL,
  permissionId TEXT NOT NULL,
  name TEXT DEFAULT NULL,
  description TEXT DEFAULT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS permission_uk_perm_id
    ON permission (permissionId);

CREATE INDEX IF NOT EXISTS objectId
    ON permission (objectId);

CREATE INDEX IF NOT EXISTS permission_uk_created_at_permission_id
    ON permission (createdAt, permissionId);

CREATE INDEX IF NOT EXISTS permission_uk_name_permission_id
    ON permission (name, permissionId);

INSERT INTO permission(objectId, permissionId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "permission";

CREATE TABLE IF NOT EXISTS role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectId INTEGER NOT NULL,
  roleId TEXT NOT NULL,
  name TEXT DEFAULT NULL,
  description TEXT DEFAULT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS role_uk_role_id
    ON role (roleId);

CREATE INDEX IF NOT EXISTS objectId
    ON role (objectId);

CREATE INDEX IF NOT EXISTS role_uk_created_at_role_id
    ON role (createdAt, roleId);

CREATE INDEX IF NOT EXISTS role_uk_name_role_id
    ON role (name, roleId);

INSERT INTO role(objectId, roleId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "role";

CREATE TABLE IF NOT EXISTS tenant (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectId INTEGER NOT NULL,
  tenantId TEXT NOT NULL,
  name TEXT DEFAULT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS tenant_uk_tenant_id
    ON tenant (tenantId);

CREATE INDEX IF NOT EXISTS objectId
    ON tenant (objectId);

CREATE INDEX IF NOT EXISTS tenant_uk_created_at_tenant_id
    ON tenant (createdAt, tenantId);

CREATE INDEX IF NOT EXISTS tenant_uk_name_tenant_id
    ON tenant (name, tenantId);

INSERT INTO tenant(objectId, tenantId, name, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "tenant";

CREATE TABLE IF NOT EXISTS user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  userId TEXT NOT NULL,
  email TEXT DEFAULT NULL,
  objectId INTEGER NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS user_uk_user_id
    ON user (userId);

CREATE INDEX IF NOT EXISTS objectId
    ON user (objectId);

CREATE INDEX IF NOT EXISTS user_uk_created_at_user_id
    ON user (createdAt, userId);

CREATE INDEX IF NOT EXISTS user_uk_email_user_id
    ON user (email, userId);

INSERT INTO user(objectId, userId, email, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.email', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "user";

CREATE TABLE IF NOT EXISTS feature (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectId INTEGER NOT NULL,
  featureId TEXT NOT NULL,
  name TEXT DEFAULT NULL,
  description TEXT DEFAULT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS feature_uk_feature_id
    ON feature (featureId);

CREATE INDEX IF NOT EXISTS objectId
    ON feature (objectId);

CREATE INDEX IF NOT EXISTS feature_uk_created_at_feature_id
    ON feature (createdAt, featureId);

CREATE INDEX IF NOT EXISTS feature_uk_name_feature_id
    ON feature (name, featureId);

INSERT INTO feature(objectId, featureId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "feature";

CREATE TABLE IF NOT EXISTS pricingTier (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectId INTEGER NOT NULL,
  pricingTierId TEXT NOT NULL,
  name TEXT DEFAULT NULL,
  description TEXT DEFAULT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (objectId) REFERENCES object (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS pricing_tier_uk_pricing_tier_id
    ON pricingTier (pricingTierId);

CREATE INDEX IF NOT EXISTS objectId
    ON pricingTier (objectId);

CREATE INDEX IF NOT EXISTS pricing_tier_uk_created_at_pricing_tier_id
    ON pricingTier (createdAt, pricingTierId);

CREATE INDEX IF NOT EXISTS pricing_tier_uk_name_pricing_tier_id
    ON pricingTier (name, pricingTierId);

INSERT INTO pricingTier(objectId, pricingTierId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "pricing-tier";

ALTER TABLE object
DROP COLUMN meta;
