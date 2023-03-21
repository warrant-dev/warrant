BEGIN;

INSERT INTO objectType (typeId, definition)
VALUES
    ('role', '{"type": "role", "relations": {"member": {"inheritIf": "member", "ofType": "role", "withRelation": "member"}}}'),
    ('permission', '{"type": "permission", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "permission", "withRelation": "member"}, {"inheritIf": "member", "ofType": "role", "withRelation": "member"}]}}}'),
    ('tenant', '{"type": "tenant", "relations": {"admin": {}, "member": {"inheritIf": "manager"}, "manager": {"inheritIf": "admin"}}}'),
    ('user', '{"type": "user", "relations": {}}'),
    ('pricing-tier', '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}'),
    ('feature', '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}')
ON DUPLICATE KEY UPDATE
    definition = VALUES(definition),
    deletedAt = NULL;

COMMIT;
