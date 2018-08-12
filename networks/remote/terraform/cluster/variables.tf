variable "name" {
  description = "The cluster name, e.g remotetestnet"
}

variable "regions" {
  description = "Region to launch in"
  type = "list"
  default = ["us-east-2", "ca-central-1", "eu-central-1", "ap-southeast-1", "us-west-1", "eu-west-2", "ap-northeast-2"]
}

variable "aws_secret_key" {
  description = "AWS secret key"
  type = "string"
}

variable "aws_access_key" {
  description = "AWS access key"
  type = "string"
}

variable "ssh_key_name" {
  description = "Name of the SSH key in AWS"
  type = "string"
}

variable "ssh_private_file" {
  description = "SSH private key to connect to nodes"
  type = "string"
}

variable "ssh_public_file" {
  description = "SSH public key filename to copy to the nodes"
  type = "string"
}

variable "instance_type" {
  description = "The AWS instance type to use"
  default = "t2.small"
}

variable "servers" {
  description = "Desired instance count"
  default     = 4
}

