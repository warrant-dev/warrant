CREATE TABLE IF NOT EXISTS objectType (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  typeId TEXT NOT NULL,
  definition TEXT NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS object_type_uk_type_id
    ON objectType (typeId);

CREATE TABLE IF NOT EXISTS warrant (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectType TEXT NOT NULL,
  objectId TEXT NOT NULL,
  relation TEXT NOT NULL,
  subjectType TEXT NOT NULL,
  subjectId TEXT NOT NULL,
  subjectRelation TEXT DEFAULT NULL,
  contextHash TEXT NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS warrant_uk_obj_rel_sub_ctx_hash
    ON warrant (objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash);

CREATE INDEX IF NOT EXISTS warrant_uk_sub_type_sub_id_sub_rel
    ON warrant (subjectType, subjectId, subjectRelation);

CREATE INDEX IF NOT EXISTS warrant_uk_obj_type_obj_id_rel
    ON warrant (objectType, objectId, relation);

CREATE TABLE IF NOT EXISTS context (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  warrantId INTEGER NOT NULL,
  name TEXT NOT NULL,
  value TEXT NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL,
  FOREIGN KEY (warrantId) REFERENCES warrant (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS context_uk_warrant_id_name
    ON context (warrantId, name);

CREATE TABLE IF NOT EXISTS object (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  objectType TEXT NOT NULL,
  objectId TEXT NOT NULL,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  deletedAt DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS object_uk_obj_type_obj_id
    ON object (objectType, objectId);

CREATE INDEX IF NOT EXISTS object_uk_created_at_object_id
    ON object (createdAt, objectId);

CREATE INDEX IF NOT EXISTS object_uk_object_type_object_id
    ON object (objectType, objectId);

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
