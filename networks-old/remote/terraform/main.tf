#Terraform Configuration

variable "TESTNET_NAME" {
  description = "Name of the testnet"
  default = "remotetestnet"
}

variable "AWS_SECRET_KEY" {
  description = "AWS secret key"
  type = "string"
}

variable "AWS_ACCESS_KEY" {
  description = "AWS access key"
  type = "string"
}

variable "SSH_KEY_NAME" {
  description = "Name of the SSH key in AWS"
  type = "string"
}

variable "SSH_PRIVATE_FILE" {
  description = "SSH private key file to be used on the nodes"
  type = "string"
}

variable "SSH_PUBLIC_FILE" {
  description = "SSH public key file to be used on the nodes"
  type = "string"
}

variable "SERVERS" {
  description = "Number of nodes in testnet"
  default = "4"
}

module "cluster" {
  source           = "./cluster"
  name             = "${var.TESTNET_NAME}"
  aws_secret_key   = "${var.AWS_SECRET_KEY}"
  aws_access_key   = "${var.AWS_ACCESS_KEY}"
  ssh_key_name     = "${var.SSH_KEY_NAME}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  servers          = "${var.SERVERS}"
}


output "public_ips" {
  value = "${module.cluster.public_ips}"
}

