BEGIN;

-- Update default pricing tier object type to old format
UPDATE object_type
SET definition = '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}'
WHERE
    type_id = 'pricing-tier' AND
    definition @> '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}'::jsonb AND
    '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}'::jsonb @> definition;

-- Update default feature object type to old format
UPDATE object_type
SET definition = '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}'
WHERE
    type_id = 'feature' AND
    definition @> '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}'::jsonb AND
    '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}'::jsonb @> definition;

COMMIT;
