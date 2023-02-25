#!/usr/bin/env bash

(psql -h $PGHOST -U $PGUSER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '${PGDATABASE}'" | grep -q 1 || psql -h $PGHOST -U $PGUSER -d postgres -c "CREATE DATABASE ${PGDATABASE};") && tern migrate -m /migrations/$PGDATABASE;
