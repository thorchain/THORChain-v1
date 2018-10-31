// The cluster name
output "name" {
  value = "${var.name}"
}

// The list of cluster instance IDs
output "instances" {
  value = ["${aws_instance.cluster.*.id}"]
}

#output "instances_count" {
#  value = "${length(aws_instance.node.*)}"
#}

// The list of cluster instance public IPs
output "public_ips" {
  value = ["${aws_eip.eip.*.public_ip}"]
}
