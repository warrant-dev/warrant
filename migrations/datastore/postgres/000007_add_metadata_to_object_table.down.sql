BEGIN;

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

INSERT INTO permission(object_id, permission_id, name, description, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'name', meta->>'description', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'permission';

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

INSERT INTO role(object_id, role_id, name, description, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'name', meta->>'description', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'role';

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

INSERT INTO tenant(object_id, tenant_id, name, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'name', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'tenant';

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

INSERT INTO "user"(object_id, user_id, email, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'email', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'user';

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

INSERT INTO feature(object_id, feature_id, name, description, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'name', meta->>'description', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'feature';

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

INSERT INTO pricing_tier(object_id, pricing_tier_id, name, description, created_at, updated_at, deleted_at)
SELECT id, object_id, meta->>'name', meta->>'description', created_at, updated_at, deleted_at
FROM object
WHERE object_type = 'pricing_tier';

ALTER TABLE object
DROP COLUMN meta;

COMMIT;
