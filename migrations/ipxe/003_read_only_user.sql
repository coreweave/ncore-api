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

GRANT CONNECT ON DATABASE ipxe TO read_only;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO read_only;

---- create above / drop below ----
REVOKE CONNECT ON DATABASE ipxe TO read_only;

REVOKE SELECT ON ALL TABLES IN SCHEMA public FROM read_only;
