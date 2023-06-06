DO
$do$
BEGIN
IF EXISTS (
  SELECT FROM pg_catalog.pg_roles
  WHERE rolname = 'read_only'
) THEN RAISE NOTICE 'read_only already exists';
ELSE
  CREATE USER read_only WITH PASSWORD '{{ env "READ_ONLY_PASSWORD" }}';
END IF;
END $do$;

GRANT CONNECT ON DATABASE payloads TO read_only;

GRANT SELECT ON node_payloads TO read_only;

---- create above / drop below ----
REVOKE CONNECT ON DATABASE payloads TO read_only;

REVOKE SELECT ON node_payloads FROM read_only;
