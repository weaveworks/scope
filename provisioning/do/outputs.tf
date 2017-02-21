output "username" {
  value = "${var.do_username}"
}

output "public_ips" {
  value = ["${digitalocean_droplet.tf_test_vm.*.ipv4_address}"]
}

output "hostnames" {
  value = "${join("\n", 
    "${formatlist("%v.%v.%v", 
      digitalocean_droplet.tf_test_vm.*.name, 
      digitalocean_droplet.tf_test_vm.*.region, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the Droplets:
# N.B.: by default Digital Ocean droplets only have public IPs, but in order to
# be consistent with other providers' recipes, we provide an output to generate
# an /etc/hosts file on the Droplets, even though it is using public IPs only.
output "private_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      digitalocean_droplet.tf_test_vm.*.ipv4_address, 
      digitalocean_droplet.tf_test_vm.*.name, 
      digitalocean_droplet.tf_test_vm.*.region, 
      var.app
    )}"
  )}"
}

# /etc/hosts file for the client:
output "public_etc_hosts" {
  value = "${join("\n", 
    "${formatlist("%v %v.%v.%v", 
      digitalocean_droplet.tf_test_vm.*.ipv4_address, 
      digitalocean_droplet.tf_test_vm.*.name, 
      digitalocean_droplet.tf_test_vm.*.region, 
      var.app
    )}"
  )}"
}

output "ansible_inventory" {
  value = "${format("[all]\n%s", join("\n",
    "${formatlist("%v private_ip=%v",
      digitalocean_droplet.tf_test_vm.*.ipv4_address,
      digitalocean_droplet.tf_test_vm.*.ipv4_address
    )}"
  ))}"
}

output "private_key_path" {
  value = "${var.do_private_key_path}"
}
