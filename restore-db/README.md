# ncore-api-db-restoration

## To apply restoration procedure

```sh
# Make sure sure you are in the ncore namespace in internal cluster

kubectl config use-context internal

kubectl config set-context --current --namespace=ncore
```

## Checklist
 - Comment out databases you DO NOT want to restore in values.yaml
 - Update image tag to latest revision in values.yaml (https://gitlab.com/coreweave/ncore-api/container_registry)
 
```sh
# Make sure to comment out databases you DO NOT want to restore in values.yaml
#To start kubernetes job
helm install restoration .
```