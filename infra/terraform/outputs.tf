output "role_arn" {
  description = "ARN of the GuardSecurityScanRole"
  value       = aws_iam_role.guard_security_scan_role.arn
}

output "external_id" {
  description = "External ID used for role assumption"
  value       = var.external_id
}