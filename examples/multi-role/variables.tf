variable "ipam_base_url" {
  type = string
}

variable "global_admin_username" {
  type = string
}

variable "global_admin_password" {
  type      = string
  sensitive = true
}

variable "tenant_admin_username" {
  type = string
}

variable "tenant_admin_password" {
  type      = string
  sensitive = true
}

variable "tenant_user_username" {
  type = string
}

variable "tenant_user_password" {
  type      = string
  sensitive = true
}
