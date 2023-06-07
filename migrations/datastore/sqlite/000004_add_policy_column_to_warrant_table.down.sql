-- NOTE: Running this down migration will result in the loss
-- of all warrant policies.
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

DELETE FROM warrant
WHERE policy != "";

ALTER TABLE warrant
ADD COLUMN contextHash TEXT NOT NULL DEFAULT "";

DROP INDEX IF EXISTS warrant_uk_obj_rel_sub_policy_hash;

ALTER TABLE warrant
DROP COLUMN policy;

ALTER TABLE warrant
DROP COLUMN policyHash;

CREATE UNIQUE INDEX IF NOT EXISTS warrant_uk_obj_rel_sub_ctx_hash
    ON warrant (objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash);
