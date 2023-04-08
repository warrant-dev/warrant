BEGIN;

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP(6);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS object_type (
  id bigserial PRIMARY KEY,
  type_id varchar(64) NOT NULL CONSTRAINT object_type_uk_type_id UNIQUE,
  definition jsonb DEFAULT NULL,
  ALTER TABLE table_name MODIFY COLUMN definition JSON NOT NULL;
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON object_type
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS warrant (
  id bigserial PRIMARY KEY,
  object_type varchar(64) NOT NULL,
  object_id varchar(64) NOT NULL,
  relation varchar(64) NOT NULL,
  subject_type varchar(64) NOT NULL,
  subject_id varchar(64) NOT NULL,
  subject_relation varchar(64) DEFAULT '',
  context_hash varchar(40) NOT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL,
  CONSTRAINT warrant_uk_obj_rel_sub_ctx_hash UNIQUE (object_type, object_id, relation, subject_type, subject_id, subject_relation, context_hash)
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON warrant
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS context (
  id bigserial PRIMARY KEY,
  warrant_id bigint NOT NULL REFERENCES warrant (id),
  name varchar(64) NOT NULL,
  value varchar(64) NOT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL,
  CONSTRAINT context_uk_warrant_id_name UNIQUE (warrant_id, name)
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON context
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS object (
  id bigserial PRIMARY KEY,
  object_type varchar(64) NOT NULL,
  object_id varchar(64) NOT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL,
  CONSTRAINT object_uk_obj_type_obj_id UNIQUE (object_type, object_id)
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON object
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS permission (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  permission_id varchar(64) NOT NULL CONSTRAINT permission_uk_perm_id UNIQUE,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON permission
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS role (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  role_id varchar(64) NOT NULL CONSTRAINT role_uk_role_id UNIQUE,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON role
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS tenant (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  tenant_id varchar(64) NOT NULL CONSTRAINT tenant_uk_tenant_id UNIQUE,
  name varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON tenant
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS "user" (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  user_id varchar(64) NOT NULL CONSTRAINT user_uk_user_id UNIQUE,
  email varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON "user"
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS feature (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  feature_id varchar(64) NOT NULL CONSTRAINT feature_uk_feature_id UNIQUE,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON feature
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

CREATE TABLE IF NOT EXISTS pricing_tier (
  id bigserial PRIMARY KEY,
  object_id bigint NOT NULL REFERENCES object (id),
  pricing_tier_id varchar(64) NOT NULL CONSTRAINT pricing_tier_uk_pricing_tier_id UNIQUE,
  name varchar(255) DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at timestamp(6) NULL DEFAULT NULL
);

CREATE TRIGGER update_updated_at
BEFORE UPDATE ON pricing_tier
FOR EACH ROW EXECUTE PROCEDURE update_updated_at();

COMMIT;
