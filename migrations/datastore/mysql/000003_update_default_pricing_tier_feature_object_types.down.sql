BEGIN;

SET @currentDefaultPricingTierDefinition := '{"type": "pricing-tier", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "pricing-tier", "withRelation": "member"}, {"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}';
SET @oldDefaultPricingTierDefinition := '{"type": "pricing-tier", "relations": {"member": {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}}}';

SET @currentDefaultFeatureDefinition := '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"},{"inheritIf": "member", "ofType": "tenant", "withRelation": "member"}]}}}';
SET @oldDefaultFeatureDefinition := '{"type": "feature", "relations": {"member": {"inheritIf": "anyOf", "rules": [{"inheritIf": "member", "ofType": "feature", "withRelation": "member"}, {"ofType": "pricing-tier", "inheritIf": "member", "withRelation": "member"}]}}}';

# Update default pricing tier object type to old format
UPDATE objectType
SET definition = @oldDefaultPricingTierDefinition
WHERE
    typeId = "pricing-tier" AND
    JSON_CONTAINS(definition, @currentDefaultPricingTierDefinition) AND
    JSON_CONTAINS(@currentDefaultPricingTierDefinition, definition);

# Update default feature object type to old format
UPDATE objectType
SET definition = @oldDefaultFeatureDefinition
WHERE
    typeId = "feature" AND
    JSON_CONTAINS(definition, @currentDefaultFeatureDefinition) AND
    JSON_CONTAINS(@currentDefaultFeatureDefinition, definition);

COMMIT;
