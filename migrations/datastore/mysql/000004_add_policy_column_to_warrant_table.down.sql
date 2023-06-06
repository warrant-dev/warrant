BEGIN;

# NOTE: Running this down migration will result in the loss
# of all existing warrant policies.
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

DELETE FROM warrant
WHERE policy != "";

ALTER TABLE warrant
ADD COLUMN contextHash varchar(40) NOT NULL AFTER subjectRelation,
DROP INDEX warrant_uk_obj_rel_sub_policy_hash,
DROP COLUMN policy,
DROP COLUMN policyHash;

CREATE INDEX warrant_uk_obj_rel_sub_ctx_hash ON warrant(objectType, objectId, relation, subjectType, subjectId, subjectRelation, contextHash);

COMMIT;
