CREATE TABLE IF NOT EXISTS resourceEvent (
  id BLOB DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-a' || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
  type TEXT NOT NULL,
  source TEXT NOT NULL,
  resourceType TEXT NOT NULL,
  resourceId TEXT NOT NULL,
  meta TEXT DEFAULT NULL,
  createdAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE INDEX resource_event_idx_created_at_id
    ON resourceEvent (createdAt, id);

CREATE TABLE IF NOT EXISTS accessEvent (
  id BLOB DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-a' || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
  type TEXT NOT NULL,
  source TEXT NOT NULL,
  objectType TEXT NOT NULL,
  objectId TEXT NOT NULL,
  relation TEXT NOT NULL,
  subjectType TEXT NOT NULL,
  subjectId TEXT NOT NULL,
  subjectRelation TEXT DEFAULT NULL,
  context TEXT DEFAULT NULL,
  meta TEXT DEFAULT NULL,
  createdAt DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);

CREATE INDEX access_event_idx_created_at_id
    ON accessEvent (createdAt, id);
