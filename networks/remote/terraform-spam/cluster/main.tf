resource "aws_key_pair" "key" {
  key_name   = "${var.name}"
  public_key = "${file(var.ssh_public_file)}"
}

data "aws_ami" "linux" {
  most_recent = true
  filter {
    name   = "name"
    values = ["${var.image_name}"]
  }
}

data "aws_availability_zones" "zones" {
  state = "available"
}
resource "aws_security_group" "secgroup" {
  name = "${var.name}"
  description = "Automated security group for spammers"
  tags {
    Name = "${var.name}"
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 1317
    to_port = 1317
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 26656
    to_port = 26657
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 26660
    to_port = 26660
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]

  }
}

resource "aws_instance" "cluster" {
#  depends_on = ["${element(aws_route_table_association.route_table_association.*,count.index)}"]
  count = "${var.SERVERS*min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  ami = "${data.aws_ami.linux.image_id}"
    instance_type = "${var.instance_type}"
  key_name = "${aws_key_pair.key.key_name}"
  associate_public_ip_address = true
  security_groups = [ "${aws_security_group.secgroup.name}" ]
  availability_zone = "${element(data.aws_availability_zones.zones.names,count.index)}"

  tags {
    Environment = "${var.name}"
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}"
  }

  volume_tags {
    Environment = "${var.name}"
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}-VOLUME"
  }

  root_block_device {
    volume_size = 40
  }

  connection {
    user = "centos"
    private_key = "${file(var.ssh_private_file)}"
    timeout = "600s"
  }
}

resource "aws_eip" "eip" {
  instance = "${element(aws_instance.cluster.*.id,count.index)}"
  vpc = true
  depends_on = ["aws_instance.cluster"]
  count = "${var.SERVERS*min(length(data.aws_availability_zones.zones.names),var.max_zones)}"

  tags {
    Environment = "${var.name}"
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}-EIP"
  }
}