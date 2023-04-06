-- Write your migrate up statements here

CREATE TABLE node_payloads (
    payload_id text NOT NULL CHECK (payload_id != ''),
    mac_address text NOT NULL CHECK (mac_address != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY(payload_id, mac_address)
    -- TODO: add node_payloads_history table to keep track of each change here.
);


CREATE TABLE payloads (
    payload_id text PRIMARY KEY CHECK (payload_id != '') NOT NULL,
    payload_directory text NOT NULL CHECK (payload_directory != ''),
    payload_schema_id text NOT NULL CHECK (payload_schema_id != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add payloads_history table to keep track of each change here.
);
CREATE INDEX payload_schema ON payloads(payload_schema_id text_pattern_ops);


CREATE TABLE payload_schemas (
    payload_schema_id text NOT NULL CHECK (payload_schema_id != ''),
    parameter_name text NOT NULL CHECK (parameter_name != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY(payload_schema_id, parameter_name)
    -- TODO: add payload_schemas_history table to keep track of each change here.
);


CREATE TABLE payload_parameters (
    payload_id text NOT NULL CHECK (payload_id != ''),
    parameter_name text NOT NULL CHECK (parameter_name != ''),
    parameter_value text NOT NULL CHECK (parameter_value != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY(payload_id, parameter_name)
    -- TODO: add payload_parameters_history table to keep track of each change here.
);

CREATE INDEX payload_parameter ON payload_parameters(parameter_name text_pattern_ops);


-- https://faraday.ai/blog/how-to-aggregate-jsonb-in-postgres/
CREATE AGGREGATE jsonb_object_agg(jsonb) (
  SFUNC = 'jsonb_concat',
  STYPE = jsonb,
  INITCOND = '{}'
);

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
DROP TABLE node_payloads;
DROP TABLE payloads;
DROP TABLE payload_schemas;
DROP TABLE payload_parameters;

DROP AGGREGATE jsonb_object_agg(jsonb);
