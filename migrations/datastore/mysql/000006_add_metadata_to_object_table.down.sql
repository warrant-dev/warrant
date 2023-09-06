BEGIN;

CREATE TABLE IF NOT EXISTS permission (
  id int NOT NULL AUTO_INCREMENT,
  objectId int NOT NULL,
  permissionId varchar(64) NOT NULL,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY permission_uk_perm_id (permissionId),
  KEY objectId (objectId),
  KEY permission_uk_created_at_permission_id (createdAt, permissionId),
  KEY permission_uk_name_permission_id (name, permissionId),
  CONSTRAINT permission_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO permission(objectId, permissionId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "permission";

CREATE TABLE IF NOT EXISTS role (
  id int NOT NULL AUTO_INCREMENT,
  objectId int NOT NULL,
  roleId varchar(64) NOT NULL,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY role_uk_role_id (roleId),
  KEY objectId (objectId),
  KEY role_uk_created_at_role_id (createdAt, roleId),
  KEY role_uk_name_role_id (name, roleId),
  CONSTRAINT role_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO role(objectId, roleId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "role";

CREATE TABLE IF NOT EXISTS tenant (
  id int NOT NULL AUTO_INCREMENT,
  objectId int NOT NULL,
  tenantId varchar(64) NOT NULL,
  name varchar(255) DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY objectId (objectId),
  KEY tenant_uk_tenant_id (tenantId),
  KEY tenant_uk_created_at_tenant_id (createdAt, tenantId),
  KEY tenant_uk_name_tenant_id (name, tenantId),
  CONSTRAINT tenant_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO tenant(objectId, tenantId, name, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "tenant";

CREATE TABLE IF NOT EXISTS user (
  id int NOT NULL AUTO_INCREMENT,
  userId varchar(64) NOT NULL,
  email varchar(255) DEFAULT NULL,
  objectId int NOT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY user_uk_user_id (userId),
  KEY objectId (objectId),
  KEY user_uk_created_at_user_id (createdAt, userId),
  KEY user_uk_email_user_id (email, userId),
  CONSTRAINT user_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO user(objectId, userId, email, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.email', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "user";

CREATE TABLE IF NOT EXISTS feature (
  id int NOT NULL AUTO_INCREMENT,
  objectId int NOT NULL,
  featureId varchar(64) NOT NULL,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY feature_uk_feature_id (featureId),
  KEY objectId (objectId),
  KEY feature_uk_created_at_feature_id (createdAt, featureId),
  KEY feature_uk_name_feature_id (name, featureId),
  CONSTRAINT feature_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO feature(objectId, featureId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "feature";

CREATE TABLE IF NOT EXISTS pricingTier (
  id int NOT NULL AUTO_INCREMENT,
  objectId int NOT NULL,
  pricingTierId varchar(64) NOT NULL,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY pricing_tier_uk_pricing_tier_id (pricingTierId),
  KEY objectId (objectId),
  KEY pricing_tier_uk_created_at_pricing_tier_id (createdAt, pricingTierId),
  KEY pricing_tier_uk_name_pricing_tier_id (name, pricingTierId),
  CONSTRAINT pricing_tier_fk_object_id FOREIGN KEY (objectId) REFERENCES object (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO pricingTier(objectId, pricingTierId, name, description, createdAt, updatedAt, deletedAt)
SELECT id, objectId, meta->>'$.name', meta->>'$.description', createdAt, updatedAt, deletedAt
FROM object
WHERE objectType = "pricing-tier";

ALTER TABLE object
DROP COLUMN meta;

COMMIT;
