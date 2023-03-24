BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS resource_event (
  id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  type varchar(64) NOT NULL,
  source varchar(64) NOT NULL,
  resource_type varchar(64) NOT NULL,
  resource_id varchar(64) NOT NULL,
  meta jsonb DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE INDEX resource_event_idx_created_at_id ON resource_event (created_at, id);

CREATE TABLE IF NOT EXISTS access_event (
  id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  type varchar(64) NOT NULL,
  source varchar(64) NOT NULL,
  object_type varchar(64) NOT NULL,
  object_id varchar(64) NOT NULL,
  relation varchar(64) NOT NULL,
  subject_type varchar(64) NOT NULL,
  subject_id varchar(64) NOT NULL,
  subject_relation varchar(64) DEFAULT NULL,
  context jsonb DEFAULT NULL,
  meta jsonb DEFAULT NULL,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE INDEX access_event_idx_created_at_id ON access_event (created_at, id);

COMMIT;
