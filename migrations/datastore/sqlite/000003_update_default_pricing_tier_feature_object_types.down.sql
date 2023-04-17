-- Update default pricing tier object type to old format
WITH
    currentDefaultPricingTierDefinition AS (SELECT '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}' AS value),
    oldDefaultPricingTierDefinition AS (SELECT '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}' AS value)
UPDATE objectType
SET definition = (SELECT value FROM oldDefaultPricingTierDefinition)
WHERE
    typeId = 'pricing-tier' AND
    definition = (SELECT value FROM currentDefaultPricingTierDefinition);

-- Update default feature object type to old format
WITH
    currentDefaultFeatureDefinition AS (SELECT '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}' AS value),
    oldDefaultFeatureDefinition AS (SELECT '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}' AS value)
UPDATE objectType
SET definition = (SELECT value FROM oldDefaultFeatureDefinition)
WHERE
    typeId = 'feature' AND
    definition = (SELECT value FROM currentDefaultFeatureDefinition);
