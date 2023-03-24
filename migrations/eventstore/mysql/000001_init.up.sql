BEGIN;

CREATE TABLE IF NOT EXISTS resourceEvent (
  id BINARY(16) DEFAULT (UUID_TO_BIN(UUID(), 1)),
  type varchar(64) NOT NULL,
  source varchar(64) NOT NULL,
  resourceType varchar(64) NOT NULL,
  resourceId varchar(64) NOT NULL,
  meta json DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE INDEX resource_event_idx_created_at_id ON resourceEvent (createdAt, id);

CREATE TABLE IF NOT EXISTS accessEvent (
  id BINARY(16) DEFAULT (UUID_TO_BIN(UUID(), 1)),
  type varchar(64) NOT NULL,
  source varchar(64) NOT NULL,
  objectType varchar(64) NOT NULL,
  objectId varchar(64) NOT NULL,
  relation varchar(64) NOT NULL,
  subjectType varchar(64) NOT NULL,
  subjectId varchar(64) NOT NULL,
  subjectRelation varchar(64) DEFAULT NULL,
  context json DEFAULT NULL,
  meta json DEFAULT NULL,
  createdAt timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE INDEX access_event_idx_created_at_id ON accessEvent (createdAt, id);

COMMIT;
