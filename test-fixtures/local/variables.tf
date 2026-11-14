variable "operating_system" {
  description = "On which operating system is Docker running?"
  default     = "Linux"

  validation {
    condition     = contains(["MacOS", "Windows", "Linux"], var.operating_system)
    error_message = "Value must be MacOS, Windows, or Linux."
  }
}

variable "docker_host" {
  description = "Docker API socket; overrides the default for operating_system (for example a Podman socket)"
  type        = string
  default     = null
}
