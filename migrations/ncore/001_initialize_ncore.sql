
CREATE TABLE nodes (
    mac_address text PRIMARY KEY CHECK (mac_address != '') NOT NULL,
    hostname text NOT NULL CHECK (system_id != ''),
    ip_address text NOT NULL CHECK (ip_address != ''),
    first_seen timestamp with time zone NOT NULL DEFAULT now(),
    last_seen timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add ncore_systems_history table to keep track of each change here.
);

---- create above / drop below ----
DROP TABLE nodes;
