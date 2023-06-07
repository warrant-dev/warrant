BEGIN;

-- NOTE: This migration is not fully reversible.
-- It will drop all existing context data in favor of policies.
ALTER TABLE warrant
ADD COLUMN policy text NOT NULL DEFAULT '',
ADD COLUMN policy_hash varchar(64) NOT NULL DEFAULT '';

-- All existing context can be converted into policies that return the
-- intersection (&&) of a strict equality comparison with each context value.
UPDATE warrant w
SET
    policy = (
        SELECT STRING_AGG(CONCAT(name, ' == ', '"', value, '"'), ' && ')
        FROM context
        WHERE warrant_id = w.id
    )
WHERE context_hash != '';

UPDATE warrant
SET policy_hash = encode(sha256(policy::bytea), 'hex')
WHERE policy != '';

ALTER TABLE warrant
DROP CONSTRAINT warrant_uk_obj_rel_sub_ctx_hash,
DROP COLUMN context_hash,
ADD CONSTRAINT warrant_uk_obj_rel_sub_policy_hash UNIQUE (object_type, object_id, relation, subject_type, subject_id, subject_relation, policy_hash);

DROP TABLE IF EXISTS context;

COMMIT;
