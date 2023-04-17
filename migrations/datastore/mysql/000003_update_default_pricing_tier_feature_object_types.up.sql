BEGIN;

SET @currentDefaultPricingTierDefinition := '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}';
SET @newDefaultPricingTierDefinition := '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}';

SET @currentDefaultFeatureDefinition := '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}';
SET @newDefaultFeatureDefinition := '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}';

# Update the default pricing-tier object type to new format
UPDATE objectType
SET definition = @newDefaultPricingTierDefinition
WHERE
    typeId = "pricing-tier" AND
    JSON_CONTAINS(definition, @currentDefaultPricingTierDefinition) AND
    JSON_CONTAINS(@currentDefaultPricingTierDefinition, definition);

# Update the default feature object type to new format
UPDATE objectType
SET definition = @newDefaultFeatureDefinition
WHERE
    typeId = "feature" AND
    JSON_CONTAINS(definition, @currentDefaultFeatureDefinition) AND
    JSON_CONTAINS(@currentDefaultFeatureDefinition, definition);

COMMIT;
