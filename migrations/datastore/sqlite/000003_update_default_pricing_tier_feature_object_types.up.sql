-- Update the default pricing-tier object type
WITH
    currentDefaultPricingTierDefinition AS (SELECT '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}' AS value),
    newDefaultPricingTierDefinition AS (SELECT '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}' AS value)
UPDATE objectType
SET definition = (SELECT value FROM newDefaultPricingTierDefinition)
WHERE
    typeId = 'pricing-tier' AND
    definition = (SELECT value FROM currentDefaultPricingTierDefinition);

-- Update the default feature object type
WITH
    currentDefaultFeatureDefinition AS (SELECT '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}' AS value),
    newDefaultFeatureDefinition AS (SELECT '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}' AS value)
UPDATE objectType
SET definition = (SELECT value FROM newDefaultFeatureDefinition)
WHERE
    typeId = 'feature' AND
    definition = (SELECT value FROM currentDefaultFeatureDefinition);
