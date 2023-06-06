BEGIN;

-- NOTE: Running this down migration will result in the loss
-- of all existing warrants with policies.
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

DELETE FROM warrant
WHERE policy != '';

ALTER TABLE warrant
ADD COLUMN context_hash varchar(40) NOT NULL DEFAULT '',
DROP CONSTRAINT warrant_uk_obj_rel_sub_policy_hash,
DROP COLUMN policy,
DROP COLUMN policy_hash,
ADD CONSTRAINT warrant_uk_obj_rel_sub_ctx_hash UNIQUE (object_type, object_id, relation, subject_type, subject_id, subject_relation, context_hash);

COMMIT;
