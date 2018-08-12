Using Terraform
===============

This is a [Terraform](https://www.terraform.io/) configuration that sets up AWS infrastructure.

Prerequisites
-------------

- Install [HashiCorp Terraform](https://www.terraform.io) on a linux machine or on MacOS.
- Create a [AWS API token](https://console.aws.amazon.com/iam/home) with read and write capability.
- Create SSH keys to access the created nodes (or use existing keys).

Build
-----

```sh
    export TESTNET_NAME="remotetestnet"
    export SERVERS="4"
    export AWS_SECRET_KEY=""
    export AWS_ACCESS_KEY=""
    export SSH_KEY_NAME="remotetestnet-deployer"
    export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
    export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"

    terraform init
    terraform apply -var TESTNET_NAME="$TESTNET_NAME" -var SERVERS="$SERVERS" -var AWS_SECRET_KEY="$AWS_SECRET_KEY" -var AWS_ACCESS_KEY="$AWS_ACCESS_KEY" -var SSH_KEY_NAME="$SSH_KEY_NAME" -var SSH_PRIVATE_FILE="$SSH_PRIVATE_FILE" -var SSH_PUBLIC_FILE="$SSH_PUBLIC_FILE"
```

At the end you will get a list of IP addresses that belongs to your new droplets.

Destroy
-------

Run the below:

```sh
    export AWS_SECRET_KEY=""
    export AWS_ACCESS_KEY=""
    export SSH_KEY_NAME="remotetestnet-deployer"
    export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
    export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"

    terraform destroy -var AWS_SECRET_KEY="$AWS_SECRET_KEY" -var AWS_ACCESS_KEY="$AWS_ACCESS_KEY" -var SSH_KEY_NAME="$SSH_KEY_NAME" -var SSH_PRIVATE_FILE="$SSH_PRIVATE_FILE" -var SSH_PUBLIC_FILE="$SSH_PUBLIC_FILE"
```
