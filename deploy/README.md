# local development with bosh-lite

to create a local bosh-lite environment
```bash
./local_bosh_lite_create
```

deploy things to it
```bash
./local_deploy_cf
./local_deploy_kubo
```

to use the various clis:
```bash
source ./bosh_target  # this has to be sourced
./local_login_credhub
./local_login_cf
```
