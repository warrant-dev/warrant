BEGIN;

CREATE TABLE IF NOT EXISTS objectType (
  id int NOT NULL AUTO_INCREMENT,
  typeId varchar(64) NOT NULL,
  definition json DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY object_type_uk_type_id (typeId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS warrant (
  id int NOT NULL AUTO_INCREMENT,
  objectType varchar(64) NOT NULL,
  objectId varchar(64) NOT NULL,
  relation varchar(64) NOT NULL,
  subjectType varchar(64) NOT NULL,
  subjectId varchar(64) NOT NULL,
  subjectRelation varchar(64) DEFAULT "",
  contextHash varchar(40) NOT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  KEY warrant_uk_sub_type_sub_id_sub_rel (subjectType, subjectId, subjectRelation),
  KEY warrant_uk_obj_rel_sub_ctx_hash (objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash),
  KEY warrant_uk_obj_type_obj_id_rel (objectType, objectId, relation)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS context (
  id int NOT NULL AUTO_INCREMENT,
  warrantId int NOT NULL,
  name varchar(64) NOT NULL,
  value varchar(64) NOT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY context_uk_warrant_id_name (warrantId, name),
  CONSTRAINT context_fk_warrant_id FOREIGN KEY (warrantId) REFERENCES warrant (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS object (
  id int NOT NULL AUTO_INCREMENT,
  objectType varchar(64) NOT NULL,
  objectId varchar(64) NOT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updatedAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deletedAt timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY object_uk_obj_type_obj_id (objectType, objectId),
  KEY object_uk_created_at_object_id (createdAt, objectId),
  KEY object_uk_object_type_object_id (objectType, objectId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

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

COMMIT;
