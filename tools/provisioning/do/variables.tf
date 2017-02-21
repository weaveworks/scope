variable "client_ip" {
  description = "IP address of the client machine"
}

variable "app" {
  description = "Name of the application using the created droplet(s)."
  default     = "default"
}

variable "name" {
  description = "Name of the droplet(s)."
  default     = "test"
}

variable "num_hosts" {
  description = "Number of droplet(s)."
  default     = 1
}

variable "do_private_key_path" {
  description = "Digital Ocean SSH private key path"
  default     = "~/.ssh/id_rsa"
}

variable "do_public_key_id" {
  description = "Digital Ocean ID for your SSH public key"

  # You can retrieve it and set it as an environment variable this way:

  # $ export TF_VAR_do_public_key_id=$(curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" "https://api.digitalocean.com/v2/account/keys" | jq -c --arg key_name "$DIGITALOCEAN_SSH_KEY_NAME" '.ssh_keys | .[] | select(.name==$key_name) | .id')
}

variable "do_username" {
  description = "Digital Ocean SSH username"
  default     = "root"
}

variable "do_os" {
  description = "Digital Ocean OS"
  default     = "ubuntu-16-04-x64"
}

# curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" "https://api.digitalocean.com/v2/images?page=1&per_page=999999" | jq ".images | .[] | .slug" | grep -P "ubuntu|coreos|centos" | grep -v alpha | grep -v beta
# "ubuntu-16-04-x32"
# "ubuntu-16-04-x64"
# "ubuntu-16-10-x32"
# "ubuntu-16-10-x64"
# "ubuntu-14-04-x32"
# "ubuntu-14-04-x64"
# "ubuntu-12-04-x64"
# "ubuntu-12-04-x32"
# "coreos-stable"
# "centos-6-5-x32"
# "centos-6-5-x64"
# "centos-7-0-x64"
# "centos-7-x64"
# "centos-6-x64"
# "centos-6-x32"
# "centos-5-x64"
# "centos-5-x32"

# Digital Ocean datacenters
# See also: 
#   $ curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" "https://api.digitalocean.com/v2/regions"  | jq -c ".regions | .[] | .slug" | sort -u

variable "do_dc_ams2" {
  description = "Digital Ocean Amsterdam Datacenter 2"
  default     = "ams2"
}

variable "do_dc_ams3" {
  description = "Digital Ocean Amsterdam Datacenter 3"
  default     = "ams3"
}

variable "do_dc_blr1" {
  description = "Digital Ocean Bangalore Datacenter 1"
  default     = "blr1"
}

variable "do_dc_fra1" {
  description = "Digital Ocean Frankfurt Datacenter 1"
  default     = "fra1"
}

variable "do_dc_lon1" {
  description = "Digital Ocean London Datacenter 1"
  default     = "lon1"
}

variable "do_dc_nyc1" {
  description = "Digital Ocean New York Datacenter 1"
  default     = "nyc1"
}

variable "do_dc_nyc2" {
  description = "Digital Ocean New York Datacenter 2"
  default     = "nyc2"
}

variable "do_dc_nyc3" {
  description = "Digital Ocean New York Datacenter 3"
  default     = "nyc3"
}

variable "do_dc_sfo1" {
  description = "Digital Ocean San Francisco Datacenter 1"
  default     = "sfo1"
}

variable "do_dc_sfo2" {
  description = "Digital Ocean San Francisco Datacenter 2"
  default     = "sfo2"
}

variable "do_dc_sgp1" {
  description = "Digital Ocean Singapore Datacenter 1"
  default     = "sgp1"
}

variable "do_dc_tor1" {
  description = "Digital Ocean Toronto Datacenter 1"
  default     = "tor1"
}

variable "do_dc" {
  description = "Digital Ocean's selected datacenter"
  default     = "lon1"
}

variable "do_size" {
  description = "Digital Ocean's selected machine size"
  default     = "4gb"
}

# Digital Ocean sizes


# See also: 


#   $ curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" "https://api.digitalocean.com/v2/sizes"  | jq -c ".sizes | .[] | .slug"


# "512mb"


# "1gb"


# "2gb"


# "4gb"


# "8gb"


# "16gb"


# "m-16gb"


# "32gb"


# "m-32gb"


# "48gb"


# "m-64gb"


# "64gb"


# "m-128gb"


# "m-224gb"

