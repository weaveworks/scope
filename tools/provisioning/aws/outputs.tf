output "username" {
  value = "${lookup(var.aws_usernames, "${lookup(var.aws_amis, var.aws_dc)}")}"
}

output "public_ips" {
  value = ["${aws_instance.tf_test_vm.*.public_ip}"]
}

output "hostnames" {
  value = "${join("\n", 
    "${formatlist("%v.%v.%v", 
      aws_instance.tf_test_vm.*.tags.Name, 
      aws_instance.tf_test_vm.*.availability_zone, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the Droplets:
output "private_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      aws_instance.tf_test_vm.*.private_ip, 
      aws_instance.tf_test_vm.*.tags.Name, 
      aws_instance.tf_test_vm.*.availability_zone, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the client:
output "public_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      aws_instance.tf_test_vm.*.public_ip, 
      aws_instance.tf_test_vm.*.tags.Name, 
      aws_instance.tf_test_vm.*.availability_zone, 
      var.app
    )}"
  )}"
}

output "ansible_inventory" {
  value = "${format("[all]\n%s", join("\n",
    "${formatlist("%v private_ip=%v",
      aws_instance.tf_test_vm.*.public_ip,
      aws_instance.tf_test_vm.*.private_ip,
    )}"
  ))}"
}

output "private_key_path" {
  value = "${var.aws_private_key_path}"
}
