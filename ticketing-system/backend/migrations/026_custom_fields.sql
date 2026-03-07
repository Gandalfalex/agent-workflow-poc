-- Custom field definitions per project
CREATE TABLE custom_fields (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        text NOT NULL,        -- internal key, unique per project
    label       text NOT NULL,        -- display label
    field_type  text NOT NULL CHECK (field_type IN ('text','number','date','enum','user','boolean')),
    options     jsonb NOT NULL DEFAULT '[]',  -- [{value,label}] for enum type
    "order"     int NOT NULL DEFAULT 0,
    required    boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (project_id, name)
);

-- Mapping: which custom fields appear for which ticket types
CREATE TABLE custom_field_schemas (
    field_id           uuid NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    ticket_type        text NOT NULL,
    required_state_ids jsonb NOT NULL DEFAULT '[]',  -- array of workflow state UUIDs where field is required
    PRIMARY KEY (field_id, ticket_type)
);

-- Per-ticket custom field values
CREATE TABLE ticket_custom_field_values (
    ticket_id     uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    field_id      uuid NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    value_text    text,
    value_number  numeric,
    value_date    date,
    value_boolean boolean,
    value_user_id uuid,
    PRIMARY KEY (ticket_id, field_id)
);

CREATE INDEX custom_fields_project_id_idx ON custom_fields(project_id);
CREATE INDEX ticket_custom_field_values_ticket_id_idx ON ticket_custom_field_values(ticket_id);
