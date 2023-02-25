-- Write your migrate up statements here

CREATE TABLE node_payloads (
    payload_id text NOT NULL CHECK (payload_id != ''),
    mac_address text NOT NULL CHECK (mac_address != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY(payload_id, mac_address)
    -- TODO: add node_payloads_history table to keep track of each change here.
);

-- Flavors of a kube-worker payload (and there use cases)
-- Original payload parameters, values, and directory (master)
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker', 'test-mac-a');
-- Orignal payload parameters and values, but different directory (other branches)
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker-directory-alt', 'test-mac-a');
-- Orignal payload parameters and directory, but different values (other clusters)
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker-values-alt', 'test-mac-a');
-- Orignal payload parameters, but different values and directory (other branches in other clusters)
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker-values-alt-directory-alt', 'test-mac-a');
--
-- Schema changes when parameters change
-- Orignal payload values and directory, but different parameters -- NOT APPLICABLE
-- Orignal payload values, but different parameters and directory -- NOT APPLICABLE
--
-- Original payload directory, but different parameters and values (testing additional parameters)
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker-parameters-alt', 'test-mac-a');
-- Different payload parameters, values, and directory (other branches of testing additional parameters
INSERT into node_payloads (payload_id, mac_address) VALUES ('kube-worker-parameters-alt-directory-alt', 'test-mac-a');


CREATE TABLE payloads (
    payload_id text PRIMARY KEY CHECK (payload_id != '') NOT NULL,
    payload_directory text NOT NULL CHECK (payload_directory != ''),
    payload_schema_id text NOT NULL CHECK (payload_schema_id != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add payloads_history table to keep track of each change here.
);
CREATE INDEX payload_schema ON payloads(payload_schema_id text_pattern_ops);

INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker', 'kube-worker', 'kube-worker');
INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker-directory-alt', 'kube-worker-alt', 'kube-worker');
INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker-values-alt', 'kube-worker', 'kube-worker');
INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker-values-alt-directory-alt', 'kube-worker-alt', 'kube-worker');
INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker-parameters-alt', 'kube-worker', 'kube-worker-alt');
INSERT into payloads (payload_id, payload_directory, payload_schema_id) VALUES ('kube-worker-parameters-alt-directory-alt', 'kube-worker-alt', 'kube-worker-alt');


CREATE TABLE payload_schemas (
    payload_schema_id text NOT NULL CHECK (payload_schema_id != ''),
    parameter_name text NOT NULL CHECK (parameter_name != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY(payload_schema_id, parameter_name)
    -- TODO: add payload_schemas_history table to keep track of each change here.
);

INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker', 'apiserver');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker', 'ca_cert_hash');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker', 'join_token');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker-alt', 'apiserver');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker-alt', 'ca_cert_hash');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker-alt', 'join_token');
INSERT into payload_schemas (payload_schema_id, parameter_name) VALUES ('kube-worker-alt', 'other_parameter_abc');


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

INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker', 'apiserver', 'server');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker', 'ca_cert_hash', 'hash');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker', 'join_token', 'token');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-directory-alt', 'apiserver', 'server');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-directory-alt', 'ca_cert_hash', 'hash');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-directory-alt', 'join_token', 'token');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt', 'apiserver', 'server-alt');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt', 'ca_cert_hash', 'hash-alt');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt', 'join_token', 'token-alt');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt-directory-alt', 'apiserver', 'server-alt');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt-directory-alt', 'ca_cert_hash', 'hash-alt');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-values-alt-directory-alt', 'join_token', 'token-alt');

INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt', 'apiserver', 'server-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt', 'ca_cert_hash', 'hash-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt', 'join_token', 'token-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt', 'other_parameter_abc', 'other-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt-directory-alt', 'apiserver', 'server-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt-directory-alt', 'ca_cert_hash', 'hash-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt-directory-alt', 'join_token', 'token-xyz');
INSERT into payload_parameters (payload_id, parameter_name, parameter_value) VALUES ('kube-worker-parameters-alt-directory-alt', 'other_parameter_abc', 'other-xyz');


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
