# local development with bosh-lite

install clis (mac only)
```
brew bundle
```

ensure you don't have other bosh-lite VMs on your system
- open up Virtualbox, delete anything running

create a local bosh-lite environment
```bash
./local_bosh_lite_create
```

deploy things to it
```bash
./local_deploy_cf
./local_deploy_kubo
```

login with the clis:
```bash
source ./bosh_target  # for bosh cli, this has to be sourced
./local_login_credhub # for credhub cli
./local_login_cf      # for cf cli
./local_login_k8s     # for kubectl cli
```


when you're done
```bash
./local_bosh_lite_delete
```
