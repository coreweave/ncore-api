CREATE USER read_only WITH PASSWORD '{{ env "READ_ONLY_PASSWORD" }}';

GRANT CONNECT ON DATABASE nodes TO read_only;

GRANT USAGE ON SCHEMA public TO read_only;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO read_only;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO read_only;

CREATE TABLE nodes (
    mac_address text PRIMARY KEY CHECK (mac_address != '') NOT NULL,
    hostname text NOT NULL CHECK (hostname != ''),
    ip_address text NOT NULL CHECK (ip_address != ''),
    first_seen timestamp with time zone NOT NULL DEFAULT now(),
    last_seen timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add ncore_systems_history table to keep track of each change here.
);

---- create above / drop below ----
DROP TABLE nodes;