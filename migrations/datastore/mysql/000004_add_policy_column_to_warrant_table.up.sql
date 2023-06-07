BEGIN;

# NOTE: This migration is not fully reversible.
# It will drop all existing context data in favor of policies.

ALTER TABLE warrant
ADD COLUMN policy TEXT NOT NULL AFTER subjectRelation,
ADD COLUMN policyHash VARCHAR(64) NOT NULL AFTER policy;

# All existing context can be converted into policies that return the
# intersection (&&) of a strict equality comparison with each context value.
UPDATE warrant
SET
    warrant.policy = (
        SELECT GROUP_CONCAT(CONCAT(context.name, ' == ', '"', context.value, '"') SEPARATOR ' && ')
        FROM context
        WHERE context.warrantId = warrant.id
    ),
    warrant.policyHash = SHA2(warrant.policy, 256)
WHERE warrant.contextHash != "";

ALTER TABLE warrant
DROP INDEX warrant_uk_obj_rel_sub_ctx_hash,
DROP COLUMN contextHash,
ADD UNIQUE KEY warrant_uk_obj_rel_sub_policy_hash (objectType, objectId, relation, subjectType, subjectId, subjectRelation, policyHash);

DROP TABLE IF EXISTS context;

COMMIT;
