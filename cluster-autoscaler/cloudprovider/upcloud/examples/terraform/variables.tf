terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = "~> 3.0"
    }
  }
}

variable "autoscaler_username" {
  type = string
}

variable "autoscaler_password" {
  type      = string
  sensitive = true
}

variable "cluster_zone" {
  type    = string
  default = "de-fra1"
}
