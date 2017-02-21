output "username" {
  value = "${var.gcp_username}"
}

output "public_ips" {
  value = ["${google_compute_instance.tf_test_vm.*.network_interface.0.access_config.0.assigned_nat_ip}"]
}

output "hostnames" {
  value = "${join("\n", 
    "${formatlist("%v.%v.%v", 
      google_compute_instance.tf_test_vm.*.name, 
      google_compute_instance.tf_test_vm.*.zone, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the Compute Engine instances:
output "private_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      google_compute_instance.tf_test_vm.*.network_interface.0.address, 
      google_compute_instance.tf_test_vm.*.name, 
      google_compute_instance.tf_test_vm.*.zone, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the client:
output "public_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      google_compute_instance.tf_test_vm.*.network_interface.0.access_config.0.assigned_nat_ip, 
      google_compute_instance.tf_test_vm.*.name, 
      google_compute_instance.tf_test_vm.*.zone, 
      var.app
    )}"
  )}"
}

output "ansible_inventory" {
  value = "${format("[all]\n%s", join("\n",
    "${formatlist("%v private_ip=%v",
      google_compute_instance.tf_test_vm.*.network_interface.0.access_config.0.assigned_nat_ip,
      google_compute_instance.tf_test_vm.*.network_interface.0.address
    )}"
  ))}"
}

output "private_key_path" {
  value = "${var.gcp_private_key_path}"
}

output "instances_names" {
  value = ["${google_compute_instance.tf_test_vm.*.name}"]
}

output "image" {
  value = "${var.gcp_image}"
}

output "zone" {
  value = "${var.gcp_zone}"
}
