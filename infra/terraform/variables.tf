variable "region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "external_id" {
  description = "External ID for role assumption"
  type        = string
}