CREATE TABLE subnet_default_payloads (
    subnet cidr NOT NULL PRIMARY KEY,
    payload_id text NOT NULL CHECK (payload_id != ''),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    modified_at timestamp with time zone NOT NULL DEFAULT now()
    -- TODO: add subnet_default_payloads table to keep track of each change here.
);

---- create above / drop below ----

DROP TABLE subnet_default_payloads;
