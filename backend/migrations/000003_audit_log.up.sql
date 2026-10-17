-- Audit trail for privileged back-office mutations.
-- One append-only row per admin API mutation, capturing who did what to which
-- entity, plus the request context needed for accountability.
BEGIN;

CREATE TABLE IF NOT EXISTS audit_log (
    id          uuid        NOT NULL DEFAULT gen_random_uuid(),
    actor_id    uuid        NULL,
    actor_role  text        NOT NULL DEFAULT '',
    action      text        NOT NULL,
    entity_type text        NOT NULL DEFAULT '',
    entity_id   text        NOT NULL DEFAULT '',
    method      text        NOT NULL DEFAULT '',
    path        text        NOT NULL DEFAULT '',
    status_code integer     NOT NULL DEFAULT 0,
    ip_address  text        NOT NULL DEFAULT '',
    user_agent  text        NOT NULL DEFAULT '',
    before_state jsonb      NULL,
    after_state  jsonb      NULL,
    metadata     jsonb      NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor_id   ON audit_log (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action     ON audit_log (action);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity     ON audit_log (entity_type, entity_id);

COMMIT;
