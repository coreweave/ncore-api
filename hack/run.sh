#!/bin/bash
SCRIPT_DIR=$(dirname ${BASH_SOURCE[0]})

export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""
export AWS_REGION=""
# payloads database connection string
export PAYLOADS_PGUSER=postgres
export PAYLOADS_PGPASSWORD=""
export PAYLOADS_PGHOST=127.0.0.1
export PAYLOADS_PGPORT=5432
export PAYLOADS_PGDATABASE=payloads
# ipxe database connection string
export IPXE_PGUSER=postgres
export IPXE_PGPASSWORD=""
export IPXE_PGHOST=127.0.0.1
export IPXE_PGPORT=5432
export IPXE_PGDATABASE=ipxe
export PGX_LOG_LEVEL=warn
go run ${SCRIPT_DIR}/..
