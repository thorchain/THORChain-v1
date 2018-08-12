provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${element(var.regions, 0)}"
}

resource "aws_key_pair" "ssh" {
  key_name   = "${var.ssh_key_name}"
  public_key = "${file(var.ssh_public_file)}"
}

resource "aws_instance" "cluster" {
  ami           = "ami-8c122be9"
  instance_type = "${var.instance_type}"
  # region = "${element(var.regions, count.index)}"
  key_name = "${var.ssh_key_name}"
  tags {
    Name = "${var.name}-node${count.index}"
    Role = "full-node"
  }
  count = "${var.servers}"

  lifecycle = {
	  prevent_destroy = false
  }

  connection {
    private_key = "${file(var.ssh_private_file)}"
    user = "ec2-user"
    timeout = "30s"
  }

  provisioner "file" {
    source = "files/terraform.sh"
    destination = "/tmp/terraform.sh"
  }

  provisioner "file" {
    source = "files/thorchaind.service"
    destination = "/tmp/thorchaind.service"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/thorchaind.service /etc/systemd/system/thorchaind.service",
      "sudo chmod +x /tmp/terraform.sh",
      "sudo /tmp/terraform.sh ${var.name} ${count.index}",
    ]
  }
}
