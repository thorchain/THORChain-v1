Terraform & Ansible
===================

Automated deployments are done using [Terraform](https://www.terraform.io/) to create servers on AWS then
[Ansible](http://www.ansible.com/) to create and manage testnets on those servers.

Prerequisites
-------------

- Install [Terraform](https://www.terraform.io/downloads.html) and [Ansible](http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html) on a Linux machine or on MacOS.
- Install [terraform-inventory](https://github.com/adammck/terraform-inventory/releases) or use `brew install terraform-inventory` on MacOS (needed to determine inventory form terraform for ansible).
- Create a [AWS API token](https://console.aws.amazon.com/iam/home) with read and write capabilities.
- Create SSH keys to access the created nodes (or use existing keys).

```sh
    export AWS_SECRET_KEY=""
    export AWS_ACCESS_KEY=""
    export SSH_KEY_NAME="remotetestnet-deployer"
    export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
    export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

These will be used by both `terraform` and `ansible`.

Create a remote network
-----------------------

```sh
    make remotenet-start
```

Optionally, you can set the number of servers you want to launch (defaults to 4) and the name of the testnet (which defaults to remotetestnet):

```sh
    TESTNET_NAME="mytestnet" SERVERS=7 make remotenet-start
```

Quickly see the /status endpoint
--------------------------------

```sh
    make remotenet-status
```

Delete servers
--------------

```sh
    make remotenet-stop
```

Logging
-------

You can ship logs to Logz.io, an Elastic stack (Elastic search, Logstash and Kibana) service provider. You can set up your nodes to log there automatically. Create an account and get your API key from the notes on [this page](https://app.logz.io/#/dashboard/data-sources/Filebeat).

On a Mac, follow these steps: https://github.com/mheese/journalbeat/issues/92#issuecomment-328309117:
1. `cd $GOPATH/src/github.com/mheese/journalbeat`
1. `mkdir -p vendor/github.com/coreos/go-systemd/sdjournal/systemd`
1. Download and unzip: https://github.com/systemd/systemd
1. `cp ~/Downloads/systemd-master/src/systemd/* vendor/github.com/coreos/go-systemd/sdjournal/systemd`
1. `go install`

Then:

```sh
   yum install systemd-devel || echo "This will only work on RHEL-based systems."
   apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

   go get github.com/mheese/journalbeat
   cd networks/remote/terraform
   ansible-playbook -i /usr/local/bin/terraform-inventory ../ansible/logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```
