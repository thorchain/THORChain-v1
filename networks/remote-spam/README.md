Terraform & Ansible
===================

Spam scripts are deployed using [Terraform](https://www.terraform.io/) to create servers on AWS then
[Ansible](http://www.ansible.com/) to create and manage spam scripts on those servers.

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

Create a number of spam nodes
-----------------------------

```sh
    make remotespam-start
```

Optionally, you can set the number of spam nodes you want to launch (defaults to 4) and the name of the testnet spam cluster (which defaults to remotetestnet-spam):

```sh
    CLUSTER_NAME="mytestnet-spam" SERVERS=7 make remotespam-start
```

Delete servers
--------------

```sh
    make remotespam-stop
```

Logging
-------

You can ship logs to Logz.io, an Elastic stack (Elastic search, Logstash and Kibana) service provider. You can set up your nodes to log there automatically. Create an account and get your API key from the notes on [this page](https://app.logz.io/#/dashboard/data-sources/Filebeat), then:

```sh
   yum install systemd-devel || echo "This will only work on RHEL-based systems."
   apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

   go get github.com/mheese/journalbeat
   cd networks/remote/terraform
   ansible-playbook -i /usr/local/bin/terraform-inventory logzio.yml ../ansible/-e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```
