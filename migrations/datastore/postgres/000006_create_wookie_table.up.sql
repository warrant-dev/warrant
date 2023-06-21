BEGIN;

CREATE TABLE IF NOT EXISTS wookie (
  id bigserial PRIMARY KEY,
  ver bigserial,
  created_at timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6)
);

COMMIT;
