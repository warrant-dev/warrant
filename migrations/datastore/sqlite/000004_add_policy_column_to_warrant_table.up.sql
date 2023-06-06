-- NOTE: This migration is not fully reversible.
-- It will drop all existing context data in favor of policies.
DROP INDEX IF EXISTS warrant_uk_obj_rel_sub_ctx_hash;
DROP TABLE IF EXISTS context;

ALTER TABLE warrant
ADD COLUMN policy text NOT NULL DEFAULT '';

ALTER TABLE warrant
ADD COLUMN policyHash varchar(64) NOT NULL DEFAULT '';

-- NOTE: All existing warrants with context will be deleted.
DELETE FROM warrant
WHERE contextHash != "";

ALTER TABLE warrant
DROP COLUMN contextHash;

CREATE UNIQUE INDEX IF NOT EXISTS warrant_uk_obj_rel_sub_policy_hash
    ON warrant (objectType, objectId, relation, subjectType, subjectId, subjectRelation, policyHash);
