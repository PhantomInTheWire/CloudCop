variable "external_id" {
  description = "External ID for role assumption (provided by CloudSentinel platform)"
  type        = string
  sensitive   = true
}

variable "platform_account_id" {
  description = "AWS account ID of the CloudSentinel platform"
  type        = string
  default     = "123456789012"
  
  validation {
    condition     = can(regex("^\\d{12}$", var.platform_account_id))
    error_message = "The platform_account_id must be a valid 12-digit AWS account ID."
  }
}