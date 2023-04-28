# ncore-api

## To run application

```sh
# Create databases
export LOCAL_PGPASSWORD=<password>
unset PGDATABASE
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=$LOCAL_PGPASSWORD
psql -h localhost -U postgres -c "CREATE DATABASE payloads;"
psql -h localhost -U postgres -c "CREATE DATABASE ipxe;"

# Run migrations
export PGDATABASE=payloads
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=$LOCAL_PGPASSWORD
tern migrate -m ./migrations/payloads

export PGDATABASE=ipxe
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=$LOCAL_PGPASSWORD
tern migrate -m ./migrations/ipxe

# Example downgrade 1 then upgrade
export PGDATABASE=ipxe
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=$LOCAL_PGPASSWORD
tern migrate --destination -+1 -m ./migrations/ipxe
# Run api
# S3 env with read access
export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""
export AWS_REGION=""
# payloads database connection string
export PAYLOADS_PGUSER=postgres
export PAYLOADS_PGPASSWORD=$LOCAL_PGPASSWORD
export PAYLOADS_PGHOST=127.0.0.1
export PAYLOADS_PGPORT=5432
export PAYLOADS_PGDATABASE=payloads
# ipxe database connection string
export IPXE_PGUSER=postgres
export IPXE_PGPASSWORD=$LOCAL_PGPASSWORD
export IPXE_PGHOST=127.0.0.1
export IPXE_PGPORT=5432
export IPXE_PGDATABASE=ipxe
export PGX_LOG_LEVEL=warn
go run .
```

### Example object storage (produced by CI)

```bash
2023-03-20 19:29           76  s3://coreweave-ncore-images/ncore-develop-ci-test/cmdline
2023-03-20 19:29    149244106  s3://coreweave-ncore-images/ncore-develop-ci-test/initrd.img
2023-03-20 19:32   1172793855  s3://coreweave-ncore-images/ncore-develop-ci-test/rootfs.cpio.gz
2023-03-20 19:32     11458952  s3://coreweave-ncore-images/ncore-develop-ci-test/vmlinuz
```

### Endpoints

- `/api/v2/payload/<macAddress>`
  - GET:
    - returns the first NodePayload as a json object for the given macAddress
    - used by [payloads.service](https://github.com/coreweave/ncore-image-tenant/blob/ca696c84cc2d3deb99d3cb61336062d22425a9da/ansible/roles/base/files/systemd/scripts/payloads.sh) in the ncore-image
    - ex. `curl +XGET localhost:8080/api/v2/payload/test_mac`

        ```json
        {
                "PayloadId": "kube-worker",
                "PayloadDirectory": "kube-worker",
                "MacAddress": "test_mac"
        }
        ```

- `/api/v2/payload/<macAddress>/<payloadId>`
  - PUT:
    - inserts/updates the node entry within node_payloads table
    - returns a list of NodePayload objects assigned as a json list for the given macAddress
    - used by [kubernetes-node.join](https://github.com/coreweave/kubernetes-node/blob/5f0ea40f0d8eb75931dfefb550c8ddf756bbb238/join.sh), [kubernetes-node.enable_virtualization](https://github.com/coreweave/kubernetes-node/blob/5f0ea40f0d8eb75931dfefb550c8ddf756bbb238/enable_virtualization.sh), [kubernetes-node.disable_virtualization](https://github.com/coreweave/kubernetes-node/blob/5f0ea40f0d8eb75931dfefb550c8ddf756bbb238/disable_virtualization.sh), and [kubernetes-node.set_nvlink](https://github.com/coreweave/kubernetes-node/blob/5f0ea40f0d8eb75931dfefb550c8ddf756bbb238/set_nvlink.sh)
    - ex: `curl +XPUT localhost:8080/api/v2/payload/test_mac/test_payload`

        ```json
        [
          {
                "PayloadId": "test_payload",
                "PayloadDirectory": "kube-worker",
                "MacAddress": "test_mac"
          }
        ]
        ```

  - DELETE
    - deletes the node entry within node_payloads table
    - returns a list of NodePayload objects assigned as a json list for the given macAddress
    - used by [kubernetes-node.leave](https://github.com/coreweave/kubernetes-node/blob/5f0ea40f0d8eb75931dfefb550c8ddf756bbb238/leave.sh)
    - ex: `curl +XDELETE localhost:8080/api/v2/payload/test_mac/test_payload`

        ```json
        [
          {
                "PayloadId": "not_test_payload",
                "PayloadDirectory": "kube-worker",
                "MacAddress": "test_mac"
          }
        ]
        ```

- `/api/v2/payload/config/<payloadId>`
  - returns the payload parameters as a json object for a given payloadId
  - used by [config.sh](https://github.com/coreweave/ncore-image-tenant/blob/ca696c84cc2d3deb99d3cb61336062d22425a9da/ansible/roles/base/files/payloads/kube-worker/config.sh) in the kube-worker payload
  - ex. `curl localhost:8080/api/v2/payload/config/kube-worker`

      ```json
      {
              "apiserver": "kube-apiserver.domain",
              "ca_cert_hash": "certhash",
              "join_token": "jointoken"
      }

- `/api/v2/ipxe/config/<macAddres>`
  - returns the IpxeConfig as a json object for a given macAddress
  - used for manual verification/image downloading
  - ex. `curl localhost:8080/api/v2/ipxe/config/test_mac`

      ```json
      {
        "ImageName": "ncore-develop-ci-test.20230320-1916",
        "ImageBucket": "coreweave-ncore-images",
        "ImageTag": "default",
        "ImageType": "default",
        "ImageInitrdUrlHttp": "http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/initrd.img?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=7e44ae67aed3bfe4f8c30125233a79d5e06aa596c65e201a19e9e03cdeac4fb7",
        "ImageInitrdUrlHttps": "https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/initrd.img?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=7e44ae67aed3bfe4f8c30125233a79d5e06aa596c65e201a19e9e03cdeac4fb7",
        "ImageKernelUrlHttp": "http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/vmlinuz?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=147e2f85f73ae2df53dd7d50de311abca6ced173cd19f53a75b6f819d4ee8cc4",
        "ImageKernelUrlHttps": "https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/vmlinuz?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=147e2f85f73ae2df53dd7d50de311abca6ced173cd19f53a75b6f819d4ee8cc4",
        "ImageRootFsUrlHttp": "http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/rootfs.cpio.gz?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=1bb194a3deb24d8a15000ebef1839d09cdb97afcc44ebe8e6404fc946da02bba",
        "ImageRootFsUrlHttps": "https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/rootfs.cpio.gz?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request\u0026X-Amz-Date=20230320T220941Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026x-id=GetObject\u0026X-Amz-Signature=1bb194a3deb24d8a15000ebef1839d09cdb97afcc44ebe8e6404fc946da02bba",
        "ImageCmdline": "root=UUID=cb2e0849-89f9-4590-a9b1-61e8b92fc308 ro console=tty1 console=ttyS0"
      }

- `/api/v2/ipxe/images/`
  - accepts a json object containing ImageName, ImageBucket, ImageTag, and ImageType and inserts it into ipxe.images where (ImageTag, ImageType) is the primary key
  - returns an IpxeConfig as a json object for the given image for verification
  - used by [stage.sh](https://github.com/coreweave/ncore-image-tenant/blob/ca696c84cc2d3deb99d3cb61336062d22425a9da/ci/stage.sh) in the gitlab-ci

      ```bash
      curl -s -XPUT "localhost:8080/api/v2/ipxe/images/" -H 'Content-Type: application/json' -d '{
        "ImageName": "ncore-develop-ci-test.20230320-1811",
        "ImageCmdline": "root=UUID=cb2e0849-89f9-4590-a9b1-61e8b92fc308 ro console=tty1 console=ttyS0",
        "ImageBucket": "coreweave-ncore-images",
        "ImageTag": "develop",
        "ImageType": "ci-test"
      }'

- `/api/v2/ipxe/template/<macAddress>`
  - returns the IpxeConfig as a templated ipxe menu
  - used by [kea](https://github.com/coreweave/pxe-infrastructure-tenant)
  - ex. `curl localhost:8080/api/v2/ipxe/template/test_mac`

      ```bash
      # Trimmed output
      #!ipxe
      ...
      :start
      menu Boot Options for ${mac}
      item --gap -------------------- Images --------------------
      item ncore-develop-ci-test.20230320-1916 ncore-develop-ci-test.20230320-1916
      ...
      # image boot
      :ncore-develop-ci-test.20230320-1916
      echo Booting ncore-develop-ci-test.20230320-1916 from https
      set conn_type http
      kernel http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/vmlinuz?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221400Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=d1d4cbf4334f1be678064aebca4fd89899be239e6e0e4b6d7e3cf1725fb5d713 root=UUID=cb2e0849-89f9-4590-a9b1-61e8b92fc308 ro console=tty1 console=ttyS0 initrd=initrd.magic root=http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/rootfs.cpio.gz?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221400Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=11c671c7d74aefa272f59859f32149bb34658a00eb8403dd09c09c210f74b967
      initrd http://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/initrd.img?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221400Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=130442c2337a7451bd52f40b9af117e40a0d74895ba437935e7d6c25ded47a9b
      boot || goto retry
      ...

- `/api/v2/ipxe/s3/<imageName>`
  - returns a presigned urls to download the image as text
  - ex. `curl localhost:8080/api/v2/ipxe/s3/ncore-develop-ci-test.20230320-1916`

      ```text
      imageInitrdUrlHttps: https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/initrd.img?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221517Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=4bca2b8dcd21e36637d9db1158d98fc39f4234a0c71e8ba270d271b6f71b0c8e
      imageKernelUrlHttps: https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/vmlinuz?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221517Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=7104ca7504264bf05681565d4f0b368344ebb08f6b00a8ff863133f279480b22
      imageRootFsUrlHttps: <https://coreweave-ncore-images.object.ord1.coreweave.com/ncore-develop-ci-test/rootfs.cpio.gz?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=1J2P027HRLMXTOGSBNIO%2F20230320%2F%2Fs3%2Faws4_request&X-Amz-Date=20230320T221517Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=b4e249ebfa0957693cf51630f90ee1576fc729833986e2615ef6ffea44e3a5f2>

### Testing

```sh
sudo apt install mockgen
```

```sh
go generate ./...
```

```sh
# Run all tests passing INTEGRATION_TESTDB explicitly
$ INTEGRATION_TESTDB=true \
    PGHOST=127.0.0.1 \
    PGPORT=5432 \
    PGUSER=postgres \
    PGPASSWORD=<password> \
    go test -v ./...
```

### Testing

```sh
sudo apt install mockgen
```

```sh
go generate ./...
```

```sh
# Run all tests passing INTEGRATION_TESTDB explicitly
$ INTEGRATION_TESTDB=true \
    PGHOST=127.0.0.1 \
    PGPORT=5432 \
    PGUSER=postgres \
    PGPASSWORD=<password> \
    go test -v ./...
```

### See also

- [Postgres Environment Variables](https://www.postgresql.org/docs/current/libpq-envars.html)
