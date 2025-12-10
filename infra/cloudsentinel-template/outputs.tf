output "role_arn" {
  description = "ARN of the GuardSecurityScanRole"
  value       = aws_iam_role.guard_security_scan_role.arn
}

output "role_name" {
  description = "Name of the GuardSecurityScanRole"
  value       = aws_iam_role.guard_security_scan_role.name
}