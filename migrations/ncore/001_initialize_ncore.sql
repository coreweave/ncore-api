CREATE ROLE readaccess LOGIN;

GRANT CONNECT ON DATABASE ncore TO readaccess;
GRANT USAGE ON SCHEMA public TO readaccess;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readaccess;

CREATE USER read_only WITH PASSWORD 'md56600a015abc595eb4d1f764df25cc7b4';

GRANT readaccess TO read_only;

CREATE TABLE nodes (
    mac_address text PRIMARY KEY CHECK (mac_address != '') NOT NULL,
    system_id text NOT NULL CHECK (system_id != ''),
    ip_address text NOT NULL CHECK (ip_address != ''),
    first_seen timestamp with time zone NOT NULL DEFAULT now(),
    last_seen timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add ncore_systems_history table to keep track of each change here.
);

---- create above / drop below ----
DROP TABLE nodes;